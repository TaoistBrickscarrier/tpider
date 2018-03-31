package main

import (
	"flag"
	"log"
	"os"
	"strings"

	"github.com/TinkerBravo/tpider/twork"
)

var (
	tproxy   = flag.String("proxy", "", "host:port to reach tumblr")
	tuser    = flag.String("user", "staff", "Tumblr user name.")
	keeppath = flag.String("path", ".", "Path to keep downloaded media files.")
	thread   = flag.Int("thread", 5, "Threads of downloading")
)

func main() {
	err := os.MkdirAll(strings.Join([]string{*keeppath, "/", *tuser}, ""), 0700)
	if err != nil {
		log.Fatal(err)
	}
	tworker.Download(*tuser, *tproxy, *keeppath, *thread)
}

func init() {
	flag.Parse()
	if *tproxy != "" && strings.Index(*tproxy, "http://") != 0 {
		*tproxy = "http://" + *tproxy
	}
}
