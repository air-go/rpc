package http

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/why444216978/codec"
	"github.com/why444216978/go-util/assert"
)

type Response interface {
	HandleResponse(ctx context.Context, rsp *http.Response) (err error)
	GetResponse() *http.Response
	GetBody() interface{}
}

type DataResponse struct {
	response *http.Response
	Body     interface{}
	Codec    codec.Codec
}

func (resp *DataResponse) HandleResponse(ctx context.Context, rsp *http.Response) (err error) {
	if assert.IsNil(resp.Codec) {
		return errors.New("DataResponse codec is nil")
	}

	bb, err := io.ReadAll(rsp.Body)
	if err != nil {
		return err
	}
	_ = rsp.Body.Close()

	resp.response = rsp
	resp.response.Body = io.NopCloser(bytes.NewBuffer(bb))

	if resp.Body != nil {
		err = resp.Codec.Decode(bytes.NewBuffer(bb), &resp.Body)
	}

	return
}

func (resp *DataResponse) GetResponse() *http.Response {
	return resp.response
}

func (resp *DataResponse) GetBody() interface{} {
	return resp.Body
}
