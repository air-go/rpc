package leakybucket

import (
	"context"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/air-go/rpc/library/limiter"
)

func TestLimiter(t *testing.T) {
	l := NewLimiter()
	r := limiter.Resource{
		Name:  "test",
		Limit: 1,
		Burst: 1,
	}

	ctl := gomock.NewController(t)
	defer ctl.Finish()

	lm := NewMockLeakyBucket(ctl)
	lm.EXPECT().Allow().Times(1).Return(true)
	lm.EXPECT().SetLimit(1).Times(1)
	lm.EXPECT().SetBurst(1).Times(1)

	patch := gomonkey.ApplyFuncSeq((*leakyBucketLimiter).getLimiter, []gomonkey.OutputCell{
		{Values: gomonkey.Params{lm}},
		{Values: gomonkey.Params{lm}},
		{Values: gomonkey.Params{lm}},
	})
	defer patch.Reset()

	ctx := context.Background()
	assert.Equal(t, true, l.Check(ctx, r).Allow())

	l.SetLimit(ctx, r)
	l.SetBurst(ctx, r)
}

func Test_leakyBucketLimiter_getLimiter(t *testing.T) {
	l := &leakyBucketLimiter{}
	r1 := limiter.Resource{
		Name:  "test1",
		Limit: 1,
		Burst: 1,
	}

	_, ok := l.getLimiter(r1).(LeakyBucket)
	assert.Equal(t, true, ok)

	r2 := limiter.Resource{
		Name:  "test2",
		Limit: 1,
		Burst: 1,
	}
	l.limiters.Store(r2.Name, "a")
	_, ok = l.getLimiter(r2).(LeakyBucket)
	assert.Equal(t, true, ok)
}
