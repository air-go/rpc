package cron

import (
	"context"
	"fmt"
	"reflect"
	"runtime"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/robfig/cron/v3"
	"github.com/why444216978/go-util/assert"
	panicErr "github.com/why444216978/go-util/panic"
	"github.com/why444216978/go-util/snowflake"

	"github.com/air-go/rpc/library/app"
	"github.com/air-go/rpc/library/lock"
	"github.com/air-go/rpc/library/logger"
)

var (
	defaultMiniLockTTL = time.Second
	defaultLockFormat  = "lock:cron:%s:%s"
)

var secondParser = cron.NewParser(
	cron.Second | cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor,
)

type Cron struct {
	*cron.Cron
	logger     logger.Logger
	lock       lock.Locker
	opts       *Options
	name       string
	rw         sync.RWMutex
	funcEntity map[string]cron.Entry
}

type Options struct {
	locker        lock.Locker
	errorCallback func(error)
	miniLockTTL   time.Duration
	lockFormat    string
}

func defaultOptions() *Options {
	return &Options{
		miniLockTTL: defaultMiniLockTTL,
		lockFormat:  defaultLockFormat,
	}
}

type Option func(*Options)

func WithLocker(l lock.Locker) Option {
	return func(o *Options) { o.locker = l }
}

func WithErrCallback(f func(error)) Option {
	return func(o *Options) { o.errorCallback = f }
}

func WithMiniLockTTL(ttl time.Duration) Option {
	return func(o *Options) { o.miniLockTTL = ttl }
}

func WithLockFormat(lockFormat string) Option {
	return func(o *Options) { o.lockFormat = lockFormat }
}

func NewCron(name string, l logger.Logger, options ...Option) (c *Cron, err error) {
	opts := defaultOptions()

	for _, o := range options {
		o(opts)
	}

	if assert.IsNil(l == nil) {
		err = errors.New("logger is nil")
		return
	}

	if opts.miniLockTTL < time.Second {
		opts.miniLockTTL = time.Second
	}

	c = &Cron{
		logger:     l,
		lock:       opts.locker,
		Cron:       cron.New(cron.WithSeconds()),
		opts:       opts,
		name:       name,
		funcEntity: make(map[string]cron.Entry),
	}

	return c, nil
}

func (c *Cron) AddJob(spec string, cmd func()) (cron.EntryID, error) {
	return c.addJob(spec, FuncJob(cmd))
}

func (c *Cron) Start() {
	c.Cron.Start()
}

func (c *Cron) Stop() {
	c.Cron.Stop()
}

func (c *Cron) Name() string {
	return c.name
}

func (c *Cron) addJob(spec string, cmd cron.Job) (cron.EntryID, error) {
	funcName := cmd.(FuncJob).FunctionName()
	lockKey := c.getLockKey(funcName)

	entityID, err := c.AddFunc(spec, c.handle(cmd, funcName, spec, lockKey))
	c.setFuncEntity(funcName, entityID)

	return entityID, err
}

func (c *Cron) handle(cmd cron.Job, funcName, spec, lockKey string) func() {
	return func() {
		var err error

		ctx := logger.InitFieldsContainer(context.Background())

		schedule := c.getFuncEntity(funcName).Schedule
		ttl := c.getLockDuration(schedule)
		random := snowflake.Generate().String()

		if !assert.IsNil(c.lock) {
			ok, err := c.lock.Lock(ctx, lockKey, random, ttl, 3)
			if err != nil {
				c.logger.Error(ctx, errors.Wrap(err, "crontab fun Lock err").Error(),
					logger.Reflect("spec", spec),
					logger.Reflect(logger.ClientIP, app.LocalIP()),
					logger.Reflect(logger.API, c.name),
					logger.Reflect(logger.Method, funcName))
				return
			}
			if !ok {
				c.logger.Error(ctx, "crontab fun Lock !ok",
					logger.Reflect("spec", spec),
					logger.Reflect(logger.ClientIP, app.LocalIP()),
					logger.Reflect(logger.API, c.name),
					logger.Reflect(logger.Method, funcName))
				return
			}
		}

		start := time.Now()
		func() {
			defer func() {
				if r := recover(); r != nil {
					err := panicErr.NewPanicError(r)
					c.logger.Error(ctx, "crontab handler panic",
						logger.Reflect("panic", err),
						logger.Reflect("spec", spec),
						logger.Reflect(logger.ClientIP, app.LocalIP()),
						logger.Reflect(logger.API, c.name),
						logger.Reflect(logger.Method, funcName))
				}
			}()
			cmd.Run()
		}()

		c.logger.Info(ctx, "handle "+c.name,
			logger.Reflect("spec", spec),
			logger.Reflect(logger.Cost, time.Since(start).Milliseconds()),
			logger.Reflect(logger.ClientIP, app.LocalIP()),
			logger.Reflect(logger.API, c.name),
			logger.Reflect(logger.Method, funcName))

		if !assert.IsNil(c.lock) {
			if err = c.lock.Unlock(ctx, lockKey, random); err != nil {
				c.logger.Error(ctx, errors.Wrap(err, "crontab fun Unlock err").Error(),
					logger.Reflect("spec", spec),
					logger.Reflect(logger.ClientIP, app.LocalIP()),
					logger.Reflect(logger.API, c.name),
					logger.Reflect(logger.Method, funcName))
				return
			}
		}
	}
}

func (c *Cron) getLockDuration(schedule cron.Schedule) time.Duration {
	now := time.Now()
	next := schedule.Next(now)
	ttl := time.Until(next)
	if ttl < c.opts.miniLockTTL {
		ttl = c.opts.miniLockTTL
	}

	return ttl
}

func (c *Cron) getLockKey(funcName string) string {
	return fmt.Sprintf(c.opts.lockFormat, c.name, funcName)
}

func (c *Cron) setFuncEntity(funcName string, entityID cron.EntryID) {
	c.rw.Lock()
	defer c.rw.Unlock()

	c.funcEntity[funcName] = c.Entry(entityID)
}

func (c *Cron) getFuncEntity(funcName string) cron.Entry {
	c.rw.RLock()
	defer c.rw.RUnlock()

	return c.funcEntity[funcName]
}

type FuncJob func()

func (f FuncJob) Run() { f() }

func (f FuncJob) Function() func() { return f }

func (f FuncJob) FunctionName() string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}
