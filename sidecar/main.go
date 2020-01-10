package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/cactus/go-statsd-client/statsd"
)

func main() {
	c, err := NewClient()
	if err != nil {
		log.Fatalf("sidecar: %s", err)
	}
	defer c.close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := c.run(ctx)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt)
	select {
	case s := <-sig:
		log.Printf("sidecar: got signal %s", s)
		cancel()
	case <-ctx.Done():
	}
	<-done
}

type Client struct {
	scrapeInterval time.Duration
	sampleRate     float32
	sds            []statsd.Statter

	name     string
	public   []string
	pagesize int
	nfdpid   int
}

func NewClient() (*Client, error) {
	c := &Client{}
	var addrs, reachable string
	var sr float64
	flag.DurationVar(&c.scrapeInterval, "interval", 15*time.Second, "scrape / flush interval")
	flag.Float64Var(&sr, "sample", 1, "sample rate 0-1")
	flag.StringVar(&addrs, "servers", "145.100.104.117:8125", "comma separated list of reporting servers host:port")
	flag.StringVar(&reachable, "addrs", "", "comma separated list of public addresses net://host:port")
	flag.StringVar(&c.name, "name", "ndn_node", "name of node in reporting")
	flag.Parse()
	c.sampleRate = float32(sr)
	c.public = strings.Split(reachable, ",")

	for _, a := range strings.Split(addrs, ",") {
		// TODO: find a name prefix
		s, err := statsd.NewClient(a, "some-name")
		if err != nil {
			return nil, fmt.Errorf("create client: %w", err)
		}
		c.sds = append(c.sds, s)
	}
	c.pagesize = os.Getpagesize()
	var err error
	c.nfdpid, err = pid("nfd")
	if err != nil {
		return nil, fmt.Errorf("create client: get nfd pid %w", err)
	}
	return c, nil
}

func (c *Client) close() {
	for _, s := range c.sds {
		s.Close()
	}
}

func (c *Client) run(ctx context.Context) chan struct{} {
	t := time.NewTicker(c.scrapeInterval)
	done := make(chan struct{})
	go func() {
		<-ctx.Done()
		t.Stop()
		done <- struct{}{}
	}()
	go func() {
		for range t.C {
			go func() {
				ctx, cancel := context.WithTimeout(ctx, c.scrapeInterval)
				defer cancel()
				c.status(ctx)
			}()
		}
	}()
	return done
}

func (c *Client) status(ctx context.Context) {
	status, err := status(ctx)
	if err != nil {
		log.Printf("sidecar: status: %s", err)
	}
	mem, err := memory(c.nfdpid, c.pagesize)
	if err != nil {
		log.Printf("sidecar: memory: %s", err)
	}
	for _, st := range c.sds {
		log.Printf("sidecar: mem %d, cs.entries %d, satisfied %d, unsatisfied %d\n", mem, status.Cs.NEntries, status.GeneralStatus.NSatisfiedInterests, status.GeneralStatus.NUnsatisfiedInterests)
		// TODO: routes (or use NLSR)
		for _, a := range c.public {
			st.Gauge("addrs."+a, 1, c.sampleRate)
		}

		st.Gauge("memory", mem, c.sampleRate)

		st.Gauge("cs.capacity", status.Cs.Capacity, c.sampleRate)
		st.Gauge("cs.entries", status.Cs.NEntries, c.sampleRate)
		st.Gauge("cs.hits", status.Cs.NHits, c.sampleRate)
		st.Gauge("cs.misses", status.Cs.NMisses, c.sampleRate)

		st.Gauge("fib.entries", status.GeneralStatus.NFibEntries, c.sampleRate)
		st.Gauge("pit.entries", status.GeneralStatus.NPitEntries, c.sampleRate)
		st.Gauge("rib.entries", int64(len(status.Rib.RibEntry)), c.sampleRate)
		st.Gauge("channel.entries", int64(len(status.Channels.Channel)), c.sampleRate)
		st.Gauge("nametree.entries", status.GeneralStatus.NNameTreeEntries, c.sampleRate)

		st.Gauge("interests.satisfied", status.GeneralStatus.NSatisfiedInterests, c.sampleRate)
		st.Gauge("interests.unsatisfied", status.GeneralStatus.NUnsatisfiedInterests, c.sampleRate)

		st.Gauge("incoming.interests", status.GeneralStatus.PacketCounters.IncomingPackets.NInterests, c.sampleRate)
		st.Gauge("incoming.data", status.GeneralStatus.PacketCounters.IncomingPackets.NData, c.sampleRate)
		st.Gauge("incoming.nacks", status.GeneralStatus.PacketCounters.IncomingPackets.NNacks, c.sampleRate)
		st.Gauge("outgoing.interests", status.GeneralStatus.PacketCounters.OutgoingPackets.NInterests, c.sampleRate)
		st.Gauge("outgoing.data", status.GeneralStatus.PacketCounters.OutgoingPackets.NData, c.sampleRate)
		st.Gauge("outgoing.nacks", status.GeneralStatus.PacketCounters.OutgoingPackets.NNacks, c.sampleRate)

		st.Gauge("face.entries", int64(len(status.Faces.Face)), c.sampleRate)
		for _, face := range status.Faces.Face {
			pref := fmt.Sprintf("face.%d", face.FaceId)
			st.Gauge(pref+".incoming.interests", face.PacketCounters.IncomingPackets.NInterests, c.sampleRate)
			st.Gauge(pref+".incoming.data", face.PacketCounters.IncomingPackets.NData, c.sampleRate)
			st.Gauge(pref+".incoming.nacks", face.PacketCounters.IncomingPackets.NNacks, c.sampleRate)
			st.Gauge(pref+".incoming.bytes", face.ByteCounters.IncomingBytes, c.sampleRate)
			st.Gauge(pref+".outgoing.interests", face.PacketCounters.OutgoingPackets.NInterests, c.sampleRate)
			st.Gauge(pref+".outgoing.data", face.PacketCounters.OutgoingPackets.NData, c.sampleRate)
			st.Gauge(pref+".outgoing.nacks", face.PacketCounters.OutgoingPackets.NNacks, c.sampleRate)
			st.Gauge(pref+".outgoing.bytes", face.ByteCounters.OutgoingBytes, c.sampleRate)
		}
	}
}
