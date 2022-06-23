package bootstrap

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pkg/errors"
	"github.com/why444216978/go-util/assert"
	"golang.org/x/sync/errgroup"

	"github.com/air-go/rpc/library/registry"
	"github.com/air-go/rpc/server"
)

type Option struct {
	registrar registry.Registrar
}

type OptionFunc func(*Option)

func defaultOption() *Option {
	return &Option{}
}

func WithRegistrar(r registry.Registrar) OptionFunc {
	return func(o *Option) { o.registrar = r }
}

type App struct {
	opt    *Option
	ctx    context.Context
	server server.Server
	cancel func()
}

func NewApp(srv server.Server, opts ...OptionFunc) *App {
	opt := defaultOption()
	for _, o := range opts {
		o(opt)
	}

	ctx, cancel := context.WithCancel(context.Background())

	app := &App{
		opt:    opt,
		ctx:    ctx,
		cancel: cancel,
		server: srv,
	}

	return app
}

func (a *App) Start() error {
	g, _ := errgroup.WithContext(a.ctx)
	g.Go(func() (err error) {
		return a.start()
	})
	g.Go(func() (err error) {
		return a.registerSignal()
	})
	g.Go(func() (err error) {
		return a.registerService()
	})
	g.Go(func() (err error) {
		return a.shutdown()
	})
	return g.Wait()
}

func (a *App) start() error {
	return a.server.Start()
}

func (a *App) registerSignal() (err error) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)

	err = errors.Errorf("%s: exit by signal %v\n", time.Now().Format("2006-01-02 15:04:05"), <-ch)

	// trigger shutdown
	a.cancel()

	return
}

func (a *App) registerService() (err error) {
	if assert.IsNil(a.opt.registrar) {
		return
	}

	return a.opt.registrar.Register(a.ctx)
}

func (a *App) shutdown() (err error) {
	<-a.ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	// server shutdown
	err = a.server.Close()

	// clean resource
	for _, f := range server.CloseFunc {
		_ = f(ctx)
	}

	return
}
