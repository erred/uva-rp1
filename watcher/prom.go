package watcher

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

func (w *Watcher) SavePromSD() error {
	cs := <-w.clusters
	sd := make([]prometheusService, 0, len(cs))
	for k, v := range cs {
		sd = append(sd, prometheusService{
			Targets: []string{v.c.Primary},
			Labels:  map[string]string{"clusterid": k},
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
	return nil
}

type prometheusService struct {
	Targets []string          `json:"targets,omitempty"`
	Labels  map[string]string `json:"labels,omitempty"`
}
