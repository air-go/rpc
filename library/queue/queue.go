package queue

import (
	"context"
)

type Consumer func(context.Context, []byte) (reject, retry bool, err error)

type ProduceResponse struct {
	Partition int32
	Offset    uint64
}

type Queue interface {
	Produce(ctx context.Context, msg interface{}) (ProduceResponse, error)
	Consume(params interface{}) error
	Shutdown() error
}
