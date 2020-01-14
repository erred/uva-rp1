package sidecar

import (
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/seankhliao/uva-rp1/management/api"
	"google.golang.org/grpc"
)

type Server struct {
	port           int
	scrapeInterval time.Duration
	faceBackoff    time.Duration
	collectors     collectors
	iaddrs         endpoints

	// memory
	nfdname  string
	nfdpid   int
	pagesize int

	// 1 buffered store
	routes chan routes
	addrs  chan endpoints

	// prometheus metrics
	memory          prometheus.Gauge
	cs_capacity     prometheus.Gauge
	cs_entries      prometheus.Gauge
	cs_hits         prometheus.Gauge
	cs_misses       prometheus.Gauge
	nt_entries      prometheus.Gauge
	fib_entries     prometheus.Gauge
	rib_entries     prometheus.Gauge
	pit_entries     prometheus.Gauge
	channel_entries prometheus.Gauge
	face_entries    prometheus.Gauge

	// statisfied: yes|no
	interests *prometheus.GaugeVec

	// direction: in|out
	// type: interest|data|nack
	pkts *prometheus.GaugeVec

	// direction: in|out
	// type: interets|data|nack
	// id: $value
	face_pkts *prometheus.GaugeVec

	// direction: in|out
	// id: $value
	face_bytes *prometheus.GaugeVec
}

func NewServer(args []string) *Server {
	s := Server{
		iaddrs:   newEndpoints(),
		pagesize: os.Getpagesize(),
		routes:   make(chan routes, 1),
		addrs:    make(chan endpoints, 1),

		memory: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "nfd_memory_bytes",
		}),
		cs_capacity: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "nfd_cs_capacity",
		}),
		cs_entries: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "nfd_cs_entries",
		}),
		cs_hits: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "nfd_cs_hits",
		}),
		cs_misses: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "nfd_cs_misses",
		}),
		nt_entries: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "nfd_nametree_entries",
		}),
		fib_entries: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "nfd_fib_entries",
		}),
		rib_entries: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "nfd_rib_entries",
		}),
		pit_entries: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "nfd_pit_entries",
		}),
		channel_entries: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "nfd_channel_entries",
		}),

		interests: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nfd_interets",
			},
			[]string{"satisfied"},
		),
		pkts: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nfd_pkts",
			},
			[]string{"direction", "type"},
		),

		face_entries: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "nfd_faces",
		}),
		face_bytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nfd_face_bytes",
			},
			[]string{"direction", "id"},
		),
		face_pkts: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nfd_face_pkts",
			},
			[]string{"direction", "type", "id"},
		),
	}

	fs := flag.NewFlagSet("sidecar", flag.ExitOnError)
	fs.IntVar(&s.port, "port", 8000, "port to serve on")

	fs.DurationVar(&s.scrapeInterval, "scrape", 5*time.Second, "time between scrapes of nfd status")
	fs.DurationVar(&s.faceBackoff, "backoff", 5*time.Minute, "time between retries of face create")
	fs.Var(&s.collectors, "collectors", "comma list of collecter host:ports")
	fs.Var(&s.iaddrs, "addrs", "comma list of public IPs announce (overrides detection)")
	fs.StringVar(&s.nfdname, "nfdname", "nfd", "name of nfd executable")
	fs.Parse(args)

	if len(s.collectors.cs) == 0 {
		log.Printf("WARNING no collectors set\n")
	}
	s.addrs <- s.iaddrs
	s.routes <- routes{}

	var err error
	s.nfdpid, err = pid("nfd")
	if err != nil {
		log.Fatalln("NewClient ", err)
	}

	prometheus.MustRegister(s.memory)
	prometheus.MustRegister(s.cs_capacity)
	prometheus.MustRegister(s.cs_entries)
	prometheus.MustRegister(s.cs_hits)
	prometheus.MustRegister(s.cs_misses)
	prometheus.MustRegister(s.nt_entries)
	prometheus.MustRegister(s.fib_entries)
	prometheus.MustRegister(s.rib_entries)
	prometheus.MustRegister(s.pit_entries)
	prometheus.MustRegister(s.channel_entries)
	prometheus.MustRegister(s.interests)
	prometheus.MustRegister(s.pkts)
	prometheus.MustRegister(s.face_entries)
	prometheus.MustRegister(s.face_bytes)
	prometheus.MustRegister(s.face_pkts)

	return &s
}

func (s *Server) Serve() {
	// scrape status
	go func() {
		for range time.NewTicker(s.scrapeInterval).C {
			go s.status()
		}
	}()
	s.status()

	// register with watchers
	// for _, a := range strings.Split(s.collectors, ",") {
	// 	go s.registerKeepAlive(s.keepAlive, a)
	// }

	httpServer := http.ServeMux{}
	httpServer.Handle("/metrics", promhttp.Handler())

	grpcServer := grpc.NewServer()
	api.RegisterDiscoveryServiceServer(grpcServer, s)

	http.ListenAndServe(fmt.Sprintf(":%d", s.port), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			httpServer.ServeHTTP(w, r)
		}
	}))
}

type collectors struct {
	cs []*url.URL
}

func (c collectors) String() string {
	us := make([]string, len(c.cs))
	for i, u := range c.cs {
		us[i] = u.Host
	}
	return strings.Join(us, ",")
}
func (c *collectors) Set(s string) error {
	for i, s := range strings.Split(s, ",") {
		u, err := url.Parse(s)
		if err != nil {
			return fmt.Errorf("parse collectors %d: %w", i, err)
		}
		c.cs = append(c.cs, u)
	}
	return nil
}

type routes struct {
	rs []api.Route
}

// func (r routes) String() string {
// 	s := make([]string, len(r.rs))
// 	for i, rr := range r.rs {
// 		s[i] = fmt.Sprintf("[%s %d %s]", rr.Source, rr.Cost, rr.Route)
// 	}
// 	return strings.Join(s, " ")
// }

func (r *routes) update(s *NFDStatus) {
	var rs []api.Route
	for _, rr := range s.Rib.RibEntry {
		rs = append(rs, api.Route{
			Source: rr.Routes.Route.Origin,
			Cost:   rr.Routes.Route.Cost,
			Route:  rr.Prefix,
		})
	}
	r.rs = rs
}

type endpoints struct {
	overrride bool
	ip4, ip6  []net.IP
	es        []api.Endpoint
}

func newEndpoints() endpoints {
	var e endpoints
	ads, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatalln("NewClient ", err)
	}
	for _, ad := range ads {
		an, ok := ad.(*net.IPNet)
		if !ok {
			continue
		}
		if !an.IP.IsGlobalUnicast() {
			continue
		}
		a4 := an.IP.To4()
		if a4 != nil {
			e.ip4 = append(e.ip4, a4)
		} else {
			e.ip6 = append(e.ip6, an.IP)
		}
	}
	return e
}

func (e endpoints) String() string {
	es := make([]string, len(e.es))
	for i, e := range e.es {
		es[i] = fmt.Sprintf("%s://%s:%s", e.Proto, e.Host, e.Port)
	}
	return strings.Join(es, ",")
}
func (e *endpoints) Set(s string) error {
	e.overrride = true
	e.es = e.es[:0]
	for i, us := range strings.Split(s, ",") {
		u, err := url.Parse(us)
		if err != nil {
			return fmt.Errorf("pparse addrs %d: %w", i, err)
		}
		e.es = append(e.es, api.Endpoint{
			Host:  u.Hostname(),
			Proto: u.Scheme,
			Port:  u.Port(),
		})
	}
	return nil
}
func (e *endpoints) update(s *NFDStatus) {
	if e.overrride {
		return
	}
	var es []api.Endpoint
	for _, ch := range s.Channels.Channel {
		u, err := url.Parse(ch.LocalUri)
		if err != nil {
			log.Printf("endpoints update parse %s: %s", ch.LocalUri, err)
			continue
		}
		var ips []net.IP
		switch u.Scheme {
		case "tcp4", "udp4":
			ips = e.ip4
		case "tcp6", "udp6":
			ips = e.ip6
		}
		for _, ip := range ips {
			es = append(es, api.Endpoint{
				Host:  ip.String(),
				Proto: u.Scheme,
				Port:  u.Port(),
			})
		}
	}
	e.es = es
}
