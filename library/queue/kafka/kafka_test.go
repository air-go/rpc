package kafka

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/air-go/rpc/library/logger"
	"github.com/air-go/rpc/library/logger/nop"
	"github.com/air-go/rpc/library/servicer"
	mk "github.com/air-go/rpc/mock/generate/kafka"
)

func TestNew(t *testing.T) {
	serviceName := "kafka"
	topic := "topic"
	group := "group"
	_ = newClient(t, serviceName, group, topic)
	time.Sleep(time.Second)
}

func TestProduce(t *testing.T) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	syncProducer := mocks.NewSyncProducer(t, config)
	asyncProducer := mocks.NewAsyncProducer(t, config)

	syncProducer.ExpectSendMessageAndSucceed()
	asyncProducer.ExpectInputAndSucceed()

	cli := &Client{
		opts: &options{
			logger: nop.Logger,
		},
		syncProducer:  syncProducer,
		asyncProducer: asyncProducer,
	}
	ctx, cancel := context.WithTimeout(logger.InitFieldsContainer(context.Background()), time.Second*1)
	defer cancel()
	_, err := cli.Produce(ctx, ProduceRequest{
		Async: false,
		Message: &sarama.ProducerMessage{
			Topic: "topic",
			Key:   sarama.StringEncoder("key"),
			Value: sarama.StringEncoder("value"),
		},
	})
	_, err = cli.Produce(ctx, ProduceRequest{
		Async: true,
		Message: &sarama.ProducerMessage{
			Topic: "topic",
			Key:   sarama.StringEncoder("key"),
			Value: sarama.StringEncoder("value"),
		},
	})
	assert.Nil(t, err)
}

type ConsumerGroupHandler struct{}

func (h ConsumerGroupHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h ConsumerGroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (h ConsumerGroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		session.MarkMessage(msg, "")
	}
	return nil
}

func TestConsumeGroup(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	m := mk.NewMockConsumerGroup(ctl)
	m.EXPECT().Errors().AnyTimes().Return(nil)
	m.EXPECT().Consume(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(nil)
	m.EXPECT().Close().Times(1).Return(nil)

	patch := gomonkey.ApplyFuncSeq(sarama.NewConsumerGroupFromClient, []gomonkey.OutputCell{
		{Values: gomonkey.Params{m, nil}},
	})
	defer patch.Reset()

	cli := &Client{
		opts: &options{
			logger: nop.Logger,
		},
	}
	ctx, cancel := context.WithTimeout(logger.InitFieldsContainer(context.Background()), time.Second*1)
	defer cancel()
	err := cli.ConsumeGroup(ConsumeGroupParams{
		Ctx: ctx,
		Consumer: func(ctx context.Context, cm *sarama.ConsumerMessage) (reject bool, err error) {
			return
		},
		GroupID: "",
	})
	assert.NotNil(t, err)
}

func TestConsumePartition(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()

	ch := make(chan *sarama.ConsumerMessage, 1)

	c := mocks.NewConsumer(t, nil)
	pc := c.ExpectConsumePartition("topic", 0, 0)

	m := mk.NewMockConsumer(ctl)
	m.EXPECT().ConsumePartition(gomock.Any(), gomock.Any(), gomock.Any()).AnyTimes().Return(pc, nil)
	// TODO 暂未mock
	// m.EXPECT().HighWaterMarks().Times(1).Return(nil)
	m.EXPECT().Close().Times(1).Return(nil)

	mpo := mk.NewMockPartitionOffsetManager(ctl)
	// TODO 暂未mock
	// mpo.EXPECT().MarkOffset(gomock.Any(), gomock.Any()).Times(1).Return()
	mpo.EXPECT().Close().Times(1).Return(nil)

	mm := mk.NewMockOffsetManager(ctl)
	mm.EXPECT().ManagePartition(gomock.Any(), gomock.Any()).Times(1).Return(mpo, nil)
	mm.EXPECT().Close().Times(1).Return(nil)

	patch := gomonkey.ApplyFuncSeq(sarama.NewConsumerFromClient, []gomonkey.OutputCell{
		{Values: gomonkey.Params{m, nil}},
	})
	patch.ApplyFuncSeq(sarama.NewOffsetManagerFromClient, []gomonkey.OutputCell{
		{Values: gomonkey.Params{mm, nil}},
	})
	defer patch.Reset()

	cli := &Client{
		opts: &options{
			logger: nop.Logger,
		},
	}
	ctx, cancel := context.WithTimeout(logger.InitFieldsContainer(context.Background()), time.Second*1)
	defer cancel()
	err := cli.ConsumePartition(&ConsumePartitionParams{
		Ctx: ctx,
		Consumer: func(ctx context.Context, cm *sarama.ConsumerMessage) (reject bool, err error) {
			return
		},
	})
	ch <- &sarama.ConsumerMessage{}
	assert.Nil(t, err)
}

func newClient(t *testing.T, serviceName, group, topic string) *Client {
	mockBroker := initMockBroker(t, serviceName, group, topic)
	defer mockBroker.Close()

	addr, _ := net.ResolveTCPAddr("tcp", mockBroker.Addr())
	node := servicer.NewNode(addr)

	// servicer mock
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	s := servicer.NewMockServicer(ctl)
	s.EXPECT().Name().AnyTimes().Return("kafka")
	s.EXPECT().GetNodes(gomock.Any()).AnyTimes().Return([]servicer.Node{node})
	_ = servicer.SetServicer(s)

	config := mocks.NewTestConfig()
	config.Version = sarama.V0_10_2_0
	config.Metadata.Full = false
	config.Producer.Return.Successes = true
	config.Consumer.Fetch.Max = sarama.MaxResponseSize

	config.Consumer.Group.Rebalance.Strategy = sarama.BalanceStrategyRange
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = sarama.OffsetNewest
	cli, _ := New(serviceName,
		WithConfig(config),
		WithRefreshInterval(time.Millisecond),
	)
	return cli
}

func initMockBroker(t *testing.T, serviceName, group, topic string) *sarama.MockBroker {
	mockBroker := sarama.NewMockBroker(t, 0)
	mockMetadataResponse := sarama.NewMockMetadataResponse(t).
		SetBroker(mockBroker.Addr(), mockBroker.BrokerID()).
		SetLeader(topic, 0, mockBroker.BrokerID())
	mockProducerResponse := sarama.NewMockProduceResponse(t).
		SetError(topic, 0, sarama.ErrNoError)
	mockOffsetResponse := sarama.NewMockOffsetResponse(t).
		SetOffset(topic, 0, sarama.OffsetOldest, 0).
		SetOffset(topic, 0, sarama.OffsetNewest, 1)
	mockFetchResponse := sarama.NewMockFetchResponse(t, 1).
		SetMessage(topic, 0, 0, sarama.StringEncoder("testing 123")).
		SetMessage(topic, 0, 1, sarama.StringEncoder("testing 123")).
		SetMessage(topic, 0, 2, sarama.StringEncoder("testing 123")).
		SetMessage(topic, 0, 3, sarama.StringEncoder("testing 123")).
		SetMessage(topic, 0, 4, sarama.StringEncoder("testing 123")).
		SetMessage(topic, 0, 5, sarama.StringEncoder("testing 123"))
	mockCoordinatorResponse := sarama.NewMockFindCoordinatorResponse(t).
		SetCoordinator(sarama.CoordinatorType(0), group, mockBroker)
	mockJoinGroupResponse := sarama.NewMockJoinGroupResponse(t)
	mockSyncGroupResponse := sarama.NewMockSyncGroupResponse(t).
		SetMemberAssignment(&sarama.ConsumerGroupMemberAssignment{
			Version:  0,
			Topics:   map[string][]int32{topic: {0}},
			UserData: nil,
		})
	mockHeartbeatResponse := sarama.NewMockHeartbeatResponse(t)
	mockOffsetFetchResponse := sarama.NewMockOffsetFetchResponse(t).
		SetOffset(group, topic, 0, 0, "", sarama.KError(0))

	mockBroker.SetHandlerByMap(map[string]sarama.MockResponse{
		"MetadataRequest":        mockMetadataResponse,
		"ProduceRequest":         mockProducerResponse,
		"OffsetRequest":          mockOffsetResponse,
		"OffsetFetchRequest":     mockOffsetFetchResponse,
		"FetchRequest":           mockFetchResponse,
		"FindCoordinatorRequest": mockCoordinatorResponse,
		"JoinGroupRequest":       mockJoinGroupResponse,
		"SyncGroupRequest":       mockSyncGroupResponse,
		"HeartbeatRequest":       mockHeartbeatResponse,
	})

	addr, _ := net.ResolveTCPAddr("tcp", mockBroker.Addr())
	node := servicer.NewNode(addr)

	ctl := gomock.NewController(t)
	defer ctl.Finish()
	s := servicer.NewMockServicer(ctl)
	s.EXPECT().Name().AnyTimes().Return(serviceName)
	s.EXPECT().GetNodes(gomock.Any()).AnyTimes().Return([]servicer.Node{node})
	_ = servicer.SetServicer(s)

	return mockBroker
}
