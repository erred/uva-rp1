package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/cactus/go-statsd-client/statsd"
)

const (
	// comma separated list of hostname:ports
	// of servers to send stats to
	// overrriden by positional arguments
	CollectorEnv = "SIDECAR_COLLECTORS"
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
	signal.Notify(sig)
	select {
	case s := <-sig:
		log.Printf("sidecar: got signal %s", s)
	case <-ctx.Done():
	}
	<-done
}

type Client struct {
	// TODO: expose config options
	scapeInterval time.Duration
	sampleRate    float32
	sds           []statsd.Statter

	pagesize int
	nfdpid   int
}

func NewClient() (*Client, error) {
	var sds []statsd.Statter
	var addrs = strings.Split(os.Getenv(CollectorEnv), ",")
	if len(os.Args) > 1 {
		addrs = os.Args[1:]
	}
	for _, a := range addrs {
		s, err := statsd.NewClient(a, "some-name")
		if err != nil {
			return nil, fmt.Errorf("sidecar: create client %w", err)
		}
		sds = append(sds, s)
	}
	pid, err := pid("nfd")
	if err != nil {
		return nil, fmt.Errorf("sidecar: get nfd pid %w", err)
	}
	return &Client{
		scapeInterval: 15 * time.Second,
		sampleRate:    1,
		sds:           sds,
		pagesize:      os.Getpagesize(),
		nfdpid:        pid,
	}, nil
}
func (c *Client) close() {

}

func (c *Client) run(ctx context.Context) chan struct{} {
	t := time.NewTicker(c.scapeInterval)
	done := make(chan struct{})
	go func() {
		<-ctx.Done()
		t.Stop()
		done <- struct{}{}
	}()
	go func() {
		for range t.C {
			go func() {
				ctx, cancel := context.WithTimeout(ctx, c.scapeInterval)
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
		_, _, _ = st, status, mem
		// TODO: memory
		// TODO: routes (or use NLSR)
		// st.Gauge("", , c.sampleRate)
		// status.GeneralStatus.NCsEntries
		// status.GeneralStatus.NFibEntries
		// status.GeneralStatus.NPitEntries
		// status.GeneralStatus.NMeasurementsEntries
		// status.GeneralStatus.NCsEntries
		// status.GeneralStatus.NSatisfiedInterests
		// status.GeneralStatus.NUnsatisfiedInterests
		// status.GeneralStatus.PacketCounters.IncomingPackets.NInterests
		// status.GeneralStatus.PacketCounters.IncomingPackets.NData
		// status.GeneralStatus.PacketCounters.IncomingPackets.NNacks
		// status.GeneralStatus.PacketCounters.OutgoingPackets.NInterests
		// status.GeneralStatus.PacketCounters.OutgoingPackets.NData
		// status.GeneralStatus.PacketCounters.OutgoingPackets.NNacks
		// int64(len(status.Channels.Channel))
		// int64(len(status.Faces.Face))
		// for _, face := range status.Faces.Face {
		// 	face.PacketCounters.IncomingPackets.NInterests
		// 	face.PacketCounters.IncomingPackets.NData
		// 	face.PacketCounters.IncomingPackets.NNacks
		// 	face.PacketCounters.OutgoingPackets.NInterests
		// 	face.PacketCounters.OutgoingPackets.NData
		// 	face.PacketCounters.OutgoingPackets.NNacks
		// 	face.ByteCounters.IncomingBytes
		// 	face.ByteCounters.OutgoingBytes
		// }
		// int64(len(status.Fib.FibEntry))
		// int64(len(status.Rib.RibEntry))
		// status.Cs.Capacity
		// status.Cs.NEntries
		// status.Cs.NHits
		// status.Cs.NMisses
	}
}
