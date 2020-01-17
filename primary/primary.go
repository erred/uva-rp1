package primary

import (
	"context"
	"flag"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/seankhliao/uva-rp1/api"
	"github.com/seankhliao/uva-rp1/nfdstat"
	"google.golang.org/grpc"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

type endpoint struct {
	uri  string
	cost int64
}

type secondary struct {
	s api.Control_RegisterServer
	h api.Status_StreamServer
	r map[string][]endpoint
}

type Primary struct {
	scrapeInterval time.Duration
	name           string
	strategy       string
	port           int
	localUris      []string
	localChan      []string
	watchers       []string

	// watchers
	watcher api.Gossip_ClustersClient

	// prefix - cost
	localStat    chan *api.StatusRequest
	localRoutes  chan map[string]int64
	remoteRoutes chan map[string][]endpoint
	secondaries  chan map[string]secondary

	// notification
	rebalance chan struct{}

	stat *nfdstat.Stat
	log  *zerolog.Logger
}

func New(args []string, logger *zerolog.Logger) *Primary {
	if logger == nil {
		l := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, NoColor: true, TimeFormat: time.RFC3339Nano}).With().Timestamp().Logger()
		logger = &l
	}

	p := &Primary{
		localStat:    make(chan *api.StatusRequest, 1),
		localRoutes:  make(chan map[string]int64, 1),
		remoteRoutes: make(chan map[string][]endpoint, 1),
		secondaries:  make(chan map[string]secondary, 1),
		rebalance:    make(chan struct{}, 1),

		stat: nfdstat.New(),
		log:  logger,
	}
	p.localStat <- nil
	p.localRoutes <- make(map[string]int64)
	p.remoteRoutes <- make(map[string][]endpoint)
	p.secondaries <- make(map[string]secondary)

	fw, pi := flagwatcher(p.watchers), flagwatcher(p.localUris)
	fs := flag.NewFlagSet("primary", flag.ExitOnError)
	fs.DurationVar(&p.scrapeInterval, "scrape", 15*time.Second, "scrape interval")
	fs.StringVar(&p.strategy, "strategy", "/localhost/nfd/strategy/access-router", "default strategy")
	fs.StringVar(&p.name, "name", strconv.FormatInt(rand.Int63(), 10), "overrdide randomly generated name of node")
	fs.IntVar(&p.port, "port", 8000, "port to serve on")
	fs.Var(&fw, "watcher", "(repeatable) host:port of watcher to connect to")
	fs.Var(&pi, "public", "(repeatable) public ip of this local node to advertise")
	fs.Parse(args)
	return p
}

func (p *Primary) Run() error {
	first := make(chan struct{})
	go p.scraper(first)
	<-first

	// gossip with watcher
	go p.gossipRunner(context.Background())
	go p.rebalancer()

	// TODO: register prometheus metrics
	// httpServer := http.ServeMux{}
	// httpServer.Handle("/metrics", promhttp.Handler())
	//
	grpcServer := grpc.NewServer()
	api.RegisterControlServer(grpcServer, p)
	api.RegisterStatusServer(grpcServer, p)

	return http.ListenAndServe(fmt.Sprintf(":%d", p.port), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
			// } else {
			// 	panic("Unimplemented: Run")
			// 		httpServer.ServeHTTP(w, r)
		}
	}))
}

type flagwatcher []string

func (f *flagwatcher) String() string {
	return strings.Join(*f, ",")
}
func (f *flagwatcher) Set(s string) error {
	*f = append(*f, s)
	return nil
}
