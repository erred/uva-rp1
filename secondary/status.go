package secondary

import (
	"context"

	"github.com/seankhliao/uva-rp1/api"
	"github.com/seankhliao/uva-rp1/nfdstat"
)

func (s *Secondary) pushStatus() {
	var stat *nfdstat.Status
	var err error
	for {
		stat, err = s.stat.Status()
		if err != nil {
			s.log.Error().Err(err).Msg("get status initial")
			continue
		}
		break
	}
	s.log.Info().Msg("get status initial")
	c, err := s.ctl.PushStatus(context.Background())
	if err != nil {
		s.log.Error().Err(err).Msg("push status")
		return
	}
	s.log.Info().Msg("push status connected")

	err = c.Send(&api.StatusResponse{Id: s.name})
	if err != nil {
		s.log.Error().Err(err).Msg("push initial status")
	}
	s.log.Info().Msg("push status initial sent")

	for {
		_, err := c.Recv()
		if err != nil {
			s.log.Error().Err(err).Msg("push status recv")
			return
		}

		for {
			stat, err = s.stat.Status()
			if err != nil {
				s.log.Error().Err(err).Msg("nfd status")
				continue
			}
			break
		}
		err = c.Send(stat.ToStatusResponse())
		if err != nil {
			s.log.Error().Err(err).Msg("push status")
		}
		s.log.Info().Msg("push status send")
	}
}
