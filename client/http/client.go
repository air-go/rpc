package http

import (
	"context"
	"net/http"

	"github.com/why444216978/codec"
)

type Request struct {
	URI    string
	Method string
	Header http.Header
	Body   interface{}
	Codec  codec.Codec
}

type Response struct {
	HTTPCode int
	Body     interface{}
	Codec    codec.Codec
}

type Client interface {
	Send(ctx context.Context, serviceName string, request Request, response *Response) (err error)
}
