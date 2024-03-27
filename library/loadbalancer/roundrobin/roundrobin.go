package roundrobin

import (
	"context"
	"net"
	"sync"

	"github.com/air-go/rpc/library/loadbalancer"
)

type RoundRobin struct {
	addrs []net.Addr
	index int
	lock  sync.Mutex
}

func New() *RoundRobin {
	return &RoundRobin{
		index: -1,
	}
}

func (rr *RoundRobin) Strategy() string {
	return "RoundRobin"
}

func (rr *RoundRobin) SetAddrs(addrs []net.Addr) error {
	rr.lock.Lock()
	defer rr.lock.Unlock()

	rr.addrs = addrs
	return nil
}

func (rr *RoundRobin) Pick(context.Context) (net.Addr, error) {
	rr.lock.Lock()
	defer rr.lock.Unlock()

	if len(rr.addrs) == 0 {
		return nil, loadbalancer.ErrAddrsEmpty
	}

	rr.index = (rr.index + 1) % len(rr.addrs)
	return rr.addrs[rr.index], nil
}

func (rr *RoundRobin) Back(net.Addr, error) {}
