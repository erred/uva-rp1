package watcher

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/seankhliao/uva-rp1/api"
	"google.golang.org/grpc"
)

type Server struct {
	keepAlive time.Duration
	targets   map[string]time.Time
	addrs     map[string]time.Time

	port int
	file string

	tin  chan string
	cin  chan []string
	cout chan []string
}

func NewServer(args []string) *Server {
	s := &Server{
		targets: make(map[string]time.Time),
		addrs:   make(map[string]time.Time),
		cin:     make(chan []string),
		cout:    make(chan []string),
		tin:     make(chan string),
	}
	fs := flag.NewFlagSet("watcher", flag.ExitOnError)
	fs.DurationVar(&s.keepAlive, "keepalive", time.Minute, "flush timer")
	fs.IntVar(&s.port, "port", 8000, "port to listen on")
	fs.StringVar(&s.file, "file", "/etc/prometheus/file_sd.json", "file to write to")
	fs.Parse(args)
	return s
}

func (s *Server) Serve() {
	go s.addrUpdater()
	go s.targetUpdater()

	httpServer := http.ServeMux{}
	httpServer.HandleFunc("/addrs", s.listAddrs)
	httpServer.HandleFunc("/register", s.register)

	grpcServer := grpc.NewServer()
	// api.RegisterDiscoveryServiceServer(grpcServer, s)

	http.ListenAndServe(fmt.Sprintf(":%d", s.port), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
			grpcServer.ServeHTTP(w, r)
		} else {
			httpServer.ServeHTTP(w, r)
		}
	}))
}

func (s *Server) register(w http.ResponseWriter, r *http.Request) {
	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Printf("register %s\n", err)
		return
	}
	var not Notification
	err = json.Unmarshal(b, &not)
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		log.Printf("register %s\n", err)
		return
	}
	li := strings.LastIndex(r.RemoteAddr, ":")
	s.tin <- fmt.Sprintf("%s:%d", r.RemoteAddr[:li], not.Metrics)
	s.cin <- not.Addrs
	log.Printf("registered %s\n", fmt.Sprintf("%s:%d", r.RemoteAddr[:li], not.Metrics))

	http.HandlerFunc(s.listAddrs).ServeHTTP(w, r)
}

func (s *Server) listAddrs(w http.ResponseWriter, r *http.Request) {
	addrs := <-s.cout
	w.Write([]byte(strings.Join(addrs, "\n")))
}

func (c *Server) addrUpdater() {
	var addrs []string
	for {
		select {
		case as := <-c.cin:
			for _, a := range as {
				c.addrs[a] = time.Now()
			}
			for k, v := range c.addrs {
				if v.Add(c.keepAlive).Before(time.Now()) {
					delete(c.addrs, k)
				}
			}
			addrs = addrs[:0]
			for k := range c.addrs {
				addrs = append(addrs, k)
			}
		case c.cout <- addrs:
		}
	}
}
func (c *Server) targetUpdater() {
	for a := range c.tin {
		var diff bool
		if _, ok := c.targets[a]; !ok {
			diff = true
		}
		c.targets[a] = time.Now()
		for k, v := range c.targets {
			if v.Add(c.keepAlive).Before(time.Now()) {
				diff = true
				delete(c.addrs, k)
			}
		}
		if !diff {
			continue
		}
		var svcs []Service
		for k := range c.targets {
			svcs = append(svcs, Service{
				Targets: []string{k},
			})
		}

		b, err := json.Marshal(svcs)
		if err != nil {
			log.Printf("marshal %s\n", err)
			continue
		}
		err = ioutil.WriteFile(c.file, b, 0644)
		if err != nil {
			log.Printf("write %s\n", err)
			continue
		}
	}
}

type Notification struct {
	Metrics int
	Addrs   []string
}

type Service struct {
	Targets []string          `json:"targets,omitempty"`
	Labels  map[string]string `json:"labels,omitempty"`
}
