package main

import (
	"log"
	"os"

	"github.com/seankhliao/uva-rp1/primary"
)

func main() {
	log.Fatal(primary.New(os.Args[1:], nil).Run())
}
