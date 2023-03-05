package kafka

import (
	"context"
	"errors"
	"time"

	"github.com/Shopify/sarama"
	"github.com/why444216978/go-util/assert"
	"go.uber.org/multierr"

	"github.com/air-go/rpc/library/logger"
	"github.com/air-go/rpc/library/logger/zap"
	"github.com/air-go/rpc/library/queue"
	"github.com/air-go/rpc/library/servicer"
)

type options struct {
	logger          logger.Logger
	consumeLog      bool
	syncConfig      *sarama.Config
	asyncConfig     *sarama.Config
	consumeConfig   *sarama.Config
	refreshInterval time.Duration
}

type OptionFunc func(*options)

func defaultOptions() *options {
	return &options{
		logger:          zap.StdLogger,
		consumeLog:      false,
		refreshInterval: time.Minute,
	}
}

func RefreshInterval(t time.Duration) OptionFunc {
	return func(o *options) { o.refreshInterval = t }
}

func OpenConsumeLog() OptionFunc {
	return func(o *options) { o.consumeLog = true }
}

func SyncConfig(c *sarama.Config) OptionFunc {
	return func(o *options) { o.syncConfig = c }
}

func AsyncConfig(c *sarama.Config) OptionFunc {
	return func(o *options) { o.asyncConfig = c }
}

func ConsumeConfig(c *sarama.Config) OptionFunc {
	return func(o *options) { o.consumeConfig = c }
}

type Client struct {
	opts          *options
	serviceName   string
	syncProducer  sarama.SyncProducer
	syncClient    sarama.Client
	asyncProducer sarama.AsyncProducer
	asyncClient   sarama.Client
}

var _ queue.Queue = (*Client)(nil)

func New(serviceName string, opts ...OptionFunc) (cli *Client, err error) {
	options := defaultOptions()
	for _, o := range opts {
		o(options)
	}

	cli = &Client{
		serviceName: serviceName,
		opts:        options,
	}

	if err = cli.newSyncProducer(); err != nil {
		return
	}

	if err = cli.newAsyncProducer(); err != nil {
		return
	}

	return
}

type ConsumeParams struct {
	ctx      context.Context
	GroupID  string
	Topics   []string
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

	config := cli.newConsumeConfig()

	addrs, err := cli.addrs()
	if err != nil {
		return
	}

	client, err := sarama.NewClient(addrs, config)
	if err != nil {
		return
	}
	defer client.Close()

	cli.refreshBrokers(client)

	cg, err := sarama.NewConsumerGroupFromClient(p.GroupID, client)
	if err != nil {
		return err
	}
	defer cg.Close()

	ctx := p.ctx
	if assert.IsNil(ctx) {
		ctx = context.Background()
	}

	go func() {
		for err := range cg.Errors() {
			e := &sarama.ConsumerError{}
			if ok := errors.As(err, &e); ok {
				err = e
			}
			cli.opts.logger.Error(ctx, "kafkaConsumerGroupErr", logger.Reflect("error", e.Error()))
		}
	}()

	consumer := Consumer{
		opts:     cli.opts,
		consumer: p.Consumer,
	}
	for {
		// `Consume` should be called inside an infinite loop, when a
		// server-side rebalance happens, the consumer session will need to be
		// recreated to get the new claims
		if err = cg.Consume(ctx, p.Topics, &consumer); err != nil {
			return
		}

		// check if context was cancelled, signaling that the consumer should stop
		if err = ctx.Err(); err != nil {
			return
		}
	}
}

type Consumer struct {
	opts     *options
	consumer queue.Consumer
}

func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case msg := <-claim.Messages():
			go func(msg *sarama.ConsumerMessage) {
				ctx := context.Background()
				var retry bool
				var err error
				defer func() {
					if err := recover(); err != nil {
						c.opts.logger.Error(ctx, "kafkaConsumeRecover", logger.Reflect("error", err))
						return
					}
					if err != nil {
						c.opts.logger.Error(ctx, "kafkaConsumeErr", logger.Error(err))
					}
				}()
				_, retry, err = c.consumer(ctx, msg.Value)
				if err != nil && c.opts.consumeLog {
					c.opts.logger.Error(ctx, "kafkaConsumeRejectErr",
						logger.Error(err),
						logger.Reflect("retry", retry))
				}

				if retry {
					return
				}

				session.MarkMessage(msg, "")
			}(msg)
		// Must: https://github.com/Shopify/sarama/issues/1192
		case <-session.Context().Done():
			return session.Context().Err()
		}
	}
}

type ProduceParams struct {
	Async   bool
	Message *sarama.ProducerMessage
}

func (cli *Client) Produce(ctx context.Context, msg interface{}) (
	response queue.ProduceResponse, err error,
) {
	params, ok := msg.(*ProduceParams)
	if !ok {
		err = errors.New("message assert fail")
		return
	}

	if params.Async {
		cli.asyncProducer.Input() <- params.Message
		return
	}

	if response.Partition, response.Offset, err = cli.syncProducer.SendMessage(params.Message); err != nil {
		return
	}

	return
}

func (cli *Client) Shutdown() (err error) {
	return multierr.Append(
		multierr.Append(
			cli.syncProducer.Close(),
			cli.asyncProducer.Close(),
		),
		multierr.Append(
			cli.syncClient.Close(),
			cli.asyncClient.Close(),
		),
	)
}

func (cli *Client) newAsyncProducer() (err error) {
	config := cli.newAsyncConfig()

	addrs, err := cli.addrs()
	if err != nil {
		return
	}

	if cli.asyncClient, err = sarama.NewClient(addrs, config); err != nil {
		return
	}

	cli.refreshBrokers(cli.asyncClient)

	if cli.asyncProducer, err = sarama.NewAsyncProducerFromClient(cli.asyncClient); err != nil {
		return
	}

	// We will just log to STDOUT if we're not able to produce messages.
	// Messages will only be returned here after all retry attempts are exhausted.
	ctx := context.Background()
	go func() {
		for err := range cli.asyncProducer.Errors() {
			if errors.Is(err, sarama.ErrInsufficientData) {
				continue
			}
			cli.opts.logger.Error(ctx, "kafkaAsyncProducerErr",
				logger.Reflect("error", err.Err.Error()),
				logger.Reflect("message", err.Msg),
			)
		}
	}()

	return
}

func (cli *Client) newSyncProducer() (err error) {
	config := cli.newSyncConfig()

	addrs, err := cli.addrs()
	if err != nil {
		return
	}

	if cli.syncClient, err = sarama.NewClient(addrs, config); err != nil {
		return
	}

	cli.refreshBrokers(cli.syncClient)

	if cli.syncProducer, err = sarama.NewSyncProducerFromClient(cli.syncClient); err != nil {
		return
	}

	return
}

func (cli *Client) newConsumeConfig() *sarama.Config {
	if cli.opts != nil && cli.opts.consumeConfig != nil {
		return cli.opts.consumeConfig
	}

	return sarama.NewConfig()
}

func (cli *Client) newAsyncConfig() *sarama.Config {
	if cli.opts != nil && cli.opts.asyncConfig != nil {
		return cli.opts.asyncConfig
	}

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal       // Only wait for the leader to ack
	config.Producer.Flush.Frequency = 500 * time.Millisecond // Flush batches every 500ms

	return config
}

func (cli *Client) newSyncConfig() *sarama.Config {
	if cli.opts != nil && cli.opts.syncConfig != nil {
		return cli.opts.syncConfig
	}

	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 3
	config.Producer.Return.Successes = true

	return config
}

func (cli *Client) refreshBrokers(client sarama.Client) {
	go func() {
		for range time.NewTicker(cli.opts.refreshInterval).C {
			addrs, err := cli.addrs()
			if err != nil {
				continue
			}
			if len(addrs) < 1 {
				continue
			}
			if err := client.RefreshBrokers(addrs); err != nil {
				cli.opts.logger.Error(context.Background(), "kafkaRefreshBrokersErr", logger.Error(err))
			}
		}
	}()
}

func (cli *Client) addrs() ([]string, error) {
	s, ok := servicer.GetServicer(cli.serviceName)
	if !ok {
		return nil, errors.New("servicer nil")
	}
	nodes, err := s.All(context.Background())
	if err != nil {
		return nil, err
	}
	if len(nodes) < 1 {
		return nil, errors.New("servicer node empty")
	}

	addrs := []string{}
	for _, n := range nodes {
		addrs = append(addrs, n.Address())
	}

	return addrs, nil
}
