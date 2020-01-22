package main

import (
	"log"
	"os"

	"github.com/seankhliao/uva-rp1/dash"
)

func main() {
	log.Fatal(dash.New(os.Args[1:], nil).Run())
}
