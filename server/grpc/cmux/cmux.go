package cmux

import (
	"context"
	"net"
	"net/http"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/soheilhy/cmux"
	"google.golang.org/grpc"

	"github.com/air-go/rpc/library/logger"
	"github.com/air-go/rpc/server"
	serverGRPC "github.com/air-go/rpc/server/grpc"
)

type (
	RegisterHTTP func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error
)

type Option struct {
	logger      logger.Logger
	httpHandler http.Handler
}

type OptionFunc func(*Option)

func WithHTTPHandler(h http.Handler) OptionFunc {
	return func(s *Option) { s.httpHandler = h }
}

func defaultOption() *Option {
	return &Option{httpHandler: http.NotFoundHandler()}
}

type MuxServer struct {
	*Option
	ctx           context.Context
	endpoint      string
	grpcRegisters []serverGRPC.RegisterGRPC
	tcpMux        cmux.CMux
}

var _ server.Server = (*MuxServer)(nil)

func NewMux(ctx context.Context, endpoint string, grpcRegisters []serverGRPC.RegisterGRPC, opts ...OptionFunc) *MuxServer {
	if len(grpcRegisters) < 1 {
		panic("len(registers) < 1")
	}

	option := defaultOption()
	for _, o := range opts {
		o(option)
	}

	s := &MuxServer{
		Option:        option,
		ctx:           ctx,
		grpcRegisters: grpcRegisters,
		endpoint:      endpoint,
	}

	return s
}

func (s *MuxServer) Start() (err error) {
	listener, err := net.Listen("tcp", s.endpoint)
	if err != nil {
		return
	}
	s.tcpMux = cmux.New(listener)

	go s.startGRPC()
	go s.startHTTP()

	return s.tcpMux.Serve()
}

func (s *MuxServer) startGRPC() {
	grpcServer := grpc.NewServer(serverGRPC.NewServerOption(serverGRPC.ServerOptionLogger(s.logger))...)

	for _, r := range s.grpcRegisters {
		r(grpcServer)
	}

	serverGRPC.RegisterTools(grpcServer)

	listener := s.tcpMux.MatchWithWriters(cmux.HTTP2MatchHeaderFieldPrefixSendSettings("content-type", "application/grpc"))
	if err := grpcServer.Serve(listener); err != nil {
		panic(err)
	}
}

func (s *MuxServer) startHTTP() {
	var err error
	defer func() {
		if err != nil {
			panic(err)
		}
	}()

	httpServer := &http.Server{
		Addr:    s.endpoint,
		Handler: s.httpHandler,
	}
	listener := s.tcpMux.Match(cmux.HTTP1Fast())
	if err = httpServer.Serve(listener); err != nil {
		return
	}
}

func (s *MuxServer) Close() (err error) {
	s.tcpMux.Close()
	return
}
