package factory

import (
	"errors"

	"github.com/air-go/rpc/library/loadbalancer"
	"github.com/air-go/rpc/library/loadbalancer/roundrobin"
	"github.com/air-go/rpc/library/loadbalancer/wr"
)

func NewLoadBalancer(strategy loadbalancer.LoadBalancerStrategy) (loadbalancer.LoadBalancer, error) {
	switch strategy {
	case loadbalancer.LoadBalancerStrategyRoundRobin:
		return roundrobin.New(), nil
	case loadbalancer.LoadBalancerStrategyWeightedRandom:
		return wr.New(), nil
	default:
		return nil, errors.New("unknown loadbalancer type")
	}
}
