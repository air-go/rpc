// wr is Weighted random
package wr

import (
	"errors"
	"math/rand"
	"sort"
	"sync"

	"github.com/air-go/rpc/library/selector"
	"github.com/air-go/rpc/library/servicer"
)

type nodeOffset struct {
	Address     string
	Weight      int
	OffsetStart int
	OffsetEnd   int
}

type Selector struct {
	lock        sync.RWMutex
	nodeCount   int
	nodes       map[string]servicer.Node
	list        []servicer.Node
	offsetList  []nodeOffset
	sameWeight  bool
	totalWeight int
	serviceName string
}

var _ selector.Selector = (*Selector)(nil)

type SelectorOption func(*Selector)

func NewSelector(serviceName string, opts ...SelectorOption) *Selector {
	s := &Selector{
		nodes:       make(map[string]servicer.Node),
		list:        make([]servicer.Node, 0),
		offsetList:  make([]nodeOffset, 0),
		serviceName: serviceName,
	}

	for _, o := range opts {
		o(s)
	}

	return s
}

func (s *Selector) ServiceName() string {
	return s.serviceName
}

func (s *Selector) AddNode(node servicer.Node) (err error) {
	address := node.Address()
	if _, ok := s.nodes[address]; ok {
		return
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	var (
		weight      = node.Weight()
		offsetStart = 0
		offsetEnd   = s.totalWeight + weight
	)
	if s.nodeCount > 0 {
		offsetStart = s.totalWeight + 1
	}

	offset := nodeOffset{
		Address:     address,
		Weight:      weight,
		OffsetStart: offsetStart,
		OffsetEnd:   offsetEnd,
	}

	s.totalWeight = offsetEnd
	s.nodes[node.Address()] = node
	s.list = append(s.list, node)
	s.offsetList = append(s.offsetList, offset)
	s.nodeCount = s.nodeCount + 1

	s.sortOffset()
	s.checkSameWeight()

	return
}

func (s *Selector) DeleteNode(node servicer.Node) (err error) {
	address := node.Address()
	node, ok := s.nodes[address]
	if !ok {
		return
	}

	s.lock.Lock()
	defer s.lock.Unlock()

	s.nodeCount = s.nodeCount - 1

	delete(s.nodes, address)

	for idx, n := range s.list {
		if n.Address() != address {
			continue
		}
		new := make([]servicer.Node, len(s.list)-1)
		new = append(s.list[:idx], s.list[idx+1:]...)
		s.list = new
	}

	for idx, n := range s.offsetList {
		if n.Address != address {
			continue
		}
		s.totalWeight = s.totalWeight - node.Weight()
		new := make([]nodeOffset, len(s.offsetList)-1)
		new = append(s.offsetList[:idx], s.offsetList[idx+1:]...)
		s.offsetList = new
	}

	s.sortOffset()
	s.checkSameWeight()

	return
}

func (s *Selector) GetNodes() (nodes []servicer.Node, err error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.list, nil
}

func (s *Selector) Select() (node servicer.Node, err error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	defer func() {
		if node != nil {
			return
		}
		err = errors.New("node is nil")
	}()

	if s.sameWeight {
		idx := rand.Intn(s.nodeCount)
		node = s.list[idx]
		return
	}

	idx := rand.Intn(s.totalWeight + 1)
	for _, n := range s.offsetList {
		if idx >= n.OffsetStart && idx <= n.OffsetEnd {
			node = s.nodes[n.Address]
			break
		}
	}

	return
}

func (s *Selector) AfterHandle(info selector.HandleInfo) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	node := s.nodes[info.Node.Address()]
	if node == nil {
		return
	}

	if info.Err != nil {
		node.IncrFail()
		return
	}
	node.IncrSuccess()
}

func (s *Selector) checkSameWeight() {
	s.sameWeight = true

	var last int
	for _, n := range s.list {
		cur := int(n.Weight())
		if last == 0 {
			last = cur
			continue
		}
		if last == cur {
			last = cur
			continue
		}
		s.sameWeight = false
		return
	}
}

func (s *Selector) sortOffset() {
	sort.Slice(s.offsetList, func(i, j int) bool {
		return s.offsetList[i].Weight > s.offsetList[j].Weight
	})
}
