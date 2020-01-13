package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	c := NewClient()

	flag.DurationVar(&c.interval, "interval", 5*time.Second, "scrape duration")
	flag.StringVar(&c.collectors, "collectors", "172.18.0.2:8000", "comma separated list of collecters")
	flag.StringVar(&c.public, "public", "0.0.0.0", "public ip of this instance")
	flag.IntVar(&c.port, "port", 8000, "port to serve on")
	flag.Parse()

	go c.run()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(fmt.Sprintf(":%d", c.port), nil)
}

type Client struct {
	collectors string
	interval   time.Duration
	port       int
	public     string

	pagesize int
	nfdpid   int

	memory          prometheus.Gauge
	cs_capacity     prometheus.Gauge
	cs_entries      prometheus.Gauge
	cs_hits         prometheus.Gauge
	cs_misses       prometheus.Gauge
	nt_entries      prometheus.Gauge
	fib_entries     prometheus.Gauge
	rib_entries     prometheus.Gauge
	pit_entries     prometheus.Gauge
	channel_entries prometheus.Gauge
	face_entries    prometheus.Gauge

	// statisfied: yes|no
	interests *prometheus.GaugeVec

	// direction: in|out
	// type: interest|data|nack
	pkts *prometheus.GaugeVec

	// direction: in|out
	// type: interets|data|nack
	// id: $value
	face_pkts *prometheus.GaugeVec

	// direction: in|out
	// id: $value
	face_bytes *prometheus.GaugeVec
}

func NewClient() *Client {
	var c Client
	var err error
	c.pagesize = os.Getpagesize()
	c.nfdpid, err = pid("nfd")
	if err != nil {
		log.Fatalln("sidecar: run ", err)
	}

	c.memory = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "nfd_memory_bytes",
	})
	c.cs_capacity = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "nfd_cs_capacity",
	})
	c.cs_entries = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "nfd_cs_entries",
	})
	c.cs_hits = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "nfd_cs_hits",
	})
	c.cs_misses = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "nfd_cs_misses",
	})
	c.nt_entries = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "nfd_nametree_entries",
	})
	c.fib_entries = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "nfd_fib_entries",
	})
	c.rib_entries = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "nfd_rib_entries",
	})
	c.pit_entries = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "nfd_pit_entries",
	})
	c.channel_entries = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "nfd_channel_entries",
	})

	c.interests = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nfd_interets",
		},
		[]string{"satisfied"},
	)
	c.pkts = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nfd_pkts",
		},
		[]string{"direction", "type"},
	)

	c.face_entries = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "nfd_faces",
	})
	c.face_bytes = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nfd_face_bytes",
		},
		[]string{"direction", "id"},
	)
	c.face_pkts = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "nfd_face_pkts",
		},
		[]string{"direction", "type", "id"},
	)

	prometheus.MustRegister(c.memory)
	prometheus.MustRegister(c.cs_capacity)
	prometheus.MustRegister(c.cs_entries)
	prometheus.MustRegister(c.cs_hits)
	prometheus.MustRegister(c.cs_misses)
	prometheus.MustRegister(c.nt_entries)
	prometheus.MustRegister(c.fib_entries)
	prometheus.MustRegister(c.rib_entries)
	prometheus.MustRegister(c.pit_entries)
	prometheus.MustRegister(c.channel_entries)
	prometheus.MustRegister(c.interests)
	prometheus.MustRegister(c.pkts)
	prometheus.MustRegister(c.face_entries)
	prometheus.MustRegister(c.face_bytes)
	prometheus.MustRegister(c.face_pkts)

	return &c
}

func (c *Client) run() {
	// register with servers
	var err error
	svc := Service{
		Targets: []string{fmt.Sprintf("%s:%d", c.public, c.port)},
	}
	b, err := json.Marshal(svc)
	if err != nil {
		log.Fatalln("sidecar: marshal ", err)
	}
	log.Printf("sidecar: registering %s\n", b)
	for _, a := range strings.Split(c.collectors, ",") {
		rd := bytes.NewReader(b)
		res, err := http.Post(fmt.Sprintf("http://%s/register", a), "application/json", rd)
		if err != nil {
			log.Printf("sidecar: register %s %v\n", a, err)
			continue
		}
		if res.StatusCode != http.StatusOK {
			log.Printf("sidecar: register %s %s\n", a, http.StatusText(res.StatusCode))
			continue
		}
	}

	t := time.NewTicker(c.interval)
	go func() {
		go c.status()
		for range t.C {
			go c.status()
		}
	}()
}

func (c *Client) status() {
	ctx, cancel := context.WithTimeout(context.Background(), c.interval)
	defer cancel()

	mem, err := memory(c.nfdpid, c.pagesize)
	if err != nil {
		log.Printf("sidecar: memory: %s", err)
	}
	c.memory.Set(float64(mem))

	status, err := status(ctx)
	if err != nil {
		log.Printf("sidecar: status: %s", err)
	}

	c.cs_capacity.Set(float64(status.Cs.Capacity))
	c.cs_entries.Set(float64(status.Cs.NEntries))
	c.cs_hits.Set(float64(status.Cs.NHits))
	c.cs_misses.Set(float64(status.Cs.NMisses))
	c.nt_entries.Set(float64(status.GeneralStatus.NNameTreeEntries))
	c.fib_entries.Set(float64(status.GeneralStatus.NFibEntries))
	c.rib_entries.Set(float64(len(status.Rib.RibEntry)))
	c.pit_entries.Set(float64(status.GeneralStatus.NPitEntries))
	c.channel_entries.Set(float64(len(status.Channels.Channel)))
	c.face_entries.Set(float64(len(status.Faces.Face)))

	c.interests.WithLabelValues("yes").Set(float64(status.GeneralStatus.NSatisfiedInterests))
	c.interests.WithLabelValues("no").Set(float64(status.GeneralStatus.NUnsatisfiedInterests))

	c.pkts.WithLabelValues("in", "interest").Set(float64(status.GeneralStatus.PacketCounters.IncomingPackets.NInterests))
	c.pkts.WithLabelValues("in", "data").Set(float64(status.GeneralStatus.PacketCounters.IncomingPackets.NData))
	c.pkts.WithLabelValues("in", "nack").Set(float64(status.GeneralStatus.PacketCounters.IncomingPackets.NNacks))
	c.pkts.WithLabelValues("out", "interest").Set(float64(status.GeneralStatus.PacketCounters.OutgoingPackets.NInterests))
	c.pkts.WithLabelValues("out", "data").Set(float64(status.GeneralStatus.PacketCounters.OutgoingPackets.NData))
	c.pkts.WithLabelValues("out", "nack").Set(float64(status.GeneralStatus.PacketCounters.OutgoingPackets.NNacks))

	for _, f := range status.Faces.Face {
		c.face_bytes.WithLabelValues("in", f.FaceId).Set(float64(f.ByteCounters.IncomingBytes))
		c.face_bytes.WithLabelValues("out", f.FaceId).Set(float64(f.ByteCounters.OutgoingBytes))

		c.face_pkts.WithLabelValues("in", "interest", f.FaceId).Set(float64(f.PacketCounters.IncomingPackets.NInterests))
		c.face_pkts.WithLabelValues("in", "data", f.FaceId).Set(float64(f.PacketCounters.IncomingPackets.NData))
		c.face_pkts.WithLabelValues("in", "nack", f.FaceId).Set(float64(f.PacketCounters.IncomingPackets.NNacks))
		c.face_pkts.WithLabelValues("out", "interest", f.FaceId).Set(float64(f.PacketCounters.OutgoingPackets.NInterests))
		c.face_pkts.WithLabelValues("out", "data", f.FaceId).Set(float64(f.PacketCounters.OutgoingPackets.NData))
		c.face_pkts.WithLabelValues("out", "nack", f.FaceId).Set(float64(f.PacketCounters.OutgoingPackets.NNacks))
	}
}

type Service struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}
