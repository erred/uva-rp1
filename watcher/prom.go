package watcher

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func (w *Watcher) SavePromSD() error {
	ap := w.allPrimaries()

	es := make([]string, 0, len(ap.Primaries))
	sd := make([]prometheusService, 0, len(ap.Primaries))
	for _, p := range ap.Primaries {
		es = append(es, p.Endpoint)
		sd = append(sd, prometheusService{
			Targets: []string{p.Endpoint},
			Labels:  map[string]string{"primary_id": p.PrimaryId},
		})
	}
	b, err := json.Marshal(sd)
	if err != nil {
		return fmt.Errorf("marshal prom sd: %w", err)
	}
	err = ioutil.WriteFile(w.promfile, b, 0644)
	if err != nil {
		return fmt.Errorf("write prom sd %s: %w", w.promfile, err)
	}
	w.log.Info().
		Strs("endpoints", es).
		Msg("wrote prom file_sd")
	return nil
}

type prometheusService struct {
	Targets []string          `json:"targets,omitempty"`
	Labels  map[string]string `json:"labels,omitempty"`
}
