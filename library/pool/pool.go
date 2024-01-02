package pool

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/why444216978/go-util/assert"
	uruntime "github.com/why444216978/go-util/runtime"
)

type Pool interface {
	Get() (Conn, error)
	Put(ctx context.Context, conn Conn) error
	Remove(ctx context.Context, conn Conn) error
	Close(ctx context.Context) error
	Stats() Stats
}

type Stats struct {
	Hits            uint64
	Misses          uint64
	Removes         uint64
	IDConflicts     uint64
	ErrOverMaxSizes uint64
	Timeouts        uint64
	Breakers        uint64
	ConnCount       uint32
	IdleCount       uint32
}

type connPool struct {
	opt *Options

	newFunc func(ctx context.Context) (Conn, error)

	lastConnError atomic.Value

	connQueue chan struct{} // 大小等于链接池大小，控制最大连接数，获取时写，归还时读（获得连接超时控制）

	mux sync.Mutex

	conns     map[string]Conn
	connCount uint32

	idleConns []string
	idleCount uint32

	stats Stats

	closed uint32

	isBreak        uint32
	connErrorCount uint32
}

func NewPool(newFunc func(ctx context.Context) (Conn, error), opts ...Option) (*connPool, error) {
	if assert.IsNil(newFunc) {
		return nil, ErrNewFunc
	}

	opt := defaultOptions()
	for _, o := range opts {
		o(opt)
	}

	p := &connPool{
		opt:       opt,
		newFunc:   newFunc,
		connQueue: make(chan struct{}, opt.PoolSize),
		conns:     make(map[string]Conn, opt.PoolSize),
		idleConns: make([]string, 0, opt.IdleSize),
	}

	if err := p.addIdle(); err != nil {
		return nil, err
	}

	return p, nil
}

func (p *connPool) Get(ctx context.Context) (conn Conn, err error) {
	if p.isClosed() {
		return nil, ErrClosed
	}

	if conn, err = p.getIdle(ctx); err != nil {
		return
	}

	if !assert.IsNil(conn) {
		return
	}

	if conn, err = p.newConn(ctx); err != nil {
		return
	}

	return
}

func (p *connPool) newConn(ctx context.Context) (conn Conn, err error) {
	p.mux.Lock()
	defer p.mux.Unlock()

	if p.connCount >= p.opt.PoolSize && p.opt.GetQuickFail {
		atomic.AddUint64(&p.stats.ErrOverMaxSizes, 1)
		return nil, ErrOverMaxSize
	}

	if err = p.connQueueIn(ctx); err != nil {
		return
	}

	conn, err = p.handleNew(ctx)
	if err == nil {
		return
	}

	p.connQueueOut()

	return
}

func (p *connPool) handleNew(ctx context.Context) (conn Conn, err error) {
	if atomic.LoadUint32(&p.connErrorCount) >= p.opt.BreakerThreshold {
		atomic.AddUint64(&p.stats.Breakers, 1)
		return nil, p.getLastConnError()
	}

	defer func() {
		if r := recover(); r != nil {
			err = uruntime.WrapStackError(r)
		}
	}()

	conn, err = p.newFunc(ctx)
	if err != nil {
		atomic.AddUint32(&p.connErrorCount, 1)
		p.setLastConnError(err)
		return
	}

	id := conn.GetID()
	if _, ok := p.conns[id]; ok {
		_ = conn.Close(context.Background())
		atomic.AddUint64(&p.stats.IDConflicts, 1)
		return nil, ErrIDConflict
	}
	p.conns[id] = conn
	p.connCount++

	atomic.StoreUint32(&p.connErrorCount, 0)

	return
}

func (p *connPool) addIdle() (err error) {
	if p.opt.IdleSize == 0 {
		return
	}

	var conn Conn
	for {
		if atomic.LoadUint32(&p.connErrorCount) >= p.opt.BreakerThreshold {
			return p.getLastConnError()
		}

		if isBreak := func() bool {
			p.mux.Lock()
			defer p.mux.Unlock()

			// before check avoid block
			if p.connCount >= p.opt.PoolSize || p.idleCount >= p.opt.IdleSize {
				return true
			}

			p.connQueueIn(context.Background())

			conn, err = p.handleNew(context.Background())
			if errors.Is(err, ErrIDConflict) {
				return true
			}
			if err != nil {
				return false
			}

			p.idleConns = append(p.idleConns, conn.GetID())
			p.idleCount++

			return false
		}(); isBreak {
			return
		}

	}
}

func (p *connPool) getIdle(ctx context.Context) (conn Conn, err error) {
	if p.opt.IdleSize == 0 {
		return
	}

	p.mux.Lock()
	defer p.mux.Unlock()

	if p.idleCount == 0 {
		atomic.AddUint64(&p.stats.Misses, 1)
		return
	}

	atomic.AddUint64(&p.stats.Hits, 1)

	conn = p.conns[p.idleConns[0]]

	copy(p.idleConns, p.idleConns[1:])
	p.idleConns = p.idleConns[:p.idleCount-1]
	p.idleCount--

	go func() {
		_ = p.addIdle()
	}()

	return
}

func (p *connPool) returnIdle(conn Conn) bool {
	p.mux.Lock()
	defer p.mux.Unlock()

	if p.idleCount >= p.opt.IdleSize {
		return false
	}

	p.idleConns = append(p.idleConns, conn.GetID())
	p.idleCount++

	return true
}

func (p *connPool) connQueueIn(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	select {
	// write success
	case p.connQueue <- struct{}{}:
		return nil
	default:
	}

	timer := timerPool.Get().(*time.Timer)
	defer timerPool.Put(timer)
	timer.Reset(p.opt.GetConnTimeout)

	select {
	case <-ctx.Done():
		// ensure empty
		if !timer.Stop() {
			<-timer.C
		}
		return ctx.Err()
	case p.connQueue <- struct{}{}:
		if !timer.Stop() {
			<-timer.C
		}
		return nil
	case <-timer.C:
		atomic.AddUint64(&p.stats.Timeouts, 1)
		return ErrGetTimeout
	}
}

func (p *connPool) connQueueOut() {
	<-p.connQueue
}

func (p *connPool) Remove(ctx context.Context, conn Conn) (err error) {
	if p.isClosed() {
		return ErrClosed
	}

	p.mux.Lock()
	defer p.mux.Unlock()

	if ok := p.removeConn(ctx, conn); !ok {
		return
	}
	atomic.AddUint64(&p.stats.Removes, 1)

	id := conn.GetID()
	for i, key := range p.idleConns {
		if key == id {
			p.idleConns = append(p.idleConns[:i], p.idleConns[i+1:]...)
			p.idleCount--
		}
	}

	return
}

func (p *connPool) Put(ctx context.Context, conn Conn) (err error) {
	if p.isClosed() {
		return ErrClosed
	}

	if p.returnIdle(conn) {
		return
	}

	p.removeConnWithLock(ctx, conn)

	return
}

func (p *connPool) removeConn(ctx context.Context, conn Conn) bool {
	if _, ok := p.conns[conn.GetID()]; !ok {
		return false
	}
	delete(p.conns, conn.GetID())
	p.connCount--
	p.connQueueOut()
	go func() {
		_ = conn.Close(ctx)
	}()
	return true
}

func (p *connPool) removeConnWithLock(ctx context.Context, conn Conn) {
	p.mux.Lock()
	p.removeConn(ctx, conn)
	p.mux.Unlock()
}

func (p *connPool) Close(ctx context.Context) error {
	if !atomic.CompareAndSwapUint32(&p.closed, 0, 1) {
		return ErrClosed
	}

	p.mux.Lock()
	for _, c := range p.conns {
		go func(c Conn) {
			_ = c.Close(ctx)
		}(c)
	}
	p.conns = nil
	p.connCount = 0
	p.idleConns = nil
	p.idleCount = 0
	p.mux.Unlock()

	return nil
}

func (p *connPool) isClosed() bool {
	return atomic.LoadUint32(&p.closed) == 1
}

func (p *connPool) Stats() Stats {
	p.mux.Lock()
	s := Stats{
		Hits:      atomic.LoadUint64(&p.stats.Hits),
		Misses:    atomic.LoadUint64(&p.stats.Misses),
		Timeouts:  atomic.LoadUint64(&p.stats.Timeouts),
		ConnCount: atomic.LoadUint32(&p.connCount),
		IdleCount: atomic.LoadUint32(&p.idleCount),
	}
	p.mux.Unlock()
	return s
}

type lastConnError struct {
	err error
}

func (e *lastConnError) Error() string {
	if e.err == nil {
		return ""
	}
	return e.err.Error()
}

func (p *connPool) setLastConnError(err error) {
	p.lastConnError.Store(&lastConnError{err: err})
}

func (p *connPool) getLastConnError() error {
	err, _ := p.lastConnError.Load().(*lastConnError)
	if err != nil {
		return err.err
	}
	return nil
}

var timerPool = sync.Pool{
	New: func() interface{} {
		t := time.NewTimer(time.Hour)
		t.Stop()
		return t
	},
}
