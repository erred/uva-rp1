package main

import (
	"context"
	"log"
	"os"

	"github.com/seankhliao/uva-rp1/secondary"
)

func main() {
	log.Fatal(secondary.New(os.Args, nil).Run(context.Background()))
}
