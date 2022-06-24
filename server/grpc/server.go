package grpc

import (
	"context"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/reflection"
)

type (
	RegisterGRPC func(s *grpc.Server)
	RegisterMux  func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) error
)

// RegisterGateway generate grpc-gateway mux http.Handler
func RegisterGateway(ctx context.Context, endpoint string, registers []RegisterMux) (handler http.Handler, err error) {
	mux := http.NewServeMux()
	gateway := runtime.NewServeMux()

	// register http server by grpc-gateway
	for _, r := range registers {
		if err = r(ctx, gateway, endpoint, NewDialOption()); err != nil {
			return
		}
	}

	mux.Handle("/", gateway)
	handler = mux

	return
}

// RegisterTools register common grpc tools
func RegisterTools(s *grpc.Server) {
	reflection.Register(s)
	service.RegisterChannelzServiceToServer(s)
}
