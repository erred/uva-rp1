package main

import (
	"log"
	"os"

	"github.com/seankhliao/uva-rp1/secondary"
)

func main() {
	log.Fatal(secondary.New(os.Args[1:], nil).Run())
}
