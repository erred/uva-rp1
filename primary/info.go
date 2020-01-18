package primary

import (
	"context"
	"fmt"
	"github.com/seankhliao/uva-rp1/api"
	"math/rand"
	"strconv"
)

func (p *Primary) Identity(ctx context.Context, r *api.IdentityRequest) (*api.IdentityResponse, error) {
	return &api.IdentityResponse{
		PrimaryId: p.name,
	}, nil
}

func (p *Primary) Channels(ctx context.Context, r *api.ChannelRequest) (*api.ChannelResponse, error) {
	ch := <-p.localChan
	nch := make([]string, len(ch))
	copy(nch, ch)
	p.localChan <- ch
	return &api.ChannelResponse{
		Channels: nch,
	}, nil
}

func (p *Primary) Routes(r *api.RouteRequest, s api.Info_RoutesServer) error {
	err := s.Send(p.getRoutes())
	if err != nil {
		return fmt.Errorf("routes send: %w", err)
	}

	id := strconv.FormatInt(rand.Int63(), 10)
	wr := <-p.wantRoutes
	wr[id] = wantRoute{s}
	p.wantRoutes <- wr

	defer func(id string) {
		wr := <-p.wantRoutes
		delete(wr, id)
		p.wantRoutes <- wr
	}(id)

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
		return fmt.Errorf("PushStatus initial recv: %w", err)
	}
	if sr.Id == "" {
		return fmt.Errorf("PushStatus no id")
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
	}(sr.Id)

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
	}
}

