package nfdstat

import (
	"context"
	"encoding/xml"
	"fmt"
	"os/exec"
	"strconv"
)

func (s *Server) Status() (int64, *NFDStatus) {
	ctx := context.Background()

	mem, err := memory(s.nfdpid, s.pagesize)
	if err != nil {
		s.log.Error().Err(err).Msg("nfd memory")
	} else {
		s.memory.Set(float64(mem))
	}

	status, err := getStatus(ctx)
	if err != nil {
		s.log.Error().Err(err).Msg("nfd status")
		return 0, nil
	}

	s.cs_capacity.Set(float64(status.Cs.Capacity))
	s.cs_entries.Set(float64(status.Cs.NEntries))
	s.cs_hits.Set(float64(status.Cs.NHits))
	s.cs_misses.Set(float64(status.Cs.NMisses))
	s.nt_entries.Set(float64(status.GeneralStatus.NNameTreeEntries))
	s.fib_entries.Set(float64(status.GeneralStatus.NFibEntries))
	s.rib_entries.Set(float64(len(status.Rib.RibEntry)))
	s.pit_entries.Set(float64(status.GeneralStatus.NPitEntries))
	s.channel_entries.Set(float64(len(status.Channels.Channel)))
	s.face_entries.Set(float64(len(status.Faces.Face)))

	s.interests.WithLabelValues("yes").Set(float64(status.GeneralStatus.NSatisfiedInterests))
	s.interests.WithLabelValues("no").Set(float64(status.GeneralStatus.NUnsatisfiedInterests))

	s.pkts.WithLabelValues("in", "interest").Set(float64(status.GeneralStatus.PacketCounters.IncomingPackets.NInterests))
	s.pkts.WithLabelValues("in", "data").Set(float64(status.GeneralStatus.PacketCounters.IncomingPackets.NData))
	s.pkts.WithLabelValues("in", "nack").Set(float64(status.GeneralStatus.PacketCounters.IncomingPackets.NNacks))
	s.pkts.WithLabelValues("out", "interest").Set(float64(status.GeneralStatus.PacketCounters.OutgoingPackets.NInterests))
	s.pkts.WithLabelValues("out", "data").Set(float64(status.GeneralStatus.PacketCounters.OutgoingPackets.NData))
	s.pkts.WithLabelValues("out", "nack").Set(float64(status.GeneralStatus.PacketCounters.OutgoingPackets.NNacks))

	for _, f := range status.Faces.Face {
		s.face_bytes.WithLabelValues("in", strconv.FormatInt(f.FaceId, 10)).Set(float64(f.ByteCounters.IncomingBytes))
		s.face_bytes.WithLabelValues("out", strconv.FormatInt(f.FaceId, 10)).Set(float64(f.ByteCounters.OutgoingBytes))

		s.face_pkts.WithLabelValues("in", "interest", strconv.FormatInt(f.FaceId, 10)).Set(float64(f.PacketCounters.IncomingPackets.NInterests))
		s.face_pkts.WithLabelValues("in", "data", strconv.FormatInt(f.FaceId, 10)).Set(float64(f.PacketCounters.IncomingPackets.NData))
		s.face_pkts.WithLabelValues("in", "nack", strconv.FormatInt(f.FaceId, 10)).Set(float64(f.PacketCounters.IncomingPackets.NNacks))
		s.face_pkts.WithLabelValues("out", "interest", strconv.FormatInt(f.FaceId, 10)).Set(float64(f.PacketCounters.OutgoingPackets.NInterests))
		s.face_pkts.WithLabelValues("out", "data", strconv.FormatInt(f.FaceId, 10)).Set(float64(f.PacketCounters.OutgoingPackets.NData))
		s.face_pkts.WithLabelValues("out", "nack", strconv.FormatInt(f.FaceId, 10)).Set(float64(f.PacketCounters.OutgoingPackets.NNacks))
	}
	return mem, status
}

func getStatus(ctx context.Context) (*NFDStatus, error) {
	b, err := exec.CommandContext(ctx, "nfdc", "status", "report", "xml").CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("nfdc: %w: %s", err, b)
	}
	var status NFDStatus
	err = xml.Unmarshal(b, &status)
	if err != nil {
		return nil, fmt.Errorf("status unmarshal: %w", err)
	}
	return &status, nil
}

type NFDStatus struct {
	XMLName         xml.Name        `xml:"nfdStatus"`
	Text            string          `xml:",chardata"`
	Xmlns           string          `xml:"xmlns,attr"`
	GeneralStatus   GeneralStatus   `xml:"generalStatus"`
	Channels        Channels        `xml:"channels"`
	Faces           Faces           `xml:"faces"`
	Fib             Fib             `xml:"fib"`
	Rib             Rib             `xml:"rib"`
	Cs              Cs              `xml:"cs"`
	StrategyChoices StrategyChoices `xml:"strategyChoices"`
}

type GeneralStatus struct {
	Text                  string         `xml:",chardata"`
	Version               string         `xml:"version"`
	StartTime             string         `xml:"startTime"`
	CurrentTime           string         `xml:"currentTime"`
	Uptime                string         `xml:"uptime"`
	NNameTreeEntries      int64          `xml:"nNameTreeEntries"`
	NFibEntries           int64          `xml:"nFibEntries"`
	NPitEntries           int64          `xml:"nPitEntries"`
	NMeasurementsEntries  int64          `xml:"nMeasurementsEntries"`
	NCsEntries            int64          `xml:"nCsEntries"`
	PacketCounters        PacketCounters `xml:"packetCounters"`
	NSatisfiedInterests   int64          `xml:"nSatisfiedInterests"`
	NUnsatisfiedInterests int64          `xml:"nUnsatisfiedInterests"`
}
type PacketCounters struct {
	Text            string        `xml:",chardata"`
	IncomingPackets PacketCounter `xml:"incomingPackets"`
	OutgoingPackets PacketCounter `xml:"outgoingPackets"`
}
type PacketCounter struct {
	Text       string `xml:",chardata"`
	NInterests int64  `xml:"nInterests"`
	NData      int64  `xml:"nData"`
	NNacks     int64  `xml:"nNacks"`
}
type Channels struct {
	Text    string    `xml:",chardata"`
	Channel []Channel `xml:"channel"`
}
type Channel struct {
	Text     string `xml:",chardata"`
	LocalUri string `xml:"localUri"`
}
type Faces struct {
	Text string `xml:",chardata"`
	Face []Face `xml:"face"`
}
type Face struct {
	Text             string         `xml:",chardata"`
	FaceId           int64          `xml:"faceId"`
	RemoteUri        string         `xml:"remoteUri"`
	LocalUri         string         `xml:"localUri"`
	FaceScope        string         `xml:"faceScope"`
	FacePersistency  string         `xml:"facePersistency"`
	LinkType         string         `xml:"linkType"`
	Congestion       Congestion     `xml:"congestion"`
	Mtu              int            `xml:"mtu"`
	Flags            FaceFlags      `xml:"flags"`
	PacketCounters   PacketCounters `xml:"packetCounters"`
	ByteCounters     ByteCounters   `xml:"byteCounters"`
	ExpirationPeriod string         `xml:"expirationPeriod"`
}
type Congestion struct {
	Text                string `xml:",chardata"`
	BaseMarkingInterval string `xml:"baseMarkingInterval"`
	DefaultThreshold    int    `xml:"defaultThreshold"`
}
type FaceFlags struct {
	Text                     string `xml:",chardata"`
	LocalFieldsEnabled       string `xml:"localFieldsEnabled"`
	CongestionMarkingEnabled string `xml:"congestionMarkingEnabled"`
}
type ByteCounters struct {
	Text          string `xml:",chardata"`
	IncomingBytes int64  `xml:"incomingBytes"`
	OutgoingBytes int64  `xml:"outgoingBytes"`
}
type Fib struct {
	Text     string     `xml:",chardata"`
	FibEntry []FibEntry `xml:"fibEntry"`
}
type FibEntry struct {
	Text     string   `xml:",chardata"`
	Prefix   string   `xml:"prefix"`
	NextHops NextHops `xml:"nextHops"`
}
type NextHops struct {
	Text    string  `xml:",chardata"`
	NextHop NextHop `xml:"nextHop"`
}
type NextHop struct {
	Text   string `xml:",chardata"`
	FaceId int64  `xml:"faceId"`
	Cost   int64  `xml:"cost"`
}
type Rib struct {
	Text     string     `xml:",chardata"`
	RibEntry []RibEntry `xml:"ribEntry"`
}
type RibEntry struct {
	Text   string `xml:",chardata"`
	Prefix string `xml:"prefix"`
	Routes Routes `xml:"routes"`
}
type Routes struct {
	Text  string `xml:",chardata"`
	Route Route  `xml:"route"`
}
type Route struct {
	Text   string     `xml:",chardata"`
	FaceId int64      `xml:"faceId"`
	Origin string     `xml:"origin"`
	Cost   int64      `xml:"cost"`
	Flags  RouteFlags `xml:"flags"`
}
type RouteFlags struct {
	Text         string `xml:",chardata"`
	ChildInherit string `xml:"childInherit"`
}
type Cs struct {
	Text         string `xml:",chardata"`
	Capacity     int64  `xml:"capacity"`
	AdmitEnabled string `xml:"admitEnabled"`
	ServeEnabled string `xml:"serveEnabled"`
	NEntries     int64  `xml:"nEntries"`
	NHits        int64  `xml:"nHits"`
	NMisses      int64  `xml:"nMisses"`
}
type StrategyChoices struct {
	Text           string           `xml:",chardata"`
	StrategyChoice []StrategyChoice `xml:"strategyChoice"`
}
type StrategyChoice struct {
	Text      string   `xml:",chardata"`
	Namespace string   `xml:"namespace"`
	Strategy  Strategy `xml:"strategy"`
}
type Strategy struct {
	Text string `xml:",chardata"`
	Name string `xml:"name"`
}
