package secondary

import (
	"context"
	"flag"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/seankhliao/uva-rp1/api"
	"github.com/seankhliao/uva-rp1/nfdstat"
	"google.golang.org/grpc"
	"net"
	"os"
	"time"
)

type route struct {
	prefix string
	uri    string
}
type primary struct {
	p    api.Primary
	c    api.InfoClient
	ch   string
	conn *grpc.ClientConn
}

type Secondary struct {
	strategy  string
	primary   string
	localAddr string
	// port     int

	localChan chan []string
	primaries chan map[string]primary

	ctl  api.InfoClient
	stat *nfdstat.Stat
	log  *zerolog.Logger
}

func New(args []string, logger *zerolog.Logger) *Secondary {
	if logger == nil {
		l := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, NoColor: true, TimeFormat: time.RFC3339Nano}).With().Timestamp().Logger()
		logger = &l
	}

	s := &Secondary{
		localChan: make(chan []string, 1),
		primaries: make(chan map[string]primary, 1),

		stat: nfdstat.New(),
		log:  logger,
	}

	s.localChan <- nil
	s.primaries <- make(map[string]primary)

	ips, err := localIPs()
	if err != nil {
		s.log.Fatal().Err(err).Msg("no known public ip to announce")
	} else if len(ips) > 0 {
		s.localAddr = ips[0]
	}

	fs := flag.NewFlagSet("secondary", flag.ExitOnError)
	fs.StringVar(&s.strategy, "strategy", "/localhost/nfd/strategy/asf", "set routing strategy")
	fs.StringVar(&s.primary, "primary", "0.0.0.0:8000", "host:port of primaary to connect to")
	fs.Parse(args)

	s.log.Info().
		Str("strategy", s.strategy).
		Str("primary", s.primary).
		Str("addr", s.localAddr).
		Msg("initialized")
	return s
}

func (s *Secondary) Run() error {
	s.log.Info().
		Str("id", s.localAddr).
		Msg("starting secondary")

	s.mustDefaultStrategy()
	s.mustGetChannels()

	s.registerRunner()
	return nil
}

func (s *Secondary) mustDefaultStrategy() {
	for {
		err := nfdstat.RouteStrategy(context.Background(), "/", s.strategy)
		if err != nil {
			s.log.Error().
				Err(err).
				Str("strategy", s.strategy).
				Msg("set default strategy")
			time.Sleep(time.Second)
			continue
		}
		break
	}
	s.log.Info().
		Str("strategy", s.strategy).
		Msg("default strategy set")
}

func (s *Secondary) registerRunner() {
	retry := time.Second
	for {
		connected, err := s.register()
		if connected {
			retry = time.Second

		}
		s.log.Error().
			Err(err).
			Dur("backoff", retry).
			Msg("register runner")
		time.Sleep(retry)
		if retry < 16*time.Second {
			retry *= 2
		}
	}

}

func (s *Secondary) register() (bool, error) {
	// establish connection
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, s.primary, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return false, fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	s.ctl = api.NewInfoClient(conn)

	ch := <-s.localChan
	chs := make([]string, len(ch))
	copy(chs, ch)
	s.localChan <- ch

	c, err := s.ctl.Register(ctx, &api.RegisterRequest{
		SecondaryId: s.localAddr,
		Channels:    chs,
	})
	if err != nil {
		return true, fmt.Errorf("register: %w", err)
	}

	s.log.Info().
		Str("primary", s.primary).
		Msg("registered")

	go s.statusPusher()

	for {
		err = s.recvCmd(c)
		if err != nil {
			return true, err
		}
	}
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
