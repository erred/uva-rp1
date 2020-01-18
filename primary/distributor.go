package primary

import (
	"context"
	"fmt"
	"time"

	"github.com/seankhliao/uva-rp1/api"
	"github.com/seankhliao/uva-rp1/nfdstat"
	"google.golang.org/grpc"
)

func (p *Primary) Register(r *api.RegisterRequest, s api.Info_RegisterServer) error {
	sec := <-p.secondaries
	if _, ok := sec[r.SecondaryId]; ok {
		p.secondaries <- sec
		return fmt.Errorf("Register duplicate secondary: %s", r.SecondaryId)
	}
	sec[r.SecondaryId] = secondary{
		p: make(map[string]primary),
		c: s,
	}
	p.secondaries <- sec
	select {
	case p.secondaryNotify <- struct{}{}:
	default:
		// don't block
	}

	defer func(id string) {
		sec := <-p.secondaries
		delete(sec, id)
		p.secondaries <- sec
		select {
		case p.secondaryNotify <- struct{}{}:
		default:
			// don't block
		}
	}(r.SecondaryId)

	<-s.Context().Done()
	return nil
}

func (p *Primary) primaryDiscoverer() {
	retry := time.Second
	for {
		connected := p.watcherConnect()
		if connected {
			retry = time.Second
		}
		time.Sleep(retry)
		if retry < 30 {
			retry *= 2
		}
	}
}
func (p *Primary) watcherConnect() bool {
	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, p.watcherAddr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		p.log.Error().Err(err).Str("watcher", p.watcherAddr).Msg("watcher dial")
		return false
	}
	defer conn.Close()

	rc := api.NewReflectorClient(conn)
	c, err := rc.Primaries(ctx, &api.Primary{
		PrimaryId: p.name,
		Endpoint:  fmt.Sprintf("%s:%d", p.localAddr, p.port),
	})
	if err != nil {
		p.log.Error().Err(err).Str("watcher", p.watcherAddr).Msg("watcher primaries")
		return true
	}
	for {
		ap, err := c.Recv()
		if err != nil {
			p.log.Error().Err(err).Str("watcher", p.watcherAddr).Msg("watcher recv")
			return true
		}
		np := make(map[string]primary, len(ap.Primaries))
		for _, pr := range ap.Primaries {
			np[pr.PrimaryId] = primary{*pr}
		}

		var diff bool
		op := <-p.primaries
		for pid := range op {
			if _, ok := np[pid]; !ok {
				diff = true
				break
			}
		}
		if !diff {
			for pid := range np {
				if _, ok := op[pid]; !ok {
					diff = true
					break
				}
			}
		}
		p.primaries <- np
		if diff {
			select {
			case p.secondaryNotify <- struct{}{}:
			default:
				// don't block
			}
		}
	}
}

func (p *Primary) distributor() {
	// set initial strategy
	nsec := 1
	pr := make(map[string]primary)
	for range p.secondaryNotify {
		var all, connect, disconnect []primary

		prs := <-p.primaries
		for k, v := range prs {
			all = append(all, v)
			if _, ok := pr[k]; !ok {
				connect = append(connect, v)
				pr[k] = v
			}
		}
		for k, v := range pr {
			if _, ok := prs[k]; !ok {
				disconnect = append(disconnect, v)
				delete(pr, k)
			}
		}
		p.primaries <- prs

		secs := <-p.secondaries
		if nsec < 2 && len(secs) >= 2 {
			// apply multi strategy
			err := nfdstat.RouteStrategy(context.Background(), "/", p.multiStrategy)
			if err != nil {
				p.log.Error().Err(err).Msg("apply strategy")
			}
		} else if nsec >= 2 && len(secs) < 2 {
			// apply single strategy
			err := nfdstat.RouteStrategy(context.Background(), "/", p.singleStrategy)
			if err != nil {
				p.log.Error().Err(err).Msg("apply strategy")
			}
		}

		ctr := make(map[string]int, len(secs))
		for k, v := range secs {
			for _, pri := range disconnect {
				if _, ok := v.p[pri.p.PrimaryId]; ok {
					delete(v.p, pri.p.PrimaryId)
				}
			}
			secs[k] = v
			ctr[k] = len(v.p)
		}
		for _, pri := range connect {
			k := mapmin(ctr)
			ctr[k]++
			v := secs[k]
			v.p[k] = pri
			secs[k] = v
		}
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
