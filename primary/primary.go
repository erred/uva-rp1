package primary

import (
	"context"
	// "errors"
	// "flag"
	// "fmt"
	// "io"
	// "math/rand"
	// "net/url"
	// "strconv"
	// "strings"
	// "time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	// "github.com/seankhliao/uva-rp1/api"
	// "google.golang.org/grpc"
)

type Primary struct {
	name string

	log *zerolog.Logger
}

func New(args []string, logger *zerolog.Logger) *Primary {
	if logger == nil {
		*logger = log.With().Str("mod", "primary").Logger()
	}

	// p := &Primary{}
	panic("Unimplemented: New")
	// return p
}

func (p *Primary) Run(ctx context.Context) error {

	// httpServer := http.ServeMux{}
	// httpServer.Handle("/metrics", promhttp.Handler())
	//
	// grpcServer := grpc.NewServer()
	// // api.RegisterDiscoveryServiceServer(grpcServer, s)
	//
	// http.ListenAndServe(fmt.Sprintf(":%d", s.port), http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
	// 		grpcServer.ServeHTTP(w, r)
	// 	} else {
	// 		httpServer.ServeHTTP(w, r)
	// 	}
	// }))
	panic("Unimplemented: Run")
}
