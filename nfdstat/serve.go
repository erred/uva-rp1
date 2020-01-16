package nfdstat

import (
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Server struct {
	log     *zerolog.Logger
	handler http.Handler

	// memory
	nfdname  string
	nfdpid   int
	pagesize int

	// prometheus metrics
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

func New(logger *zerolog.Logger) *Server {
	if logger == nil {
		*logger = log.With().Str("mod", "nfdstat").Logger()
	}
	s := Server{
		log:     logger,
		handler: promhttp.Handler(),

		nfdname:  "nfd",
		pagesize: os.Getpagesize(),

		memory: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "nfd_memory_bytes",
		}),
		cs_capacity: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "nfd_cs_capacity",
		}),
		cs_entries: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "nfd_cs_entries",
		}),
		cs_hits: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "nfd_cs_hits",
		}),
		cs_misses: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "nfd_cs_misses",
		}),
		nt_entries: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "nfd_nametree_entries",
		}),
		fib_entries: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "nfd_fib_entries",
		}),
		rib_entries: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "nfd_rib_entries",
		}),
		pit_entries: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "nfd_pit_entries",
		}),
		channel_entries: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "nfd_channel_entries",
		}),

		interests: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nfd_interets",
			},
			[]string{"satisfied"},
		),
		pkts: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nfd_pkts",
			},
			[]string{"direction", "type"},
		),

		face_entries: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "nfd_faces",
		}),
		face_bytes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nfd_face_bytes",
			},
			[]string{"direction", "id"},
		),
		face_pkts: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "nfd_face_pkts",
			},
			[]string{"direction", "type", "id"},
		),
	}

	var err error
	s.nfdpid, err = pid("nfd")
	if err != nil {
		s.log.Error().Err(err).Msg("can't find nfd")
		return nil
	}

	prometheus.MustRegister(s.memory)
	prometheus.MustRegister(s.cs_capacity)
	prometheus.MustRegister(s.cs_entries)
	prometheus.MustRegister(s.cs_hits)
	prometheus.MustRegister(s.cs_misses)
	prometheus.MustRegister(s.nt_entries)
	prometheus.MustRegister(s.fib_entries)
	prometheus.MustRegister(s.rib_entries)
	prometheus.MustRegister(s.pit_entries)
	prometheus.MustRegister(s.channel_entries)
	prometheus.MustRegister(s.interests)
	prometheus.MustRegister(s.pkts)
	prometheus.MustRegister(s.face_entries)
	prometheus.MustRegister(s.face_bytes)
	prometheus.MustRegister(s.face_pkts)

	return &s
}
