package main

import (
	"log"
	"os"

	"github.com/seankhliao/uva-rp1/sidecar"
)

func main() {
	log.SetPrefix("sidecar | ")
	sidecar.NewServer(os.Args).Serve()
}
