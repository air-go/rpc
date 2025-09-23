// wr is Weighted random
package wr

import (
	"context"
	"math/rand"
	"sort"
	"sync"

	"github.com/pkg/errors"

	"github.com/air-go/rpc/library/loadbalancer"
	"github.com/air-go/rpc/library/servicer"
)

type nodeOffset struct {
	Address     string
	Weight      int
	OffsetStart int
	OffsetEnd   int
}

type WeightedRandom struct {
	options     *loadbalancer.LoadBalancerOptions
	lock        sync.RWMutex
	nodeCount   int
	nodes       map[string]servicer.Node
	list        []servicer.Node
	offsetList  []nodeOffset
	sameWeight  bool
	totalWeight int
}

var _ loadbalancer.LoadBalancer = (*WeightedRandom)(nil)

func New(opts ...loadbalancer.LoadBalancerOptionFunc) *WeightedRandom {
	s := &WeightedRandom{
		options:    &loadbalancer.LoadBalancerOptions{},
		nodes:      make(map[string]servicer.Node),
		list:       make([]servicer.Node, 0),
		offsetList: make([]nodeOffset, 0),
	}

	for _, o := range opts {
		o(s.options)
	}

	return s
}

func (wr *WeightedRandom) Strategy() loadbalancer.LoadBalancerStrategy {
	return loadbalancer.LoadBalancerStrategyWeightedRandom
}

func (wr *WeightedRandom) SetNodes(nodes []servicer.Node) error {
	wr.lock.Lock()
	defer wr.lock.Unlock()

	s2n := map[string]servicer.Node{}
	for _, n := range nodes {
		wr.addNode(n)
		s2n[n.Addr().String()] = n
	}

	for s, n := range wr.nodes {
		if _, ok := s2n[s]; !ok {
			wr.deleteNode(n)
		}
	}

	return nil
}

func (wr *WeightedRandom) addNode(node servicer.Node) {
	address := node.Addr().String()
	if _, ok := wr.nodes[address]; ok {
		return
	}

	var (
		weight      = node.Weight()
		offsetStart = 0
		offsetEnd   = wr.totalWeight + weight
	)
	if wr.nodeCount > 0 {
		offsetStart = wr.totalWeight + 1
	}

	offset := nodeOffset{
		Address:     address,
		Weight:      weight,
		OffsetStart: offsetStart,
		OffsetEnd:   offsetEnd,
	}

	wr.totalWeight = offsetEnd
	wr.nodes[address] = node
	wr.list = append(wr.list, node)
	wr.offsetList = append(wr.offsetList, offset)
	wr.nodeCount = wr.nodeCount + 1

	wr.sortOffset()
	wr.checkSameWeight()
}

func (wr *WeightedRandom) deleteNode(node servicer.Node) {
	address := node.Addr().String()
	node, ok := wr.nodes[address]
	if !ok {
		return
	}

	wr.nodeCount = wr.nodeCount - 1

	delete(wr.nodes, address)

	for idx, n := range wr.list {
		if n.Addr().String() != address {
			continue
		}
		new := append(wr.list[:idx], wr.list[idx+1:]...)
		wr.list = new
	}

	for idx, n := range wr.offsetList {
		if n.Address != address {
			continue
		}
		wr.totalWeight = wr.totalWeight - node.Weight()
		new := append(wr.offsetList[:idx], wr.offsetList[idx+1:]...)
		wr.offsetList = new
	}

	wr.sortOffset()
	wr.checkSameWeight()
}

func (wr *WeightedRandom) Pick(ctx context.Context) (node servicer.Node, err error) {
	wr.lock.RLock()
	defer wr.lock.RUnlock()

	defer func() {
		if node != nil {
			return
		}
		err = errors.New("node is nil")
	}()

	if wr.sameWeight {
		idx := rand.Intn(wr.nodeCount)
		node = wr.list[idx]
		return
	}

	idx := rand.Intn(wr.totalWeight + 1)
	for _, n := range wr.offsetList {
		if idx >= n.OffsetStart && idx <= n.OffsetEnd {
			node = wr.nodes[n.Address]
			break
		}
	}

	return
}

func (wr *WeightedRandom) Back(n servicer.Node, err error) {
	wr.lock.Lock()
	defer wr.lock.Unlock()

	node := wr.nodes[n.Addr().String()]
	if node == nil {
		return
	}

	if err != nil {
		node.IncrFail()
		return
	}
	node.IncrSuccess()
}

func (wr *WeightedRandom) GetNodes() []servicer.Node {
	wr.lock.Lock()
	defer wr.lock.Unlock()

	nodes := make([]servicer.Node, len(wr.list))
	copy(nodes, wr.list)

	return nodes
}

func (wr *WeightedRandom) checkSameWeight() {
	wr.sameWeight = true

	var last int
	for _, n := range wr.list {
		cur := int(n.Weight())
		if last == 0 {
			last = cur
			continue
		}
		if last == cur {
			last = cur
			continue
		}
		wr.sameWeight = false
		return
	}
}

func (wr *WeightedRandom) sortOffset() {
	sort.Slice(wr.offsetList, func(i, j int) bool {
		return wr.offsetList[i].Weight > wr.offsetList[j].Weight
	})
}
