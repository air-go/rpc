package http

import (
	"context"
	"net/http"
	"runtime"
	"time"

	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"

	"github.com/air-go/rpc/server"
	"github.com/air-go/rpc/server/http/response"
)

type Server struct {
	*http.Server
	middleware []gin.HandlerFunc
	router     RegisterRouter
	pprofTurn  bool
	isDebug    bool
	onShutdown []func()
}

var _ server.Server = (*Server)(nil)

type RegisterRouter func(server *gin.Engine)

type Option func(s *Server)

func WithReadTimeout(timeout time.Duration) Option {
	return func(s *Server) { s.Server.ReadTimeout = timeout }
}

func WithWriteTimeout(timeout time.Duration) Option {
	return func(s *Server) { s.Server.WriteTimeout = timeout }
}

func WithMiddleware(middleware ...gin.HandlerFunc) Option {
	return func(s *Server) { s.middleware = middleware }
}

func WithPprof(pprofTurn bool) Option {
	return func(s *Server) { s.pprofTurn = pprofTurn }
}

func WithDebug(isDebug bool) Option {
	return func(s *Server) { s.isDebug = isDebug }
}

func WithOnShutDown(onShutdown []func()) Option {
	return func(s *Server) { s.onShutdown = onShutdown }
}

func New(addr string, router RegisterRouter, opts ...Option) *Server {
	s := &Server{
		Server: &http.Server{
			Addr: addr,
		},
		router: router,
	}

	for _, o := range opts {
		o(s)
	}

	for _, f := range s.onShutdown {
		s.Server.RegisterOnShutdown(f)
	}

	s.Handler = s.initHandler()

	return s
}

func (s *Server) Start() (err error) {
	err = s.Server.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return
}

func (s *Server) Close() (err error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()
	return s.Server.Shutdown(ctx)
}

func (s *Server) initHandler() *gin.Engine {
	server := gin.New()

	if !s.isDebug {
		gin.SetMode(gin.ReleaseMode)
	}

	s.startPprof(server)

	server.Use(s.middleware...)

	s.router(server)

	server.NoRoute(func(c *gin.Context) {
		response.ResponseJSON(c, http.StatusNotFound, nil, response.WrapToast(nil, http.StatusText(http.StatusNotFound)))
		c.AbortWithStatus(http.StatusNotFound)
	})

	return server
}

func (s *Server) startPprof(server *gin.Engine) {
	if !s.pprofTurn {
		return
	}

	runtime.SetBlockProfileRate(1)
	runtime.SetMutexProfileFraction(1)
	pprof.Register(server)
}
