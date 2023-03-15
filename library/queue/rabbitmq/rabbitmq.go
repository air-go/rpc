package rabbitmq

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"github.com/why444216978/go-util/assert"
	panicErr "github.com/why444216978/go-util/panic"

	"github.com/air-go/rpc/library/logger"
	"github.com/air-go/rpc/library/logger/zap"
	"github.com/air-go/rpc/library/queue"
)

type Config struct {
	ServiceName string
	Host        string
	Port        int
	Virtual     string
	User        string
	Pass        string
}

type option struct {
	logger     logger.Logger
	consumeLog bool
	qos        int
}

type optionFunc func(*option)

func defaultOption() *option {
	return &option{
		logger:     zap.StdLogger,
		consumeLog: false,
		qos:        10,
	}
}

func WithLogger(l logger.Logger) optionFunc {
	return func(o *option) { o.logger = l }
}

func OpenConsumeLog() optionFunc {
	return func(o *option) { o.consumeLog = true }
}

func WithQos(qos int) optionFunc {
	return func(o *option) { o.qos = qos }
}

type Client struct {
	opts        *option
	connection  *amqp.Connection
	serviceName string
	url         string
	close       chan struct{}
}

var _ queue.Queue = (*Client)(nil)

func New(config *Config, opts ...optionFunc) (cli *Client, err error) {
	if config == nil {
		err = errors.New("config is nil")
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

	cli = &Client{
		opts:        opt,
		serviceName: config.ServiceName,
		url:         fmt.Sprintf("amqp://%s:%s@%s:%d/%s", config.User, config.Pass, config.Host, config.Port, config.Virtual),
	}

	if err = cli.connect(); err != nil {
		return
	}

	return
}

type ProduceMessage struct {
	Exchange  string // exchange
	Key       string // routing key
	Mandatory bool   // set true, when no queue match Basic.Return
	Immediate bool   // set false, not dependent consumers
	// amqp.Publishing{
	// 	DeliveryMode: amqp.Persistent,
	// 	ContentType:  "text/plain",
	// 	Body:         m,
	// })
	Message     amqp.Publishing
	OpenConfirm bool
}

func (cli *Client) Produce(ctx context.Context, msg interface{}) (
	response queue.ProduceResponse, err error,
) {
	m, ok := msg.(*ProduceMessage)
	if !ok {
		err = errors.New("message assert fail")
		return
	}

	if m.OpenConfirm {
		return cli.publishConfirm(ctx, m)
	}

	return cli.publish(ctx, m)
}

type publishResult struct {
	deliverTag uint64
	success    bool
}

func (cli *Client) publish(ctx context.Context, m *ProduceMessage) (
	response queue.ProduceResponse, err error,
) {
	channel, err := cli.newChannel(ctx)
	if err != nil {
		return
	}
	defer channel.Close()
	if err = channel.Publish(
		m.Exchange,
		m.Key,
		m.Mandatory,
		m.Immediate,
		m.Message,
	); err != nil {
		return
	}

	return
}

func (cli *Client) publishConfirm(ctx context.Context, m *ProduceMessage) (
	response queue.ProduceResponse, err error,
) {
	channel, ack, nack, err := cli.newConfirmChannel(ctx)
	if err != nil {
		return
	}

	if err = channel.Publish(
		m.Exchange,
		m.Key,
		m.Mandatory,
		m.Immediate,
		m.Message,
	); err != nil {
		return
	}

	result := make(chan publishResult)
	select {
	case r := <-ack:
		result <- publishResult{
			deliverTag: r,
			success:    true,
		}
	case r := <-nack:
		result <- publishResult{
			deliverTag: r,
			success:    false,
		}
	}
	r := <-result
	response.Offset = r.deliverTag
	if !r.success {
		err = errors.New("publish confirm nack")
		return
	}

	return
}

//	params := ConsumeParams{
//		queue:     "queue_name",
//		AutoAck:   false,
//		Exclusive: false,
//		NoLocal:   false,
//		NoWait:    false,
//		Args:      nil,
//		Consumer:  queue.Consumer,
//	}
type ConsumeParams struct {
	Context     context.Context
	Queue       string
	AutoAck     bool
	Exclusive   bool
	NoLocal     bool
	NoWait      bool
	Args        amqp.Table
	MultipleAck bool
	Consumer    queue.Consumer
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

	channel, err := cli.newChannel(p.Context)
	if err != nil {
		return
	}
	defer channel.Close()

	deliveries, err := channel.Consume(
		p.Queue,
		p.Queue,
		p.AutoAck,
		p.Exclusive,
		p.NoLocal,
		p.NoWait,
		p.Args,
	)
	if err != nil {
		return
	}

	for d := range deliveries {
		go func(d amqp.Delivery) {
			ctx := context.Background()
			var reject bool
			var retry bool
			var err error
			defer func() {
				if r := recover(); r != nil {
					err := panicErr.NewPanicError(r)
					cli.opts.logger.Error(ctx, "rabbitMQConsumeRecover", logger.Reflect("error", err))
					return
				}
			}()

			reject, retry, err = p.Consumer(ctx, d.Body)
			if err != nil && cli.opts.consumeLog {
				cli.opts.logger.Error(ctx, "rabbitMQHandleConsumerFuncErr",
					logger.Error(err),
					logger.Reflect("retry", retry))
			}

			// If your queue set x-dead-letter-exchange, reject with requeue false can republish to dead letter queue.
			if reject {
				if err = d.Reject(false); err != nil {
					cli.opts.logger.Error(ctx, "rabbitMQConsumeRejectFalseErr", logger.Error(err))
				}
				return
			}

			if retry {
				// Resend to other consumers.
				if err = d.Reject(true); err != nil {
					cli.opts.logger.Error(ctx, "rabbitMQConsumeRejectTrueErr", logger.Error(err))
				}
			} else {
				if err = d.Ack(p.MultipleAck); err != nil {
					cli.opts.logger.Error(ctx, "rabbitMQConsumeAckErr", logger.Error(err))
				}
			}
		}(d)
	}

	return
}

func (cli *Client) Shutdown() (err error) {
	cli.close <- struct{}{}
	return cli.connection.Close()
}

func (cli *Client) newConfirmChannel(ctx context.Context) (
	channel *amqp.Channel, ack chan uint64, nack chan uint64, err error,
) {
	if err = cli.reconnect(); err != nil {
		return
	}

	if channel, err = cli.connection.Channel(); err != nil {
		return
	}

	if err = channel.Qos(cli.opts.qos, 0, false); err != nil {
		_ = channel.Close()
		return
	}

	if err = channel.Confirm(false); err != nil {
		return
	}

	go func() {
		select {
		// Connection.Close or Channel.Close
		case r := <-channel.NotifyClose(make(chan *amqp.Error)):
			cli.opts.logger.Error(ctx, "rabbitMQChannelNotifyCloseErr",
				logger.Error(errors.New(r.Error())),
				logger.Reflect("amqpError", r),
			)
		// Basic.Cancel (consume cancel)
		case r := <-channel.NotifyCancel(make(chan string)):
			cli.opts.logger.Error(ctx, "rabbitMQChannelNotifyCancelErr", logger.Error(errors.New(r)))
		// Basic.Return (publish return)
		case r := <-channel.NotifyReturn(make(chan amqp.Return)):
			cli.opts.logger.Error(ctx, "rabbitMQChannelNotifyReturnErr", logger.Reflect("return", r))
		// shutdown over
		case <-cli.close:
			return
		}
	}()

	ack, nack = channel.NotifyConfirm(make(chan uint64), make(chan uint64))

	return
}

func (cli *Client) newChannel(ctx context.Context) (channel *amqp.Channel, err error) {
	if err = cli.reconnect(); err != nil {
		return
	}
	if channel, err = cli.connection.Channel(); err != nil {
		return
	}

	if err = channel.Qos(cli.opts.qos, 0, false); err != nil {
		_ = channel.Close()
		return
	}

	go func() {
		select {
		// Connection.Close or Channel.Close
		case r := <-channel.NotifyClose(make(chan *amqp.Error)):
			cli.opts.logger.Error(ctx, "rabbitMQChannelNotifyCloseErr",
				logger.Error(errors.New(r.Error())),
				logger.Reflect("amqpError", r),
			)
			_ = cli.reconnect()
		// Basic.Cancel (consume cancel)
		case r := <-channel.NotifyCancel(make(chan string)):
			cli.opts.logger.Error(ctx, "rabbitMQChannelNotifyCancelErr", logger.Error(errors.New(r)))
		// Basic.Return (publish return)
		case r := <-channel.NotifyReturn(make(chan amqp.Return)):
			cli.opts.logger.Error(ctx, "rabbitMQChannelNotifyReturnErr", logger.Reflect("return", r))
		// shutdown over
		case <-cli.close:
			return
		}
	}()

	return
}

func (cli *Client) connect() (err error) {
	cli.connection, err = amqp.Dial(cli.url)
	if err != nil {
		return errors.Wrap(err, "amqp.Dial fail")
	}

	go func() {
		select {
		case r := <-cli.connection.NotifyClose(make(chan *amqp.Error)):
			cli.opts.logger.Error(context.Background(), "rabbitMQConnectionNotifyCloseErr",
				logger.Error(errors.New(r.Error())),
				logger.Reflect("amqpError", r))

			_ = cli.reconnect()
		// shutdown over
		case <-cli.close:
			return
		}
	}()

	return
}

// TODO Implement connection pool.
// Connection and channel are one-to-many.
// Support to control the number of connections.
// Support to control the number of channels in per pool.
// Note: discriminate channel confirm
func (cli *Client) reconnect() (err error) {
	if !cli.connection.IsClosed() {
		return
	}

	if cli.connection, err = amqp.Dial(cli.url); err != nil {
		cli.opts.logger.Error(context.Background(), "rabbitMQReconnectErr", logger.Error(err))
		return
	}

	return
}
