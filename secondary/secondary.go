package secondary

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/seankhliao/uva-rp1/api"
	"google.golang.org/grpc"
)

type Secondary struct {
	primary string
	name    string

	scrape time.Duration
	faces  chan map[string]bool
	routes chan map[string][]*url.URL

	client api.ControlClient
	log    *zerolog.Logger
}

func New(args []string, logger *zerolog.Logger) *Secondary {
	if logger == nil {
		*logger = log.With().Str("mod", "secondary").Logger()
	}

	s := &Secondary{
		log:    logger,
		faces:  make(chan map[string]bool, 1),
		routes: make(chan map[string][]*url.URL, 1),
	}

	s.faces <- make(map[string]bool)
	s.routes <- make(map[string][]*url.URL)

	fs := flag.NewFlagSet("secondary", flag.ExitOnError)
	fs.StringVar(&s.primary, "primary", "145.100.104.117:8000", "host:port of primaary to connect to")
	fs.StringVar(&s.name, "name", strconv.FormatInt(rand.Int63(), 10), "overrdide randomly generated name of node")
	fs.DurationVar(&s.scrape, "scrape", 15*time.Second, "scrape / reporting interval")
	fs.Parse(args)
	return s
}

func (s *Secondary) Run(ctx context.Context) error {
	conn, err := grpc.DialContext(ctx, s.primary, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("run: %w", err)
	}
	defer conn.Close()

	s.client = api.NewControlClient(conn)
	retry := time.Second
	for {
		c, err := s.client.RegisterSecondary(ctx)
		if err != nil {
			s.log.Error().Err(err).Dur("backoff", retry).Msg("register")
			time.Sleep(retry)
			if retry < 32*time.Second {
				retry *= 2
			}
			continue
		}
		retry = time.Second
		s.handle(ctx, c)
	}
}

func (s *Secondary) handle(ctx context.Context, c api.Control_RegisterSecondaryClient) {
	go func() {
		for {
			stat, err := status(ctx)
			if err != nil {
				log.Error().Err(err).Msg("status")
				time.Sleep(s.scrape)
				continue
			}

			err = c.Send(s.update(stat))
			if err != nil {
				log.Error().Err(err).Msg("send")
			}
			time.Sleep(s.scrape)
		}
	}()
	for {
		cm, err := c.Recv()
		if errors.Is(err, io.EOF) {
			log.Info().Msg("receive EOF")
			return
		} else if err != nil {
			log.Error().Err(err).Msg("receive")
			continue
		}
		_ = cm
		panic("Unimplemented: apply controls")
	}
}

func (s *Secondary) update(stat *NFDStatus) *api.SecondaryInfo {
	m := make(map[int64]*url.URL, len(stat.Faces.Face))
	mf := make(map[string]bool, len(stat.Faces.Face))
	<-s.faces
	for _, f := range stat.Faces.Face {
		if !strings.HasPrefix(f.RemoteUri, "tcp") && !strings.HasPrefix(f.RemoteUri, "udp") {
			continue
		}
		u, err := url.Parse(f.RemoteUri)
		if err != nil {
			continue
		}
		m[f.FaceId] = u
		mf[f.RemoteUri] = true
	}
	s.faces <- mf

	rs := make([]*api.Route, len(stat.Rib.RibEntry))
	mr := make(map[string][]*url.URL, len(stat.Rib.RibEntry))
	<-s.routes
	for _, r := range stat.Rib.RibEntry {
		if u, ok := m[r.Routes.Route.FaceId]; ok {
			mr[r.Prefix] = append(mr[r.Prefix], u)
			rs = append(rs, &api.Route{
				Prefix: r.Prefix,
				Cost:   r.Routes.Route.Cost,
				Endpoint: &api.Endpoint{
					Scheme: u.Scheme,
					Host:   u.Hostname(),
					Port:   u.Port(),
				},
			})
		}
	}
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

func newsets(sc *api.SecondaryControl) (map[string]bool, map[string][]*url.URL) {
	fs := make(map[string]bool)
	rs := make(map[string][]*url.URL, len(sc.Routes))
	for _, r := range sc.Routes {
		fs[r.Endpoint.Scheme+"://"+r.Endpoint.Host+":"+r.Endpoint.Port] = true
		rs[r.Prefix] = append(rs[r.Prefix], &url.URL{
			Scheme: r.Endpoint.Scheme,
			Host:   r.Endpoint.Host + ":" + r.Endpoint.Port,
		})
	}
	return fs, rs
}
