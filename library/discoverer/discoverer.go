package discoverer

import (
	"context"
)

type ServiceConfig struct {
	Name string
	IDC  IDC
}

type IDC struct {
	IDC   string
	Modes []Node
}

type Node struct {
	Host     string
	Port     int
	Priority int
}

// Discovery is service discovery
type Discoverer interface {
	Start(ctx context.Context) error
	Stop() error
}
