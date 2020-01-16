package nfdstat

import (
	"context"
	"fmt"
	"os/exec"
	"strconv"
)

func AddFace(ctx context.Context, uri string) error {
	args := []string{"face", "create", "remote", uri, "persistency", "persistent"}
	b, err := exec.CommandContext(ctx, "nfdc", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("addFace: %s: %w", b, err)
	}
	return nil
}
func DelFace(ctx context.Context, uri string) error {
	args := []string{"face", "destroy", uri}
	b, err := exec.CommandContext(ctx, "nfdc", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("delFace: %s: %w", b, err)
	}
	return nil
}

func AddRoute(ctx context.Context, prefix, uri string, cost int64) error {
	args := []string{"route", "add", prefix, "nexthop", uri, "origin", "secondary"}
	if cost > 0 {
		args = append(args, "cost", strconv.FormatInt(cost, 10))
	}
	b, err := exec.CommandContext(ctx, "nfdc", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("addRoute: %s: %w", b, err)
	}
	return nil
}

func DelRoute(ctx context.Context, prefix, uri string) error {
	args := []string{"route", "remove", prefix, "nexthop", uri}
	b, err := exec.CommandContext(ctx, "nfdc", args...).CombinedOutput()
	if err != nil {
		return fmt.Errorf("delRoute: %s: %w", b, err)
	}
	return nil
}
