package transport

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	httpClient "github.com/air-go/rpc/client/http"
	"github.com/air-go/rpc/library/logger/zap/service"
	"github.com/air-go/rpc/library/servicer"
	"github.com/air-go/rpc/library/servicer/mock"
	jsonCodec "github.com/why444216978/codec/json"
)

func testNew() *RPC {
	l, _ := service.NewServiceLogger("test", &service.Config{})
	return New(WithLogger(l))
}

func TestRPC_Send(t *testing.T) {
	l := testNew()

	convey.Convey("TestRPC_Send", t, func() {
		convey.Convey("response == nil", func() {
			req := httpClient.Request{
				URI:    "/test",
				Method: http.MethodGet,
				Header: nil,
				Body:   map[string]interface{}{},
				Codec:  jsonCodec.JSONCodec{},
			}
			ctx := context.Background()
			err := l.Send(ctx, "test", req, nil)
			assert.NotNil(t, err)
		})
		convey.Convey("assert.IsNil(request.Codec)", func() {
			req := httpClient.Request{
				URI:    "/test",
				Method: http.MethodGet,
				Header: nil,
				Body:   map[string]interface{}{},
			}
			resp := &httpClient.Response{
				Body: new(map[string]interface{}),
			}
			ctx := context.Background()
			err := l.Send(ctx, "test", req, resp)
			assert.NotNil(t, err)
		})
		convey.Convey("assert.IsNil(response.Codec)", func() {
			req := httpClient.Request{
				URI:    "/test",
				Method: http.MethodGet,
				Header: nil,
				Body:   map[string]interface{}{},
				Codec:  jsonCodec.JSONCodec{},
			}
			resp := &httpClient.Response{
				Body: new(map[string]interface{}),
			}
			ctx := context.Background()
			err := l.Send(ctx, "test", req, resp)
			assert.NotNil(t, err)
		})
		convey.Convey("success", func() {
			ctx := context.Background()
			node := servicer.NewNode("127.0.0.1", 80)

			// servicer mock
			ctl := gomock.NewController(t)
			defer ctl.Finish()
			s := mock.NewMockServicer(ctl)
			s.EXPECT().Name().AnyTimes().Return("test")
			s.EXPECT().Pick(gomock.Any()).Times(1).Return(node, nil)
			// s.EXPECT().GetCaCrt().Times(1).Return([]byte(""))
			// s.EXPECT().GetClientPem().Times(1).Return([]byte(""))
			// s.EXPECT().GetClientKey().Times(1).Return([]byte(""))
			s.EXPECT().Done(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)
			_ = servicer.SetServicer(s)

			// http mock
			patch := gomonkey.ApplyMethodSeq(reflect.TypeOf(&http.Client{}), "Do", []gomonkey.OutputCell{
				{Values: gomonkey.Params{&http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewBufferString("{}")),
				}, nil}},
			})
			defer patch.Reset()

			req := httpClient.Request{
				URI:    "/test",
				Method: http.MethodGet,
				Header: nil,
				Body:   map[string]interface{}{},
				Codec:  jsonCodec.JSONCodec{},
			}
			resp := &httpClient.Response{
				Body:  new(map[string]interface{}),
				Codec: jsonCodec.JSONCodec{},
			}
			err := l.Send(ctx, "test", req, resp)
			assert.Nil(t, err)
		})
	})
}

func TestRPC_getClient(t *testing.T) {
}
