package tworker

import (
	"encoding/json"
	"strconv"
	"strings"
)

// tumblrPost struct refines all the information I need.
type tumblrPost struct {
	// ID and Video are interface{}.
	// ID used to be string, but sometimes randomly swithes to int.
	// Video used to be HTML, but sometimes it's bool.
	ID            interface{}  `json:"id"`
	URL           string       `json:"url"`
	Slug          string       `json:"slug"`
	Type          string       `json:"type"`
	PhotoURL      string       `json:"photo-url-1280"`
	Photos        []tumblrPost `json:"photos,omitempty"`
	UnixTimestamp int64        `json:"unix-timestamp"`
	PhotoCaption  string       `json:"photo-caption"`
	Video         interface{}  `json:"video-player"`
	VideoCaption  string       `json:"video-caption"` // For links to outside sites.
	RegularBody   string       `json:"regular-body"`
}

// A tumblrBlog is the outer container for Posts.
type tumblrBlog struct {
	Posts      []tumblrPost `json:"posts"`
	TotalPosts int          `json:"posts-total"`
}

type tumblrResource struct {
	slug      string
	timeStamp int64
	resType   string
	resURL    []string
}

func getTotal(j []byte) (int, error) {
	j = j[22 : len(j)-2]
	var blog tumblrBlog
	err := json.Unmarshal(j, &blog)
	if err != nil {
		return 0, err
	}
	return blog.TotalPosts, nil
}

func refine(j []byte) (map[string]tumblrResource, error) {
	resources := make(map[string]tumblrResource)
	j = j[22 : len(j)-2]
	var blog tumblrBlog
	err := json.Unmarshal(j, &blog)
	if err != nil {
		return nil, err
	}

	for _, post := range blog.Posts {
		var u tumblrResource
		u.timeStamp = post.UnixTimestamp
		u.slug = post.Slug
		u.resType = post.Type
		switch u.resType {
		case "photo":
			u.resURL = refinePhoto(post, u.resURL)
		case "video":
			switch post.Video.(type) {
			case string:
				u.resURL = refineVideo(post.Video.(string), u.resURL)
			}
		case "regular":
			u.resURL = refineRegular(post.RegularBody, u.resURL)
		}
		switch post.ID.(type) {
		case string:
			resources[post.ID.(string)] = u
		case int:
			resources[strconv.Itoa(post.ID.(int))] = u
		case int64:
			resources[strconv.FormatInt(post.ID.(int64), 10)] = u
		}
	}

	return resources, nil
}

func refineMapKey(m map[string]struct{}, u []string) []string {
	for k := range m {
		u = append(u, k)
	}
	return u
}

func refinePhoto(p tumblrPost, u []string) []string {
	m := make(map[string]struct{})
	m[p.PhotoURL] = struct{}{}
	for _, u := range p.Photos {
		m[u.PhotoURL] = struct{}{}
	}
	return refineMapKey(m, u)
}

func refineRegular(r string, u []string) []string {
	m := make(map[string]struct{})
	for {
		index := strings.Index(r, "<img src=\"")
		if index == -1 {
			break
		}
		r = r[index+10:]
		index = strings.Index(r, "\"")
		if index == -1 {
			break
		}
		m[r[:index]] = struct{}{}
		r = r[index+1:]
	}
	return refineMapKey(m, u)
}

func refineVideo(r string, u []string) []string {
	index := strings.Index(r, "<source src=\"")
	if index == -1 {
		return u
	}
	r = r[index+13:]
	index = strings.Index(r, "\"")
	if index == -1 {
		return u
	}
	if strings.Index(r, "tumblr.com/video_file/") == -1 {
		return u
	}

	u = append(u, r[:index])
	return u
}
