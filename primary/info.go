package primary

import (
	"context"
	"fmt"
	"github.com/seankhliao/uva-rp1/api"
	"sync"
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

func (p *Primary) SecondaryStatus(s api.Info_SecondaryStatusServer) error {
	sr, err := s.Recv()
	if err != nil {
		return fmt.Errorf("SecondaryStatus recv init: %w", err)
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
			Msg("SecondaryStatus unregistered")
	}(sr.Id)

	p.log.Info().
		Str("id", sr.Id).
		Msg("SecondaryStatus registered")
	<-s.Context().Done()
	return nil
}

func (p *Primary) PrimaryStatus(s api.Info_PrimaryStatusServer) error {
	for {
		_, err := s.Recv()
		if err != nil {
			return fmt.Errorf("PrimaryStatus recv: %w", err)
		}

		var wg sync.WaitGroup
		lstat, sstat := make(chan *api.StatusNFD), make(chan *api.StatusNFD)

		go func() {
			stat, err := p.stat.Status()
			if err != nil {
				p.log.Error().Err(err).Msg("PrimaryStatus local")
				close(lstat)
				return
			}

			lstat <- stat.ToStatusNFD(p.localAddr, nil)
			close(lstat)
		}()

		psec := <-p.secondaries
		secs := make([]string, 0, len(psec))
		for sid, sec := range psec {
			secs = append(secs, sid)
			wg.Add(1)
			go func(id string, ss api.Info_SecondaryStatusServer) {
				defer wg.Done()
				err := ss.Send(&api.StatusRequest{})
				if err != nil {
					p.log.Error().Err(err).Str("id", id).Msg("PrimaryStatus secondary send")
					return
				}
				stat, err := ss.Recv()
				if err != nil {
					p.log.Error().Err(err).Str("id", id).Msg("PrimaryStatus secondary recv")
				}
				sstat <- stat
			}(sid, sec.s)
		}
		p.secondaries <- psec

		go func() {
			wg.Wait()
			close(sstat)
		}()

		pstat := &api.StatusPrimary{
			Id:    p.localAddr,
			Local: <-lstat,
		}
		for sec := range sstat {
			pstat.Secondaries = append(pstat.Secondaries, sec)
		}
		if len(pstat.Secondaries) == 0 {
			pstat.Secondaries = []*api.StatusNFD{}
		}

		err = s.Send(pstat)
		if err != nil {
			return fmt.Errorf("PrimaryStatus send: %w", err)
		}
		// p.log.Info().Msg("PrimaryStatus sent")
	}
}
