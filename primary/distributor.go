package primary

import (
	"context"
	"fmt"
	"time"

	"github.com/seankhliao/uva-rp1/api"
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
	for range p.secondaryNotify {
		// TODO: distribute endpoints to secondaries
	}
}
