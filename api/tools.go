// +build tools

package api

import (
	_ "github.com/golang/protobuf/protoc-gen-go"
)

//go:generate protoc --go_out=plugins=grpc:. primary.proto
