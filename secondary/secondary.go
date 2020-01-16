package secondary

import (
	"context"
	"flag"
	"math/rand"
	"strconv"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/seankhliao/uva-rp1/api"
	"github.com/seankhliao/uva-rp1/nfdstat"
)

type route struct {
	prefix string
	uri    string
}

type Secondary struct {
	primary  string
	name     string
	strategy string

	faces  chan map[int64]string
	routes chan map[route]struct{}

	client api.ControlClient
	stat   *nfdstat.Server
	log    *zerolog.Logger
}

func New(args []string, logger *zerolog.Logger) *Secondary {
	if logger == nil {
		*logger = log.With().Str("mod", "secondary").Logger()
	}

	s := &Secondary{
		faces:  make(chan map[int64]string, 1),
		routes: make(chan map[route]struct{}, 1),

		stat: nfdstat.New(logger),
		log:  logger,
	}

	s.faces <- make(map[int64]string)
	s.routes <- make(map[route]struct{})

	fs := flag.NewFlagSet("secondary", flag.ExitOnError)
	fs.StringVar(&s.primary, "primary", "145.100.104.117:8000", "host:port of primaary to connect to")
	fs.StringVar(&s.name, "name", strconv.FormatInt(rand.Int63(), 10), "overrdide randomly generated name of node")
	fs.StringVar(&s.strategy, "strategy", "/localhost/nfd/strategy/asf", "set routing strategy")
	fs.Parse(args)
	return s
}

func (s *Secondary) Run() error {
	// set default prefix
	var err error
	for err != nil {
		err = nfdstat.RouteStrategy(context.Background(), "/", s.strategy)
		if err != nil {
			s.log.Error().Err(err).Str("prefix", "/").Str("strategy", s.strategy).Msg("set strategy")
		}
	}

	retry := time.Second
	for {
		err := s.handle(context.Background())
		s.log.Error().Err(err).Dur("backoff", retry).Msg("register")
		time.Sleep(retry)
		if err != nil {
			if retry < 32*time.Second {
				retry *= 2
			}
		} else {
			retry = time.Second
		}
	}
}

