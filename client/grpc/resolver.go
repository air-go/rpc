package grpc

import (
	"context"

	"google.golang.org/grpc/resolver"

	"github.com/air-go/rpc/library/servicer"
)

type registryResolver struct {
	serviceName string
	target      resolver.Target
	cc          resolver.ClientConn
}

func (r *registryResolver) start() {
	r.ResolveNow(resolver.ResolveNowOptions{})
}

func (r *registryResolver) ResolveNow(o resolver.ResolveNowOptions) {
	srv, has := servicer.GetServicer(r.serviceName)
	if !has {
		return
	}

	nodes, err := srv.All(context.Background())
	if err != nil {
		return
	}

	address := make([]resolver.Address, len(nodes))
	for i, node := range nodes {
		address[i] = resolver.Address{Addr: node.Address()}
	}
	r.cc.UpdateState(resolver.State{Addresses: address})
}

func (r *registryResolver) Close() {}

func init() {
	resolver.Register(&registryBuilder{})
}
