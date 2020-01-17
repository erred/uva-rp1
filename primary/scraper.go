package primary

import "time"

import "strings"

import "github.com/seankhliao/uva-rp1/api"

import "net"

import "fmt"

func (p *Primary) scraper(first chan struct{}) {
	for {
		err := p.scrape(true)
		if err == nil {
			break
		}
		p.log.Error().Err(err).Msg("first scrape")
	}
	first <- struct{}{}
	for range time.NewTicker(p.scrapeInterval).C {
		go p.scrape(false)
	}
}
func (p *Primary) scrape(first bool) error {
	r, n, err := p.stat.Stats()
	if err != nil {
		return err
	}
	<-p.localStat
	p.localStat <- r

	if first {
		if len(p.localUris) > 0 {
			p.localUris, err = localIPs()
			if err != nil {
				return fmt.Errorf("scrape: %w", err)
			}
		}
		for _, ch := range n.Channels.Channel {
			scheme := ch.LocalUri[:3]
			if scheme == "tcp" || scheme == "udp" {
				i := strings.LastIndex(ch.LocalUri, ":")
				for _, u := range p.localUris {
					p.localChan = append(p.localChan, scheme+"://"+u+":"+ch.LocalUri[i+1:])
				}
			}
		}
	}

	rt := make(map[string]int64)
	var rts []*api.Route
	for _, re := range n.Rib.RibEntry {
		if strings.HasPrefix(re.Prefix, "/local") {
			continue
		}
		rt[re.Prefix] = re.Routes.Route.Cost
		for _, u := range p.localUris {
			rts = append(rts, &api.Route{
				Prefix:   re.Prefix,
				Endpoint: u,
				Cost:     re.Routes.Route.Cost,
			})
		}
	}

	ort := <-p.localRoutes
	var diff bool
	for k := range ort {
		if _, ok := rt[k]; !ok {
			diff = true
		}
	}
	for k := range rt {
		if _, ok := ort[k]; !ok {
			diff = true
		}
	}
	p.localRoutes <- rt

	if diff && p.watcher != nil {
		c := &api.Cluster{
			Id:     p.name,
			Routes: rts,
		}
		err = p.watcher.Send(c)
		if err != nil {
			p.log.Error().Err(err).Msg("watcher send")
		}
	}

	return nil
}

func localIPs() ([]string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return nil, fmt.Errorf("interface addrs: %w", err)
	}
	var ips []string
	for _, addr := range addrs {
		an, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		if !an.IP.IsGlobalUnicast() {
			continue
		}
		ips = append(ips, an.IP.String())
	}
	if len(ips) == 0 {
		return nil, fmt.Errorf("interface addrs: no addrs found")
	}
	return ips, nil
}
