package pool

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
)

var (
	id     int64
	idLock sync.Mutex
)

func generateID() string {
	idLock.Lock()
	id++
	s := cast.ToString(id)
	idLock.Unlock()
	return s
}

type TestConn struct {
	id string
}

func (c *TestConn) GetID() string { return c.id }

func (*TestConn) Close(ctx context.Context) error { return nil }

type TestIDErrConn struct {
	id string
}

func (c *TestIDErrConn) GetID() string { return "1" }

func (*TestIDErrConn) Close(ctx context.Context) error { return nil }

func TestConnPool(t *testing.T) {
	ctx := context.Background()

	// test quick fail
	func() {
		p, err := NewPool(func(ctx context.Context) (Conn, error) {
			return &TestConn{id: generateID()}, nil
		},
			WithPoolSize(3),
			WithIdleSize(2),
			WithBreakerThreshold(1),
			WithGetQuickFail(),
			WithGetConnTimeout(time.Second),
		)
		assert.Nil(t, err)
		stats := p.Stats()
		assert.Equal(t, uint32(2), stats.ConnCount)
		assert.Equal(t, uint32(2), stats.IdleCount)

		conn1, err := p.Get(ctx)
		assert.Nil(t, err)
		assert.Equal(t, "1", conn1.GetID())

		conn2, err := p.Get(ctx)
		assert.Nil(t, err)
		assert.Equal(t, "2", conn2.GetID())

		conn3, err := p.Get(ctx)
		assert.Nil(t, err)
		assert.Equal(t, "3", conn3.GetID())

		_, err = p.Get(ctx)
		assert.Equal(t, ErrOverMaxSize, err)

		err = p.Put(ctx, conn3)
		assert.Nil(t, err)

		err = p.Remove(ctx, conn3)
		assert.Nil(t, err)

		conn4, err := p.Get(ctx)
		assert.Nil(t, nil)
		assert.Equal(t, "4", conn4.GetID())

		err = p.Put(ctx, conn4)
		assert.Nil(t, err)

		err = p.Put(ctx, conn2)
		assert.Nil(t, err)

		err = p.Put(ctx, conn1)
		assert.Nil(t, err)

		// close
		err = p.Close(ctx)
		assert.Nil(t, err)
	}()

	// test get block timeout
	func() {
		p, err := NewPool(func(ctx context.Context) (Conn, error) {
			return &TestConn{id: generateID()}, nil
		},
			WithPoolSize(3),
			WithIdleSize(2),
			WithBreakerThreshold(1),
			WithGetConnTimeout(time.Second*1),
		)
		assert.Nil(t, err)
		stats := p.Stats()
		assert.Equal(t, uint32(2), stats.ConnCount)
		assert.Equal(t, uint32(2), stats.IdleCount)
		_, err = p.Get(ctx)
		assert.Nil(t, err)
		_, err = p.Get(ctx)
		assert.Nil(t, err)
		_, err = p.Get(ctx)
		assert.Nil(t, err)

		// timeout
		_, err = p.Get(ctx)
		assert.Equal(t, ErrGetTimeout, err)

		// close
		err = p.Close(ctx)
		assert.Nil(t, err)
	}()

	// test get block wait success
	func() {
		p, err := NewPool(func(ctx context.Context) (Conn, error) {
			return &TestConn{id: generateID()}, nil
		},
			WithPoolSize(3),
			WithIdleSize(0),
			WithBreakerThreshold(1),
			WithGetConnTimeout(time.Second*3),
		)
		assert.Nil(t, err)

		p.connQueueIn(ctx)
		p.connQueueIn(ctx)
		p.connQueueIn(ctx)
		go func() {
			time.Sleep(time.Second * 1)
			p.connQueueOut()
		}()
		p.connQueueIn(ctx)

		// close
		err = p.Close(ctx)
		assert.Nil(t, err)
	}()

	// test new err
	func() {
		e := errors.New("error")
		_, err := NewPool(func(ctx context.Context) (Conn, error) {
			return nil, e
		},
			WithPoolSize(3),
			WithIdleSize(2),
			WithBreakerThreshold(1),
			WithGetConnTimeout(time.Second),
		)
		assert.Equal(t, e, err)
	}()

	// test id err
	func() {
		_, err := NewPool(func(ctx context.Context) (Conn, error) {
			return &TestIDErrConn{}, nil
		},
			WithPoolSize(3),
			WithIdleSize(2),
			WithBreakerThreshold(1),
			WithGetConnTimeout(time.Second),
		)
		assert.Equal(t, ErrIDConflict, err)
	}()
}
