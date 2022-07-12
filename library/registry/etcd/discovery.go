package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/air-go/rpc/library/registry"
	"github.com/air-go/rpc/library/servicer"

	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
)

type DiscoveryOption struct {
	refreshDuration time.Duration
	cmdTimeout      time.Duration
	decode          registry.Decode
}

type DiscoveryOptionFunc func(*DiscoveryOption)

func defaultDiscoveryOption() *DiscoveryOption {
	return &DiscoveryOption{
		refreshDuration: time.Second * 10,
		cmdTimeout:      time.Second * 3,
		decode:          JSONDecode,
	}
}

// EtcdDiscovery 服务发现
type EtcdDiscovery struct {
	opts        *DiscoveryOption
	cli         *clientv3.Client
	nodeList    map[string]*registry.Node
	lock        sync.RWMutex
	updateTime  time.Time
	ticker      *time.Ticker
	serviceName string
}

var _ registry.Discovery = (*EtcdDiscovery)(nil)

func WithRefreshDuration(d int) DiscoveryOptionFunc {
	return func(o *DiscoveryOption) { o.refreshDuration = time.Duration(d) * time.Second }
}

func WithCmdTimeOut(t time.Duration) DiscoveryOptionFunc {
	return func(o *DiscoveryOption) { o.cmdTimeout = t }
}

// NewDiscovery
func NewDiscovery(cli *clientv3.Client, name string, opts ...DiscoveryOptionFunc) (registry.Discovery, error) {
	if cli == nil {
		return nil, errors.New("cli is nil")
	}

	if name = strings.TrimSpace(name); name == "" {
		return nil, errors.New("serviceName is nil")
	}

	opt := defaultDiscoveryOption()
	for _, o := range opts {
		o(opt)
	}

	ed := &EtcdDiscovery{
		opts:        opt,
		cli:         cli,
		nodeList:    make(map[string]*registry.Node),
		serviceName: name,
	}

	if err := ed.init(); err != nil {
		return nil, err
	}

	return ed, nil
}

// GetNodes
func (s *EtcdDiscovery) GetNodes() []servicer.Node {
	s.lock.RLock()
	defer s.lock.RUnlock()
	nodes := make([]servicer.Node, 0)

	for _, node := range s.nodeList {
		nodes = append(nodes, servicer.NewNode(node.Host, node.Port, servicer.WithWeight(node.Weight)))
	}
	return nodes
}

// GetUpdateTime
func (s *EtcdDiscovery) GetUpdateTime() time.Time {
	return s.updateTime
}

// Close
func (s *EtcdDiscovery) Close() error {
	if s.ticker != nil {
		s.ticker.Stop()
	}
	if s.cli != nil {
		s.cli.Close()
	}
	return nil
}

// WatchService
func (s *EtcdDiscovery) init() error {
	// set all nodes
	s.setNodes()

	// start etcd watcher
	go s.watcher()

	// start refresh ticker
	go s.refresh()

	return nil
}

// loadKVs
func (s *EtcdDiscovery) loadKVs() (kvs []*mvccpb.KeyValue) {
	ctx, cancel := s.context()
	defer cancel()

	resp, err := s.cli.Get(ctx, s.serviceName, clientv3.WithPrefix())
	if err != nil {
		s.logErr("get by prefix", s.serviceName, "", err)
		return
	}
	kvs = resp.Kvs
	return
}

// watcher
func (s *EtcdDiscovery) watcher() {
	ctx, cancel := s.context()
	defer cancel()

	rch := s.cli.Watch(ctx, s.serviceName, clientv3.WithPrefix())
	s.log("Watch", "")
	for wresp := range rch {
		for _, ev := range wresp.Events {
			key := string(ev.Kv.Key)
			val := string(ev.Kv.Value)

			switch ev.Type {
			case mvccpb.PUT:
				node, err := s.opts.decode(val)
				if err != nil {
					s.logErr("decode val", key, val, err)
					return
				}
				s.setNode(key, node)
				s.log("mvccpb.PUT", key)
			case mvccpb.DELETE:
				s.delNode(key)
				s.log("mvccpb.DELETE", key)
			}
		}
	}
}

// refresh
func (s *EtcdDiscovery) refresh() {
	if s.opts.refreshDuration == -1 {
		return
	}

	s.ticker = time.NewTicker(s.opts.refreshDuration)
	for range s.ticker.C {
		s.setNodes()
		s.log("refresh", "all")
	}
}

// setNodes
func (s *EtcdDiscovery) setNodes() {
	nodeList := make(map[string]*registry.Node)
	kvs := s.loadKVs()
	for _, kv := range kvs {
		key := string(kv.Key)
		val := string(kv.Value)

		node, err := s.opts.decode(val)
		if err != nil {
			s.logErr("decode val", key, val, err)
			continue
		}
		nodeList[key] = node
	}

	s.lock.Lock()
	defer s.lock.Unlock()
	s.nodeList = nodeList
	s.updateTime = time.Now()
}

// setNode
func (s *EtcdDiscovery) setNode(key string, node *registry.Node) {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.nodeList[key] = node
	s.updateTime = time.Now()
}

// delNode
func (s *EtcdDiscovery) delNode(key string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	delete(s.nodeList, key)
	s.updateTime = time.Now()
}

func (s *EtcdDiscovery) logErr(action, key, val string, err error) {
	log.Printf("[%s]: action:%s, err:%s, service:%s, key:%s, val:%s\n", time.Now().Format("2006-01-02 15:04:05"), action, s.serviceName, key, val, err.Error())
}

func (s *EtcdDiscovery) log(action, key string) {
	log.Printf("[%s]: [action:%s, service:%s, key:%s]\n", time.Now().Format("2006-01-02 15:04:05"), action, s.serviceName, key)
}

func (s *EtcdDiscovery) context() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), s.opts.cmdTimeout)
}

func JSONDecode(val string) (*registry.Node, error) {
	node := &registry.Node{}
	err := json.Unmarshal([]byte(val), node)
	if err != nil {
		return nil, errors.New("Unmarshal val " + err.Error())
	}

	return node, nil
}
