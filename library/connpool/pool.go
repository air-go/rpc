package connpool

import (
	"context"
	"net"
)

type ConnPool interface {
	Get(context.Context, net.Addr) (net.Conn, error)
	Back(context.Context, net.Addr) error
}
