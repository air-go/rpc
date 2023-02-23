package rabbitmq

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	"github.com/streadway/amqp"
	"github.com/why444216978/go-util/assert"

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

	channel, err := cli.channel()
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
	Queue     string
	AutoAck   bool
	Exclusive bool
	NoLocal   bool
	NoWait    bool
	Args      amqp.Table
	Consumer  queue.Consumer
}

func (cli *Client) Consume(params interface{}) (err error) {
	p, ok := params.(*ConsumeParams)
	if !ok {
		return errors.New("params assert fail")
	}

	if assert.IsNil(p.Consumer) {
		return errors.New("consumer is nil")
	}

	channel, err := cli.channel()
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
			var retry bool
			var err error
			defer func() {
				if err := recover(); err != nil {
					cli.opts.logger.Error(ctx, "rabbitMQConsumeRecover", logger.Reflect("error", err))
					return
				}
				if err != nil {
					cli.opts.logger.Error(ctx, "rabbitMQConsumeErr", logger.Error(err))
				}
			}()

			retry, err = p.Consumer(ctx, d.Body)
			if err != nil && cli.opts.consumeLog {
				cli.opts.logger.Error(ctx, "rabbitMQConsumeRejectErr",
					logger.Error(err),
					logger.Reflect("retry", retry))
			}

			if retry {
				err = d.Reject(true)
				cli.opts.logger.Error(ctx, "rabbitMQConsumeRejectErr", logger.Error(err))
				return
			}

			if !retry {
				err = d.Ack(true)
				cli.opts.logger.Error(ctx, "rabbitMQConsumeAckErr", logger.Error(err))
			}
		}(d)
	}

	return
}

func (cli *Client) Shutdown() (err error) {
	return cli.connection.Close()
}

func (cli *Client) channel() (channel *amqp.Channel, err error) {
	// TODO channel pool
	if channel, err = cli.connection.Channel(); err != nil {
		return
	}

	if err = channel.Qos(cli.opts.qos, 0, false); err != nil {
		_ = channel.Close()
		return
	}

	ctx := context.Background()
	select {
	case r := <-channel.NotifyClose(make(chan *amqp.Error)):
		cli.opts.logger.Error(ctx, "rabbitMQChannelNotifyCloseErr",
			logger.Error(errors.New(r.Error())),
			logger.Reflect("amqpError", r),
		)
	case r := <-channel.NotifyCancel(make(chan string)):
		cli.opts.logger.Error(ctx, "rabbitMQChannelNotifyCancelErr", logger.Error(errors.New(r)))
	case r := <-channel.NotifyReturn(make(chan amqp.Return)):
		cli.opts.logger.Error(ctx, "rabbitMQChannelNotifyReturnErr", logger.Reflect("return", r))
	default:
	}

	return
}

func (cli *Client) connect() (err error) {
	// TODO connection pool
	cli.connection, err = amqp.Dial(cli.url)
	if err != nil {
		return errors.Wrap(err, "amqp.Dial fail")
	}

	ctx := context.Background()
	select {
	case r := <-cli.connection.NotifyClose(make(chan *amqp.Error)):
		cli.opts.logger.Error(ctx, "rabbitMQConnectionNotifyCloseErr",
			logger.Error(errors.New(r.Error())),
			logger.Reflect("amqpError", r),
		)
	default:
	}

	return
}
