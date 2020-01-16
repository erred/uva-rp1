package watcher

import (
	"github.com/seankhliao/uva-rp1/api"
	"io"
	"time"
)

func (w *Watcher) Clusters(s api.Gossip_ClustersServer) error {
	// wait for hello
	ci, err := s.Recv()
	if err != nil {
		w.log.Error().Err(err).Msg("gossip initial receive")
		return err
	}
	cid := ci.Id
	w.log.Info().Str("cluster", cid).Msg("gossip connected")

	for {
		ocs := <-w.clusters
		// update cluster info
		c := ocs[ci.Id]
		c.t = time.Now()
		c.c = *ci
		c.s = s
		ocs[ci.Id] = c

		// create new clusters array
		ncs := &api.ClusterList{
			Clusters: make([]*api.Cluster, 0, len(ocs)),
		}
		for _, v := range ocs {
			vc := &v.c
			ncs.Clusters = append(ncs.Clusters, vc)
		}
		for _, v := range ocs {
			go v.send(ncs, w.log)
		}
		w.clusters <- ocs

		go func() {
			err = w.SavePromSD()
			if err != nil {
				w.log.Error().Err(err).Msg("save prom sd")
			}
		}()

		ci, err = s.Recv()
		if err == io.EOF {
			w.log.Error().Err(err).Str("cluster", cid).Msg("EOF")
			return nil
		} else if err != nil {
			w.log.Error().Err(err).Str("cluster", cid).Msg("gossip receive")
			return err
		}

	}
}
