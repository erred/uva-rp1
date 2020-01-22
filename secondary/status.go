package secondary

import (
	"context"
	"github.com/seankhliao/uva-rp1/nfdstat"
	"strings"
)

func (s *Secondary) statusPusher() {
	c, err := s.ctl.SecondaryStatus(context.Background())
	if err != nil {
		s.log.Error().
			Err(err).
			Msg("statusPusher connect")
		return
	}
	s.log.Info().Msg("statusPusher connected")

	first := true
	for {
		if !first {
			_, err := c.Recv()
			if err != nil {
				s.log.Error().
					Err(err).
					Msg("statusPusher recv")
				return
			}
		}

		var stat *nfdstat.Status
		var err error
		for {
			stat, err = s.stat.Status()
			if err != nil {
				s.log.Error().
					Err(err).
					Msg("statusPusher nfd status")
				continue
			}
			break
		}

		p := <-s.primaries
		prims := make([]string, 0, len(p))
		for pid := range p {
			prims = append(prims, pid)
		}
		s.primaries <- p

		err = c.Send(stat.ToStatusNFD(s.localAddr, prims))
		if err != nil {
			s.log.Error().Err(err).Msg("statusPusher send")
		}
		s.log.Debug().Msg("statusPusher sent")
	}
}

func (s *Secondary) mustGetChannels() {
	for {
		stat, err := s.stat.Status()
		if err != nil {
			s.log.Error().
				Err(err).
				Msg("getChannels nfd status")
			continue
		}
		chans := make(map[string]struct{}, len(stat.Channels.Channel))
		for _, ch := range stat.Channels.Channel {
			scheme := ch.LocalUri[:3]
			if scheme == "tcp" || scheme == "udp" {
				i := strings.LastIndex(ch.LocalUri, ":")
				chans[scheme+"://"+s.localAddr+":"+ch.LocalUri[i+1:]] = struct{}{}
			}
		}
		chs := make([]string, 0, len(chans))
		for k := range chans {
			chs = append(chs, k)
		}
		s.log.Info().
			Strs("channels", chs).
			Msg("getChannels got")

		<-s.localChan
		s.localChan <- chs
		break
	}
}
