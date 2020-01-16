package watcher

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"strings"
	"time"

	"google.golang.org/grpc"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/seankhliao/uva-rp1/api"
)

type cluster struct {
	c api.Cluster
	t time.Time
	s api.Gossip_GossipClustersServer
}

func (c *cluster) send(cs *api.Clusters, log *zerolog.Logger) {
	err := c.s.Send(cs)
	if err != nil {
		log.Error().Err(err).Str("cluster", c.c.Id).Msg("send update")
	}
}

type Watcher struct {
	port     int
	promfile string

	clusters chan map[string]cluster

	log *zerolog.Logger
}

func New(args []string, logger *zerolog.Logger) *Watcher {
	if logger == nil {
		*logger = log.With().Str("mod", "watcher").Logger()
	}

	w := &Watcher{
		clusters: make(chan map[string]cluster),
		log:      logger,
	}
	w.clusters <- make(map[string]cluster)

	fs := flag.NewFlagSet("watcher", flag.ExitOnError)
	fs.IntVar(&w.port, "port", 8000, "port to serve on")
	fs.StringVar(&w.promfile, "file", "/etc/prometheus/file_sd.json", "prometheus file sd path")
	fs.Parse(args)

	return w
}

func (w *Watcher) Run(ctx context.Context) error {
	grpcServer := grpc.NewServer()
	api.RegisterGossipServer(grpcServer, w)

	return http.ListenAndServe(fmt.Sprintf(":%d", w.port), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		}
	}))
}
