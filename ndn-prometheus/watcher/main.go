package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

func main() {
	var svc ServiceDiscovery
	var port = 8000
	var err error

	flag.IntVar(&port, "port", port, "port to listen on")
	flag.StringVar(&svc.file, "file", "/etc/prometheus/file_sd.json", "file to write to")
	flag.Parse()

	svc.services, err = readSD(svc.file)
	if err != nil {
		log.Println("watcher: ", err)
	}

	http.HandleFunc("/register", func(w http.ResponseWriter, r *http.Request) {
		err := svc.register(r.Body)
		defer r.Body.Close()
		if err != nil {
			log.Println("watcher: ", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

type ServiceDiscovery struct {
	services []Service
	file     string
}

func (s *ServiceDiscovery) register(r io.Reader) error {
	b, err := ioutil.ReadAll(r)
	if err != nil {
		return fmt.Errorf("register: read body %w", err)
	}
	log.Printf("watcher: received %s\n", b)
	var svc Service
	err = json.Unmarshal(b, &svc)
	if err != nil {
		return fmt.Errorf("register: unmarshal %w", err)
	}
	if len(svc.Targets) == 0 {
		return fmt.Errorf("register: no targets")
	}

	s.services = append(s.services, svc)

	log.Printf("watcher: registered %v", svc.Targets)
	return writeSD(s.file, s.services)
}

// func (s *ServiceDiscovery) register(addr string, r io.Reader) error {
// 	for _, service := range s.services {
// 		for _, t := range service.Targets {
// 			if t == addr {
// 				return fmt.Errorf("register: duplicate addr %s", addr)
// 			}
// 		}
// 	}
//
// 	b, err := ioutil.ReadAll(r)
// 	if err != nil {
// 		return fmt.Errorf("register: read body")
// 	}
// 	var m map[string]string
// 	err = json.Unmarshal(b, &m)
// 	if err != nil {
// 		return fmt.Errorf("register unmarshal")
// 	}
//
// 	s.services = append(s.services, Service{
// 		Targets: []string{addr},
// 		Labels:  m,
// 	})
// 	return writeSD(s.file, s.services)
// }

type Service struct {
	Targets []string          `json:"targets"`
	Labels  map[string]string `json:"labels"`
}

func readSD(file string) ([]Service, error) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, fmt.Errorf("readSD: read %w", err)
	}
	var svc []Service
	err = json.Unmarshal(b, &svc)
	if err != nil {
		return nil, fmt.Errorf("readSD: unmarshal %w", err)
	}
	return svc, nil
}

func writeSD(file string, svc []Service) error {

	b, err := json.Marshal(svc)
	if err != nil {
		return fmt.Errorf("writeSD: marshal %w", err)
	}
	err = ioutil.WriteFile(file, b, 0644)
	if err != nil {
		return fmt.Errorf("writeSD: write %w", err)
	}
	return nil
}
