package tokenbucket

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/air-go/rpc/library/limiter"
)

func TestTokenBucket(t *testing.T) {
	l := NewLimiter()

	r := limiter.Resource{
		Name:  "test",
		Limit: 1,
		Burst: 2,
	}
	ctx := context.Background()
	assert.Equal(t, true, l.Check(ctx, r).Allow())
	assert.Equal(t, true, l.Check(ctx, r).Allow())
	assert.Equal(t, false, l.Check(ctx, r).Allow())

	time.Sleep(time.Second)
	assert.Equal(t, true, l.Check(ctx, r).Allow())
	assert.Equal(t, false, l.Check(ctx, r).Allow())

	r.Limit = 2
	l.SetLimit(ctx, r)
	time.Sleep(time.Second)
	assert.Equal(t, true, l.Check(ctx, r).Allow())
	assert.Equal(t, true, l.Check(ctx, r).Allow())
	assert.Equal(t, false, l.Check(ctx, r).Allow())

	r.Burst = 3
	l.SetBurst(ctx, r)
	time.Sleep(time.Second * 3)
	assert.Equal(t, true, l.Check(ctx, r).Allow())
	assert.Equal(t, true, l.Check(ctx, r).Allow())
	assert.Equal(t, true, l.Check(ctx, r).Allow())
	assert.Equal(t, false, l.Check(ctx, r).Allow())
}

func Test_tokenBucketLimiter_getLimiter(t *testing.T) {
	l := &tokenBucketLimiter{}
	r := limiter.Resource{
		Name:  "test",
		Limit: 1,
		Burst: 2,
	}
	lim := l.getLimiter(r)
	assert.NotNil(t, lim)
	lim = l.getLimiter(r)
	assert.NotNil(t, lim)

	l.limiters.Store("test", 1)
	lim = l.getLimiter(r)
	assert.NotNil(t, lim)
}
