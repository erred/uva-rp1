// +build tools

package management

import (
	_ "github.com/golang/protobuf/protoc-gen-go"
)

//go:generate protoc --go_out=plugins=grpc:. ./api/api.proto
