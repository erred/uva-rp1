package watcher

import (
	"context"
	"fmt"
	"time"

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
			err := fmt.Errorf("duplicate id")
			w.log.Error().
				Err(err).
				Str("id", p.PrimaryId).
				Msg("handle primaries")
			return err
		}
		pr[p.PrimaryId] = primary{
			p, s,
		}
		w.notify()
		w.log.Info().
			Str("id", p.PrimaryId).
			Msg("handle primaries registered")
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
		w.log.Info().
			Str("id", id).
			Msg("handle primaries unregistered")
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
		w.log.Info().
			Str("id", id).
			Msg("handle gossip unregistered")
	}(id)

	return w.gossipRecv(id, s)
}

func (w *Watcher) gossipRunner(watcher string) {
	for {
		w.gossiper(watcher)
		w.log.Info().
			Dur("backoff", time.Second).
			Str("id", watcher).
			Msg("gossipRunner")
		time.Sleep(time.Second)
	}
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
	if err != nil {
		w.log.Error().
			Err(err).
			Str("id", id).
			Msg("handle gossip recv init")
		return

	}

	defer func(id string) {
		r := <-w.reflectors
		delete(r, id)
		w.reflectors <- r
		w.notify()
		w.log.Info().Str("id", id).Msg("gossiper unregistered")
	}(id)

	err = w.gossipRecv(id, c)
	if err != nil {
		w.log.Error().
			Err(err).
			Msg("gossiper recv error")
	}
}

func (w *Watcher) gossipInitRecv(g gossiper) (string, error) {
	ap, err := g.Recv()
	if err != nil {
		return "unknown", err
	}

	id := ap.WatcherId
	r := <-w.reflectors
	defer func() {
		w.reflectors <- r
	}()
	if _, ok := r[ap.WatcherId]; ok {
		err = fmt.Errorf("duplicate id")
		return id, err
	}
	r[id] = reflector{
		ap, g,
	}
	w.notify()
	w.log.Info().
		Str("id", id).
		Msg("handle gossip registered")
	return id, nil
}

func (w *Watcher) gossipRecv(id string, g gossipRecver) error {
	known := make(map[string]struct{})

	for {
		ap, err := g.Recv()
		if err != nil {
			w.log.Error().
				Err(err).
				Str("id", id).
				Msg("handle gossip recv")
			return err
		}
		var diff bool
		for _, p := range ap.Primaries {
			if _, ok := known[p.PrimaryId]; !ok {
				diff = true
			}
		}
		curr := make(map[string]struct{}, len(ap.Primaries))
		for _, p := range ap.Primaries {
			curr[p.PrimaryId] = struct{}{}
		}
		for pid := range known {
			if _, ok := curr[pid]; !ok {
				diff = true
			}
		}
		known = curr
		if diff {
			r := <-w.reflectors
			re := r[id]
			re.a = ap
			r[id] = re
			w.reflectors <- r
			w.notify()
		}

		w.log.Info().
			Str("id", id).
			Int("primaries", len(ap.Primaries)).
			Bool("diff", diff).
			Msg("gosssip recv")
	}
}
