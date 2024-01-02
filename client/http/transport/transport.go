package transport

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/pkg/errors"
	"github.com/why444216978/go-util/assert"

	client "github.com/air-go/rpc/client/http"
	"github.com/air-go/rpc/library/app"
	"github.com/air-go/rpc/library/logger"
	"github.com/air-go/rpc/library/servicer"
)

type RPC struct {
	logger        logger.Logger
	beforePlugins []client.BeforeRequestPlugin
	afterPlugins  []client.AfterRequestPlugin
}

var _ client.Client = (*RPC)(nil)

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

func (r *RPC) Send(ctx context.Context, request client.Request, response client.Response) (err error) {
	if err = r.beforeCheck(ctx, request, response); err != nil {
		return
	}

	ctx = logger.ForkContextOnlyMeta(ctx)

	serviceName := request.GetServiceName()

	logger.AddField(ctx,
		logger.Reflect(logger.ServiceName, serviceName),
		logger.Reflect(logger.Method, request.GetMethod()),
		logger.Reflect(logger.ClientIP, app.LocalIP()),
		logger.Reflect(logger.ClientPort, app.Port()),
		logger.Reflect(logger.Request, request.GetBody()),
		logger.Reflect(logger.API, request.GetPath()),
		logger.Reflect(logger.Request, request.GetBody()))

	defer func() {
		if assert.IsNil(r.logger) {
			return
		}

		if err != nil {
			r.logger.Error(ctx, err.Error())
			return
		}
		r.logger.Info(ctx, "rpc success")
	}()

	// get servicer
	service, ok := servicer.GetServicer(serviceName)
	if !ok {
		err = errors.Errorf("get [%s] servicer is nil", serviceName)
		return
	}

	// construct client
	cli, node, err := r.getClient(ctx, serviceName, service)
	if err != nil {
		return
	}
	logger.AddField(ctx,
		logger.Reflect(logger.ServerIP, node.Host()),
		logger.Reflect(logger.ServerPort, node.Port()))

	if assert.IsNil(node) {
		err = errors.New("node nil")
		return
	}

	// build url
	uu, err := r.buildURL(request, node)
	if err != nil {
		return
	}

	uri := r.formatURI(ctx, uu)
	logger.AddField(ctx, logger.Reflect(logger.URI, uri))

	// build http request
	req, err := r.buildRequest(ctx, request, uu)
	if err != nil {
		return
	}

	_, err = r.send(ctx, cli, service, node, req, response)

	return
}

func (r *RPC) beforeCheck(ctx context.Context, request client.Request, response client.Response) error {
	if assert.IsNil(request) {
		return errors.New("request is nil")
	}

	if assert.IsNil(request.GetCodec()) {
		return errors.New("request codec is nil")
	}

	if assert.IsNil(response) {
		return errors.New("response is nil")
	}

	return nil
}

func (r *RPC) formatURI(ctx context.Context, uu *url.URL) string {
	if uu.RawQuery == "" {
		return uu.Path
	}
	return fmt.Sprintf("%s?%s", uu.Path, uu.RawQuery)
}

func (r *RPC) buildURL(request client.Request, node servicer.Node) (u *url.URL, err error) {
	u = &url.URL{
		Scheme:   "http",
		Host:     fmt.Sprintf("%s:%d", node.Host(), node.Port()),
		Path:     request.GetPath(),
		RawQuery: request.GetQuery().Encode(),
	}

	return
}

func (r *RPC) buildRequest(ctx context.Context, request client.Request, uu *url.URL) (req *http.Request, err error) {
	encode := request.GetCodec()
	if assert.IsNil(encode) {
		err = errors.New("request.Codec is nil")
		return
	}

	if assert.IsNil(request.GetHeader()) {
		request.SetHeader(http.Header{})
	}

	var body io.Reader
	switch r := request.(type) {
	case *client.DefaultRequest:
		if body, err = encode.Encode(r.GetBody()); err != nil {
			return
		}
	case *client.MultiRequest:
		if body, err = encode.Encode(nil); err != nil {
			return
		}
	default:
		err = errors.New("build Request type err")
		return
	}

	if req, err = http.NewRequestWithContext(ctx, request.GetMethod(), uu.String(), body); err != nil {
		return
	}

	// multi encode will set header
	// so set http.Request header must after encode
	req.Header = request.GetHeader()

	return
}

func (r *RPC) beforeSend(ctx context.Context, req *http.Request) (err error) {
	// check context cancel
	if err = ctx.Err(); err != nil {
		return
	}

	// before plugins
	var errH error
	for _, p := range r.beforePlugins {
		ctx, errH = p.Handle(ctx, req)
		if errH != nil && !assert.IsNil(r.logger) {
			r.logger.Warn(ctx, p.Name(), logger.Error(errH))
		}
	}

	return
}

func (r *RPC) send(ctx context.Context, cli *http.Client, service servicer.Servicer, node servicer.Node,
	req *http.Request, response client.Response,
) (resp *http.Response, err error) {
	defer func() {
		// Ensure plugin fields are written to the log.
		logger.AddField(ctx, logger.Reflect(logger.RequestHeader, req.Header))
	}()

	if err = r.beforeSend(ctx, req); err != nil {
		return
	}

	start := time.Now()
	resp, err = cli.Do(req)

	logger.AddField(ctx, logger.Reflect(logger.Cost, time.Since(start).Milliseconds()))
	_ = service.Done(ctx, node, err)
	_ = r.afterSend(ctx, req, resp)

	if err != nil {
		return
	}
	// This don't close body !!!

	logger.AddField(ctx, logger.Reflect(logger.ResponseHeader, resp.Header))
	logger.AddField(ctx, logger.Reflect(logger.Status, resp.StatusCode))

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf("http code is %d", resp.StatusCode)
		return
	}

	err = response.HandleResponse(ctx, resp)

	logger.AddField(ctx, logger.Reflect(logger.Response, response.GetBody()))

	return
}

func (r *RPC) afterSend(ctx context.Context, req *http.Request, resp *http.Response) (err error) {
	// after plugins
	var errH error
	for _, p := range r.afterPlugins {
		ctx, errH = p.Handle(ctx, req, resp)
		if errH != nil && !assert.IsNil(r.logger) {
			r.logger.Warn(ctx, p.Name(), logger.Error(errH))
		}
	}

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
