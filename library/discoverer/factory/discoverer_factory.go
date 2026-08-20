package factory

import (
	"errors"

	"github.com/air-go/rpc/library/discoverer"
	"github.com/air-go/rpc/library/discoverer/dns"
	"github.com/air-go/rpc/library/loadbalancer"
)

func NewDiscoverer(strategy discoverer.DiscovererStrategy, serviceName string,
	lb loadbalancer.LoadBalancer, opts ...discoverer.OptionFunc,
) (discoverer.Discoverer, error) {
	switch strategy {
	case discoverer.DiscovererStrategyDNS:
		return dns.NewDNSDiscoverer(serviceName, lb, opts...)
	default:
		return nil, errors.New("unknown discoverer type")
	}
}
