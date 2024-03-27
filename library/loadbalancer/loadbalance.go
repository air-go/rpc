package loadbalancer

import (
	"context"
	"net"
)

type LoadBalancer interface {
	Strategy() string
	SetAddrs([]net.Addr) error
	Pick(context.Context) (net.Addr, error)
	Back(net.Addr, error)
}
