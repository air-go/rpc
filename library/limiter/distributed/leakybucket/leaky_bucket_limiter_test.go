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
	ctx := context.Background()
	l := NewLimiter(nil)
	r := limiter.Resource{
		Name:  "test",
		Limit: 1,
		Burst: 1,
	}

	ctl := gomock.NewController(t)
	defer ctl.Finish()

	lm := NewMockLeakyBucket(ctl)
	lm.EXPECT().Allow(ctx, r.Name, 1).Times(1).Return(true, nil)
	lm.EXPECT().SetRate(1).Times(1)
	lm.EXPECT().SetVolume(1).Times(1)

	patch := gomonkey.ApplyFuncSeq((*leakyBucketLimiter).getLimiter, []gomonkey.OutputCell{
		{Values: gomonkey.Params{lm}},
		{Values: gomonkey.Params{lm}},
		{Values: gomonkey.Params{lm}},
	})
	defer patch.Reset()

	assert.Equal(t, true, l.Check(ctx, r).Allow())

	l.SetLimit(ctx, r)
	l.SetBurst(ctx, r)
}

func Test_getLimiter(t *testing.T) {
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
