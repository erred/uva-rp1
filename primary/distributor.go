package primary

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/seankhliao/uva-rp1/api"
	"github.com/seankhliao/uva-rp1/nfdstat"
	"google.golang.org/grpc"
)

func (p *Primary) distributor() {
	// set initial strategy
	for {
		err := nfdstat.RouteStrategy(context.Background(), "/", p.singleStrategy)
		if err != nil {
			p.log.Error().Err(err).Msg("apply strategy")
			time.Sleep(time.Second)
			continue
		}
		break
	}

	localSec := Secondary{
		primaries: make(chan map[string]primaryInfo, 1),
		log:       p.log,
	}
	localSec.primaries <- make(map[string]primaryInfo)

	nsec := 0
	pr := make(map[string]primary)
	for range p.secondaryNotify {
		all, disconnect := make(map[string]primary), make(map[string]primary)

		prs := <-p.primaries
		for k, v := range prs {
			all[k] = v
			if _, ok := pr[k]; !ok {
				pr[k] = v
			}
		}
		for k, v := range pr {
			if _, ok := prs[k]; !ok {
				disconnect[k] = v
				delete(pr, k)
			}
		}
		p.primaries <- prs

		p.log.Info().Int("all", len(pr)).Int("disconnect", len(disconnect)).Msg("distributor notify")

		secs := <-p.secondaries
		if nsec == 0 && len(secs) == 0 {
			// remove existing from all
			lprs := <-localSec.primaries
			for k, v := range all {
				if _, ok := lprs[k]; !ok {
					go localSec.connect(v)
				}
			}
			localSec.primaries <- lprs
			for k := range disconnect {
				go localSec.disconnect(k)
			}
		} else if nsec == 0 && len(secs) > 0 {
			// apply multi strategy
			err := nfdstat.RouteStrategy(context.Background(), "/", p.multiStrategy)
			if err != nil {
				p.log.Error().Err(err).Msg("apply strategy")
			}
			for k := range all {
				go localSec.disconnect(k)
			}
		} else if nsec > 0 && len(secs) == 0 {
			// apply single strategy
			err := nfdstat.RouteStrategy(context.Background(), "/", p.singleStrategy)
			if err != nil {
				p.log.Error().Err(err).Msg("apply strategy")
			}
			for _, v := range all {
				go localSec.connect(v)
			}
		} else {
			// remove existing from all
			// remove disconnect from secs
			ctr := make(map[string]int, len(secs))
			for sid, sec := range secs {
				for pid := range disconnect {
					if _, ok := sec.p[pid]; ok {
						delete(sec.p, pid)
					}
				}
				for pid := range sec.p {
					if _, ok := all[pid]; ok {
						delete(all, pid)
					}
				}
				secs[sid] = sec
				ctr[sid] = len(sec.p)
			}

			// add remaining from all to secs
			for pid, p := range all {
				sid := mapmin(ctr)
				ctr[sid]++
				sec := secs[sid]
				sec.p[pid] = p
				secs[sid] = sec
			}

			// send
			for id, sec := range secs {
				go func(id string, sec secondary) {
					prims := make([]*api.Primary, len(sec.p))
					for _, pri := range sec.p {
						prims = append(prims, &pri.p)
					}
					err := sec.c.Send(&api.RegisterControl{
						Primaries: prims,
					})
					if err != nil {
						p.log.Error().Err(err).Str("secondary", id).Msg("send primaries")
					}
				}(id, sec)
			}
		}

		nsec = len(secs)
		p.secondaries <- secs
	}
}

func mapmin(d map[string]int) string {
	s, m := "", 0
	for k, v := range d {
		s, m = k, v
		break
	}
	for k, v := range d {
		if v < m {
			s, m = k, v
		}
	}
	return s
}

type primaryInfo struct {
	p    api.Primary
	c    api.InfoClient
	ch   string
	conn *grpc.ClientConn
}

type Secondary struct {
	primaries chan map[string]primaryInfo
	log       *zerolog.Logger
}

func (s *Secondary) disconnect(pid string) {
	pr := <-s.primaries
	p := pr[pid]
	delete(pr, pid)
	s.primaries <- pr

	p.conn.Close()

	ctx := context.Background()
	err := nfdstat.DelFace(ctx, p.ch)
	if err != nil {
		s.log.Error().Err(err).Str("id", p.p.PrimaryId).Msg("disconnect del face")
	}
	s.log.Info().Str("primary", p.p.PrimaryId).Msg("disconnected")
}

func (s *Secondary) connect(ps primary) {
	p := primaryInfo{
		p: ps.p,
	}
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, p.p.Endpoint, grpc.WithInsecure())
	// conn, err := grpc.DialContext(ctx, p.p.Endpoint, grpc.WithInsecure(), grpc.WithBlock())
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
