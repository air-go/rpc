package kafka

import (
	"context"
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	"github.com/pkg/errors"
	"github.com/why444216978/go-util/assert"
	"github.com/why444216978/go-util/nopanic"
	uruntime "github.com/why444216978/go-util/runtime"
	"go.uber.org/multierr"

	"github.com/air-go/rpc/library/logger"
	"github.com/air-go/rpc/library/logger/nop"
	"github.com/air-go/rpc/library/servicer"
)

type ConsumerFunc func(context.Context, *sarama.ConsumerMessage) (reject bool, err error)

type options struct {
	logger          logger.Logger
	config          *sarama.Config
	refreshInterval time.Duration
	concurrent      int
}

type OptionFunc func(*options)

func defaultOptions() *options {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Flush.Frequency = 500 * time.Millisecond
	config.Producer.Return.Successes = true
	config.Consumer.Return.Errors = true
	config.ChannelBufferSize = 1000

	return &options{
		logger:          nop.Logger,
		config:          config,
		refreshInterval: 30 * time.Second,
		concurrent:      100,
	}
}

func WithRefreshInterval(t time.Duration) OptionFunc {
	return func(o *options) { o.refreshInterval = t }
}

func WithConfig(c *sarama.Config) OptionFunc {
	return func(o *options) { o.config = c }
}

func WithLogger(l logger.Logger) OptionFunc {
	return func(o *options) { o.logger = l }
}

func WithConcurrent(c int) OptionFunc {
	return func(o *options) { o.concurrent = c }
}

type Client struct {
	sarama.Client
	serviceName   string
	opts          *options
	syncProducer  sarama.SyncProducer
	asyncProducer sarama.AsyncProducer
}

func New(serviceName string, opts ...OptionFunc) (cli *Client, err error) {
	options := defaultOptions()
	for _, o := range opts {
		o(options)
	}

	cli = &Client{
		serviceName: serviceName,
		opts:        options,
	}

	addrs, err := cli.getAddrs()
	if err != nil {
		return
	}

	if cli.Client, err = sarama.NewClient(addrs, cli.opts.config); err != nil {
		return
	}

	if err = cli.newSyncProducer(); err != nil {
		return
	}

	if err = cli.newAsyncProducer(); err != nil {
		return
	}

	cli.refreshBrokers()

	return
}

type ConsumeGroupParams struct {
	Ctx      context.Context
	Consumer ConsumerFunc
	GroupID  string
	Topics   []string
}

func (cli *Client) ConsumeGroup(p ConsumeGroupParams) (err error) {
	if assert.IsNil(p.Consumer) {
		return errors.New("consumer is nil")
	}

	cg, err := sarama.NewConsumerGroupFromClient(p.GroupID, cli.Client)
	if err != nil {
		return err
	}
	defer cg.Close()

	ctx := p.Ctx
	if assert.IsNil(ctx) {
		ctx = context.Background()
	}

	go func() {
		for err := range cg.Errors() {
			e := &sarama.ConsumerError{}
			if ok := errors.As(err, &e); ok {
				err = e
			}
			cli.opts.logger.Error(ctx, "kafkaConsumerGroupErr",
				logger.Reflect(logger.Module, logger.ModuleKafka),
				logger.Reflect(logger.Method, "ConsumePartition"),
				logger.Error(err),
			)
		}
	}()

	consumer := Consumer{
		p:        p,
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

type ConsumePartitionParams struct {
	Ctx       context.Context
	Consumer  ConsumerFunc
	Topic     string
	GroupID   string
	Partition int32
	Offset    int64
}

func (cli *Client) ConsumePartition(p *ConsumePartitionParams) (err error) {
	if assert.IsNil(p.Consumer) {
		return errors.New("consumer is nil")
	}

	consumer, err := sarama.NewConsumerFromClient(cli.Client)
	if err != nil {
		return
	}
	defer consumer.Close()

	offsetManager, err := sarama.NewOffsetManagerFromClient(p.GroupID, cli.Client)
	if err != nil {
		return
	}
	defer offsetManager.Close()

	partitionOffsetManager, err := offsetManager.ManagePartition(p.Topic, p.Partition)
	if err != nil {
		return
	}
	defer partitionOffsetManager.Close()

	pc, err := consumer.ConsumePartition(p.Topic, p.Partition, p.Offset)
	if err != nil {
		return
	}
	defer pc.AsyncClose()

	ctx := p.Ctx
	if assert.IsNil(ctx) {
		ctx = context.Background()
	}

	go func() {
		for err := range pc.Errors() {
			cli.opts.logger.Error(ctx, "kafkaConsumePartitionErr",
				logger.Reflect(logger.Module, logger.ModuleKafka),
				logger.Reflect(logger.Method, "ConsumePartition"),
				logger.Reflect(logger.API, p.Topic),
				logger.Reflect("group_id", p.GroupID),
				logger.Reflect("partition", p.Partition),
				logger.Error(err),
			)
		}
	}()

	//g := nopanic.New(context.Background(),
	//	nopanic.SetConcurrent(100),
	//)
	for {
		select {
		case msg := <-pc.Messages():
			ctx := logger.InitFieldsContainer(context.Background())
			logger.AddLogID(ctx)

			uuid := fmt.Sprintf("%s-%s-%d-%d", msg.Topic, p.GroupID, msg.Partition, msg.Offset)

			start := time.Now()
			for i := 1; i <= 3; i++ {
				reject, err := handleMessage(ctx, p.Consumer, msg)
				if err != nil {
					cli.opts.logger.Error(ctx, "ConsumePartitionMessageError",
						logger.Reflect(logger.Module, logger.ModuleKafka),
						logger.Reflect(logger.Method, "ConsumePartition"),
						logger.Reflect(logger.API, p.Topic),
						logger.Reflect(logger.UUID, uuid),
						logger.Reflect(logger.Cost, time.Since(start).Milliseconds()),
						logger.Reflect("group_id", p.GroupID),
						logger.Reflect("partition", p.Partition),
						logger.Reflect("offset", msg.Offset),
						logger.Reflect("msg_value", string(msg.Value)),
						logger.Error(err),
					)
				}
				if reject {
					break
				}
			}
			consumer.HighWaterMarks()
			if msg.Offset%100 == 0 {
				partitionOffsetManager.MarkOffset(msg.Offset, "")
			}

			cli.opts.logger.Info(ctx, "ConsumeClaimInfo",
				logger.Reflect(logger.Module, logger.ModuleKafka),
				logger.Reflect(logger.Method, "ConsumePartition"),
				logger.Reflect(logger.API, p.Topic),
				logger.Reflect(logger.UUID, uuid),
				logger.Reflect(logger.Cost, time.Since(start).Milliseconds()),
				logger.Reflect("group_id", p.GroupID),
				logger.Reflect("partition", p.Partition),
				logger.Reflect("offset", msg.Offset),
				logger.Error(err),
			)
			return nil
		case <-ctx.Done():
			cli.opts.logger.Error(ctx, "ConsumePartitionContextDoneErr",
				logger.Reflect(logger.Module, logger.ModuleKafka),
				logger.Reflect(logger.Method, "ConsumePartition"),
				logger.Reflect(logger.API, p.Topic),
				logger.Reflect("group_id", p.GroupID),
				logger.Reflect("partition", p.Partition),
				logger.Error(err),
			)

			return
		}
	}
}

type Consumer struct {
	p        ConsumeGroupParams
	opts     *options
	consumer ConsumerFunc
}

func (c *Consumer) Setup(s sarama.ConsumerGroupSession) error {
	c.opts.logger.Info(context.Background(), "SetupInfo",
		logger.Reflect(logger.Module, logger.ModuleKafka),
		logger.Reflect(logger.Method, "Setup"),
		logger.Reflect("topics", c.p.Topics),
		logger.Reflect("group_id", c.p.GroupID),
		logger.Reflect("claims", s.Claims()),
	)
	return nil
}

func (c *Consumer) Cleanup(s sarama.ConsumerGroupSession) error {
	s.Commit()
	c.opts.logger.Info(context.Background(), "CleanupInfo",
		logger.Reflect(logger.Module, logger.ModuleKafka),
		logger.Reflect(logger.Method, "Cleanup"),
		logger.Reflect("topics", c.p.Topics),
		logger.Reflect("group_id", c.p.GroupID),
		logger.Reflect("claims", s.Claims()),
	)
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	g := nopanic.New(context.Background(),
		nopanic.SetConcurrent(c.opts.concurrent),
	)
	for _msg := range claim.Messages() {
		if err := session.Context().Err(); err != nil {
			c.opts.logger.Error(session.Context(), "ConsumeClaimSessionContextDone",
				logger.Reflect(logger.Module, logger.ModuleKafka),
				logger.Reflect(logger.Method, "ConsumeClaim"),
				logger.Reflect("group_id", c.p.GroupID),
				logger.Reflect("topics", c.p.Topics),
				logger.Error(err),
			)
			return err
		}
		msg := _msg
		// https://github.com/Shopify/sarama/blob/main/consumer_group.go#L29-L31
		g.Go(func() error {
			ctx := logger.InitFieldsContainer(context.Background())
			logger.AddLogID(ctx)

			uuid := fmt.Sprintf("%s-%s-%d-%d", msg.Topic, c.p.GroupID, msg.Partition, msg.Offset)

			start := time.Now()
			for i := 1; i <= 3; i++ {
				reject, err := handleMessage(ctx, c.consumer, msg)
				if err != nil {
					c.opts.logger.Error(ctx, "ConsumeClaimHandleMessageError",
						logger.Reflect(logger.Module, logger.ModuleKafka),
						logger.Reflect(logger.Method, "ConsumeClaim"),
						logger.Reflect(logger.API, msg.Topic),
						logger.Reflect(logger.UUID, uuid),
						logger.Reflect("group_id", c.p.GroupID),
						logger.Reflect("partition", msg.Partition),
						logger.Reflect("offset", msg.Offset),
						logger.Error(err),
					)
				}
				if reject {
					break
				}
			}
			session.MarkMessage(msg, "")
			if msg.Offset%100 == 0 {
				session.Commit()
			}

			c.opts.logger.Info(ctx, "ConsumeClaimInfo",
				logger.Reflect(logger.Module, logger.ModuleKafka),
				logger.Reflect(logger.Method, "ConsumeClaim"),
				logger.Reflect(logger.API, msg.Topic),
				logger.Reflect(logger.UUID, uuid),
				logger.Reflect("group_id", c.p.GroupID),
				logger.Reflect("partition", msg.Partition),
				logger.Reflect("offset", msg.Offset),
				logger.Reflect("cost", time.Since(start).Milliseconds()),
			)
			return nil
		})
	}
	return g.Wait()
}

func handleMessage(ctx context.Context, f ConsumerFunc, msg *sarama.ConsumerMessage) (reject bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = uruntime.WrapStackError(r)
			return
		}
	}()
	fReject, fErr := f(ctx, msg)
	if fErr != nil {
		return fReject, fErr
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

type ProduceRequest struct {
	Async   bool
	Message *sarama.ProducerMessage
}

type ProduceResponse struct {
	Partition int32
	Offset    uint64
}

func (cli *Client) Produce(ctx context.Context, request ProduceRequest) (
	response ProduceResponse, err error,
) {
	if request.Async {
		cli.asyncProducer.Input() <- request.Message
		return
	}

	var offset int64
	if response.Partition, offset, err = cli.syncProducer.SendMessage(request.Message); err != nil {
		cli.opts.logger.Error(ctx, "ProduceError",
			logger.Reflect(logger.Module, logger.ModuleKafka),
			logger.Reflect(logger.Method, "Produce"),
			logger.Reflect(logger.API, request.Message.Topic),
			logger.Error(err),
		)
		return
	}
	cli.opts.logger.Info(ctx, "ProduceInfo",
		logger.Reflect(logger.Module, logger.ModuleKafka),
		logger.Reflect(logger.Method, "Produce"),
		logger.Reflect(logger.API, request.Message.Topic),
		logger.Reflect("partition", response.Partition),
		logger.Reflect("offset", offset),
		logger.Error(err),
	)
	response.Offset = uint64(offset)

	return
}

type BatchProduceRequest struct {
	Async    bool
	Messages []*sarama.ProducerMessage
}

type BatchProduceResponse struct{}

func (cli *Client) BatchProduce(ctx context.Context, request BatchProduceRequest) (
	response BatchProduceResponse, err error,
) {
	if request.Async {
		for _, msg := range request.Messages {
			cli.asyncProducer.Input() <- msg
		}
		return
	}

	if err = cli.syncProducer.SendMessages(request.Messages); err != nil {
		cli.opts.logger.Error(ctx, "BatchProduceError",
			logger.Reflect(logger.Module, logger.ModuleKafka),
			logger.Reflect(logger.Method, "BatchProduce"),
			logger.Error(err),
		)
		return
	}

	cli.opts.logger.Info(ctx, "BatchProduceInfo",
		logger.Reflect(logger.Module, logger.ModuleKafka),
		logger.Reflect(logger.Method, "BatchProduce"),
		logger.Error(err),
	)

	return
}

func (cli *Client) Shutdown() (err error) {
	return multierr.Append(
		multierr.Append(
			cli.syncProducer.Close(),
			cli.asyncProducer.Close(),
		),
		cli.Close(),
	)
}

func (cli *Client) newAsyncProducer() (err error) {
	if cli.asyncProducer, err = sarama.NewAsyncProducerFromClient(cli.Client); err != nil {
		return
	}

	// We will just log to STDOUT if we're not able to produce messages.
	// Messages will only be returned here after all retry attempts are exhausted.
	// ctx := context.Background()
	go func() {
		for err := range cli.asyncProducer.Errors() {
			if errors.Is(err, sarama.ErrInsufficientData) {
				continue
			}
			cli.opts.logger.Error(context.Background(), "asyncProducerError",
				logger.Reflect(logger.Module, logger.ModuleKafka),
				logger.Reflect(logger.Method, "asyncProduce"),
				logger.Error(err),
			)
		}
	}()

	return
}

func (cli *Client) newSyncProducer() (err error) {
	if cli.syncProducer, err = sarama.NewSyncProducerFromClient(cli.Client); err != nil {
		return
	}

	return
}

func (cli *Client) refreshBrokers() {
	go func() {
		for range time.NewTicker(cli.opts.refreshInterval).C {
			addrs, err := cli.getAddrs()
			if err != nil {
				continue
			}
			if len(addrs) < 1 {
				continue
			}
			if err := cli.RefreshBrokers(addrs); err != nil {
				cli.opts.logger.Error(context.Background(), "refreshBrokersError",
					logger.Reflect(logger.Module, logger.ModuleKafka),
					logger.Reflect(logger.Method, "refreshBrokers"),
					logger.Error(err),
				)
			}
		}
	}()
}

func (cli *Client) getAddrs() ([]string, error) {
	s, ok := servicer.GetServicer(cli.serviceName)
	if !ok {
		return nil, errors.New("servicer nil")
	}
	nodes := s.GetNodes(context.Background())
	if len(nodes) < 1 {
		return nil, errors.New("servicer node empty")
	}

	addrs := []string{}
	for _, n := range nodes {
		addrs = append(addrs, n.Addr().String())
	}

	return addrs, nil
}

func (cli *Client) Stop() error {
	return cli.Close()
}
