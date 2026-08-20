package load

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/air-go/rpc/library/discoverer"
	"github.com/air-go/rpc/library/loadbalancer"
	"github.com/air-go/rpc/library/servicer"
	"github.com/air-go/rpc/library/servicer/service"
)

func TestLoadService(t *testing.T) {
	t.Run("dns with idc nodes", func(t *testing.T) {
		cfg := &service.Config{
			ServiceName:          "it_load_dns_with_idcnodes",
			Host:                 "127.0.0.1",
			Port:                 80,
			DiscovererStrategy:   discoverer.DiscovererStrategyDNS,
			LoadBalancerStrategy: loadbalancer.LoadBalancerStrategyRoundRobin,
			IDCNodes: []discoverer.IDCNode{
				{IDC: "bj", Network: "tcp", Host: "127.0.0.1", Port: 80},
			},
		}

		err := LoadService(cfg)
		require.Nil(t, err)

		s, ok := servicer.GetServicer(cfg.ServiceName)
		require.True(t, ok)
		t.Cleanup(func() { servicer.DelServicer(s) })
		assert.Equal(t, cfg.ServiceName, s.Name())
	})

	t.Run("dns without idc nodes", func(t *testing.T) {
		cfg := &service.Config{
			ServiceName:          "it_load_dns_no_idcnodes",
			Host:                 "127.0.0.1",
			Port:                 80,
			DiscovererStrategy:   discoverer.DiscovererStrategyDNS,
			LoadBalancerStrategy: loadbalancer.LoadBalancerStrategyRoundRobin,
		}

		err := LoadService(cfg)
		assert.EqualError(t, err, "new dns discoverer nodes nil")
	})

	t.Run("unknown loadbalancer", func(t *testing.T) {
		cfg := &service.Config{
			ServiceName:          "it_load_unknown_lb",
			Host:                 "127.0.0.1",
			Port:                 80,
			DiscovererStrategy:   discoverer.DiscovererStrategyDNS,
			LoadBalancerStrategy: "unknown",
		}

		err := LoadService(cfg)
		assert.EqualError(t, err, "unknown loadbalancer type")
	})
}