package secondary

import "context"

import "github.com/seankhliao/uva-rp1/nfdstat"

func (s *Secondary) pushStatus() {
	var stat *nfdstat.Status
	var err error
	for {
		stat, err = s.stat.Status()
		if err != nil {
			s.log.Error().Err(err).Msg("nfd status")
			continue
		}
		break
	}

	c, err := s.ctl.PushStatus(context.Background())
	if err != nil {
		s.log.Error().Err(err).Msg("push status")
		return
	}

	err = c.Send(stat.ToStatusResponse())
	if err != nil {
		s.log.Error().Err(err).Msg("push initial status")
	}

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

	}
}
