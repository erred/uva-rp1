package primary

import (
	"flag"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/seankhliao/uva-rp1/api"
	"github.com/seankhliao/uva-rp1/nfdstat"
	"google.golang.org/grpc"
)

type wantRoute struct {
	s api.Info_RoutesServer
}
type primary struct {
	p api.Primary
}
type secondary struct {
	p map[string]primary
	c api.Info_RegisterServer
	s api.Info_PushStatusServer
}

type Primary struct {
	scrapeInterval time.Duration
	name           string
	strategy       string
	localAddr      string
	watcherAddr    string
	port           int

	watcher api.Reflector_PrimariesClient

	localChan   chan []string
	localRoutes chan map[string]int64
	primaries   chan map[string]primary
	secondaries chan map[string]secondary
	wantRoutes  chan map[string]wantRoute
	status      chan *nfdstat.Status

	// notification
	routeNotify     chan struct{}
	secondaryNotify chan struct{}

	stat *nfdstat.Stat
	log  *zerolog.Logger
}

func New(args []string, logger *zerolog.Logger) *Primary {
	if logger == nil {
		l := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, NoColor: true, TimeFormat: time.RFC3339Nano}).With().Timestamp().Logger()
		logger = &l
	}
	var defaultIP string
	ips, err := localIPs()
	if err != nil {
		// handle error
	} else if len(ips) > 0 {
		defaultIP = ips[0]
	}

	p := &Primary{
		localChan:   make(chan []string, 1),
		localRoutes: make(chan map[string]int64, 1),
		primaries:   make(chan map[string]primary, 1),
		secondaries: make(chan map[string]secondary, 1),
		wantRoutes:  make(chan map[string]wantRoute, 1),
		status:      make(chan *nfdstat.Status, 1),

		routeNotify:     make(chan struct{}, 1),
		secondaryNotify: make(chan struct{}, 1),

		stat: nfdstat.New(),
		log:  logger,
	}

	p.localChan <- nil
	p.localRoutes <- make(map[string]int64)
	p.primaries <- make(map[string]primary)
	p.secondaries <- make(map[string]secondary)
	p.wantRoutes <- make(map[string]wantRoute)
	p.status <- nil

	fs := flag.NewFlagSet("primary", flag.ExitOnError)
	fs.DurationVar(&p.scrapeInterval, "scrape", 15*time.Second, "scrape interval")
	fs.StringVar(&p.name, "name", strconv.FormatInt(rand.Int63(), 10), "overrdide randomly generated name of node")
	fs.StringVar(&p.strategy, "strategy", "/localhost/nfd/strategy/access-router", "default strategy")
	fs.StringVar(&p.watcherAddr, "watcher", "145.100.104.117:8000", "host:port of watcher to connect to")
	fs.StringVar(&p.localAddr, "addr", defaultIP, "public ip to announce")
	fs.IntVar(&p.port, "port", 8000, "port to serve on")
	fs.Parse(args)

	if p.localAddr == "" {
		p.log.Fatal().Msg("no public ip addr found")
	}
	return p
}

func (p *Primary) Run() error {
	go p.distributor()
	first := make(chan struct{})
	go p.scraper(first)
	<-first
	go p.routeAdvertiser()
	go p.primaryDiscoverer()

	// TODO: register prometheus metrics
	// httpServer := http.ServeMux{}
	// httpServer.Handle("/metrics", promhttp.Handler())
	//
	grpcServer := grpc.NewServer()
	api.RegisterInfoServer(grpcServer, p)

	return http.ListenAndServe(fmt.Sprintf(":%d", p.port), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
			// } else {
			// 	panic("Unimplemented: Run")
			// 		httpServer.ServeHTTP(w, r)
		}
	}))
}

// type flagSlice []string
//
// func (f *flagSlice) String() string {
// 	return strings.Join(*f, ",")
// }
// func (f *flagSlice) Set(s string) error {
// 	*f = append(*f, s)
// 	return nil
// }

func localIPs() ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("interface addrs: %w", err)
	}
	var ip4, ip6 []string
	for _, addr := range addrs {
		an, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		if !an.IP.IsGlobalUnicast() {
			continue
		}
		if i4 := an.IP.To4(); i4 != nil {
			ip4 = append(ip4, i4.String())
		} else {
			ip6 = append(ip6, an.IP.String())
		}
	}
	return append(ip4, ip6...), nil
}
