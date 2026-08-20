package service

import (
	"context"
	"errors"
	"time"

	"github.com/why444216978/go-util/assert"
	"github.com/why444216978/go-util/validate"

	"github.com/air-go/rpc/library/discoverer"
	"github.com/air-go/rpc/library/loadbalancer"
	"github.com/air-go/rpc/library/servicer"
)

type Config struct {
	ServiceName          string                            `validate:"required"`
	Host                 string                            `validate:"required"`
	Port                 int                               `validate:"required"`
	DiscovererStrategy   discoverer.DiscovererStrategy     `validate:"required"`
	LoadBalancerStrategy loadbalancer.LoadBalancerStrategy `validate:"required"`
	IDCNodes             []discoverer.IDCNode
	CaCrt                string
	ClientPem            string
	ClientKey            string
}

type Service struct {
	discovery    discoverer.Discoverer
	loadBalancer loadbalancer.LoadBalancer
	updateTime   time.Time
	caCrt        []byte
	clientPem    []byte
	clientKey    []byte
	config       *Config
}

type Option func(*Service)

var _ servicer.Servicer = (*Service)(nil)

func NewService(config *Config,
	discovery discoverer.Discoverer,
	loadBalancer loadbalancer.LoadBalancer,
	opts ...Option,
) (*Service, error) {
	if err := validate.Validate(config); err != nil {
		return nil, err
	}

	if assert.IsNil(discovery) {
		return nil, errors.New("discovery is nil")
	}

	if assert.IsNil(loadBalancer) {
		return nil, errors.New("loadBalancer is nil")
	}

	s := &Service{
		discovery:    discovery,
		loadBalancer: loadBalancer,
		config:       config,
		caCrt:        []byte(config.CaCrt),
		clientPem:    []byte(config.ClientPem),
		clientKey:    []byte(config.ClientKey),
	}

	for _, o := range opts {
		o(s)
	}

	return s, nil
}

func (s *Service) Start(ctx context.Context) error {
	return s.discovery.Start(ctx)
}

func (s *Service) Name() string {
	return s.config.ServiceName
}

func (s *Service) Pick(ctx context.Context) (node servicer.Node, err error) {
	return s.loadBalancer.Pick(ctx)
}

func (s *Service) GetNodes(ctx context.Context) (node []servicer.Node) {
	return s.loadBalancer.GetNodes()
}

func (s *Service) Done(ctx context.Context, node servicer.Node, err error) error {
	s.loadBalancer.Back(node, err)
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
