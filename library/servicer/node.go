package servicer

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type Statistics struct {
	Success uint64
	Fail    uint64
}

type Node interface {
	Address() string
	Host() string
	Port() int
	Weight() int
	FloatWeight() float64
	Statistics() Statistics
	IncrSuccess()
	IncrFail()
}

func GenerateAddress(host string, port int) string {
	return fmt.Sprintf("%s:%d", host, port)
}

func ExtractAddress(address string) (string, int) {
	arr := strings.Split(address, ":")
	if len(arr) != 2 {
		return "", 0
	}

	port, _ := strconv.Atoi(arr[1])
	return arr[0], port
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
	address     string
	host        string
	port        int
	weight      int
	floatWeight float64
	statistics  Statistics
}

var _ Node = (*node)(nil)

func Empty() *node {
	return &node{}
}

func NewNode(host string, port int, opts ...Option) *node {
	n := &node{
		address:    GenerateAddress(host, port),
		host:       host,
		port:       port,
		statistics: Statistics{},
	}
	for _, o := range opts {
		o(n)
	}

	return n
}

func (n *node) Address() string {
	return n.address
}

func (n *node) Host() string {
	return n.host
}

func (n *node) Port() int {
	return n.port
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
