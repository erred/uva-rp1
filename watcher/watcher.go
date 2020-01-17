package watcher

import (
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"

	"github.com/rs/zerolog"
	"github.com/seankhliao/uva-rp1/api"
)

type Watcher struct {
	promfile string
	name     string
	port     int
	watchers []string

	primaries     chan map[string]primary
	reflectors    chan map[string]reflector
	notififcation chan struct{}
	log           *zerolog.Logger
}

func New(args []string, logger *zerolog.Logger) *Watcher {
	if logger == nil {
		l := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, NoColor: true, TimeFormat: time.RFC3339Nano}).With().Timestamp().Logger()
		logger = &l
	}

	w := &Watcher{
		primaries:     make(chan map[string]primary, 1),
		reflectors:    make(chan map[string]reflector, 1),
		notififcation: make(chan struct{}, 1),
		log:           logger,
	}
	w.primaries <- make(map[string]primary)
	w.reflectors <- make(map[string]reflector)

	ws := flagslice(w.watchers)
	fs := flag.NewFlagSet("watcher", flag.ExitOnError)
	fs.StringVar(&w.promfile, "file", "/etc/prometheus/file_sd.json", "prometheus file sd path")
	fs.StringVar(&w.name, "name", strconv.FormatInt(rand.Int63(), 10), "overrdide randomly generated name of node")
	fs.IntVar(&w.port, "port", 8000, "port to serve on")
	fs.Var(&ws, "watcher", "(repeatable) other watchers to connect to (full mesh)")
	fs.Parse(args)

	return w
}

func (w *Watcher) Run() error {
	w.log.Info().Msg("starting run")

	go w.notifier()

	for _, wa := range w.watchers {
		go w.gossiper(wa)
	}

	grpcServer := grpc.NewServer()
	api.RegisterReflectorServer(grpcServer, w)

	w.log.Info().Int("port", w.port).Msg("serving")
	return http.ListenAndServe(fmt.Sprintf(":%d", w.port), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		}
	}))
}

type gossiper interface {
	gossipRecver
	gossipSender
}
type gossipRecver interface {
	Recv() (*api.AllPrimaries, error)
}
type gossipSender interface {
	Send(*api.AllPrimaries) error
}

type primary struct {
	p *api.Primary
	s api.Reflector_PrimariesServer
}
type reflector struct {
	a *api.AllPrimaries
	s gossipSender
}

type flagslice []string

func (f *flagslice) String() string {
	return strings.Join(*f, ",")
}
func (f *flagslice) Set(s string) error {
	*f = append(*f, s)
	return nil
}
