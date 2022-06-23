package client

import (
	"context"

	"google.golang.org/grpc"

	serverGRPC "github.com/air-go/rpc/server/grpc"
)

func Conn(ctx context.Context, target string) (cc *grpc.ClientConn, err error) {
	// TODO resolver
	cc, err = grpc.DialContext(ctx, target, serverGRPC.NewDialOption()...)
	if err != nil {
		return
	}

	return
}
