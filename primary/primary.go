package primary

import (
	"context"
	// "errors"
	"flag"
	"fmt"
	// "io"
	"math/rand"
	// "net/url"
	"net/http"
	"strconv"
	"strings"
	// "time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/seankhliao/uva-rp1/api"
	"google.golang.org/grpc"
	// "github.com/seankhliao/uva-rp1/api"
	// "google.golang.org/grpc"
)

type Primary struct {
	name string
	port int

	// watchers

	// secondaries
	// + route assignments
	// + status

	// clusters

	// routes

	log *zerolog.Logger
}

func New(args []string, logger *zerolog.Logger) *Primary {
	if logger == nil {
		*logger = log.With().Str("mod", "primary").Logger()
	}

	p := &Primary{
		log: logger,
	}

	fs := flag.NewFlagSet("primary", flag.ExitOnError)
	fs.StringVar(&p.name, "name", strconv.FormatInt(rand.Int63(), 10), "overrdide randomly generated name of node")
	fs.IntVar(&p.port, "port", 8000, "port to serve on")
	fs.Parse(args)
	return p
}

func (p *Primary) Run(ctx context.Context) error {
	// ensure get first info

	//  start route manager

	// start gossip with watcher

	// httpServer := http.ServeMux{}
	// httpServer.Handle("/metrics", promhttp.Handler())
	//
	grpcServer := grpc.NewServer()
	// api.RegisterGossipServer(grpcServer, p)
	api.RegisterControlServer(grpcServer, p)
	api.RegisterInfoServer(grpcServer, p)

	return http.ListenAndServe(fmt.Sprintf(":%d", p.port), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			panic("Unimplemented: Run")
			// 		httpServer.ServeHTTP(w, r)
		}
	}))
}
