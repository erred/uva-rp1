package primary

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/seankhliao/uva-rp1/api"
	"github.com/seankhliao/uva-rp1/nfdstat"
	"google.golang.org/grpc"
)

type primaryInfo struct {
	p    api.Primary
	c    api.InfoClient
	ch   string
	conn *grpc.ClientConn
}

type Secondary struct {
	localAddr string
	primaries chan map[string]primaryInfo
	log       *zerolog.Logger
}

func (s *Secondary) disconnect(pid string) {
	pr := <-s.primaries
	p := pr[pid]
	delete(pr, pid)
	s.primaries <- pr

	ctx := context.Background()
	err := nfdstat.DelFace(ctx, p.ch)
	if err != nil {
		s.log.Error().
			Err(err).
			Str("id", p.p.PrimaryId).
			Msg("localSec disconnect del face")
	}

	if p.conn != nil {
		p.conn.Close()
	}
	s.log.Info().
		Str("id", p.p.PrimaryId).
		Msg("localSec disconnected")
}

func (s *Secondary) connect(ps primary) {
	p := primaryInfo{
		p: ps.p,
	}
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, p.p.Endpoint, grpc.WithInsecure())
	if err != nil {
		s.log.Error().
			Err(err).
			Str("id", p.p.PrimaryId).
			Msg("localSec connect primary")
		return
	}
	p.conn = conn
	p.c = api.NewInfoClient(conn)

	cr, err := p.c.Channels(ctx, &api.ChannelRequest{})
	if err != nil {
		s.log.Error().
			Err(err).
			Str("id", p.p.PrimaryId).
			Msg("localSec primary channels")
		p.conn.Close()
		return
	}
	rtc, err := p.c.Routes(ctx, &api.RouteRequest{Id: s.localAddr})
	if err != nil {
		s.log.Error().
			Err(err).
			Str("id", p.p.PrimaryId).
			Msg("localSec primary routes")
		p.conn.Close()
		return
	}

	s.log.Info().
		Str("id", p.p.PrimaryId).
		Strs("chans", cr.Channels).
		Msg("localSec primary connected")

	for _, ch := range cr.Channels {
		err = nfdstat.AddFace(ctx, ch)
		if err != nil {
			s.log.Error().
				Err(err).
				Str("id", p.p.PrimaryId).
				Str("channel", ch).
				Msg("localSec primary add face")
		} else {
			p.ch = ch
			break
		}
	}
	if p.ch == "" {
		s.log.Error().
			Str("id", p.p.PrimaryId).
			Msg("localSec primary no channels")
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
			s.log.Error().
				Err(err).
				Str("id", id).
				Msg("routeUpdater recv")
			return
		}
		nrts := make(map[string]int64, len(rt.Routes))
		lrts := make([]string, 0, len(rt.Routes))
		for _, r := range rt.Routes {
			lrts = append(lrts, r.Prefix)
			nrts[r.Prefix] = r.Cost
		}
		s.log.Debug().
			Str("id", id).
			Str("chan", ch).
			Strs("routes", lrts).
			Msg("routeUpdater recv")

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

		s.log.Info().
			Strs("connect", connect).
			Strs("disconnect", disconnect).
			Msg("routeUpdater delta")
		for _, r := range connect {
			err = nfdstat.AddRoute(ctx, r, ch, nrts[r])
			if err != nil {
				s.log.Error().
					Err(err).
					Str("id", id).
					Str("chan", ch).
					Str("prefix", r).
					Msg("routeUpdater nfd add")
				continue
			}
			rts[r] = nrts[r]
		}
		for _, r := range disconnect {
			err = nfdstat.DelRoute(ctx, r, ch)
			if err != nil {
				s.log.Error().
					Err(err).
					Str("id", id).
					Str("chan", ch).
					Str("prefix", r).
					Msg("routeUpdater nfd del")
			}
			delete(rts, r)
		}
	}
}
