package primary

import (
	"context"
	"github.com/seankhliao/uva-rp1/api"
)

func (p *Primary) RegisterSecondary(s api.Control_RegisterSecondaryServer) error {
	panic("Unimplemented: RegisterSecondary")
}

func (p *Primary) UnregisterSecondary(ctx context.Context, r *api.UnregisterRequest) (*api.UnregisterResponse, error) {
	panic("Unimplemented: UnregisterSecondary")
}
