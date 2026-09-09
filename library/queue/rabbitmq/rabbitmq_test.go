package rabbitmq

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"github.com/stretchr/testify/assert"

	"github.com/air-go/rpc/library/logger"
	"github.com/air-go/rpc/library/logger/nop"
	"github.com/air-go/rpc/library/queue"
)

// newTestClient 构造一个不含真实连接的客户端，用于测试不依赖 broker 的分支。
func newTestClient(scenes ...Scene) *Client {
	cli := &Client{
		opts: &option{
			logger:        nop.Logger,
			prefetchCount: 10,
		},
		serviceName: "rabbitmq",
		host:        "127.0.0.1",
		port:        5672,
		scheme:      schemeAMQP,
		scenes:      make(map[string]*Scene),
		close:       make(chan struct{}),
	}
	for i := range scenes {
		s := &scenes[i]
		cli.scenes[s.Name] = s
	}
	return cli
}

func TestNew_ConfigNil(t *testing.T) {
	cli, err := New(nil)
	assert.Nil(t, cli)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "config is nil")
}

func TestNew_ScenesEmpty(t *testing.T) {
	cli, err := New(&Config{Host: "127.0.0.1", Port: 5672})
	assert.Nil(t, cli)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "scenes is empty")
}

func TestNew_UnsupportedScheme(t *testing.T) {
	cli, err := New(&Config{
		Scheme: "http",
		Scenes: []Scene{{Name: "s"}},
	})
	assert.Nil(t, cli)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "unsupported scheme")
}

func TestNew_LoggerNil(t *testing.T) {
	cli, err := New(&Config{Scenes: []Scene{{Name: "s"}}}, WithLogger(nil))
	assert.Nil(t, cli)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "logger is nil")
}

// TestNew_URL 校验 URL 拼装：密码特殊字符转义、vhost 归一化为 path。
func TestNew_URL(t *testing.T) {
	cases := []struct {
		name  string
		cfg   *Config
		wantU string
	}{
		{
			name: "default scheme and empty vhost",
			cfg: &Config{
				Host:     "127.0.0.1",
				Port:     5672,
				Username: "user",
				Password: "pass#1",
			},
			wantU: "amqp://user:pass%231@127.0.0.1:5672/",
		},
		{
			name: "amqps with vhost",
			cfg: &Config{
				Scheme:   schemeAMQPS,
				Host:     "127.0.0.1",
				Port:     5671,
				Vhost:    "/vh",
				Username: "user",
				Password: "pass",
			},
			wantU: "amqps://user:pass@127.0.0.1:5671/vh",
		},
		{
			name: "vhost without leading slash",
			cfg: &Config{
				Host:     "127.0.0.1",
				Port:     5672,
				Vhost:    "vh",
				Username: "user",
				Password: "pass",
			},
			wantU: "amqp://user:pass@127.0.0.1:5672/vh",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			c.cfg.Scenes = []Scene{{Name: "s"}}
			// 无 broker，connect 必然失败，但 cli 已构造完成，可校验其字段
			cli, err := New(c.cfg, WithLogger(nop.Logger), OpenConsumeLog(), WithPrefetchCount(3))
			assert.NotNil(t, err)
			assert.Contains(t, err.Error(), "amqp.Dial fail")
			assert.NotNil(t, cli)
			assert.Equal(t, c.wantU, cli.url)
			assert.Equal(t, 3, cli.opts.prefetchCount)
			assert.True(t, cli.opts.consumeLog)
			assert.NotNil(t, cli.scenes["s"])
		})
	}
}

func TestIsClosingAndShutdown(t *testing.T) {
	cli := newTestClient()
	assert.False(t, cli.isClosing())

	// connection 为 nil 时 Shutdown 不 panic，且可重复调用
	assert.Nil(t, cli.Shutdown())
	assert.True(t, cli.isClosing())
	assert.Nil(t, cli.Shutdown())
}

func TestReconnect_Closing(t *testing.T) {
	cli := newTestClient()
	_ = cli.Shutdown()

	err := cli.reconnect()
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "client is closing")
}

func TestConn(t *testing.T) {
	cli := newTestClient()
	assert.Nil(t, cli.conn())
}

func TestProduce_AssertFail(t *testing.T) {
	cli := newTestClient()
	_, err := cli.Produce(context.Background(), "msg")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "message assert fail")
}

func TestProduce_SceneNotFound(t *testing.T) {
	cli := newTestClient(Scene{Name: "exist"})
	_, err := cli.Produce(context.Background(), &ProduceMessage{Scene: "not_exist"})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "scene not_exist not found")
}

// TestProduce_ReconnectFail 覆盖 publish/publishConfirm 取 channel 失败的分支。
func TestProduce_ReconnectFail(t *testing.T) {
	cli := newTestClient(
		Scene{Name: "normal"},
		Scene{Name: "confirm", OpenConfirm: true},
	)
	_ = cli.Shutdown()

	for _, scene := range []string{"normal", "confirm"} {
		_, err := cli.Produce(context.Background(), &ProduceMessage{Scene: scene})
		assert.NotNil(t, err)
		assert.Contains(t, err.Error(), "client is closing")
	}
}

func TestConsume_ParamsInvalid(t *testing.T) {
	cli := newTestClient(Scene{Name: "exist"})

	err := cli.Consume("params")
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "params assert fail")

	err = cli.Consume(&ConsumeParams{Context: context.Background()})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "consumer is nil")

	consumer := queue.Consumer(func(ctx context.Context, msg interface{}) (bool, bool, error) {
		return false, false, nil
	})

	err = cli.Consume(&ConsumeParams{Consumer: consumer})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "context is nil")

	err = cli.Consume(&ConsumeParams{
		Context:  context.Background(),
		Scene:    "not_exist",
		Consumer: consumer,
	})
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "scene not_exist not found")
}

// withSmallBackoff 临时调小退避参数，返回恢复函数。
func withSmallBackoff(initial, max, stable time.Duration) func() {
	oi, om, os := consumeInitialBackoff, consumeMaxBackoff, consumeStableDuration
	consumeInitialBackoff, consumeMaxBackoff, consumeStableDuration = initial, max, stable
	return func() {
		consumeInitialBackoff, consumeMaxBackoff, consumeStableDuration = oi, om, os
	}
}

// TestConsume_LoopRetryUntilContextCancel 覆盖订阅中断后持续重试，直到 ctx 取消。
func TestConsume_LoopRetryUntilContextCancel(t *testing.T) {
	defer withSmallBackoff(time.Millisecond, 2*time.Millisecond, time.Hour)()

	cli := newTestClient(Scene{Name: "s", Queue: "q"})
	_ = cli.Shutdown() // 让 newChannel 快速失败，模拟订阅不断中断
	// 重新打开 close，使 consumeLoop 依赖 ctx 结束而非 shutdown
	cli.close = make(chan struct{})
	cli.closeOnce = sync.Once{}
	cli.connection = nil

	var calls int32
	ctx, cancel := context.WithTimeout(logger.InitFieldsContainer(context.Background()), 50*time.Millisecond)
	defer cancel()

	err := cli.consumeLoop(ctx, func() error {
		atomic.AddInt32(&calls, 1)
		return errDeliveriesClosed
	})
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded))
	assert.True(t, atomic.LoadInt32(&calls) > 1, "subscribe should be retried")
}

// TestConsumeLoop_ShutdownReturnNil 覆盖 shutdown 时退出且不返回 error。
func TestConsumeLoop_ShutdownReturnNil(t *testing.T) {
	defer withSmallBackoff(10*time.Millisecond, 10*time.Millisecond, time.Hour)()

	cli := newTestClient()
	ctx := logger.InitFieldsContainer(context.Background())

	// 首次订阅返回后进入退避，退避期间 shutdown
	go func() {
		time.Sleep(5 * time.Millisecond)
		_ = cli.Shutdown()
	}()

	err := cli.consumeLoop(ctx, func() error { return errDeliveriesClosed })
	assert.Nil(t, err)
}

// TestConsumeLoop_ClosedBeforeStart 覆盖进入循环前已 shutdown 的短路分支。
func TestConsumeLoop_ClosedBeforeStart(t *testing.T) {
	cli := newTestClient()
	_ = cli.Shutdown()

	called := false
	err := cli.consumeLoop(logger.InitFieldsContainer(context.Background()), func() error {
		called = true
		return nil
	})
	assert.Nil(t, err)
	assert.False(t, called)
}

// TestConsumeLoop_ContextCanceledBeforeStart 覆盖进入循环前 ctx 已取消的分支。
func TestConsumeLoop_ContextCanceledBeforeStart(t *testing.T) {
	cli := newTestClient()
	ctx, cancel := context.WithCancel(logger.InitFieldsContainer(context.Background()))
	cancel()

	called := false
	err := cli.consumeLoop(ctx, func() error {
		called = true
		return nil
	})
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
	assert.False(t, called)
}

// TestConsumeOnce_NewChannelFail 覆盖订阅前取 channel 失败。
func TestConsumeOnce_NewChannelFail(t *testing.T) {
	cli := newTestClient(Scene{Name: "s", Queue: "q"})
	_ = cli.Shutdown()

	err := cli.consumeOnce(logger.InitFieldsContainer(context.Background()),
		cli.scenes["s"],
		func(ctx context.Context, msg interface{}) (bool, bool, error) { return false, false, nil })
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "client is closing")
}

// fakeAcknowledger 记录 ack/reject 调用，用于校验投递处理的确认行为。
type fakeAcknowledger struct {
	mu       sync.Mutex
	acks     []bool // multiple
	rejects  []bool // requeue
	ackErr   error
	rejctErr error
}

func (f *fakeAcknowledger) Ack(tag uint64, multiple bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.acks = append(f.acks, multiple)
	return f.ackErr
}

func (f *fakeAcknowledger) Nack(tag uint64, multiple, requeue bool) error {
	return nil
}

func (f *fakeAcknowledger) Reject(tag uint64, requeue bool) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rejects = append(f.rejects, requeue)
	return f.rejctErr
}

func TestHandleDelivery(t *testing.T) {
	cases := []struct {
		name        string
		handler     queue.Consumer
		ackErr      error
		rejectErr   error
		wantAcks    []bool
		wantRejects []bool
	}{
		{
			name: "ack",
			handler: func(ctx context.Context, msg interface{}) (bool, bool, error) {
				assert.Equal(t, []byte("body"), msg)
				return false, false, nil
			},
			wantAcks: []bool{false},
		},
		{
			name: "ack error only logs",
			handler: func(ctx context.Context, msg interface{}) (bool, bool, error) {
				return false, false, nil
			},
			ackErr:   errors.New("ack fail"),
			wantAcks: []bool{false},
		},
		{
			name: "reject to dead letter",
			handler: func(ctx context.Context, msg interface{}) (bool, bool, error) {
				return true, false, errors.New("handle fail")
			},
			wantRejects: []bool{false},
		},
		{
			name: "reject error only logs",
			handler: func(ctx context.Context, msg interface{}) (bool, bool, error) {
				return true, false, nil
			},
			rejectErr:   errors.New("reject fail"),
			wantRejects: []bool{false},
		},
		{
			name: "retry requeue",
			handler: func(ctx context.Context, msg interface{}) (bool, bool, error) {
				return false, true, errors.New("handle fail")
			},
			wantRejects: []bool{true},
		},
		{
			name: "reject wins over retry",
			handler: func(ctx context.Context, msg interface{}) (bool, bool, error) {
				return true, true, nil
			},
			wantRejects: []bool{false},
		},
		{
			// panic 后必须归还 prefetch 窗口，requeue=false 避免毒消息无限重投
			name: "panic reject without requeue",
			handler: func(ctx context.Context, msg interface{}) (bool, bool, error) {
				panic("boom")
			},
			wantRejects: []bool{false},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cli := newTestClient()
			cli.opts.consumeLog = true

			ack := &fakeAcknowledger{ackErr: c.ackErr, rejctErr: c.rejectErr}
			d := amqp.Delivery{
				Acknowledger: ack,
				DeliveryTag:  1,
				MessageId:    "msg-id",
				Body:         []byte("body"),
			}

			assert.NotPanics(t, func() {
				cli.handleDelivery(logger.InitFieldsContainer(context.Background()), d, c.handler)
			})

			assert.Equal(t, c.wantAcks, nilIfEmpty(ack.acks))
			assert.Equal(t, c.wantRejects, nilIfEmpty(ack.rejects))
		})
	}
}

func nilIfEmpty(s []bool) []bool {
	if len(s) == 0 {
		return nil
	}
	return s
}

func TestWaitConfirm_Ack(t *testing.T) {
	ack := make(chan uint64, 1)
	nack := make(chan uint64, 1)
	const tag uint64 = 7
	ack <- tag

	got, err := waitConfirm(context.Background(), ack, nack)
	assert.Nil(t, err)
	assert.Equal(t, tag, got)
}

func TestWaitConfirm_Nack(t *testing.T) {
	ack := make(chan uint64, 1)
	nack := make(chan uint64, 1)
	const tag uint64 = 3
	nack <- tag

	got, err := waitConfirm(context.Background(), ack, nack)
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, errPublishConfirmNack))
	assert.Contains(t, err.Error(), "nack")
	assert.Equal(t, tag, got)
}

func TestWaitConfirm_ContextDeadlineExceeded(t *testing.T) {
	ack := make(chan uint64, 1)
	nack := make(chan uint64, 1)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	got, err := waitConfirm(ctx, ack, nack)
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, context.DeadlineExceeded))
	assert.Equal(t, uint64(0), got)
}

func TestWaitConfirm_ContextCanceled(t *testing.T) {
	ack := make(chan uint64, 1)
	nack := make(chan uint64, 1)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	got, err := waitConfirm(ctx, ack, nack)
	assert.NotNil(t, err)
	assert.True(t, errors.Is(err, context.Canceled))
	assert.Equal(t, uint64(0), got)
}

// TestWatchNotify_NotifyClose channel 关闭事件触发重连并退出监听。
func TestWatchNotify_NotifyClose(t *testing.T) {
	cli := newTestClient()
	_ = cli.Shutdown() // reconnect 直接返回 error，避免真实拨号

	notifyClose := make(chan *amqp.Error, 1)
	notifyClose <- &amqp.Error{Code: 504, Reason: "channel/connection is not open"}

	done := make(chan struct{})
	go func() {
		cli.watchNotify(logger.InitFieldsContainer(context.Background()),
			notifyClose, make(chan string, 1), make(chan amqp.Return, 1))
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("watchNotify not return after notifyClose")
	}
}

// TestWatchNotify_ClosedChannelReturn amqp 关闭 notify channel 时必须直接退出，
// 否则会读到零值并打出伪造的 error 日志。
func TestWatchNotify_ClosedChannelReturn(t *testing.T) {
	cases := []struct {
		name  string
		build func() (chan *amqp.Error, chan string, chan amqp.Return)
	}{
		{
			name: "notifyClose closed",
			build: func() (chan *amqp.Error, chan string, chan amqp.Return) {
				c := make(chan *amqp.Error, 1)
				close(c)
				return c, make(chan string, 1), make(chan amqp.Return, 1)
			},
		},
		{
			name: "notifyCancel closed",
			build: func() (chan *amqp.Error, chan string, chan amqp.Return) {
				c := make(chan string, 1)
				close(c)
				return make(chan *amqp.Error, 1), c, make(chan amqp.Return, 1)
			},
		},
		{
			name: "notifyReturn closed",
			build: func() (chan *amqp.Error, chan string, chan amqp.Return) {
				c := make(chan amqp.Return, 1)
				close(c)
				return make(chan *amqp.Error, 1), make(chan string, 1), c
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			cli := newTestClient()
			nc, ncancel, nreturn := c.build()

			done := make(chan struct{})
			go func() {
				cli.watchNotify(logger.InitFieldsContainer(context.Background()), nc, ncancel, nreturn)
				close(done)
			}()

			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("watchNotify not return on closed notify channel")
			}
		})
	}
}

// TestWatchNotify_CancelAndReturnKeepWatching cancel/return 事件不结束监听，直到 shutdown。
func TestWatchNotify_CancelAndReturnKeepWatching(t *testing.T) {
	cli := newTestClient()

	notifyCancel := make(chan string, 1)
	notifyReturn := make(chan amqp.Return, 1)
	notifyCancel <- "consumer-tag"
	notifyReturn <- amqp.Return{ReplyCode: 312, ReplyText: "NO_ROUTE"}

	done := make(chan struct{})
	go func() {
		cli.watchNotify(logger.InitFieldsContainer(context.Background()),
			make(chan *amqp.Error, 1), notifyCancel, notifyReturn)
		close(done)
	}()

	select {
	case <-done:
		t.Fatal("watchNotify should keep watching after cancel/return")
	case <-time.After(50 * time.Millisecond):
	}

	_ = cli.Shutdown()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("watchNotify not return after shutdown")
	}
}

// TestWatchConnection_Shutdown shutdown 后连接监听协程退出。
func TestWatchConnection_Shutdown(t *testing.T) {
	cli := newTestClient()
	conn := &amqp.Connection{}

	cli.watchConnection(conn)
	_ = cli.Shutdown()
	// 无 panic、无 goroutine 泄漏地退出
	time.Sleep(10 * time.Millisecond)
}

// TestNewChannel_ReconnectFail 覆盖 newChannel/newConfirmChannel 的重连失败分支。
func TestNewChannel_ReconnectFail(t *testing.T) {
	cli := newTestClient()
	_ = cli.Shutdown()
	ctx := logger.InitFieldsContainer(context.Background())

	channel, err := cli.newChannel(ctx)
	assert.Nil(t, channel)
	assert.NotNil(t, err)

	channel, ack, nack, err := cli.newConfirmChannel(ctx)
	assert.Nil(t, channel)
	assert.Nil(t, ack)
	assert.Nil(t, nack)
	assert.NotNil(t, err)
}

// TestDialAmqps amqps 场景走 TLS 拨号，无 broker 时返回 error 而非 panic。
func TestDialAmqps(t *testing.T) {
	cli := newTestClient()
	cli.scheme = schemeAMQPS
	cli.insecureSkipVerify = true
	cli.url = "amqps://user:pass@127.0.0.1:1/"

	conn, err := cli.dial()
	assert.Nil(t, conn)
	assert.NotNil(t, err)
}

// TestAddFields 校验日志字段注入。
func TestAddFields(t *testing.T) {
	cli := newTestClient(Scene{Name: "s", Exchange: "e", Queue: "q", Key: "k"})

	ctx := logger.InitFieldsContainer(context.Background())
	cli.addSceneFields(ctx, cli.scenes["s"])
	fields := map[string]any{}
	for _, f := range logger.ExtractFields(ctx) {
		fields[f.Key()] = f.Value()
	}
	assert.Equal(t, "rabbitmq", fields["service"])
	assert.Equal(t, "127.0.0.1", fields["ip"])
	assert.Equal(t, 5672, fields["port"])
	assert.Equal(t, "s", fields["scene"])
	assert.Equal(t, "e", fields["exchange"])
	assert.Equal(t, "q", fields["queue"])
	assert.Equal(t, "k", fields["key"])

	ctx = logger.InitFieldsContainer(context.Background())
	cli.addClientFields(ctx)
	keys := []string{}
	for _, f := range logger.ExtractFields(ctx) {
		keys = append(keys, f.Key())
	}
	assert.Equal(t, "service,ip,port", strings.Join(keys, ","))
}
