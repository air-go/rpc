package grpc

import (
	"context"
	"fmt"

	"google.golang.org/grpc"

	serverGRPC "github.com/air-go/rpc/server/grpc"
)

func Conn(ctx context.Context, serviceName string) (cc *grpc.ClientConn, err error) {
	if cc, err = grpc.Dial(
		fmt.Sprintf("%s:///%s", scheme, serviceName),
		serverGRPC.NewDialOption(serverGRPC.DialOptionResolver(NewRegistryBuilder(serviceName)))...,
	); err != nil {
		return
	}

	return
}
