package dash

import (
	"context"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gosuri/uilive"
	"github.com/gosuri/uitable"
	"github.com/rs/zerolog"
	"github.com/seankhliao/uva-rp1/api"
	"google.golang.org/grpc"
)

type primary struct {
	cli  api.Info_PrimaryStatusClient
	stat *api.StatusPrimary
}

type Dash struct {
	interval time.Duration
	watcher  string

	primaries chan map[string]primary

	log *zerolog.Logger
}

func New(args []string, logger *zerolog.Logger) *Dash {
	if logger == nil {
		l := zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, NoColor: true, TimeFormat: time.RFC3339Nano}).With().Timestamp().Logger()
		logger = &l
	}

	d := &Dash{
		log:       logger,
		primaries: make(chan map[string]primary, 1),
	}

	d.primaries <- make(map[string]primary)

	fs := flag.NewFlagSet("dash", flag.ExitOnError)
	fs.DurationVar(&d.interval, "interval", 2*time.Second, "refresh interval")
	fs.StringVar(&d.watcher, "watcher", "", "watcher host:port to get primaries from")
	fs.Parse(args)

	return d
}

func (d *Dash) Run() error {
	d.log.Info().Msg("starting dash")
	conn, err := grpc.Dial(d.watcher, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return fmt.Errorf("run dial %s: %w", d.watcher, err)
	}
	defer conn.Close()
	d.log.Info().Msg("connected to reflector")
	rcli := api.NewReflectorClient(conn)
	cli, err := rcli.Primaries(context.Background(), &api.Primary{
		PrimaryId: "dash-" + strconv.Itoa(rand.Int()),
	})
	if err != nil {
		return fmt.Errorf("run get primaries: %w", err)
	}
	go d.primariesUpdater(cli)

	d.log.Info().Msg("starting uilive")
	ui := uilive.New()
	ui.RefreshInterval = 100 * time.Millisecond
	ui.Start()
	defer ui.Stop()

	rows := draw(ui, d.update(), 0)
	for range time.NewTicker(d.interval).C {
		rows = draw(ui, d.update(), rows)
	}

	return nil
}

var cacheh map[string][]int = make(map[string][]int)
var cachem map[string][]int = make(map[string][]int)

func draw(ui *uilive.Writer, up []*api.StatusPrimary, prevrows int) int {
	tab := uitable.New()
	tab.MaxColWidth = 80
	tab.RightAlign(1)
	tab.RightAlign(2)
	tab.RightAlign(3)
	tab.AddRow("")
	tab.AddRow("NODE", "CACHE", "MEMORY (B)", "IN/OUT (B)", "ROUTES")
	sort.Slice(up, func(i, j int) bool {
		return up[i].Id < up[j].Id
	})
	var rowcnt int
	for p := range up {
		sort.Slice(up[p].Secondaries, func(i, j int) bool {
			return up[p].Secondaries[i].Id < up[p].Secondaries[j].Id
		})
		sort.Slice(up[p].Local.Connected, func(i, j int) bool {
			return up[p].Local.Connected[i] < up[p].Local.Connected[j]
		})
		sort.Slice(up[p].Local.Routes, func(i, j int) bool {
			return up[p].Local.Routes[i] < up[p].Local.Routes[j]
		})

		tab.AddRow(
			up[p].Id,
			fmt.Sprintf("%d / %d", up[p].Local.CsEntries, up[p].Local.CsCapacity),
			strconv.FormatInt(up[p].Local.Memory, 10),
			fmt.Sprintf("%d / %d", up[p].Local.BytesIn, up[p].Local.BytesOut),
			strings.Join(up[p].Local.Routes, ", "),
		)
		rowcnt++
		for s := range up[p].Secondaries {
			sort.Slice(up[p].Secondaries[s].Connected, func(i, j int) bool {
				return up[p].Secondaries[s].Connected[i] < up[p].Secondaries[s].Connected[j]
			})
			sort.Slice(up[p].Secondaries[s].Routes, func(i, j int) bool {
				return up[p].Secondaries[s].Routes[i] < up[p].Secondaries[s].Routes[j]
			})

			prefix := " ├ "
			if s == len(up[p].Secondaries)-1 {
				prefix = " └ "
			}
			tab.AddRow(
				prefix+up[p].Secondaries[s].Id,
				fmt.Sprintf("%d / %d", up[p].Secondaries[s].CsEntries, up[p].Secondaries[s].CsCapacity),
				strconv.FormatInt(up[p].Secondaries[s].Memory, 10),
				fmt.Sprintf("%d / %d", up[p].Secondaries[s].BytesIn, up[p].Secondaries[s].BytesOut),
				strings.Join(up[p].Secondaries[s].Routes, ", "),
			)
			rowcnt++
		}
	}
	for i := rowcnt; i < prevrows; i++ {
		tab.AddRow("", "", "", "", "")
	}

	fmt.Fprint(ui.Newline(), tab)
	return rowcnt
}

func (d *Dash) update() []*api.StatusPrimary {
	var wg sync.WaitGroup

	p := <-d.primaries
	stats := make(chan *api.StatusPrimary, len(p))
	os := make([]*api.StatusPrimary, 0, len(p))

	for pid, prim := range p {
		wg.Add(1)
		go func(pid string, prim primary) {
			defer wg.Done()
			err := prim.cli.Send(&api.StatusRequest{})
			if err != nil {
				d.log.Error().Err(err).Str("id", pid).Msg("update send")
				return
			}
			s, err := prim.cli.Recv()
			if err != nil {
				d.log.Error().Err(err).Str("id", pid).Msg("update recv")
				// d.log.Fatal().Err(err).Str("id", pid).Msg("update recv")
				return
			}
			stats <- s
		}(pid, prim)
	}
	d.primaries <- p

	go func() {
		wg.Wait()
		close(stats)
	}()

	for stat := range stats {
		os = append(os, stat)
	}
	return os
}

func (d *Dash) primariesUpdater(cli api.Reflector_PrimariesClient) {
	for {
		ap, err := cli.Recv()
		if err != nil {
			d.log.Error().Err(err).Msg("primaries updater recv")
			return
		}
		p := <-d.primaries
		for _, pr := range ap.Primaries {
			if _, ok := p[pr.PrimaryId]; ok {
				continue
			}
			go d.addPrimary(pr.PrimaryId, pr.Endpoint)
		}
		d.primaries <- p
	}
}

func (d *Dash) addPrimary(id, ep string) {
	conn, err := grpc.Dial(ep, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		d.log.Error().Err(err).Str("id", id).Msg("addPrimary dial")
		return
	}
	defer conn.Close()
	icli := api.NewInfoClient(conn)
	cli, err := icli.PrimaryStatus(context.Background())
	if err != nil {
		d.log.Error().Err(err).Str("id", id).Msg("addPrimary get status")
		return
	}
	p := <-d.primaries
	p[id] = primary{
		cli: cli,
	}
	d.primaries <- p

	defer func() {
		p := <-d.primaries
		delete(p, id)
		d.primaries <- p
	}()

	<-cli.Context().Done()
}

// type flagslice struct {
// 	s []string
// }
//
// func (f *flagslice) String() string {
// 	return strings.Join(f.s, ",")
// }
// func (f *flagslice) Set(s string) error {
// 	f.s = append(f.s, s)
// 	return nil
// }
