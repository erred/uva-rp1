package primary

import (
	"fmt"

	"github.com/seankhliao/uva-rp1/api"
)

func (p *Primary) Register(r *api.RegisterRequest, s api.Control_RegisterServer) error {
	secs := <-p.secondaries
	if _, ok := secs[r.Id]; ok {
		p.secondaries <- secs
		p.log.Error().Str("register", r.Id).Msg("duplicate secondary id")
		return fmt.Errorf("duplicate secondary id: %s", r.Id)
	}
	secs[r.Id] = secondary{
		s: s,
		r: make(map[string][]endpoint),
	}
	p.rebalance <- struct{}{}
	go func(id string) {
		<-s.Context().Done()
		secs := <-p.secondaries
		delete(secs, id)
		p.secondaries <- secs
	}(r.Id)
	return nil
}
