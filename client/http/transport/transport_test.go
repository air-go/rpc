package transport

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	httpClient "github.com/air-go/rpc/client/http"
	"github.com/air-go/rpc/library/logger/zap/service"
	"github.com/air-go/rpc/library/servicer"
	"github.com/air-go/rpc/library/servicer/mock"
	"github.com/air-go/rpc/mock/third/server"
	jsonCodec "github.com/why444216978/codec/json"
)

func testNew() *RPC {
	l, _ := service.NewServiceLogger("test", &service.Config{})
	return New(WithLogger(l))
}

func TestRPC_Send(t *testing.T) {
	l := testNew()

	convey.Convey("TestRPC_Send", t, func() {
		convey.Convey("request is nil", func() {
			ctx := context.Background()
			err := l.Send(ctx, nil, nil)
			assert.NotNil(t, err)
		})
		convey.Convey("response is nil", func() {
			req := &httpClient.DefaultRequest{
				ServiceName: "test",
				Path:        "/test",
				Method:      http.MethodGet,
				Header:      nil,
				Body:        map[string]interface{}{},
				Codec:       jsonCodec.JSONCodec{},
			}
			ctx := context.Background()
			err := l.Send(ctx, req, nil)
			assert.NotNil(t, err)
		})
		convey.Convey("request codec is nil", func() {
			req := &httpClient.DefaultRequest{
				ServiceName: "test",
				Path:        "/test",
				Method:      http.MethodGet,
				Header:      nil,
				Body:        map[string]interface{}{},
			}
			resp := &httpClient.Response{
				Body: new(map[string]interface{}),
			}
			ctx := context.Background()
			err := l.Send(ctx, req, resp)
			assert.NotNil(t, err)
		})
		convey.Convey("response codec is nil", func() {
			req := &httpClient.DefaultRequest{
				ServiceName: "test",
				Path:        "/test",
				Method:      http.MethodGet,
				Header:      nil,
				Body:        map[string]interface{}{},
				Codec:       jsonCodec.JSONCodec{},
			}
			resp := &httpClient.Response{
				Body: new(map[string]interface{}),
			}
			ctx := context.Background()
			err := l.Send(ctx, req, resp)
			assert.NotNil(t, err)
		})
		convey.Convey("success default request", func() {
			ctx := context.Background()

			// http mock
			srv, err := server.NewHTTP(func(server *gin.Engine) {
				server.GET("/test", func(c *gin.Context) {
					c.JSON(http.StatusOK, nil)
					c.Abort()
				})
			})
			assert.Nil(t, err)
			go func() {
				_ = srv.Start()
			}()
			time.Sleep(time.Second * 1)
			defer func() {
				_ = srv.Stop()
			}()

			arr := strings.Split(srv.Addr(), ":")
			port, _ := strconv.Atoi(arr[1])
			node := servicer.NewNode(arr[0], port)

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

			req := &httpClient.DefaultRequest{
				ServiceName: "test",
				Path:        "/test",
				Method:      http.MethodGet,
				Header:      nil,
				Body:        map[string]interface{}{},
				Codec:       jsonCodec.JSONCodec{},
			}
			resp := &httpClient.Response{
				Body:  new(map[string]interface{}),
				Codec: jsonCodec.JSONCodec{},
			}
			// TODO check resp data
			err = l.Send(ctx, req, resp)
			assert.Nil(t, err)
		})
	})
}

func TestRPC_getClient(t *testing.T) {
}
