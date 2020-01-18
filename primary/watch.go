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
		p.log.Info().Str("id", id).Msg("secondary unregistered")
	}(r.SecondaryId)

	p.log.Info().Str("id", r.SecondaryId).Msg("secondary registered")
	<-s.Context().Done()
	return nil
}

func (p *Primary) primaryDiscoverer() {
	p.log.Info().Msg("starting primaryDiscoverer")
	retry := time.Second
	for {
		p.watcherConnect()
		p.log.Error().Dur("backoff", retry).Msg("watcher connect")
		time.Sleep(retry)
	}
}
func (p *Primary) watcherConnect() bool {
	p.log.Info().Str("watcher", p.watcherAddr).Msg("starting watcherConnect")
	ctx := context.Background()
	// conn, err := grpc.DialContext(ctx, p.watcherAddr, grpc.WithInsecure(), grpc.WithBlock())
	conn, err := grpc.DialContext(ctx, p.watcherAddr, grpc.WithInsecure())
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
	p.log.Info().Str("watcher", p.watcherAddr).Msg("watcher connected")

	p.watcher = c

	for {
		ap, err := c.Recv()
		if err != nil {
			p.log.Error().Err(err).Str("watcher", p.watcherAddr).Msg("watcher recv")
			return true
		}
		p.log.Info().Str("watcher", p.watcherAddr).Int("primaries", len(ap.Primaries)).Msg("watcher recv")
		np := make(map[string]primary, len(ap.Primaries))
		for _, pr := range ap.Primaries {
			if pr.PrimaryId == p.name {
				continue
			}
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
