package secondary

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/seankhliao/uva-rp1/api"
	"github.com/seankhliao/uva-rp1/nfdstat"
	"google.golang.org/grpc"
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

func (s *Secondary) Run(ctx context.Context) error {
	// set default prefix
	var err error
	for err != nil {
		err = nfdstat.RouteStrategy(ctx, "/", s.strategy)
		if err != nil {
			s.log.Error().Err(err).Str("prefix", "/").Str("strategy", s.strategy).Msg("set strategy")
		}
	}

	retry := time.Second
	for {
		err := s.handle(ctx)
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

func (s *Secondary) handle(ctx context.Context) error {
	// establish connection
	conn, err := grpc.DialContext(ctx, s.primary, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	s.client = api.NewControlClient(conn)
	c, err := s.client.RegisterSecondary(ctx)
	if err != nil {
		return fmt.Errorf("register: %w", err)
	}
	_, stat := s.stat.Status()

	// send initial
	err = c.Send(s.update(stat))
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}
	// respond to commands
	for {
		cm, err := c.Recv()
		if errors.Is(err, io.EOF) {
			log.Info().Msg("receive EOF")
			return nil
		} else if err != nil {
			log.Error().Err(err).Msg("receive")
			continue
		}

		// get copy of current state
		mf := <-s.faces
		// faces - count of routes
		nmf := make(map[string]int, len(mf))
		for _, v := range mf {
			nmf[v] = 0
		}
		s.faces <- mf
		mr := <-s.routes
		// routes - set
		nmr := make(map[route]struct{}, len(mr))
		for k, v := range mr {
			nmr[k] = v
			nmf[k.uri]++
		}
		s.routes <- mr

		// filter face/route changes
		for _, r := range cm.Routes {
			rt := route{r.Prefix, r.Endpoint}
			if _, ok := nmr[rt]; ok {
				delete(nmr, rt)
				nmf[r.Endpoint]--
			} else {
				if _, ok := nmf[r.Endpoint]; !ok {
					s.AddFace(r.Endpoint)
				}
				s.AddRoute(rt, r.Cost)
			}
		}
		for r := range nmr {
			s.DelRoute(r)
			nmf[r.uri]--
		}
		for f, c := range nmf {
			if c <= 0 {
				s.DelFace(f)
			}
		}

		// send response
		_, stat := s.stat.Status()
		err = c.Send(s.update(stat))
		if err != nil {
			log.Error().Err(err).Msg("send")
		}
	}
}

func (s *Secondary) update(stat *nfdstat.NFDStatus) *api.SecondaryInfo {
	mf := make(map[int64]string, len(stat.Faces.Face))
	rs := make([]*api.Route, len(stat.Rib.RibEntry))
	mr := make(map[route]struct{}, len(stat.Rib.RibEntry))

	<-s.faces
	<-s.routes

	for _, f := range stat.Faces.Face {
		if !strings.HasPrefix(f.RemoteUri, "tcp") && !strings.HasPrefix(f.RemoteUri, "udp") {
			continue
		}
		mf[f.FaceId] = f.RemoteUri
	}
	for _, r := range stat.Rib.RibEntry {
		mr[route{r.Prefix, mf[r.Routes.Route.FaceId]}] = struct{}{}
		rs = append(rs, &api.Route{
			Prefix:   r.Prefix,
			Endpoint: mf[r.Routes.Route.FaceId],
			Cost:     r.Routes.Route.Cost,
		})
	}

	s.faces <- mf
	s.routes <- mr

	return &api.SecondaryInfo{
		Id:            s.name,
		Routes:        rs,
		CacheCapacity: stat.Cs.Capacity,
		CacheEntries:  stat.Cs.NEntries,
		CacheHits:     stat.Cs.NHits,
		CacheMisses:   stat.Cs.NMisses,
	}
}

func (s *Secondary) AddFace(uri string) {
	ctx := context.Background()
	err := nfdstat.AddFace(ctx, uri)
	if err != nil {
		s.log.Error().Err(err).Str("uri", uri).Msg("add face")
	}
}
func (s *Secondary) DelFace(uri string) {
	ctx := context.Background()
	err := nfdstat.DelFace(ctx, uri)
	if err != nil {
		s.log.Error().Err(err).Str("uri", uri).Msg("delete face")
	}
}

func (s *Secondary) AddRoute(r route, c int64) {
	ctx := context.Background()
	err := nfdstat.AddRoute(ctx, r.prefix, r.uri, c)
	if err != nil {
		s.log.Error().Err(err).Str("prefix", r.prefix).Str("uri", r.uri).Int64("cost", c).Msg("add route")
	}
	// err = nfdstat.RouteStrategy(ctx, r.prefix, s.strategy)
	// if err != nil {
	// 	s.log.Error().Err(err).Str("prefix", r.prefix).Str("strategy", s.strategy).Msg("set strategy")
	// }
}

func (s *Secondary) DelRoute(r route) {
	ctx := context.Background()
	err := nfdstat.DelRoute(ctx, r.prefix, r.uri)
	if err != nil {
		s.log.Error().Err(err).Str("prefix", r.prefix).Str("uri", r.uri).Msg("delete route")
	}
}
