package watcher

import (
	"github.com/seankhliao/uva-rp1/api"
)

func (w *Watcher) notify() {
	select {
	case w.notififcation <- struct{}{}:
	default:
		// don't block if there's already a pending notififcation
	}
}

func (w *Watcher) notifier() {
	for range w.notififcation {
		lp := w.localPrimaries()
		r := <-w.reflectors
		rc := len(r)
		for _, v := range r {
			go func(v reflector) {
				err := v.s.Send(lp)
				if err != nil {
					w.log.Error().
						Err(err).
						Str("id", v.a.WatcherId).
						Msg("notifier send gossip")
				}
			}(v)
		}
		w.reflectors <- r

		ap := w.allPrimaries()
		p := <-w.primaries
		pc := len(p)
		for _, v := range p {
			go func(v primary) {
				err := v.s.Send(ap)
				if err != nil {
					w.log.Error().
						Err(err).
						Str("id", v.p.PrimaryId).
						Msg("notifier send primary")
				}
			}(v)
		}
		w.primaries <- p

		w.log.Info().
			Int("reflectors", rc).
			Int("primaries", pc).
			Int("locprimaries", len(lp.Primaries)).
			Int("allprimaries", len(ap.Primaries)).
			Msg("notified")
	}
}

func (w *Watcher) localPrimaries() *api.AllPrimaries {
	ap := &api.AllPrimaries{
		WatcherId: w.localAddr,
	}
	p := <-w.primaries
	for _, v := range p {
		v := v
		ap.Primaries = append(ap.Primaries, v.p)
	}
	w.primaries <- p
	return ap
}

func (w *Watcher) allPrimaries() *api.AllPrimaries {
	ap := &api.AllPrimaries{
		WatcherId: w.localAddr,
	}
	p := <-w.primaries
	for _, v := range p {
		v := v
		ap.Primaries = append(ap.Primaries, v.p)
	}
	w.primaries <- p
	r := <-w.reflectors
	for _, v := range r {
		v := v
		ap.Primaries = append(ap.Primaries, v.a.Primaries...)
	}
	w.reflectors <- r
	return ap
}
