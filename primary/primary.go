package primary

import (
	"context"
	"time"
	// "errors"
	"flag"
	"fmt"

	// "io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	// "time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/seankhliao/uva-rp1/api"
	"google.golang.org/grpc"
	// "google.golang.org/grpc"
)

type endpoint struct {
	uri  string
	cost int64
}

type secondary struct {
	s api.Control_RegisterServer
	r map[string][]endpoint
}

type Primary struct {
	name   string
	port   int
	scrape time.Duration

	// watchers
	watchers []string
	watcher  api.Gossip_ClustersClient

	// prefix - cost
	localRoutes  chan map[string]int64
	localUris    chan []string
	remoteRoutes chan map[string][]endpoint

	secondaries chan map[string]secondary
	// notification
	rebalance chan struct{}

	log *zerolog.Logger
}

func New(args []string, logger *zerolog.Logger) *Primary {
	if logger == nil {
		*logger = log.With().Str("mod", "primary").Logger()
	}

	p := &Primary{
		log: logger,
	}

	var watcher string
	fs := flag.NewFlagSet("primary", flag.ExitOnError)
	fs.DurationVar(&p.scrape, "scrape", 15*time.Second, "scrape interval")
	fs.StringVar(&watcher, "watcher", "145.100.104.117:8000", "host:port of watcher to connect to")
	fs.StringVar(&p.name, "name", strconv.FormatInt(rand.Int63(), 10), "overrdide randomly generated name of node")
	fs.IntVar(&p.port, "port", 8000, "port to serve on")
	fs.Parse(args)
	p.watchers = strings.Split(watcher, ",")
	return p
}

func (p *Primary) Run() error {
	first := make(chan struct{})
	go p.scraper(first)
	<-first

	// gossip with watcher
	go p.gossipRunner(context.Background())
	go p.rebalancer()

	// httpServer := http.ServeMux{}
	// httpServer.Handle("/metrics", promhttp.Handler())
	//
	grpcServer := grpc.NewServer()
	api.RegisterControlServer(grpcServer, p)

	return http.ListenAndServe(fmt.Sprintf(":%d", p.port), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
			// } else {
			// 	panic("Unimplemented: Run")
			// 		httpServer.ServeHTTP(w, r)
		}
	}))
}
