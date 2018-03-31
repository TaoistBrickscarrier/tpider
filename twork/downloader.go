package tworker

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

type downloading struct {
	name, proxy, path string
	m                 map[string]tumblrResource
}

// Download fetch one tumblr user's all media files.
func Download(name, proxy, path string, th int) {
	var start, total int

	base := fmt.Sprintf("https://%s.tumblr.com/api/read/json", name)

	type queryResp struct {
		data  []byte
		start int
		err   error
	}

	conn := make(chan queryResp, 2)

	query := func(start, next int) {
		frameURL, err := url.Parse(base)
		if err != nil {
			conn <- queryResp{nil, start, err}
			return
		}
		vals := url.Values{}
		vals.Set("num", strconv.Itoa(next))
		vals.Set("start", strconv.Itoa(start))
		frameURL.RawQuery = vals.Encode()
		var resp *http.Response
		if proxy != "" {
			setProxy := func(_ *http.Request) (*url.URL, error) {
				return url.Parse(proxy)
			}
			transport := &http.Transport{Proxy: setProxy}
			client := &http.Client{Transport: transport}
			resp, err = client.Get(frameURL.String())
		} else {
			resp, err = http.Get(frameURL.String())
		}
		if err != nil {
			conn <- queryResp{nil, start, err}
			return
		}
		defer resp.Body.Close()
		contents, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			conn <- queryResp{nil, start, err}
			return
		}
		conn <- queryResp{contents, start, nil}
	}

	// Get user info, especially the total posts.
	for {
		go query(start, 0)
		status := <-conn
		if status.err != nil {
			fmt.Println("Download: query user info faild:", status.err)
			continue
		}
		var err error
		total, err = getTotal(status.data)
		if err != nil {
			fmt.Println("Download: query user info faild:", err)
			continue
		}
		if total == 0 {
			fmt.Println("Download: User not exists, or post nothing.")
			return
		}
		break
	}

	d := downloading{name: name, proxy: proxy, path: path + "/" + name}

	go query(start, 50)
	for start < total {
		start += 50
		go query(start, 50)
		status := <-conn
		if status.err != nil {
			fmt.Println("Download: query user posts faild:", status.err)
			continue
		}
		var err error
		d.m, err = refine(status.data)
		if err != nil {
			go query(status.start, 50)
			continue
		}
		d.feed(th)
	}
}

type task struct {
	url      string
	filename string
	succeed  bool
}

func (d *downloading) feed(concurrency int) {
	taskbegin := make(chan task, concurrency)
	taskdone := make(chan task, concurrency)
	taskend := make(chan struct{}, concurrency)

	go d.recycle(taskdone)

	for i := 0; i < concurrency; i++ {
		go d.fetch(taskbegin, taskdone, taskend)
	}

	for index, urls := range d.m {
		path := strings.Join([]string{d.path, "/", urls.slug, "_", index}, "")
		os.MkdirAll(path, 0700)
		for _, url := range urls.resURL {
			index := strings.LastIndex(url, "/")
			if index == -1 {
				continue
			}
			var t task
			t.succeed = false
			t.url = url
			t.filename = path + "/" + url[index+1:]
			if urls.resType == "video" {
				t.filename += ".mp4"
			}
			taskbegin <- t
		}
	}
	close(taskbegin)
	for i := 0; i < concurrency; i++ {
		<-taskend
	}
	close(taskend)
	close(taskdone)
}

func (d *downloading) fetch(in <-chan task, out chan<- task, end chan<- struct{}) {
	for t := range in {
		var resp *http.Response
		var err error
		t.succeed = false
		for !t.succeed {
			if d.proxy != "" {
				proxy := func(_ *http.Request) (*url.URL, error) {
					return url.Parse(d.proxy)
				}
				transport := &http.Transport{Proxy: proxy}
				client := &http.Client{Transport: transport}
				resp, err = client.Get(t.url)
			} else {
				resp, err = http.Get(t.url)
			}
			if err != nil {
				fmt.Printf("Download %s failed: %v!", t.url, err)
				time.Sleep(500 * time.Millisecond)
			} else {
				defer resp.Body.Close()
				stream, err := ioutil.ReadAll(resp.Body)
				if err != nil {
					fmt.Printf("Download %s failed: %v!", t.url, err)
					time.Sleep(500 * time.Millisecond)
				} else {
					err = ioutil.WriteFile(t.filename, stream, 0700)
					if err != nil {
						fmt.Printf("Download %s failed: %v!", t.url, err)
						time.Sleep(500 * time.Millisecond)
					} else {
						t.succeed = true
					}
				}
			}
		}
		out <- t
	}
	end <- struct{}{}
}

func (d *downloading) recycle(f <-chan task) {
	for t := range f {
		if t.succeed {
			fmt.Println("Download", t.filename, "succeed!")
		} else {
			fmt.Println("Download", t.filename, "failed!")
		}
	}
}
