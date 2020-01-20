package watcher

import (
	"flag"
	"fmt"
	"net"
	// "net/http"
	"os"
	"strings"
	"time"

	"google.golang.org/grpc"

	"github.com/rs/zerolog"
	"github.com/seankhliao/uva-rp1/api"
)

type Watcher struct {
	watchers  []string
	promfile  string
	localAddr string
	port      int

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

	ips, err := localIPs()
	if err != nil {
		w.log.Fatal().
			Err(err).
			Msg("no known public ip to announce")
	} else if len(ips) > 0 {
		w.localAddr = ips[0]
	}

	ws := flagslice{w.watchers}
	fs := flag.NewFlagSet("watcher", flag.ExitOnError)
	fs.StringVar(&w.promfile, "file", "/etc/prometheus/file_sd.json", "prometheus file sd path")
	fs.IntVar(&w.port, "port", 8000, "port to serve on")
	fs.Var(&ws, "watcher", "(repeatable) other watchers to connect to (full mesh)")
	fs.Parse(args)
	w.watchers = ws.s

	w.log.Info().
		Str("file_sd", w.promfile).
		Strs("watchers", w.watchers).
		Str("addr", w.localAddr).
		Int("port", w.port).
		Msg("initialized")
	return w
}

func (w *Watcher) Run() error {
	w.log.Info().
		Str("id", w.localAddr).
		Msg("starting watcher")

	go w.notifier()

	for _, wa := range w.watchers {
		go w.gossipRunner(wa)
	}

	grpcServer := grpc.NewServer()
	api.RegisterReflectorServer(grpcServer, w)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", w.port))
	if err != nil {
		w.log.Fatal().
			Err(err).
			Int("port", w.port).
			Msg("can't listen to port")
	}
	return grpcServer.Serve(lis)
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

type flagslice struct {
	s []string
}

func (f *flagslice) String() string {
	return strings.Join(f.s, ",")
}
func (f *flagslice) Set(s string) error {
	f.s = append(f.s, s)
	return nil
}

func localIPs() ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("localIPs: %w", err)
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
