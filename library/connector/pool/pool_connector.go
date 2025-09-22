package pool

import (
	"context"
	"net"

	"github.com/air-go/rpc/library/connector"
)

func defaultOptions() *connector.Options {
	return &connector.Options{
		ConnectFunc: func(ctx context.Context, addr net.Addr) (net.Conn, error) {
			return (&net.Dialer{}).DialContext(ctx, addr.Network(), addr.String())
		},
	}
}

type poolConnector struct {
	options     *connector.Options
	serviceName string
}

func NewConnector(serviceName string, opts ...connector.OptionFunc) (*poolConnector, error) {
	opt := defaultOptions()
	for _, o := range opts {
		o(opt)
	}

	return &poolConnector{
		options:     opt,
		serviceName: serviceName,
	}, nil
}

func (s *poolConnector) Connect(ctx context.Context, addr net.Addr) (net.Conn, error) {
	if s.options.ConnPool != nil {
		return s.options.ConnPool.Get(ctx, addr)
	}

	return s.options.ConnectFunc(ctx, addr)
}

func (s *poolConnector) Back(ctx context.Context, conn net.Conn) error {
	if s.options.ConnPool != nil {
		return s.options.ConnPool.Back(ctx, conn.RemoteAddr())
	}
	return conn.Close()
}
