package nfdstat

import (
	"context"
	"fmt"
	"github.com/seankhliao/uva-rp1/api"
	"os"
)

type Status struct {
	NFDStatus
	Memory int64
}

type Stat struct {
	ps int
}

func New() *Stat {
	s := &Stat{
		ps: os.Getpagesize(),
	}
	return s
}

func (s *Stat) Status() (*Status, error) {
	p, err := pid("nfd")
	if err != nil {
		return nil, fmt.Errorf("nfdstat pid: %w", err)
	}
	mem, err := memory(p, s.ps)
	if err != nil {
		return nil, fmt.Errorf("nfdstat mem: %w", err)
	}
	n, err := getStatus(context.Background())
	if err != nil {
		return nil, fmt.Errorf("nfdstat status: %w", err)
	}
	return &Status{*n, mem}, nil

}
func (s Status) ToStatusResponse() *api.StatusResponse {
	var bin, bout int64
	for _, f := range s.NFDStatus.Faces.Face {
		bin += f.ByteCounters.IncomingBytes
		bout += f.ByteCounters.OutgoingBytes
	}
	return &api.StatusResponse{
		Memory: s.Memory,

		CsCapacity: s.NFDStatus.Cs.Capacity,
		CsEntries:  s.NFDStatus.Cs.NEntries,
		CsHits:     s.NFDStatus.Cs.NHits,
		CsMisses:   s.NFDStatus.Cs.NMisses,

		FibEntries:  s.NFDStatus.GeneralStatus.NFibEntries,
		RibEntries:  int64(len(s.NFDStatus.Rib.RibEntry)),
		PitEntries:  s.NFDStatus.GeneralStatus.NPitEntries,
		FaceEntries: int64(len(s.NFDStatus.Faces.Face)),

		IntSatisifed:   s.NFDStatus.GeneralStatus.NSatisfiedInterests,
		IntUnsatisfied: s.NFDStatus.GeneralStatus.NUnsatisfiedInterests,
		PktInInt:       s.NFDStatus.GeneralStatus.PacketCounters.IncomingPackets.NInterests,
		PktInData:      s.NFDStatus.GeneralStatus.PacketCounters.IncomingPackets.NData,
		PktInNack:      s.NFDStatus.GeneralStatus.PacketCounters.IncomingPackets.NNacks,
		PktOutInt:      s.NFDStatus.GeneralStatus.PacketCounters.OutgoingPackets.NInterests,
		PktOutData:     s.NFDStatus.GeneralStatus.PacketCounters.OutgoingPackets.NData,
		PktOutNack:     s.NFDStatus.GeneralStatus.PacketCounters.OutgoingPackets.NNacks,
		BytesIn:        bin,
		BytesOut:       bout,
	}
}
