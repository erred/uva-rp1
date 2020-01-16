package main

import (
	"context"
	"log"
	"os"

	"github.com/seankhliao/uva-rp1/primary"
)

func main() {
	log.Fatal(primary.New(os.Args, nil).Run(context.Background()))
}
