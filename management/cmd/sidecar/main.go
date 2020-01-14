package main

import (
	"github.com/seankhliao/uva-rp1/management/sidecar"
	"log"
)

func main() {
	log.SetPrefix("sidecar | ")
	c := sidecar.NewClient()
	c.Run()
}
