package primary

import (
	"context"
	"time"

	"github.com/seankhliao/uva-rp1/api"
	"github.com/seankhliao/uva-rp1/nfdstat"
)

func (p *Primary) distributor() {
	// set initial strategy
	for {
		err := nfdstat.RouteStrategy(context.Background(), "/", p.singleStrategy)
		if err != nil {
			p.log.Error().Err(err).Msg("distributor initial strategy")
			time.Sleep(time.Second)
			continue
		}
		break
	}

	localSec := Secondary{
		localAddr: p.localAddr,
		primaries: make(chan map[string]primaryInfo, 1),
		log:       p.log,
	}
	localSec.primaries <- make(map[string]primaryInfo)

	nsec := 0
	pr := make(map[string]primary)
	for range p.secondaryNotify {
		all, disconnect := make(map[string]primary), make(map[string]primary)

		prs := <-p.primaries
		for k, v := range prs {
			all[k] = v
			if _, ok := pr[k]; !ok {
				pr[k] = v
			}
		}
		for k, v := range pr {
			if _, ok := prs[k]; !ok {
				disconnect[k] = v
				delete(pr, k)
			}
		}
		p.primaries <- prs

		log := p.log.Info().
			Int("all", len(pr)).
			Int("disconnect", len(disconnect))

		secs := <-p.secondaries
		if nsec == 0 && len(secs) == 0 {
			log.Msg("distributor apply local -> local")

			lprs := <-localSec.primaries
			for k, v := range all {
				if _, ok := lprs[k]; !ok {
					go localSec.connect(v)
				}
			}
			localSec.primaries <- lprs
			for k := range disconnect {
				go localSec.disconnect(k)
			}

			nsec = len(secs)
			p.secondaries <- secs
			continue
		} else if nsec > 0 && len(secs) == 0 {
			// apply single strategy
			log.Msg("distributor apply secondaries -> local")

			err := nfdstat.RouteStrategy(context.Background(), "/", p.singleStrategy)
			if err != nil {
				p.log.Error().Err(err).Msg("apply strategy")
			}
			for _, v := range all {
				go localSec.connect(v)
			}

			nsec = len(secs)
			p.secondaries <- secs
			continue
		} else if nsec == 0 {
			log.Msg("distributor apply local -> secondaries")

			err := nfdstat.RouteStrategy(context.Background(), "/", p.multiStrategy)
			if err != nil {
				p.log.Error().Err(err).Msg("apply strategy")
			}
			for k := range all {
				go localSec.disconnect(k)
			}
		} else {
			log.Msg("distributor apply secondaries -> secondaries")
		}

		// remove existing from all
		// remove disconnect from secs
		ctr := make(map[string]int, len(secs))
		for sid, sec := range secs {
			for pid := range disconnect {
				if _, ok := sec.p[pid]; ok {
					delete(sec.p, pid)
				}
			}
			for pid := range sec.p {
				if _, ok := all[pid]; ok {
					delete(all, pid)
				}
			}

			secs[sid] = sec
			ctr[sid] = len(sec.p)
		}

		// add remaining from all to secs
		for pid, p := range all {
			sid := mapmin(ctr)
			if sid == "" {
				continue
			}
			ctr[sid]++
			sec := secs[sid]
			sec.p[pid] = p
			secs[sid] = sec
		}

		for {
			max, min := mapmaxmin(ctr)
			if max == "" {
				break
			}
			ctr[max]--
			ctr[min]++
			var pid string
			var p primary

			s := secs[max]
			for k, v := range s.p {
				pid, p = k, v
				delete(s.p, k)
				break
			}
			secs[max] = s

			s = secs[min]
			s.p[pid] = p
			secs[min] = s
		}

		// send
		dbg := p.log.Debug()
		for id, sec := range secs {
			prims := make([]*api.Primary, 0, len(sec.p))
			primsd := make([]string, 0, len(sec.p))
			for _, pri := range sec.p {
				pri := pri
				prims = append(prims, &pri.p)
				primsd = append(primsd, pri.p.PrimaryId)
			}
			err := sec.c.Send(&api.RegisterControl{
				Primaries: prims,
			})
			if err != nil {
				p.log.Error().Err(err).Str("secondary", id).Msg("send primaries")
			}
			dbg = dbg.Strs(id, primsd)
			// }(id, sec)
		}
		dbg.Msg("distribor results")

		nsec = len(secs)
		p.secondaries <- secs
	}
}

func mapmin(d map[string]int) string {
	s, m := "", 0
	for k, v := range d {
		s, m = k, v
		break
	}
	for k, v := range d {
		if v < m {
			s, m = k, v
		}
	}
	return s
}

func mapmaxmin(d map[string]int) (max, min string) {
	sa, a, si, i := "", 0, "", 0
	for k, v := range d {
		sa, a, si, i = k, v, k, v
	}
	for k, v := range d {
		if v < i {
			si, i = k, v
		}
		if v > a {
			sa, a = k, v
		}
	}
	if a-i <= 1 {
		return "", ""
	}
	return sa, si
}
