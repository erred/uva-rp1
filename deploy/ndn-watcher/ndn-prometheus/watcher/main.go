package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

func main() {
	log.SetPrefix("watcher | ")
	c := NewClient()

	flag.DurationVar(&c.keepAlive, "keepalive", time.Minute, "flush timer")
	flag.IntVar(&c.port, "port", 8000, "port to listen on")
	flag.StringVar(&c.file, "file", "/etc/prometheus/file_sd.json", "file to write to")
	flag.Parse()

	go c.addrUpdater()
	go c.targetUpdater()

	http.HandleFunc("/addrs", c.listAddrs)
	http.HandleFunc("/register", c.register)
	http.ListenAndServe(fmt.Sprintf(":%d", c.port), nil)
}

type Client struct {
	keepAlive time.Duration
	targets   map[string]time.Time
	addrs     map[string]time.Time

	port int
	file string

	tin  chan string
	cin  chan []string
	cout chan []string
}

func NewClient() *Client {
	return &Client{
		targets: make(map[string]time.Time),
		addrs:   make(map[string]time.Time),
		cin:     make(chan []string),
		cout:    make(chan []string),
		tin:     make(chan string),
	}
}

func (c *Client) register(w http.ResponseWriter, r *http.Request) {
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
	c.tin <- fmt.Sprintf("%s:%d", r.RemoteAddr[:li], not.Metrics)
	c.cin <- not.Addrs
	log.Printf("registered %s\n", fmt.Sprintf("%s:%d", r.RemoteAddr[:li], not.Metrics))

	http.HandlerFunc(c.listAddrs).ServeHTTP(w, r)
}

func (c *Client) listAddrs(w http.ResponseWriter, r *http.Request) {
	addrs := <-c.cout
	w.Write([]byte(strings.Join(addrs, "\n")))
}

func (c *Client) addrUpdater() {
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
func (c *Client) targetUpdater() {
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
