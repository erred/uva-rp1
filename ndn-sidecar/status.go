package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"os/exec"
)

func status(ctx context.Context) (*NFDStatus, error) {

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
	FaceId           string         `xml:"faceId"`
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
	FaceId int    `xml:"faceId"`
	Cost   int    `xml:"cost"`
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
	FaceId int        `xml:"faceId"`
	Origin string     `xml:"origin"`
	Cost   int        `xml:"cost"`
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
