// Package server implements different protocol for accessing the cache in a different process over network.
package server

import (
	"context"
)

type Server interface {
	Start(ctx context.Context, addr string, port int) error
	Stop(ctx context.Context) error
}
