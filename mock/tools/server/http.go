package server

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/air-go/rpc/mock/tools/listener"
)

type httpServer struct {
	*http.Server
	listener *listener.Listener
}

func NewHTTP(handler func(server *gin.Engine)) (*httpServer, error) {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}

	s := &httpServer{
		Server: &http.Server{},
		listener: &listener.Listener{
			Listener: l,
		},
	}

	h := gin.New()
	handler(h)
	s.Server.Handler = h

	return s, nil
}

func (s *httpServer) Addr() string {
	return s.listener.Addr().String()
}

func (s *httpServer) Start() error {
	return s.Serve(s.listener)
}

func (s *httpServer) Stop() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	_ = cancel
	return s.Shutdown(ctx)
}
