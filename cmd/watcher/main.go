package main

import (
	"context"
	"log"
	"os"

	"github.com/seankhliao/uva-rp1/watcher"
)

func main() {
	log.Fatal(watcher.New(os.Args, nil).Run(context.Background()))
}
