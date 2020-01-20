package main

import (
	"log"
	"os"

	"github.com/seankhliao/uva-rp1/watcher"
)

func main() {
	log.Fatal(watcher.New(os.Args[1:], nil).Run())
}
