package grpc

import (
	"google.golang.org/grpc/resolver"
)

const scheme = "registry"

type registryBuilder struct {
	serviceName string
}

func NewRegistryBuilder(serviceName string) *registryBuilder {
	return &registryBuilder{serviceName: serviceName}
}

func (rb *registryBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	r := &registryResolver{
		serviceName: rb.serviceName,
		target:      target,
		cc:          cc,
	}
	r.start()
	return r, nil
}

func (rb *registryBuilder) Scheme() string {
	return scheme
}
