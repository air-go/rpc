package roundrobin

import (
	"context"
	"sync"

	"github.com/air-go/rpc/library/loadbalancer"
	"github.com/air-go/rpc/library/servicer"
)

type RoundRobin struct {
	nodes []servicer.Node
	index int
	lock  sync.Mutex
}

var _ loadbalancer.LoadBalancer = (*RoundRobin)(nil)

func New() *RoundRobin {
	return &RoundRobin{
		index: -1,
	}
}

func (rr *RoundRobin) Strategy() loadbalancer.LoadBalancerStrategy {
	return loadbalancer.LoadBalancerStrategyRoundRobin
}

func (rr *RoundRobin) SetNodes(nodes []servicer.Node) error {
	rr.lock.Lock()
	defer rr.lock.Unlock()

	rr.nodes = nodes
	return nil
}

func (rr *RoundRobin) GetNodes() []servicer.Node {
	rr.lock.Lock()
	defer rr.lock.Unlock()

	nodes := make([]servicer.Node, len(rr.nodes))
	copy(nodes, rr.nodes)

	return nodes
}

func (rr *RoundRobin) Pick(context.Context) (servicer.Node, error) {
	rr.lock.Lock()
	defer rr.lock.Unlock()

	if len(rr.nodes) == 0 {
		return nil, loadbalancer.ErrNodesEmpty
	}

	rr.index = (rr.index + 1) % len(rr.nodes)
	return rr.nodes[rr.index], nil
}

func (rr *RoundRobin) Back(servicer.Node, error) {}
