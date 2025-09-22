package loadbalancer

import (
	"context"

	"github.com/air-go/rpc/library/servicer"
)

type LoadBalancerStrategy string

const (
	LoadBalancerStrategyRoundRobin     LoadBalancerStrategy = "RoundRobin"
	LoadBalancerStrategyWeightedRandom LoadBalancerStrategy = "WeightedRandom"
)

type LoadBalancerOptions struct{}

type LoadBalancerOptionFunc func(*LoadBalancerOptions)

type LoadBalancer interface {
	Strategy() LoadBalancerStrategy
	SetNodes([]servicer.Node) error
	GetNodes() []servicer.Node
	Pick(context.Context) (servicer.Node, error)
	Back(servicer.Node, error)
}
