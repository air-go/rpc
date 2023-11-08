package transport

import (
	"context"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"

	httpClient "github.com/air-go/rpc/client/http"
	"github.com/air-go/rpc/library/logger"
	"github.com/air-go/rpc/library/logger/zap"
	"github.com/air-go/rpc/library/servicer"
	"github.com/air-go/rpc/library/servicer/mock"
	"github.com/air-go/rpc/mock/tools/server"
	jsonCodec "github.com/why444216978/codec/json"
)

type testBeforeCheckPlugin struct {
	t *testing.T
}

var _ httpClient.BeforeRequestPlugin = (*testBeforeCheckPlugin)(nil)

func (*testBeforeCheckPlugin) Handle(ctx context.Context, req *http.Request) (context.Context, error) {
	return ctx, nil
}

func (*testBeforeCheckPlugin) Name() string {
	return "testBeforeCheckPlugin"
}

type testAfterCheckPlugin struct {
	t *testing.T
}

var _ httpClient.AfterRequestPlugin = (*testAfterCheckPlugin)(nil)

func (p *testAfterCheckPlugin) Handle(ctx context.Context, req *http.Request, resp *http.Response) (context.Context, error) {
	m := map[string]string{}
	logger.RangeFields(ctx, func(f logger.Field) {
		m[f.Key()] = cast.ToString(f.Value())
	})
	assert.Equal(p.t, "/test", m[logger.API])
	assert.Equal(p.t, "/test?data=data", m[logger.URI])
	assert.Equal(p.t, "GET", m[logger.Method])
	assert.Equal(p.t, "test", m[logger.ServiceName])
	return ctx, nil
}

func (*testAfterCheckPlugin) Name() string {
	return "testAfterCheckPlugin"
}

func TestDefaultRequest(t *testing.T) {
	convey.Convey("TestDefaultRequest", t, func() {
		convey.Convey("request is nil", func() {
			l := New()
			ctx := logger.InitFieldsContainer(context.Background())
			err := l.Send(ctx, nil, nil)
			assert.NotNil(t, err)
		})
		convey.Convey("response is nil", func() {
			l := New()
			req := &httpClient.DefaultRequest{
				ServiceName: "test",
				Path:        "/test",
				Method:      http.MethodGet,
				Header:      nil,
				Body:        map[string]interface{}{},
				Codec:       jsonCodec.JSONCodec{},
			}
			ctx := logger.InitFieldsContainer(context.Background())
			err := l.Send(ctx, req, nil)
			assert.NotNil(t, err)
		})
		convey.Convey("request codec is nil", func() {
			l := New()
			req := &httpClient.DefaultRequest{
				ServiceName: "test",
				Path:        "/test",
				Method:      http.MethodGet,
				Header:      nil,
				Body:        map[string]interface{}{},
			}
			resp := &httpClient.DataResponse{
				Body: new(map[string]interface{}),
			}
			ctx := logger.InitFieldsContainer(context.Background())
			err := l.Send(ctx, req, resp)
			assert.NotNil(t, err)
		})
		convey.Convey("success default request", func() {
			l := New(
				WithLogger(zap.StdLogger),
				WithBeforePlugins(&testBeforeCheckPlugin{t: t}),
				WithAfterPlugins(&testAfterCheckPlugin{t: t}),
			)

			ctx := logger.InitFieldsContainer(context.Background())

			logger.AddField(ctx, logger.Reflect(logger.ServiceName, "old_service_name"))

			// http mock
			respBody := map[string]string{"data": "data"}
			srv, err := server.NewHTTP(func(server *gin.Engine) {
				server.GET("/test", func(c *gin.Context) {
					c.JSON(http.StatusOK, respBody)
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
				Query:       url.Values{"data": []string{"data"}},
				Method:      http.MethodGet,
				Header:      nil,
				Body:        map[string]string{},
				Codec:       jsonCodec.JSONCodec{},
			}
			resp := &httpClient.DataResponse{
				Body:  new(map[string]string),
				Codec: jsonCodec.JSONCodec{},
			}
			err = l.Send(ctx, req, resp)
			assert.Nil(t, err)
			assert.Equal(t, &respBody, resp.Body)
		})
	})
}
