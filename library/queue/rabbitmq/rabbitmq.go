package rabbitmq

import (
	"context"
	"crypto/tls"
	"net"
	"net/url"
	"runtime/debug"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"github.com/why444216978/go-util/assert"
	"github.com/why444216978/go-util/nopanic"

	"github.com/air-go/rpc/library/logger"
	"github.com/air-go/rpc/library/logger/zap"
	"github.com/air-go/rpc/library/queue"
)

const (
	schemeAMQP  = "amqp"
	schemeAMQPS = "amqps"
)

// 退避参数。声明为 var 仅为便于单测临时改小后恢复，生产运行期不修改。
var (
	consumeInitialBackoff = time.Second
	consumeMaxBackoff     = 30 * time.Second
	// consumeStableDuration 订阅存活超过该时长视为一次成功订阅，退避重置。
	consumeStableDuration = 30 * time.Second
)

// errDeliveriesClosed 表示 broker 侧投递通道被关闭（channel 或连接死亡），需重新订阅。
var errDeliveriesClosed = errors.New("deliveries channel closed")

type Config struct {
	ServiceName string `toml:"service_name"`
	Host        string `toml:"host"`
	Port        int    `toml:"port"`
	Vhost       string `toml:"vhost"`
	Username    string `toml:"username"`
	Password    string `toml:"password"`
	// Scheme 为 "amqp"（明文）或 "amqps"（TLS），为空时默认 "amqp"。
	Scheme string `toml:"scheme"`
	// InsecureSkipVerify 为 true 时跳过 broker 证书校验，仅 amqps 生效，用于自签名证书的内网环境。
	InsecureSkipVerify bool    `toml:"insecure_skip_verify"`
	Scenes             []Scene `toml:"scenes"`
}

// Scene represents one complete produce-consume pipeline.
// Topology (exchange, queue, binding) is declared out-of-band in the management
// console; the client only produces and consumes.
type Scene struct {
	// Name identifies this scene. Produce and Consume calls reference it by name.
	Name string `toml:"name" yaml:"name"`

	Exchange string `toml:"exchange" yaml:"exchange"`
	Queue    string `toml:"queue" yaml:"queue"`
	Key      string `toml:"key" yaml:"key"`         // routing key
	NoWait   bool   `toml:"no_wait" yaml:"no_wait"` // consumed by Channel.Consume

	// --- produce parameters (Channel.Publish) ---
	Mandatory   bool `toml:"mandatory" yaml:"mandatory"`
	OpenConfirm bool `toml:"open_confirm" yaml:"open_confirm"`

	// --- consume parameters (Channel.Consume) ---
	ConsumerTag string     `toml:"consumer_tag" yaml:"consumer_tag"`
	AutoAck     bool       `toml:"auto_ack" yaml:"auto_ack"`
	Exclusive   bool       `toml:"exclusive" yaml:"exclusive"`
	NoLocal     bool       `toml:"no_local" yaml:"no_local"`
	ConsumeArgs amqp.Table `toml:"consume_args" yaml:"consume_args"`
}

type option struct {
	logger        logger.Logger
	consumeLog    bool
	prefetchCount int
}

type optionFunc func(*option)

func defaultOption() *option {
	return &option{
		logger:        zap.StdLogger,
		consumeLog:    false,
		prefetchCount: 10,
	}
}

func WithLogger(l logger.Logger) optionFunc {
	return func(o *option) { o.logger = l }
}

func OpenConsumeLog() optionFunc {
	return func(o *option) { o.consumeLog = true }
}

func WithPrefetchCount(prefetchCount int) optionFunc {
	return func(o *option) { o.prefetchCount = prefetchCount }
}

type Client struct {
	opts               *option
	mu                 sync.RWMutex // 保护 connection
	connection         *amqp.Connection
	serviceName        string
	url                string
	scheme             string
	insecureSkipVerify bool
	host               string
	port               int
	scenes             map[string]*Scene
	close              chan struct{}
	closeOnce          sync.Once // 防重复 Shutdown 造成 close panic
}

var _ queue.Queue = (*Client)(nil)

// conn 并发安全地获取当前连接。
func (cli *Client) conn() *amqp.Connection {
	cli.mu.RLock()
	defer cli.mu.RUnlock()
	return cli.connection
}

// isClosing 报告客户端是否已进入关闭流程。
func (cli *Client) isClosing() bool {
	select {
	case <-cli.close:
		return true
	default:
		return false
	}
}

func New(config *Config, opts ...optionFunc) (cli *Client, err error) {
	if config == nil {
		err = errors.New("config is nil")
		return
	}

	if len(config.Scenes) == 0 {
		err = errors.New("scenes is empty")
		return
	}

	scheme := config.Scheme
	if scheme == "" {
		scheme = schemeAMQP
	}
	if scheme != schemeAMQP && scheme != schemeAMQPS {
		err = errors.Errorf("unsupported scheme %s", config.Scheme)
		return
	}

	opt := defaultOption()
	for _, o := range opts {
		o(opt)
	}

	if assert.IsNil(opt.logger) {
		err = errors.New("logger is nil")
		return
	}

	// 密码含 # 等特殊字符，由 net/url 负责转义；vhost 作为 path，空值等价于默认 vhost "/"
	amqpURL := url.URL{
		Scheme: scheme,
		User:   url.UserPassword(config.Username, config.Password),
		Host:   net.JoinHostPort(config.Host, strconv.Itoa(config.Port)),
		Path:   "/" + strings.TrimPrefix(config.Vhost, "/"),
	}

	cli = &Client{
		opts:               opt,
		serviceName:        config.ServiceName,
		url:                amqpURL.String(),
		scheme:             scheme,
		insecureSkipVerify: config.InsecureSkipVerify,
		host:               config.Host,
		port:               config.Port,
		scenes:             make(map[string]*Scene),
		close:              make(chan struct{}),
	}

	for i := range config.Scenes {
		s := &config.Scenes[i]
		cli.scenes[s.Name] = s
	}

	if err = cli.connect(); err != nil {
		return
	}

	return
}

// ProduceMessage is the produce payload of queue.Queue.
// Topology and publish parameters come from the scene, only the message body is per-call.
//
//	amqp.Publishing{
//		DeliveryMode: amqp.Persistent,
//		ContentType:  "text/plain",
//		Body:         m,
//	}
type ProduceMessage struct {
	// Scene references a scene declared in Config.Scenes by name.
	Scene   string
	Message amqp.Publishing
}

func (cli *Client) Produce(ctx context.Context, msg interface{}) (
	response queue.ProduceResponse, err error,
) {
	m, ok := msg.(*ProduceMessage)
	if !ok {
		err = errors.New("message assert fail")
		return
	}

	s, ok := cli.scenes[m.Scene]
	if !ok {
		err = errors.Errorf("scene %s not found", m.Scene)
		return
	}

	ctx = logger.InitFieldsContainer(ctx)
	cli.addSceneFields(ctx, s)

	if s.OpenConfirm {
		return cli.publishConfirm(ctx, s, m.Message)
	}

	return cli.publish(ctx, s, m.Message)
}

// addSceneFields 为日志上下文补充服务与场景维度字段。
func (cli *Client) addSceneFields(ctx context.Context, s *Scene) {
	logger.AddField(ctx,
		logger.Reflect("service", cli.serviceName),
		logger.Reflect("ip", cli.host),
		logger.Reflect("port", cli.port),
		logger.Reflect("scene", s.Name),
		logger.Reflect("exchange", s.Exchange),
		logger.Reflect("queue", s.Queue),
		logger.Reflect("key", s.Key),
	)
}

// addClientFields 为无场景上下文（连接级）的日志补充服务维度字段。
func (cli *Client) addClientFields(ctx context.Context) {
	logger.AddField(ctx,
		logger.Reflect("service", cli.serviceName),
		logger.Reflect("ip", cli.host),
		logger.Reflect("port", cli.port),
	)
}

func (cli *Client) publish(ctx context.Context, s *Scene, msg amqp.Publishing) (
	response queue.ProduceResponse, err error,
) {
	channel, err := cli.newChannel(ctx)
	if err != nil {
		cli.opts.logger.Error(ctx, "rabbitMQPublishNewChannelErr", logger.Error(err))
		return
	}
	defer channel.Close()
	if err = channel.Publish(
		s.Exchange,
		s.Key,
		s.Mandatory,
		false,
		msg,
	); err != nil {
		cli.opts.logger.Error(ctx, "rabbitMQPublishErr", logger.Error(err))
		return
	}

	cli.opts.logger.Info(ctx, "rabbitMQPublishSucess")

	return
}

var errPublishConfirmNack = errors.New("publish confirm nack")

// waitConfirm blocks until the broker acknowledges (ack) or negatively
// acknowledges (nack) a published message, or ctx is cancelled. It returns the
// delivery tag together with a nil error on ack, a wrapped errPublishConfirmNack
// on nack, or a wrapped ctx.Err() on context cancellation/timeout.
func waitConfirm(ctx context.Context, ack, nack chan uint64) (tag uint64, err error) {
	select {
	case tag = <-ack:
		return
	case tag = <-nack:
		err = errors.Wrapf(errPublishConfirmNack, "deliveryTag=%d", tag)
		return
	case <-ctx.Done():
		err = errors.Wrap(ctx.Err(), "publish confirm wait ack timeout")
		return
	}
}

func (cli *Client) publishConfirm(ctx context.Context, s *Scene, msg amqp.Publishing) (
	response queue.ProduceResponse, err error,
) {
	channel, ack, nack, err := cli.newConfirmChannel(ctx)
	if err != nil {
		cli.opts.logger.Error(ctx, "rabbitMQPublishConfirmNewChannelErr", logger.Error(err))
		return
	}
	defer channel.Close()

	if err = channel.Publish(
		s.Exchange,
		s.Key,
		s.Mandatory,
		false,
		msg,
	); err != nil {
		cli.opts.logger.Error(ctx, "rabbitMQPublishConfirmErr", logger.Error(err))
		return
	}

	tag, err := waitConfirm(ctx, ack, nack)
	if err != nil {
		if errors.Is(err, errPublishConfirmNack) {
			cli.opts.logger.Error(ctx, "rabbitMQPublishConfirmNack", logger.Reflect("deliveryTag", tag))
		} else {
			cli.opts.logger.Error(ctx, "rabbitMQPublishConfirmTimeout", logger.Error(err))
		}
		return
	}

	response.DeliveryTag = tag
	cli.opts.logger.Info(ctx, "rabbitMQPublishConfirmSucess")

	return
}

// ConsumeParams is the consume parameter of queue.Queue.
// Queue and consume flags come from the scene, only the handler is per-call.
type ConsumeParams struct {
	Context context.Context
	// Scene references a scene declared in Config.Scenes by name.
	Scene    string
	Consumer queue.Consumer
}

func (cli *Client) Consume(params interface{}) (err error) {
	p, ok := params.(*ConsumeParams)
	if !ok {
		return errors.New("params assert fail")
	}

	if assert.IsNil(p.Consumer) {
		return errors.New("consumer is nil")
	}

	if assert.IsNil(p.Context) {
		return errors.New("context is nil")
	}

	s, ok := cli.scenes[p.Scene]
	if !ok {
		return errors.Errorf("scene %s not found", p.Scene)
	}

	ctx := logger.InitFieldsContainer(p.Context)
	cli.addSceneFields(ctx, s)

	return cli.consumeLoop(ctx, func() error {
		return cli.consumeOnce(ctx, s, p.Consumer)
	})
}

// consumeLoop 持续调用 subscribe 订阅，直到 ctx 取消或客户端关闭。
// subscribe 正常情况下会一直阻塞；一旦返回即视为订阅中断，退避后重新订阅。
func (cli *Client) consumeLoop(ctx context.Context, subscribe func() error) error {
	backoff := consumeInitialBackoff
	for {
		if cli.isClosing() {
			return nil
		}
		if err := ctx.Err(); err != nil {
			return err
		}

		start := time.Now()
		err := subscribe()
		elapsed := time.Since(start)
		cli.opts.logger.Error(ctx, "rabbitMQConsumeInterrupted",
			logger.Error(err),
			logger.Reflect("elapsed", elapsed.String()))

		// 订阅曾稳定存活，说明不是持续性故障，退避重置
		if elapsed >= consumeStableDuration {
			backoff = consumeInitialBackoff
		}

		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		// shutdown over
		case <-cli.close:
			timer.Stop()
			return nil
		case <-timer.C:
		}

		if backoff < consumeMaxBackoff {
			backoff *= 2
			if backoff > consumeMaxBackoff {
				backoff = consumeMaxBackoff
			}
		}
	}
}

// consumeOnce 重开 channel 并重新发起订阅，阻塞直到投递通道被关闭。
// 每次调用都新建 channel + 重新 Consume，因此 channel 级与连接级断开都能恢复。
func (cli *Client) consumeOnce(ctx context.Context, s *Scene, handler queue.Consumer) error {
	channel, err := cli.newChannel(ctx)
	if err != nil {
		return err
	}
	defer channel.Close()

	deliveries, err := channel.Consume(
		s.Queue,
		s.ConsumerTag,
		s.AutoAck,
		s.Exclusive,
		s.NoLocal,
		s.NoWait,
		s.ConsumeArgs,
	)
	if err != nil {
		return err
	}

	for _d := range deliveries {
		d := _d
		go nopanic.GoVoid(ctx, func() {
			cli.handleDelivery(ctx, d, handler)
		})
	}

	// deliveries 被关闭说明 channel 或连接已死亡，返回哨兵 error 交由上层重新订阅
	return errDeliveriesClosed
}

// handleDelivery 处理单条投递，负责 ack/reject。
// 任何返回路径都必须归还 prefetch 窗口，否则累计达 prefetchCount 后 broker 会静默停止投递。
func (cli *Client) handleDelivery(ctx context.Context, d amqp.Delivery, handler queue.Consumer) {
	ctx = logger.ForkContext(ctx)
	logger.AddLogID(ctx)
	logger.AddField(ctx,
		logger.Reflect("msg_id", d.MessageId),
		logger.Reflect("delivery", d),
		logger.Reflect("msg", string(d.Body)),
	)

	cli.opts.logger.Info(ctx, "rabbitMQConsumeMsg")

	defer func() {
		if r := recover(); r != nil {
			cli.opts.logger.Error(ctx, "rabbitMQConsumeRecover",
				logger.Reflect("error", r),
				logger.Stack(string(debug.Stack())))
			// panic 后必须归还 prefetch 窗口；requeue=false 避免毒消息无限重投
			if err := d.Reject(false); err != nil {
				cli.opts.logger.Error(ctx, "rabbitMQConsumeRecoverRejectErr", logger.Error(err))
			}
		}
	}()

	reject, retry, err := handler(ctx, d.Body)
	if err != nil && cli.opts.consumeLog {
		cli.opts.logger.Error(ctx, "rabbitMQHandleConsumerFuncErr",
			logger.Error(err),
			logger.Reflect("retry", retry))
	}

	// If your queue set x-dead-letter-exchange, reject with requeue false can republish to dead letter queue.
	if reject {
		cli.opts.logger.Info(ctx, "rabbitMQConsumeMsgReject")
		if err = d.Reject(false); err != nil {
			cli.opts.logger.Error(ctx, "rabbitMQConsumeRejectFalseErr", logger.Error(err))
		}
		return
	}

	if retry {
		cli.opts.logger.Info(ctx, "rabbitMQConsumeMsgRetry")
		// Resend to other consumers.
		if err = d.Reject(true); err != nil {
			cli.opts.logger.Error(ctx, "rabbitMQConsumeRejectTrueErr", logger.Error(err))
		}
		return
	}

	cli.opts.logger.Info(ctx, "rabbitMQConsumeMsgAck")
	if err = d.Ack(false); err != nil {
		cli.opts.logger.Error(ctx, "rabbitMQConsumeAckErr", logger.Error(err))
	}
}

func (cli *Client) Shutdown() (err error) {
	cli.closeOnce.Do(func() { close(cli.close) })

	cli.mu.Lock()
	defer cli.mu.Unlock()
	if cli.connection == nil {
		return nil
	}

	return cli.connection.Close()
}

func (cli *Client) newConfirmChannel(ctx context.Context) (
	channel *amqp.Channel, ack chan uint64, nack chan uint64, err error,
) {
	if err = cli.reconnect(); err != nil {
		return
	}

	if channel, err = cli.conn().Channel(); err != nil {
		return
	}

	if err = channel.Qos(cli.opts.prefetchCount, 0, false); err != nil {
		_ = channel.Close()
		return
	}

	if err = channel.Confirm(false); err != nil {
		_ = channel.Close()
		return
	}

	go cli.watchChannel(ctx, channel)

	ack, nack = channel.NotifyConfirm(make(chan uint64), make(chan uint64))

	return
}

// watchChannel 为指定 channel 注册三类 notify 监听，并常驻消费事件。
func (cli *Client) watchChannel(ctx context.Context, channel *amqp.Channel) {
	cli.watchNotify(ctx,
		channel.NotifyClose(make(chan *amqp.Error, 1)),
		channel.NotifyCancel(make(chan string, 1)),
		channel.NotifyReturn(make(chan amqp.Return, 1)),
	)
}

// watchNotify 常驻监听 channel 的关闭/取消/Return 事件，直到 channel 关闭或客户端 shutdown。
// Return 事件仅打印 error 日志，不做死信处理；多事件通过 for 循环持续接收。
// channel 关闭时 amqp 库会同时 close 这三个 notify channel，因此每个分支都必须校验 ok，
// 否则会从已关闭的 channel 读到零值并打出伪造的 error 日志。
func (cli *Client) watchNotify(ctx context.Context, notifyClose chan *amqp.Error,
	notifyCancel chan string, notifyReturn chan amqp.Return,
) {
	for {
		select {
		// Connection.Close or Channel.Close
		case r, ok := <-notifyClose:
			if !ok {
				return
			}
			cli.opts.logger.Error(ctx, "rabbitMQChannelNotifyCloseErr",
				logger.Error(errors.New(r.Error())),
				logger.Reflect("amqpError", r))
			_ = cli.reconnect()
			return
		// Basic.Cancel (consume cancel)
		case r, ok := <-notifyCancel:
			if !ok {
				return
			}
			cli.opts.logger.Error(ctx, "rabbitMQChannelNotifyCancelErr", logger.Error(errors.New(r)))
		// Basic.Return (publish return)
		case r, ok := <-notifyReturn:
			if !ok {
				return
			}
			cli.opts.logger.Error(ctx, "rabbitMQChannelNotifyReturnErr", logger.Reflect("return", r))
		// shutdown over
		case <-cli.close:
			return
		}
	}
}

func (cli *Client) newChannel(ctx context.Context) (channel *amqp.Channel, err error) {
	if err = cli.reconnect(); err != nil {
		return
	}
	if channel, err = cli.conn().Channel(); err != nil {
		return
	}

	if err = channel.Qos(cli.opts.prefetchCount, 0, false); err != nil {
		_ = channel.Close()
		return
	}

	go cli.watchChannel(ctx, channel)

	return
}

// dial 建立连接。amqps 走 TLS，是否校验 broker 证书由配置 insecure_skip_verify 决定；
// amqp 走明文连接。两者 SASL 均为默认的 PLAIN，用户名密码取自 URL 的 userinfo。
func (cli *Client) dial() (*amqp.Connection, error) {
	if cli.scheme == schemeAMQPS {
		return amqp.DialTLS(cli.url, &tls.Config{
			InsecureSkipVerify: cli.insecureSkipVerify, // nolint:gosec // 自签名证书场景由配置显式开启
		})
	}

	return amqp.Dial(cli.url)
}

func (cli *Client) connect() (err error) {
	conn, err := cli.dial()
	if err != nil {
		return errors.Wrap(err, "amqp.Dial fail")
	}

	cli.mu.Lock()
	cli.connection = conn
	cli.mu.Unlock()

	cli.watchConnection(conn)

	return
}

// watchConnection 为指定连接注册 NotifyClose 监听，每条新连接（含重连产生的）都必须调用。
// 注意：本函数不得获取 cli.mu，它会在 reconnect 持锁期间被调用。
func (cli *Client) watchConnection(conn *amqp.Connection) {
	notifyClose := conn.NotifyClose(make(chan *amqp.Error, 1))
	go func() {
		select {
		case r, ok := <-notifyClose:
			if !ok || cli.isClosing() {
				return
			}
			ctx := logger.InitFieldsContainer(context.Background())
			cli.addClientFields(ctx)
			cli.opts.logger.Error(ctx, "rabbitMQConnectionNotifyCloseErr",
				logger.Error(errors.New(r.Error())),
				logger.Reflect("amqpError", r))

			_ = cli.reconnect()
		// shutdown over
		case <-cli.close:
			return
		}
	}()
}

// reconnect 在连接已断开时重新拨号，并为新连接注册监听。
// 加锁 + double-check 保证并发调用只实际拨号一次；客户端关闭后拒绝重连。
func (cli *Client) reconnect() (err error) {
	if cli.isClosing() {
		return errors.New("client is closing")
	}

	cli.mu.Lock()
	defer cli.mu.Unlock()

	// 持锁后再判一次：Shutdown 可能在取锁期间发生
	if cli.isClosing() {
		return errors.New("client is closing")
	}

	if cli.connection != nil && !cli.connection.IsClosed() {
		return
	}

	conn, err := cli.dial()
	if err != nil {
		ctx := logger.InitFieldsContainer(context.Background())
		cli.addClientFields(ctx)
		cli.opts.logger.Error(ctx, "rabbitMQReconnectErr", logger.Error(err))
		return
	}

	cli.connection = conn
	cli.watchConnection(conn)

	return
}
