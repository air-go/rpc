package redis

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/air-go/rpc/library/lock"
	redismock "github.com/go-redis/redismock/v8"
	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

var (
	ctx      = context.Background()
	key      = "lock"
	val      = "1"
	duration = time.Second
)

func TestNew(t *testing.T) {
	r, _ := redismock.NewClientMock()

	convey.Convey("TestNew", t, func() {
		convey.Convey("success", func() {
			var err error

			rc, err := New(r)
			assert.Nil(t, err)
			assert.Equal(t, r, rc.c)
		})
		convey.Convey("fail", func() {
			rc, err := New(nil)
			assert.Equal(t, lock.ErrClientNil, err)
			assert.Nil(t, rc)
		})
	})
}

func TestRedisLock_Lock(t *testing.T) {
	convey.Convey("TestRedisLock_Lock", t, func() {
		convey.Convey("once success", func() {
			r, rc := redismock.NewClientMock()
			rl, _ := New(r)

			expect := rc.ExpectSetNX(key, val, duration)
			expect.SetVal(true)
			expect.SetErr(nil)

			ok, err := rl.Lock(ctx, key, val, duration, 1)
			assert.Nil(t, err)
			assert.Equal(t, true, ok)
		})
		convey.Convey("fail error", func() {
			r, rc := redismock.NewClientMock()
			rl, _ := New(r)

			expect := rc.ExpectSetNX(key, val, duration)
			expect.SetVal(false)
			expect.SetErr(errors.New("err"))

			ok, err := rl.Lock(ctx, key, val, duration, 3)
			assert.NotNil(t, err)
			assert.Equal(t, false, ok)
		})
		convey.Convey("fail ttl", func() {
			r, rc := redismock.NewClientMock()
			rl, _ := New(r)

			expect := rc.ExpectSetNX(key, val, duration)
			expect.SetVal(false)
			expect.SetErr(nil)

			ttl := rc.ExpectTTL(key)
			ttl.SetVal(time.Millisecond)
			ttl.SetErr(errors.New("err"))

			ok, err := rl.Lock(ctx, key, val, duration, 3)
			assert.NotNil(t, err)
			assert.Equal(t, false, ok)
		})
		convey.Convey("fail try", func() {
			r, rc := redismock.NewClientMock()
			rl, _ := New(r)

			set1 := rc.ExpectSetNX(key, val, duration)
			set1.SetVal(false)
			set1.SetErr(nil)

			ttl1 := rc.ExpectTTL(key)
			ttl1.SetVal(time.Millisecond)

			set2 := rc.ExpectSetNX(key, val, duration)
			set2.SetVal(false)
			set2.SetErr(nil)

			ttl2 := rc.ExpectTTL(key)
			ttl2.SetVal(time.Millisecond)

			set3 := rc.ExpectSetNX(key, val, duration)
			set3.SetVal(false)
			set3.SetErr(nil)

			ttl3 := rc.ExpectTTL(key)
			ttl3.SetVal(time.Millisecond)

			ok, err := rl.Lock(ctx, key, val, duration, 3)
			assert.Nil(t, err)
			assert.Equal(t, false, ok)
		})
	})
}

func TestRedisLock_Unlock(t *testing.T) {
	convey.Convey("TestRedisLock_Unlock", t, func() {
		convey.Convey("success", func() {
			r, rc := redismock.NewClientMock()
			rl, _ := New(r)

			expect := rc.ExpectEval(lockLua, []string{key}, val)
			expect.SetVal(true)
			expect.SetErr(nil)

			rl.Unlock(ctx, key, val)
		})
		convey.Convey("fail error", func() {
			r, rc := redismock.NewClientMock()
			rl, _ := New(r)

			expect := rc.ExpectEval(lockLua, []string{key}, val)
			expect.SetErr(errors.New("err"))

			rl.Unlock(ctx, key, val)
		})
		convey.Convey("fail result lockFail", func() {
			r, rc := redismock.NewClientMock()
			rl, _ := New(r)

			expect := rc.ExpectEval(lockLua, []string{key}, val)
			expect.SetVal(lockFail)
			expect.SetErr(nil)

			rl.Unlock(ctx, key, val)
		})
	})
}
