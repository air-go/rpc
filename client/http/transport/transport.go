package transport

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/why444216978/go-util/assert"

	client "github.com/air-go/rpc/client/http"
	"github.com/air-go/rpc/library/logger"
	"github.com/air-go/rpc/library/servicer"
	timeoutLib "github.com/air-go/rpc/server/http/middleware/timeout"
)

type RPC struct {
	logger        logger.Logger
	beforePlugins []client.BeforeRequestPlugin
	afterPlugins  []client.AfterRequestPlugin
}

type Option func(r *RPC)

func WithLogger(logger logger.Logger) Option {
	return func(r *RPC) { r.logger = logger }
}

func WithBeforePlugins(plugins ...client.BeforeRequestPlugin) Option {
	return func(r *RPC) { r.beforePlugins = plugins }
}

func WithAfterPlugins(plugins ...client.AfterRequestPlugin) Option {
	return func(r *RPC) { r.afterPlugins = plugins }
}

func New(opts ...Option) *RPC {
	r := &RPC{}
	for _, o := range opts {
		o(r)
	}

	return r
}

// Send is send HTTP request
//
func (r *RPC) Send(ctx context.Context, serviceName string, request client.Request, response *client.Response) (err error) {
	var (
		cost int64
		node servicer.Node
		cli  *http.Client
	)

	defer func() {
		if r.logger == nil {
			return
		}

		if assert.IsNil(node) {
			node = servicer.Empty()
		}

		if response == nil {
			response = &client.Response{}
		}

		fields := []logger.Field{
			logger.Reflect(logger.ServiceName, serviceName),
			logger.Reflect(logger.Header, request.Header),
			logger.Reflect(logger.Method, request.Method),
			logger.Reflect(logger.API, request.URI),
			logger.Reflect(logger.Request, request.Body),
			logger.Reflect(logger.Response, response.Body),
			logger.Reflect(logger.ServerIP, node.Host()),
			logger.Reflect(logger.ServerPort, node.Port()),
			logger.Reflect(logger.Code, response.HTTPCode),
			logger.Reflect(logger.Cost, cost),
		}
		if err == nil {
			r.logger.Info(ctx, "rpc success", fields...)
			return
		}
		r.logger.Error(ctx, err.Error(), fields...)
	}()

	if response == nil {
		return errors.New("response is nil")
	}

	if assert.IsNil(request.Codec) {
		return errors.New("request.Codec is nil")
	}

	if assert.IsNil(response.Codec) {
		return errors.New("request.Codec is nil")
	}

	if request.Header == nil {
		request.Header = http.Header{}
	}

	// encode request body
	reqReader, err := request.Codec.Encode(request.Body)
	if err != nil {
		return
	}

	// get servicer
	service, ok := servicer.GetServicer(serviceName)
	if !ok {
		err = errors.New("service is nil")
		return
	}

	// construct client
	cli, node, err = r.getClient(ctx, serviceName, service)
	if err != nil {
		return
	}

	// construct request
	url := fmt.Sprintf("http://%s:%d%s", node.Host(), node.Port(), request.URI)
	req, err := http.NewRequestWithContext(ctx, request.Method, url, reqReader)
	if err != nil {
		return
	}

	// timeout deliver
	if err = timeoutLib.SetHeader(ctx, request.Header); err != nil {
		return
	}

	// set header
	req.Header = request.Header

	// before plugins
	for _, plugin := range r.beforePlugins {
		_ = plugin.Handle(ctx, req)
	}

	// start time
	start := time.Now()

	// check context cancel
	if err = ctx.Err(); err != nil {
		return
	}

	// send request
	resp, err := cli.Do(req)

	_ = service.Done(ctx, node, err)

	// after plugins
	for _, plugin := range r.afterPlugins {
		_ = plugin.Handle(ctx, req, resp)
	}

	// cost duration
	cost = time.Since(start).Milliseconds()
	if err != nil {
		return
	}
	defer resp.Body.Close()

	response.HTTPCode = resp.StatusCode
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("http code is %d", resp.StatusCode)
		return
	}

	// decode response body
	err = response.Codec.Decode(resp.Body, response.Body)

	return
}

func (r *RPC) getClient(ctx context.Context, serviceName string, service servicer.Servicer) (client *http.Client, node servicer.Node, err error) {
	node, err = service.Pick(ctx)
	if err != nil {
		return
	}

	address := node.Address()

	tp := &http.Transport{
		MaxIdleConnsPerHost: 30,
		MaxConnsPerHost:     30,
		IdleConnTimeout:     time.Minute,
		DialContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
			conn, err := (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 60 * time.Second,
			}).DialContext(ctx, "tcp", address)
			if err != nil {
				return nil, err
			}
			return conn, nil
		},
		DialTLSContext: func(ctx context.Context, network, _ string) (net.Conn, error) {
			pool := x509.NewCertPool()
			pool.AppendCertsFromPEM(service.GetCaCrt())
			cliCrt, err := tls.X509KeyPair(service.GetClientPem(), service.GetClientKey())
			if err != nil {
				err = errors.New("server pem error " + err.Error())
				return nil, err
			}

			conn, err := (&net.Dialer{
				Timeout:   30 * time.Second,
				KeepAlive: 60 * time.Second,
			}).DialContext(ctx, "tcp", address)
			if err != nil {
				return nil, err
			}

			return tls.Client(conn, &tls.Config{
				RootCAs:      pool,
				Certificates: []tls.Certificate{cliCrt},
				ServerName:   serviceName,
			}), err
		},
	}
	client = &http.Client{Transport: tp}

	return
}
