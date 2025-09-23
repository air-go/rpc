package discoverer

import (
	"context"
	"time"

	"github.com/air-go/rpc/library/addr"
	"github.com/air-go/rpc/library/logger"
)

type DiscovererStrategy string

const (
	DiscovererStrategyDNS  DiscovererStrategy = "dns"
	DiscovererStrategyEtcd DiscovererStrategy = "etcd"
)

// Discovery is service discovery
type Discoverer interface {
	Start(ctx context.Context) error
	Stop() error
}

type OptionFunc func(*Options)

type Options struct {
	Logger        logger.Logger
	RefreshWindow time.Duration
	IDCNodes      []IDCNode
}

type IDCNode struct {
	IDC     string
	Network string
	Host    string
	Port    int
}

func WithIDCNodes(nodes []IDCNode) OptionFunc {
	return func(o *Options) { o.IDCNodes = nodes }
}

func WithLogger(l logger.Logger) OptionFunc {
	return func(o *Options) { o.Logger = l }
}

func WithRefreshWindow(t time.Duration) OptionFunc {
	return func(o *Options) { o.RefreshWindow = t }
}

// GetIDCNodesFunc is a function to get nodes from IDC
type GetIDCNodesFunc func(idc string) ([]addr.WeightedAddr, error)
