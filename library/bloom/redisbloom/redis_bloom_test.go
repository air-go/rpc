package redisbloom

import (
	"context"
	"testing"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"

	"github.com/air-go/rpc/mock/tools/miniredis"
)

func TestBloom(t *testing.T) {
	ctx := context.Background()

	r := miniredis.NewClient()

	b, _ := NewRedisBloom(func() *redis.Client {
		return r
	},
		SetEstimateParameters(10000, 0.01),
	)

	key := "key"

	ok, err := b.Check(ctx, key, []byte("abc"), time.Minute)
	assert.Nil(t, err)
	assert.Equal(t, false, ok)

	err = b.Add(ctx, key, []byte("abc"), time.Minute)
	assert.Nil(t, err)

	ok, err = b.Check(ctx, key, []byte("abc"), time.Minute)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)

	ok, err = b.CheckAndAdd(ctx, key, []byte("abc"), time.Minute)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)

	ok, err = b.CheckAndAdd(ctx, key, []byte("abcd"), time.Minute)
	assert.Nil(t, err)
	assert.Equal(t, false, ok)

	ok, err = b.Check(ctx, key, []byte("abcd"), time.Minute)
	assert.Nil(t, err)
	assert.Equal(t, true, ok)
}
