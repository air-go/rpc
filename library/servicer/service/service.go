package service

import (
	"context"
	"errors"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/why444216978/go-util/assert"
	"github.com/why444216978/go-util/validate"

	"github.com/air-go/rpc/library/registry"
	"github.com/air-go/rpc/library/selector"
	"github.com/air-go/rpc/library/servicer"
)

type Config struct {
	ServiceName  string `validate:"required"`
	RegistryName string
	Type         uint8  `validate:"required,oneof=1 2"`
	Host         string `validate:"required"`
	Port         int    `validate:"required"`
	Selector     string `validate:"required,oneof=wr"` // TODO support others
	CaCrt        string
	ClientPem    string
	ClientKey    string
}

type Service struct {
	sync.RWMutex
	selector   selector.Selector
	adjusting  int32
	updateTime time.Time
	discovery  registry.Discovery
	caCrt      []byte
	clientPem  []byte
	clientKey  []byte
	config     *Config
}

type Option func(*Service)

func WithDiscovery(discovery registry.Discovery) Option {
	return func(s *Service) { s.discovery = discovery }
}

func WithSelector(selector selector.Selector) Option {
	return func(s *Service) { s.selector = selector }
}

func NewService(config *Config, opts ...Option) (*Service, error) {
	s := &Service{
		adjusting: 0,
		config:    config,
		caCrt:     []byte(config.CaCrt),
		clientPem: []byte(config.ClientPem),
		clientKey: []byte(config.ClientKey),
	}

	for _, o := range opts {
		o(s)
	}

	if err := validate.ValidateCamel(config); err != nil {
		return nil, err
	}

	if !assert.IsNil(s.discovery) && s.config.RegistryName == "" {
		return nil, errors.New("RegistryName empty")
	}

	if err := s.initSelector(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *Service) Name() string {
	return s.config.ServiceName
}

func (s *Service) RegistryName() string {
	return s.config.RegistryName
}

func (s *Service) Pick(ctx context.Context) (node servicer.Node, err error) {
	switch s.config.Type {
	case servicer.TypeIPPort:
		node = servicer.NewNode(s.config.Host, s.config.Port)
		return
	case servicer.TypeDomain:
		var host *net.IPAddr
		host, err = net.ResolveIPAddr("ip", s.config.Host)
		if err != nil {
			return
		}
		node = servicer.NewNode(host.IP.String(), s.config.Port)
		return
	case servicer.TypeRegistry:
		s.adjustSelectorNode()
		return s.selector.Select()
	}

	return nil, errors.New("config type not support")
}

func (s *Service) initSelector() (err error) {
	if s.config.Type != servicer.TypeRegistry {
		return nil
	}

	if assert.IsNil(s.discovery) {
		return errors.New("discovery is nil")
	}

	if assert.IsNil(s.selector) {
		return errors.New("selector is nil")
	}

	s.adjustSelectorNode()

	return nil
}

func (s *Service) adjustSelectorNode() {
	if s.discovery.GetUpdateTime().Before(s.updateTime) {
		return
	}

	if !atomic.CompareAndSwapInt32(&s.adjusting, 0, 1) {
		return
	}

	s.Lock()
	defer s.Unlock()

	var (
		address     string
		nowNodes    = s.discovery.GetNodes()
		nowMap      = make(map[string]struct{})
		selectorMap = make(map[string]servicer.Node)
	)

	// selector add new nodes
	for _, node := range nowNodes {
		address = node.Address()
		nowMap[address] = struct{}{}
		selectorMap[address] = node

		_ = s.selector.AddNode(node)
	}

	// selector delete non-existent nodes
	selectorNodes, _ := s.selector.GetNodes()
	for _, n := range selectorNodes {
		if _, ok := nowMap[n.Address()]; ok {
			continue
		}
		_ = s.selector.DeleteNode(n)
	}

	s.updateTime = time.Now()
	atomic.StoreInt32(&s.adjusting, 0)
}

func (s *Service) Done(ctx context.Context, node servicer.Node, err error) error {
	if assert.IsNil(s.selector) {
		return errors.New("selector is nil")
	}
	s.selector.AfterHandle(selector.HandleInfo{Node: node, Err: err})
	return nil
}

func (s *Service) GetCaCrt() []byte {
	return s.caCrt
}

func (s *Service) GetClientPem() []byte {
	return s.clientPem
}

func (s *Service) GetClientKey() []byte {
	return s.clientKey
}
