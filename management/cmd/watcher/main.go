package main

import (
	"log"
	"os"

	"github.com/seankhliao/uva-rp1/management/watcher"
)

func main() {
	log.SetPrefix("watcher | ")
	watcher.NewServer(os.Args).Serve()
}
