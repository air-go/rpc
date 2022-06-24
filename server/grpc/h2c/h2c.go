package h2c

import (
	"context"
	"net/http"
	"strings"
	"time"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
	"google.golang.org/grpc"

	"github.com/air-go/rpc/library/logger"
	"github.com/air-go/rpc/server"
	serverGRPC "github.com/air-go/rpc/server/grpc"
)

type Option struct {
	logger      logger.Logger
	httpHandler http.Handler
}

type OptionFunc func(*Option)

func WithLogger(l logger.Logger) OptionFunc {
	return func(s *Option) { s.logger = l }
}

func WithHTTPHandler(h http.Handler) OptionFunc {
	return func(s *Option) { s.httpHandler = h }
}

func defaultOption() *Option {
	return &Option{httpHandler: http.NotFoundHandler()}
}

type H2CServer struct {
	*Option
	*grpc.Server
	ctx           context.Context
	endpoint      string
	grpcRegisters []serverGRPC.RegisterGRPC
	httpServer    *http.Server
}

var _ server.Server = (*H2CServer)(nil)

func NewH2C(ctx context.Context, endpoint string, grpcRegisters []serverGRPC.RegisterGRPC, opts ...OptionFunc) *H2CServer {
	if len(grpcRegisters) < 1 {
		panic("len(grpcRegisters) < 1")
	}

	option := defaultOption()
	for _, o := range opts {
		o(option)
	}

	s := &H2CServer{
		Option:        option,
		ctx:           ctx,
		endpoint:      endpoint,
		grpcRegisters: grpcRegisters,
	}

	return s
}

func (s *H2CServer) Start() (err error) {
	grpcServer := grpc.NewServer(serverGRPC.NewServerOption(serverGRPC.ServerOptionLogger(s.logger))...)

	// register grpc server
	for _, r := range s.grpcRegisters {
		r(grpcServer)
	}

	serverGRPC.RegisterTools(grpcServer)

	s.Server = grpcServer

	s.httpServer = &http.Server{
		Addr: s.endpoint,
		Handler: h2c.NewHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.ProtoMajor == 2 && strings.Contains(r.Header.Get("Content-Type"), "application/grpc") {
				grpcServer.ServeHTTP(w, r)
			} else {
				s.httpHandler.ServeHTTP(w, r)
			}
		}), &http2.Server{}),
	}

	return s.httpServer.ListenAndServe()
}

func (s *H2CServer) Close() (err error) {
	s.GracefulStop()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	return s.httpServer.Shutdown(ctx)
}
