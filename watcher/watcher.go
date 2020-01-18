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
	fs.StringVar(&w.name, "name", "", "id of node")
	fs.IntVar(&w.port, "port", 8000, "port to serve on")
	fs.Var(&ws, "watcher", "(repeatable) other watchers to connect to (full mesh)")
	fs.Parse(args)

	if w.name == "" {
		addr, err := interfaceAddrs()
		if err != nil {
			w.log.Fatal().Msg("no name found")
		}
		w.name = addr
	}

	w.log.Info().Str("name", w.name).Int("port", w.port).Str("file_sd", w.promfile).Strs("watchers", w.watchers).Msg("created")
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
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", w.port))
	if err != nil {
		w.log.Fatal().Err(err).Int("port", w.port).Msg("can't listen to port")
	}
	return grpcServer.Serve(lis)
	// return http.ListenAndServe(fmt.Sprintf(":%d", w.port), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	if strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
	// 		grpcServer.ServeHTTP(w, r)
	// 	}
	// }))
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

func interfaceAddrs() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", fmt.Errorf("interface addrs: %w", err)
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
	ips := append(ip4, ip6...)
	if len(ips) == 0 {
		return "", fmt.Errorf("no addresses found")
	}
	return ips[0], nil
}
