package primary

import (
	"strings"
	"time"
)

func (p *Primary) scraper(first chan struct{}) {
	for {
		err := p.scrape()
		if err == nil {
			break
		}
		p.log.Error().Err(err).Msg("first scrape")
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
	var chans []string
	for _, ch := range s.Channels.Channel {
		scheme := ch.LocalUri[:3]
		if scheme == "tcp" || scheme == "udp" {
			i := strings.LastIndex(ch.LocalUri, ":")
			chans = append(chans, scheme+"://"+p.localAddr+":"+ch.LocalUri[i+1:])
		}
	}
	<-p.localChan
	p.localChan <- chans

	<-p.status
	p.status <- s

	rt := make(map[string]int64)
	for _, re := range s.Rib.RibEntry {
		if strings.HasPrefix(re.Prefix, "/local") {
			continue
		}
		rt[re.Prefix] = re.Routes.Route.Cost
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
					p.log.Error().Err(err).Str("id", k).Msg("route send")
				}
			}(k, v)
		}
		p.wantRoutes <- wr
	}
}
