package watcher

import (
	"context"
	"net/url"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type Watcher struct {
	clusters chan map[string]*url.URL
	routes   chan map[string]*url.URL
}

func New(args []string, logger *zerolog.Logger) *Watcher {
	if logger == nil {
		*logger = log.With().Str("mod", "secondary").Logger()
	}
	panic("Unimplemented: New")
}
func Run(ctx context.Context) error {
	panic("Unimplemented: Run")
}

type prometheusService struct {
	Targets []string          `json:"targets,omitempty"`
	Labels  map[string]string `json:"labels,omitempty"`
}
