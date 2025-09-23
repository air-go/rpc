package connector

import (
	"context"
	"net"

	"github.com/air-go/rpc/library/connpool"
)

type ConnectorStrategy string

const (
	ConnectorStrategyPool ConnectorStrategy = "pool"
)

type Options struct {
	ConnectFunc func(ctx context.Context, addr net.Addr) (net.Conn, error)
	ConnPool    connpool.ConnPool
}

type OptionFunc func(*Options)

func WithConnectFunc(fn func(ctx context.Context, addr net.Addr) (net.Conn, error)) OptionFunc {
	return func(o *Options) {
		o.ConnectFunc = fn
	}
}

func WithConnPool(p connpool.ConnPool) OptionFunc {
	return func(o *Options) {
		o.ConnPool = p
	}
}

type Connector interface {
	// Connect is connect to the address
	Connect(ctx context.Context, addr net.Addr) (net.Conn, error)
	// Back is return the address to the connector
	Back(context.Context, net.Conn) error
}
