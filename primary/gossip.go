package primary

import (
	"context"
	"fmt"
	"time"

	"github.com/seankhliao/uva-rp1/api"
	"google.golang.org/grpc"
)

func (p *Primary) gossip(ctx context.Context, watcher string) error {
	conn, err := grpc.DialContext(ctx, watcher, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("dial: %w", err)
	}
	defer conn.Close()

	ngc := api.NewGossipClient(conn)
	c, err := ngc.Clusters(ctx)
	if err != nil {
		return fmt.Errorf("clusters: %w", err)
	}

	err = c.Send(p.getLocCluster())
	if err != nil {
		return fmt.Errorf("send: %w", err)
	}

	p.watcher = c
	go func() {
		<-c.Context().Done()
		p.watcher = nil
	}()
	for {
		cs, err := c.Recv()
		if err != nil {
			return fmt.Errorf("receive: %w", err)
		}
		nr := make(map[string][]endpoint)
		for _, ci := range cs.Clusters {
			for _, ri := range ci.Routes {
				nr[ri.Prefix] = append(nr[ri.Prefix], endpoint{
					ri.Endpoint,
					ri.Cost,
				})
			}
		}
		<-p.remoteRoutes
		p.remoteRoutes <- nr
		p.rebalance <- struct{}{}
	}
}

func (p *Primary) gossipRunner(ctx context.Context) {
	for i := 0; i < len(p.watchers); i = (i + 1) % len(p.watchers) {
		retry := time.Second
		for retry < 30 {
			err := p.gossip(ctx, p.watchers[i])
			p.log.Error().Err(err).Dur("backoff", retry).Str("watcher", p.watchers[i]).Msg("gossip")
			time.Sleep(retry)
			if err != nil {
				retry *= 2
			}
		}
	}
}

func (p *Primary) getLocCluster() *api.Cluster {
	uris := p.localUris
	urs := make([]string, len(uris))
	copy(urs, uris)

	loc := <-p.localRoutes
	rs := make([]*api.Route, 0, len(loc)*len(urs))
	for prefix, cost := range loc {
		for _, u := range urs {
			rs = append(rs, &api.Route{
				Prefix:   prefix,
				Endpoint: u,
				Cost:     cost,
			})
		}
	}
	p.localRoutes <- loc

	return &api.Cluster{
		Id: p.name,
		// TODO: find local endpoint uri
		Primary: p.localUris[0] + ":" + p.port,
		Routes:  rs,
	}
}
