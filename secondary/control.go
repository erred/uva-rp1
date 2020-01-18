package secondary

import (
	"context"
	"fmt"

	"github.com/seankhliao/uva-rp1/api"
	"github.com/seankhliao/uva-rp1/nfdstat"
	"google.golang.org/grpc"
)

func (s *Secondary) recvCmd(c api.Info_RegisterClient) error {
	rc, err := c.Recv()
	if err != nil {
		return fmt.Errorf("recv: %w", err)
	}

	cp := make(map[string]primary, len(rc.Primaries))
	for _, p := range rc.Primaries {
		cp[p.PrimaryId] = primary{
			*p, nil, "", nil,
		}
	}

	var connect, disconnect []primary
	op := <-s.primaries
	for pid, p := range op {
		if _, ok := cp[pid]; !ok {
			disconnect = append(disconnect, p)
		}
	}
	for pid, p := range cp {
		if _, ok := op[pid]; !ok {
			connect = append(connect, p)
		}
	}
	s.primaries <- op
	s.log.Info().Int("primaries", len(rc.Primaries)).Int("connect", len(connect)).Int("disconnect", len(disconnect)).Msg("cmd recv")

	for _, p := range connect {
		go s.connect(p)
	}
	for _, p := range disconnect {
		go s.disconnect(p)
	}
	return nil
}

func (s *Secondary) disconnect(p primary) {
	pr := <-s.primaries
	delete(pr, p.p.PrimaryId)
	s.primaries <- pr

	p.conn.Close()

	ctx := context.Background()
	err := nfdstat.DelFace(ctx, p.ch)
	if err != nil {
		s.log.Error().Err(err).Str("id", p.p.PrimaryId).Msg("disconnect del face")
	}
	s.log.Info().Str("primary", p.p.PrimaryId).Msg("disconnected")
}

func (s *Secondary) connect(p primary) {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, p.p.Endpoint, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		s.log.Error().Err(err).Str("id", p.p.PrimaryId).Str("endpoint", p.p.Endpoint).Msg("connect primary")
		return
	}
	p.conn = conn
	p.c = api.NewInfoClient(conn)

	cr, err := p.c.Channels(ctx, &api.ChannelRequest{})
	if err != nil {
		s.log.Error().Err(err).Str("id", p.p.PrimaryId).Str("endpoint", p.p.Endpoint).Msg("primary channels")
		p.conn.Close()
		return
	}
	rtc, err := p.c.Routes(ctx, &api.RouteRequest{})
	if err != nil {
		s.log.Error().Err(err).Str("id", p.p.PrimaryId).Str("endpoint", p.p.Endpoint).Msg("primary routes")
		p.conn.Close()
		return
	}
	s.log.Info().Str("primary", p.p.PrimaryId).Strs("chans", cr.Channels).Msg("primary connected")

	for _, ch := range cr.Channels {
		err = nfdstat.AddFace(ctx, ch)
		if err != nil {
			s.log.Error().Err(err).Str("id", p.p.PrimaryId).Str("endpoint", p.p.Endpoint).Str("channel", ch).Msg("primary face")
		} else {
			p.ch = ch
			break
		}
	}
	if p.ch == "" {
		s.log.Error().Str("id", p.p.PrimaryId).Str("endpoint", p.p.Endpoint).Msg("primary no channels")
		p.conn.Close()
		return
	}

	go s.routeUpdater(p.p.PrimaryId, p.ch, rtc)

	pr := <-s.primaries
	pr[p.p.PrimaryId] = p
	s.primaries <- pr
}

func (s *Secondary) routeUpdater(id, ch string, c api.Info_RoutesClient) {
	ctx := context.Background()
	rts := make(map[string]int64)
	for {
		rt, err := c.Recv()
		if err != nil {
			s.log.Error().Err(err).Str("id", id).Msg("routeUpdater recv")
			return
		}
		nrts := make(map[string]int64, len(rt.Routes))
		lrts := make([]string, 0, len(rt.Routes))
		for _, r := range rt.Routes {
			lrts = append(lrts, r.Prefix)
			nrts[r.Prefix] = r.Cost
		}
		s.log.Info().Str("primary", id).Str("chan", ch).Strs("routes", lrts).Msg("routes recv")

		var connect, disconnect []string
		for r := range rts {
			if _, ok := nrts[r]; !ok {
				disconnect = append(disconnect, r)
			}
		}
		for r := range nrts {
			if _, ok := rts[r]; !ok {
				connect = append(connect, r)
			}
		}

		s.log.Info().Strs("connect", connect).Strs("disconnect", disconnect).Msg("routes delta")
		for _, r := range connect {
			err = nfdstat.AddRoute(ctx, r, ch, nrts[r])
			if err != nil {
				s.log.Error().Err(err).Str("id", id).Str("chan", ch).Str("prefix", r).Msg("routeUpdater add")
				continue
			}
			rts[r] = nrts[r]
		}
		for _, r := range disconnect {
			err = nfdstat.DelRoute(ctx, r, ch)
			if err != nil {
				s.log.Error().Err(err).Str("id", id).Str("chan", ch).Str("prefix", r).Msg("routeUpdater del")
			}
			delete(rts, r)
		}
	}
}
