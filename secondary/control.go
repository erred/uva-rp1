package secondary

import (
	"context"
	"fmt"
	"strings"

	"github.com/seankhliao/uva-rp1/api"
	"github.com/seankhliao/uva-rp1/nfdstat"
	"google.golang.org/grpc"
)

func (s *Secondary) handle(ctx context.Context) error {
	// establish connection
	conn, err := grpc.DialContext(ctx, s.primary, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	s.client = api.NewControlClient(conn)
	c, err := s.client.Register(ctx, &api.RegisterRequest{
		Id: s.name,
	})
	if err != nil {
		return fmt.Errorf("register: %w", err)
	}

	// respond to commands
	for {
		cm, err := c.Recv()
		if err != nil {
			return fmt.Errorf("receive: %w", err)
		}

		_, stat := s.stat.Status()
		s.update(stat)

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
					// TODO: test connection with primary first
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
	}
}

func (s *Secondary) update(stat *nfdstat.NFDStatus) {
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
}

func (s *Secondary) AddFace(uri string) {
	err := nfdstat.AddFace(context.Background(), uri)
	if err != nil {
		s.log.Error().Err(err).Str("uri", uri).Msg("add face")
	}
}
func (s *Secondary) DelFace(uri string) {
	err := nfdstat.DelFace(context.Background(), uri)
	if err != nil {
		s.log.Error().Err(err).Str("uri", uri).Msg("delete face")
	}
}

func (s *Secondary) AddRoute(r route, c int64) {
	err := nfdstat.AddRoute(context.Background(), r.prefix, r.uri, c)
	if err != nil {
		s.log.Error().Err(err).Str("prefix", r.prefix).Str("uri", r.uri).Int64("cost", c).Msg("add route")
	}
	// err = nfdstat.RouteStrategy(ctx, r.prefix, s.strategy)
	// if err != nil {
	// 	s.log.Error().Err(err).Str("prefix", r.prefix).Str("strategy", s.strategy).Msg("set strategy")
	// }
}

func (s *Secondary) DelRoute(r route) {
	err := nfdstat.DelRoute(context.Background(), r.prefix, r.uri)
	if err != nil {
		s.log.Error().Err(err).Str("prefix", r.prefix).Str("uri", r.uri).Msg("delete route")
	}
}
