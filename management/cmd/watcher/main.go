package main

import (
	"github.com/seankhliao/uva-rp1/management/watcher"
	"log"
)

func main() {
	log.SetPrefix("watcher | ")
	c := watcher.NewClient()
	c.Run()
}
