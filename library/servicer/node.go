package servicer

import (
	"net"
	"sync"
)

type Statistics struct {
	Success uint64
	Fail    uint64
}

type Node interface {
	Addr() net.Addr
	Weight() int
	FloatWeight() float64
	Statistics() Statistics
	IncrSuccess()
	IncrFail()
}

type Option func(*node)

func WithWeight(w int) Option {
	return func(n *node) { n.weight = w }
}

func WithFloatWeight(w float64) Option {
	return func(n *node) { n.floatWeight = w }
}

type node struct {
	lock        sync.RWMutex
	addr        net.Addr
	weight      int
	floatWeight float64
	statistics  Statistics
}

var _ Node = (*node)(nil)

func Empty() *node {
	return &node{}
}

func NewNode(addr net.Addr, opts ...Option) *node {
	n := &node{
		addr:       addr,
		statistics: Statistics{},
	}
	for _, o := range opts {
		o(n)
	}

	return n
}

func (n *node) Addr() net.Addr {
	return n.addr
}

func (n *node) Statistics() Statistics {
	n.lock.RLock()
	defer n.lock.RUnlock()
	return n.statistics
}

func (n *node) Weight() int {
	return n.weight
}

func (n *node) FloatWeight() float64 {
	return n.floatWeight
}

func (n *node) IncrSuccess() {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.statistics.Success = n.statistics.Success + 1
}

func (n *node) IncrFail() {
	n.lock.Lock()
	defer n.lock.Unlock()
	n.statistics.Fail = n.statistics.Fail + 1
}
