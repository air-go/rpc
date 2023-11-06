package kafka

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/air-go/rpc/library/logger"
	"github.com/air-go/rpc/library/servicer"
	"github.com/air-go/rpc/library/servicer/mock"
)

func TestProduce(t *testing.T) {
	serviceName := "kafka"
	topic := "topic"
	group := "group"

	mockBroker := initMockBroker(t, serviceName, group, topic)
	defer mockBroker.Close()

	cli := newClient(t, serviceName, group, topic)

	sp := mocks.NewSyncProducer(t, nil)
	defer sp.Close()
	sp.ExpectSendMessageAndSucceed()
	cli.syncProducer = sp

	// ap := mocks.NewAsyncProducer(t, nil)
	// defer ap.Close()
	// ap.ExpectInputAndSucceed()
	// cli.asyncProducer = ap

	ctx, cancel := context.WithTimeout(logger.InitFieldsContainer(context.Background()), time.Second*1)
	defer cancel()
	_, err := cli.Produce(ctx, &ProduceParams{
		Async: false,
		Message: &sarama.ProducerMessage{
			Topic: topic,
			Key:   sarama.StringEncoder("key"),
			Value: sarama.StringEncoder("value"),
		},
	})
	_, err = cli.Produce(ctx, &ProduceParams{
		Async: true,
		Message: &sarama.ProducerMessage{
			Topic: topic,
			Key:   sarama.StringEncoder("key"),
			Value: sarama.StringEncoder("value"),
		},
	})
	assert.Nil(t, err)

	assert.NotNil(t, cli.Consume(nil))
	assert.NotNil(t, cli.Consume(&ConsumeParams{
		ctx:      ctx,
		GroupID:  group,
		Topics:   []string{topic},
		Consumer: nil,
	}))

	go func() {
		err = cli.Consume(&ConsumeParams{
			ctx:     ctx,
			GroupID: group,
			Topics:  []string{topic},
			Consumer: func(ctx context.Context, msg interface{}) (reject, retry bool, err error) {
				fmt.Println(msg)
				return
			},
		})
		assert.Nil(t, err)
	}()
	time.Sleep(time.Second)
	_ = cli.Shutdown()
}

func TestConfig(t *testing.T) {
	cli := &Client{}
	cli.newSyncConfig()
	cli.newAsyncProducer()
	cli.newConsumeConfig()
}

func newClient(t *testing.T, serviceName, group, topic string) *Client {
	mockBroker := initMockBroker(t, serviceName, group, topic)
	defer mockBroker.Close()

	addr := mockBroker.Addr()
	arr := strings.Split(addr, ":")
	port, _ := strconv.Atoi(arr[1])
	node := servicer.NewNode(arr[0], port)

	// servicer mock
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	s := mock.NewMockServicer(ctl)
	s.EXPECT().Name().AnyTimes().Return("kafka")
	s.EXPECT().All(gomock.Any()).AnyTimes().Return([]servicer.Node{node}, nil)
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
		SyncConfig(config),
		AsyncConfig(config),
		ConsumeConfig(config),
		RefreshInterval(time.Second),
		OpenConsumeLog(),
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

	addr := mockBroker.Addr()
	arr := strings.Split(addr, ":")
	port, _ := strconv.Atoi(arr[1])
	node := servicer.NewNode(arr[0], port)

	ctl := gomock.NewController(t)
	defer ctl.Finish()
	s := mock.NewMockServicer(ctl)
	s.EXPECT().Name().AnyTimes().Return(serviceName)
	s.EXPECT().All(gomock.Any()).AnyTimes().Return([]servicer.Node{node}, nil)
	_ = servicer.SetServicer(s)

	return mockBroker
}
