package sidecar

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (c *Client) Run() {
	go c.notificationKeeper()
	go c.faceKeeper()
	// ensure first status
	c.status()
	// register with servers
	for _, a := range strings.Split(c.collectors, ",") {
		go c.registerKeepAlive(c.keepAlive, a)
	}
	// scrape status
	go func() {
		for range time.NewTicker(c.interval).C {
			go c.status()
		}
	}()

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(fmt.Sprintf(":%d", c.port), nil)
}

type Client struct {
	collectors    string
	interval      time.Duration
	keepAlive     time.Duration
	blacklistTime time.Duration
	port          int
	// public     string
	ip4 []net.IP
	ip6 []net.IP

	pagesize int
	nfdpid   int

	blacklist map[string]time.Time
	faces     map[string]string
	fin       chan map[string]string
	fout      chan map[string]string

	notification Notification
	cin          chan []string
	cout         chan Notification

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
	c := Client{
		cin:       make(chan []string),
		cout:      make(chan Notification),
		fin:       make(chan map[string]string),
		fout:      make(chan map[string]string),
		blacklist: make(map[string]time.Time),
		pagesize:  os.Getpagesize(),

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

	flag.DurationVar(&c.blacklistTime, "blacktlisttime", 10*time.Minute, "time to blacklist after failed face creation")
	flag.DurationVar(&c.interval, "interval", 5*time.Second, "scrape duration")
	flag.DurationVar(&c.keepAlive, "keepalive", 10*time.Second, "registration keep alive")
	flag.StringVar(&c.collectors, "collectors", "172.18.0.2:8000", "comma separated list of collecters")
	// flag.StringVar(&c.public, "public", "0.0.0.0", "public ip of this instance")
	flag.IntVar(&c.port, "port", 8000, "port to serve on")
	flag.Parse()

	var err error
	c.notification.Metrics = c.port
	c.nfdpid, err = pid("nfd")
	if err != nil {
		log.Fatalln("NewClient ", err)
	}

	ads, err := net.InterfaceAddrs()
	if err != nil {
		log.Fatalln("NewClient ", err)
	}
	for _, ad := range ads {
		an, ok := ad.(*net.IPNet)
		if !ok {
			continue
		}
		if !an.IP.IsGlobalUnicast() {
			continue
		}
		a4 := an.IP.To4()
		if a4 != nil {
			c.ip4 = append(c.ip4, a4)
		} else {
			c.ip6 = append(c.ip6, an.IP)
		}
	}

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

func (c *Client) status() {
	ctx, cancel := context.WithTimeout(context.Background(), c.interval)
	defer cancel()

	mem, err := memory(c.nfdpid, c.pagesize)
	if err != nil {
		log.Printf("status memory: %s\n", err)
	} else {
		c.memory.Set(float64(mem))
	}

	status, err := status(ctx)
	if err != nil {
		log.Printf("status: %s\n", err)
		return
	}

	var uris []string
	for _, ch := range status.Channels.Channel {
		switch ch.LocalUri[:3] {
		case "tcp", "udp":
			uris = append(uris, ch.LocalUri)
		}
	}
	c.cin <- uris

	fcs := make(map[string]string)
	for _, fa := range status.Faces.Face {
		fcs[fa.RemoteUri] = fa.FaceId
	}
	c.fin <- fcs

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

func (c *Client) registerKeepAlive(d time.Duration, addr string) {
	t := time.NewTicker(d)
	go c.register(addr)
	for range t.C {
		c.register(addr)
	}
}
func (c *Client) register(addr string) {
	n := <-c.cout
	b, err := json.Marshal(n)
	if err != nil {
		log.Printf("register %s: %s\n", addr, err)
		return
	}
	u := fmt.Sprintf("http://%s/register", addr)
	res, err := http.Post(u, "application/json", bytes.NewReader(b))
	if err != nil {
		log.Printf("register %s: %s\n", addr, err)
		return
	} else if res.StatusCode != 200 {
		log.Printf("register %s: %d %s\n", addr, res.StatusCode, res.Status)
		return
	}
	defer res.Body.Close()

	b, err = ioutil.ReadAll(res.Body)
	if err != nil {
		log.Printf("register response %s\n", err)
		return
	}
	bb := bytes.Fields(b)
	faces := <-c.fout
	for _, b := range bb {
		if _, ok := faces[string(b)]; !ok {
			go c.addFace(string(b))
		}
	}
}

func (c *Client) addFace(uri string) {
	// ignore own ips
	u, err := url.Parse(uri)
	if err != nil {
		log.Printf("add face filter %s\n", err)
		return
	}
	ip := net.ParseIP(u.Hostname())
	for _, i := range c.ip4 {
		if i.Equal(ip) {
			return
		}
	}
	for _, i := range c.ip6 {
		if i.Equal(ip) {
			return
		}
	}
	// ignore blacklist
	if v, ok := c.blacklist[uri]; ok {
		if v.Add(c.blacklistTime).After(time.Now()) {
			return
		}
		delete(c.blacklist, uri)
	}

	args := []string{"face", "create", uri, "persistent"}
	log.Printf("creating face : %v\n", args)
	b, err := exec.Command("nfdc", args...).CombinedOutput()
	if err != nil {
		c.blacklist[uri] = time.Now()
		log.Printf("face create %s\n%s\n", err, b)
		return
	}
}

func (c *Client) faceKeeper() {
	m := make(map[string]string)
	for k, v := range c.faces {
		m[k] = v
	}
	for {
		select {
		case c.faces = <-c.fin:
			m = make(map[string]string)
			for k, v := range c.faces {
				m[k] = v
			}
		case c.fout <- m:
		}
	}
}

func (c *Client) notificationKeeper() {
	for {
		select {
		case uris := <-c.cin:
			var addrs []string
			for _, uri := range uris {
				u, err := url.Parse(uri)
				if err != nil {
					log.Printf("notikeep parse %s\n", err)
					continue
				}
				switch u.Scheme {
				case "tcp4", "udp4":
					for _, ip := range c.ip4 {
						u.Host = fmt.Sprintf("%s:%s", ip, u.Port())
						addrs = append(addrs, u.String())
					}
				case "tcp6", "udp6":
					for _, ip := range c.ip6 {
						u.Host = fmt.Sprintf("%s:%s", ip, u.Port())
						addrs = append(addrs, u.String())
					}
				}
			}
			c.notification.Addrs = addrs
		case c.cout <- c.notification:
		}
	}
}

type Notification struct {
	Metrics int
	uris    []string
	Addrs   []string
}
