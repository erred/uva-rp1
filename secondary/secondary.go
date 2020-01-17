package secondary

import (
	"context"
	"flag"
	"fmt"
	"github.com/rs/zerolog"
	"github.com/seankhliao/uva-rp1/api"
	"github.com/seankhliao/uva-rp1/nfdstat"
	"google.golang.org/grpc"
	"math/rand"
	"os"
	"strconv"
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
	primary  string
	strategy string
	name     string

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
		primaries: make(chan map[string]primary, 1),

		stat: nfdstat.New(),
		log:  logger,
	}

	s.primaries <- make(map[string]primary)

	fs := flag.NewFlagSet("secondary", flag.ExitOnError)
	fs.StringVar(&s.primary, "primary", "145.100.104.117:8000", "host:port of primaary to connect to")
	fs.StringVar(&s.strategy, "strategy", "/localhost/nfd/strategy/asf", "set routing strategy")
	fs.StringVar(&s.name, "name", strconv.FormatInt(rand.Int63(), 10), "overrdide randomly generated name of node")
	fs.Parse(args)
	return s
}

func (s *Secondary) Run() error {
	s.log.Info().Msg("starting run")
	// set default prefix
	for {
		err := nfdstat.RouteStrategy(context.Background(), "/", s.strategy)
		if err != nil {
			s.log.Error().Err(err).Str("prefix", "/").Str("strategy", s.strategy).Msg("set strategy")
			time.Sleep(time.Second)
			continue
		}
		break
	}
	retry := time.Second
	for {
		connected, err := s.run()
		if connected {
			retry = time.Second

		}
		s.log.Error().Err(err).Dur("backoff", retry).Msg("run")
		time.Sleep(retry)
		if retry < 32*time.Second {
			retry *= 2
		}
	}

}

func (s *Secondary) run() (bool, error) {
	// establish connection
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, s.primary, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return false, fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	s.ctl = api.NewInfoClient(conn)

	go s.pushStatus()

	c, err := s.ctl.Register(ctx, &api.RegisterRequest{})
	if err != nil {
		return true, fmt.Errorf("register: %w", err)
	}
	for {
		err = s.recvCmd(c)
		if err != nil {
			return true, err
		}
	}
}
