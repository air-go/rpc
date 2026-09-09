package rabbitmq

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

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
