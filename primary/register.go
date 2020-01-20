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
		p.log.Error().
			Str("id", r.SecondaryId).
			Msg("Register secondary duplicate")
		return fmt.Errorf("duplicate id")
	}
	sec[r.SecondaryId] = secondary{
		p:   make(map[string]primary),
		c:   s,
		chs: r.Channels,
	}

	p.secondaries <- sec
	select {
	case p.secondaryNotify <- struct{}{}:
	default:
		// don't block
	}

	go p.localAddSecondary(r.Channels)
	defer func(id string) {
		sec := <-p.secondaries
		delete(sec, id)
		p.secondaries <- sec
		select {
		case p.secondaryNotify <- struct{}{}:
		default:
			// don't block
		}
		p.log.Info().
			Str("id", id).
			Msg("secondary unregistered")

		go p.localDelSecondary(r.Channels)
	}(r.SecondaryId)

	p.log.Info().
		Str("id", r.SecondaryId).
		Msg("secondary registered")
	<-s.Context().Done()
	return nil
}

func (p *Primary) reflectorClient() {
	p.log.Info().Msg("starting reflectorClient")
	retry := time.Second
	for {
		err := p.reflectorConnect()
		p.log.Error().
			Err(err).
			Dur("backoff", retry).
			Msg("reflectorClient")
		time.Sleep(retry)
	}
}

func (p *Primary) reflectorConnect() error {
	p.log.Info().Str("watcher", p.watcherAddr).Msg("starting reflectorConnect")
	ctx := context.Background()
	// conn, err := grpc.DialContext(ctx, p.watcherAddr, grpc.WithInsecure(), grpc.WithBlock())
	conn, err := grpc.DialContext(ctx, p.watcherAddr, grpc.WithInsecure())
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	rc := api.NewReflectorClient(conn)
	c, err := rc.Primaries(ctx, &api.Primary{
		PrimaryId: p.localAddr,
		Endpoint:  fmt.Sprintf("%s:%d", p.localAddr, p.port),
	})
	if err != nil {
		return fmt.Errorf("primaries: %w", err)
	}

	p.log.Info().Str("watcher", p.watcherAddr).Msg("reflectorConnect connected")
	p.watcher = c

	for {
		ap, err := c.Recv()
		if err != nil {
			return fmt.Errorf("recv: %w", err)
		}

		p.log.Debug().
			Str("id", p.watcherAddr).
			Int("primaries", len(ap.Primaries)).
			Msg("watcher recv")
		np := make(map[string]primary, len(ap.Primaries))
		for _, pr := range ap.Primaries {
			if pr.PrimaryId == p.localAddr {
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

func (p *Primary) localAddSecondary(chs []string) {
	ctx := context.Background()
	p.log.Info().
		Strs("channels", chs).
		Msg("localAddSecondary")
	for _, ch := range chs {
		err := nfdstat.AddFace(ctx, ch)
		if err != nil {
			p.log.Error().
				Err(err).
				Str("channel", ch).
				Msg("localAddSecondary face")
			continue
		}
		err = nfdstat.AddRoute(ctx, "/", ch, 128)
		if err != nil {
			p.log.Error().
				Err(err).
				Str("channel", ch).
				Msg("localAddSecondary route")
		}
	}
}

func (p *Primary) localDelSecondary(chs []string) {
	ctx := context.Background()
	p.log.Info().
		Strs("channels", chs).
		Msg("localDelSecondary")
	for _, ch := range chs {
		err := nfdstat.DelRoute(ctx, "/", ch)
		if err != nil {
			p.log.Error().
				Err(err).
				Str("channel", ch).
				Msg("localDelSecondary route")
			continue
		}
		err = nfdstat.DelFace(ctx, ch)
		if err != nil {
			p.log.Error().
				Err(err).
				Str("channel", ch).
				Msg("localDelSecondary face")
		}
	}
}
