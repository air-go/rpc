package slidinglog

import (
	"context"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	gomock "github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/air-go/rpc/library/limiter"
)

func TestLimiter(t *testing.T) {
	ctx := context.Background()
	r := limiter.Resource{
		Name:   "test",
		Limit:  1,
		Window: time.Second,
	}

	ctl := gomock.NewController(t)
	defer ctl.Finish()

	sl := NewMockSlidingLog(ctl)
	sl.EXPECT().Allow(ctx, r.Name).Times(1).Return(true, nil)
	sl.EXPECT().SetLimit(int64(r.Limit)).Times(1)
	sl.EXPECT().SetWindow(r.Window).Times(1)

	patch := gomonkey.ApplyFuncSeq((*slidingLogLimiter).getLimiter, []gomonkey.OutputCell{
		{Values: gomonkey.Params{sl}},
		{Values: gomonkey.Params{sl}},
		{Values: gomonkey.Params{sl}},
	})
	defer patch.Reset()

	l := NewLimiter(nil)

	e := l.Check(ctx, r)
	assert.Nil(t, e.Error())
	assert.Equal(t, true, e.Allow())

	l.SetLimit(ctx, r)
	l.SetWindow(ctx, r)
}

func Test_getLimiter(t *testing.T) {
	l := &slidingLogLimiter{}
	r1 := limiter.Resource{
		Name:  "test1",
		Limit: 1,
		Burst: 1,
	}

	_, ok := l.getLimiter(r1).(SlidingLog)
	assert.Equal(t, true, ok)

	r2 := limiter.Resource{
		Name:  "test2",
		Limit: 1,
		Burst: 1,
	}
	l.limiters.Store(r2.Name, "a")
	_, ok = l.getLimiter(r2).(SlidingLog)
	assert.Equal(t, true, ok)
}
