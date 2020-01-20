package primary

import (
	"context"
	"fmt"
	"github.com/seankhliao/uva-rp1/api"
)

func (p *Primary) Identity(ctx context.Context, r *api.IdentityRequest) (*api.IdentityResponse, error) {
	p.log.Debug().Msg("identity request")
	return &api.IdentityResponse{
		PrimaryId: p.localAddr,
	}, nil
}

func (p *Primary) Channels(ctx context.Context, r *api.ChannelRequest) (*api.ChannelResponse, error) {
	ch := <-p.localChan
	nch := make([]string, len(ch))
	copy(nch, ch)
	p.localChan <- ch
	p.log.Debug().
		Strs("chans", nch).
		Msg("channel request")
	return &api.ChannelResponse{
		Channels: nch,
	}, nil
}

func (p *Primary) Routes(r *api.RouteRequest, s api.Info_RoutesServer) error {
	p.log.Debug().
		Str("id", r.Id).
		Msg("routes request")
	err := s.Send(p.getRoutes())
	if err != nil {
		p.log.Error().
			Err(err).
			Str("id", r.Id).
			Msg("routes send")
		return fmt.Errorf("routes send: %w", err)
	}

	id := r.Id
	wr := <-p.wantRoutes
	wr[id] = wantRoute{s}
	p.wantRoutes <- wr

	defer func(id string) {
		wr := <-p.wantRoutes
		delete(wr, id)
		p.wantRoutes <- wr
		p.log.Info().
			Str("id", id).
			Msg("routes request unregistered")
	}(id)

	p.log.Info().
		Str("id", id).
		Msg("routes request registered")
	<-s.Context().Done()
	return nil
}

func (p *Primary) getRoutes() *api.RouteResponse {
	rt := <-p.localRoutes
	rr := &api.RouteResponse{
		Routes: make([]*api.Route, 0, len(rt)),
	}
	for pr, c := range rt {
		rr.Routes = append(rr.Routes, &api.Route{
			Prefix: pr,
			Cost:   c,
		})
	}
	p.localRoutes <- rt
	return rr
}

func (p *Primary) PushStatus(s api.Info_PushStatusServer) error {
	sr, err := s.Recv()
	if err != nil {
		return fmt.Errorf("PushStatus recv init: %w", err)
	}

	sec := <-p.secondaries
	se := sec[sr.Id]
	se.s = s
	sec[sr.Id] = se
	p.secondaries <- sec

	defer func(id string) {
		sec := <-p.secondaries
		delete(sec, id)
		p.secondaries <- sec
		p.log.Info().
			Str("id", id).
			Msg("PushStatus unregistered")
	}(sr.Id)

	p.log.Info().
		Str("id", sr.Id).
		Msg("PushStatus registered")
	<-s.Context().Done()
	return nil
}

func (p *Primary) PullStatus(s api.Info_PullStatusServer) error {
	for {
		_, err := s.Recv()
		if err != nil {
			return fmt.Errorf("PullStatus recv: %w", err)
		}
		stat, err := p.stat.Status()
		if err != nil {
			return fmt.Errorf("PullStatus status: %w", err)
		}

		err = s.Send(stat.ToStatusResponse())
		if err != nil {
			return fmt.Errorf("PullStatus send: %w", err)
		}
		p.log.Info().Msg("PullStatus sent")
	}
}
