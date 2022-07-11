package cache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/golang/mock/gomock"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"

	"github.com/air-go/rpc/library/lock/mock"
	redismock "github.com/air-go/rpc/mock/redis"
)

func TestNew(t *testing.T) {
	convey.Convey("TestNew", t, func() {
		convey.Convey("redis nil", func() {
			ctl := gomock.NewController(t)
			defer ctl.Finish()
			l := mock.NewMockLocker(ctl)

			r, err := New(nil, l)
			assert.NotNil(t, err)
			assert.Nil(t, r)
		})
		convey.Convey("locker nil", func() {
			r, err := New(&redis.Client{}, nil)
			assert.NotNil(t, err)
			assert.Nil(t, r)
		})
		convey.Convey("success", func() {
			ctl := gomock.NewController(t)
			defer ctl.Finish()
			l := mock.NewMockLocker(ctl)

			r, err := New(&redis.Client{}, l, WithTry(3))
			assert.Nil(t, err)
			assert.NotNil(t, r)
		})
	})
}

func TestRedisCache_GetData(t *testing.T) {
	convey.Convey("TestRedisCache_GetData", t, func() {
		convey.Convey("cache no-exists", func() {
			key := "key"

			ctl := gomock.NewController(t)
			defer ctl.Finish()

			l := mock.NewMockLocker(ctl)
			l.EXPECT().Lock(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(true, nil)
			l.EXPECT().Unlock(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)

			rc := redismock.NewMockCmdable(ctl)
			rc.EXPECT().Get(gomock.Any(), gomock.Any()).Times(1).Return(redis.NewStringResult("", nil))
			rc.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(redis.NewStatusResult("", nil))

			cache, err := New(rc, l, WithTry(1))
			assert.Nil(t, err)

			data := &Data{}
			err = cache.GetData(context.Background(), key, time.Second*2, time.Second, getData, data)
			assert.Nil(t, err)
		})
		convey.Convey("cache expiration", func() {
			ctx := context.Background()
			key := "key"

			ctl := gomock.NewController(t)
			defer ctl.Finish()

			l := mock.NewMockLocker(ctl)
			l.EXPECT().Lock(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(true, nil)
			l.EXPECT().Unlock(gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(nil)

			rc := redismock.NewMockCmdable(ctl)
			rc.EXPECT().Get(gomock.Any(), gomock.Any()).Times(1).Return(redis.NewStringResult(`{"ExpireAt":1,"Data":"{\"a\":\"a\"}"}`, nil))
			rc.EXPECT().Set(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).Times(1).Return(redis.NewStatusResult("", nil))

			cache, err := New(rc, l, WithTry(1))
			assert.Nil(t, err)

			data := &Data{}
			err = cache.GetData(ctx, key, time.Second*2, time.Second, getData, data)

			time.Sleep(time.Second)
			assert.Nil(t, err)
		})
	})
}

type Data struct {
	A string `json:"a"`
}

func getData(ctx context.Context, target interface{}) (err error) {
	data, ok := target.(*Data)
	if !ok {
		err = errors.New("err assert")
		return
	}
	data.A = "a"
	return
}
