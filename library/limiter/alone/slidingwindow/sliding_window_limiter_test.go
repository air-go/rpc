package slidingwindow

import (
	"context"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/air-go/rpc/library/limiter"
)

func TestLimiter(t *testing.T) {
	ctl := gomock.NewController(t)
	defer ctl.Finish()
	sw := NewMockSlidingWindow(ctl)
	sw.EXPECT().Allow().Times(1).Return(true)
	sw.EXPECT().Allow().Times(1).Return(false)
	sw.EXPECT().SetLimit(gomock.Any()).Times(1)
	sw.EXPECT().SetWindow(gomock.Any()).Times(1)

	patch := gomonkey.ApplyFuncSeq((*slidingWindowLimiter).getLimiter, []gomonkey.OutputCell{
		{Values: gomonkey.Params{sw}},
		{Values: gomonkey.Params{sw}},
		{Values: gomonkey.Params{sw}},
		{Values: gomonkey.Params{sw}},
	})
	defer patch.Reset()

	l := NewLimiter()
	ctx := context.Background()
	r := limiter.Resource{
		Name: "test",
	}
	assert.Equal(t, true, l.Check(ctx, r).Allow())
	assert.Equal(t, false, l.Check(ctx, r).Allow())
	l.SetLimit(ctx, r)
	l.SetBurst(ctx, r)
}

func Test_slidingWindowLimiter_getLimiter(t *testing.T) {
	l := &slidingWindowLimiter{}
	r1 := limiter.Resource{
		Name: "test1",
	}
	_, ok := l.getLimiter(r1).(SlidingWindow)
	assert.Equal(t, true, ok)

	r2 := limiter.Resource{
		Name:  "test2",
		Limit: 1,
		Burst: 1,
	}
	l.limiters.Store(r2.Name, "a")
	_, ok = l.getLimiter(r2).(SlidingWindow)
	assert.Equal(t, true, ok)
}

func Test_slidingWindowLimiter_burst2Window(t *testing.T) {
	l := &slidingWindowLimiter{}
	assert.Equal(t, time.Second, l.burst2Window(1000))
}
