package factory

import (
	"github.com/air-go/rpc/library/connector"
	"github.com/air-go/rpc/library/connector/pool"
)

func New(strategy connector.ConnectorStrategy, serviceName string, opts ...connector.OptionFunc) (connector.Connector, error) {
	switch strategy {
	case connector.ConnectorStrategyPool:
		return pool.NewConnector(serviceName, opts...)
	default:
		return pool.NewConnector(serviceName, opts...)
	}
}
