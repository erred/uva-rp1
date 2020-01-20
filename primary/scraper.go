package primary

import (
	"github.com/seankhliao/uva-rp1/nfdstat"
	"strings"
	"time"
)

func (p *Primary) scraper(first chan struct{}) {
	for {
		err := p.scrape()
		if err == nil {
			break
		}
		p.log.Error().
			Err(err).
			Msg("scraper init")
		time.Sleep(time.Second)
	}
	first <- struct{}{}
	for range time.NewTicker(p.scrapeInterval).C {
		go p.scrape()
	}
}
func (p *Primary) scrape() error {
	s, err := p.stat.Status()
	if err != nil {
		return err
	}
	chans := make(map[string]struct{}, len(s.Channels.Channel))
	for _, ch := range s.Channels.Channel {
		scheme := ch.LocalUri[:3]
		if scheme == "tcp" || scheme == "udp" {
			i := strings.LastIndex(ch.LocalUri, ":")
			chans[scheme+"://"+p.localAddr+":"+ch.LocalUri[i+1:]] = struct{}{}
		}
	}
	chs := make([]string, 0, len(chans))
	for k := range chans {
		chs = append(chs, k)
	}
	<-p.localChan
	p.localChan <- chs

	<-p.status
	p.status <- s

	rt := make(map[string]int64)
	rts := make([]string, 0, len(s.Rib.RibEntry))
	for _, re := range s.Rib.RibEntry {
		if strings.HasPrefix(re.Prefix, "/local") {
			continue
		}
		if re.Routes.Route.Origin == nfdstat.Origin {
			continue
		}
		rt[re.Prefix] = re.Routes.Route.Cost
		rts = append(rts, re.Prefix)
	}

	ort := <-p.localRoutes
	var diff bool
	for k := range ort {
		if _, ok := rt[k]; !ok {
			diff = true
			break
		}
	}
	if !diff {
		for k := range rt {
			if _, ok := ort[k]; !ok {
				diff = true
			}
		}
	}
	p.localRoutes <- rt

	if diff {
		select {
		case p.routeNotify <- struct{}{}:
		default:
			// don't block
		}
	}
	return nil
}

func (p *Primary) routeAdvertiser() {
	for range p.routeNotify {
		rt := p.getRoutes()
		wr := <-p.wantRoutes
		for k, v := range wr {
			go func(k string, v wantRoute) {
				err := v.s.Send(rt)
				if err != nil {
					p.log.Error().
						Err(err).
						Str("id", k).
						Msg("routeAdvertiser send")
				}
			}(k, v)
		}
		rts := make([]string, 0, len(rt.Routes))
		for _, r := range rt.Routes {
			rts = append(rts, r.Prefix)
		}

		p.log.Info().
			Int("wantRoutes", len(wr)).
			Strs("routes", rts).
			Msg("routeAdvertiser sent")
		p.wantRoutes <- wr
	}
}
