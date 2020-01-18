package watcher

import (
	"context"
	"fmt"
	"github.com/seankhliao/uva-rp1/api"
	"google.golang.org/grpc"
)

func (w *Watcher) Primaries(p *api.Primary, s api.Reflector_PrimariesServer) error {
	err := func() error {
		pr := <-w.primaries
		defer func() {
			w.primaries <- pr
		}()
		if _, ok := pr[p.PrimaryId]; ok {
			err := fmt.Errorf("duplicate id: %s %s", p.PrimaryId, p.Endpoint)
			w.log.Error().Err(err).Str("id", p.PrimaryId).Msg("primaries duplicate")
			return err
		}
		pr[p.PrimaryId] = primary{
			p, s,
		}
		w.notify()
		w.log.Info().Str("id", p.PrimaryId).Str("endpoint", p.Endpoint).Msg("primaries registered")
		return nil
	}()
	if err != nil {
		return err
	}

	defer func(id string) {
		pr := <-w.primaries
		delete(pr, id)
		w.primaries <- pr
		w.notify()
		w.log.Info().Str("id", id).Msg("primaries unregistered")
	}(p.PrimaryId)

	<-s.Context().Done()
	return nil
}

func (w *Watcher) Gossip(s api.Reflector_GossipServer) error {
	id, err := w.gossipInitRecv(s)
	if err != nil {
		return err
	}

	defer func(id string) {
		r := <-w.reflectors
		delete(r, id)
		w.reflectors <- r
		w.notify()
		w.log.Info().Str("id", id).Msg("gossip unregistered")
	}(id)

	return w.gossipRecv(id, s)
}

func (w *Watcher) gossiper(watcher string) {
	conn, err := grpc.DialContext(context.Background(), watcher, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		w.log.Error().Err(err).Str("watcher", watcher).Msg("gossiper dial")
		return
	}
	defer conn.Close()

	c, err := api.NewReflectorClient(conn).Gossip(context.Background())
	if err != nil {
		w.log.Error().Err(err).Str("watcher", watcher).Msg("gossiper client")
		return
	}
	ap := w.allPrimaries()
	err = c.Send(ap)
	if err != nil {
		w.log.Error().Err(err).Str("watcher", watcher).Msg("gossiper initial send")
		return
	}

	id, err := w.gossipInitRecv(c)

	defer func(id string) {
		r := <-w.reflectors
		delete(r, id)
		w.reflectors <- r
		w.notify()
		w.log.Info().Str("id", id).Msg("gossiper unregistered")
	}(id)

	w.gossipRecv(id, c)
}

func (w *Watcher) gossipInitRecv(g gossiper) (string, error) {
	ap, err := g.Recv()
	if err != nil {
		w.log.Error().Err(err).Msg("gossip init receive")
	}

	id := ap.WatcherId
	r := <-w.reflectors
	defer func() {
		w.reflectors <- r
	}()
	if _, ok := r[ap.WatcherId]; ok {
		err = fmt.Errorf("duplicate id: %s", id)
		w.log.Error().Err(err).Str("id", id).Msg("gossip init duplicate")
		return id, err
	}
	r[id] = reflector{
		ap, g,
	}
	w.notify()
	w.log.Info().Str("id", id).Msg("gossip registered")
	return id, nil
}

func (w *Watcher) gossipRecv(id string, g gossipRecver) error {
	for {
		ap, err := g.Recv()
		if err != nil {
			w.log.Error().Err(err).Str("id", id).Msg("gossip recv")
			return err
		}

		r := <-w.reflectors
		re := r[id]
		re.a = ap
		r[id] = re
		w.reflectors <- r
		w.notify()
		w.log.Info().Str("id", id).Msg("gosssip recv")
	}
}
