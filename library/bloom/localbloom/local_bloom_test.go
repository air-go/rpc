package localbloom

import (
	"context"
	"testing"
	"time"

	"github.com/benbjohnson/clock"
	"github.com/stretchr/testify/assert"
)

func TestMemoryBloom(t *testing.T) {
	ctx := context.Background()

	c := clock.NewMock()
	c.Set(time.Now())

	b := NewMemoryBloom(
		SetEstimateParameters(10000, 0.01),
		SetClock(c),
		SetReleaseDuration(time.Nanosecond),
	)
	key := "key"

	ok, err := b.Check(ctx, key, []byte("abc"), time.Millisecond)
	assert.Nil(t, err)
	assert.Equal(t, false, ok)

	err = b.Add(ctx, key, []byte("abc"), time.Millisecond)
	assert.Nil(t, err)

	ok, err = b.Check(ctx, key, []byte("abc"), time.Millisecond)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)

	ok, err = b.CheckAndAdd(ctx, key, []byte("abcd"), time.Millisecond)
	assert.Nil(t, err)
	assert.Equal(t, false, ok)

	ok, err = b.Check(ctx, key, []byte("abcd"), time.Millisecond)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)

	c.Add(time.Minute * 2)
	ok, err = b.Check(ctx, key, []byte("abc"), time.Millisecond)
	assert.Nil(t, err)
	assert.Equal(t, false, ok)
	ok, err = b.Check(ctx, key, []byte("abcd"), time.Millisecond)
	assert.Nil(t, err)
	assert.Equal(t, false, ok)
}
