package miniredis

import (
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
)

var cli *miniredis.Miniredis

func init() {
	cli = miniredis.NewMiniRedis()
	if err := cli.Start(); err != nil {
		panic(err)
	}
	go func() {
		for range time.NewTicker(time.Millisecond * 100).C {
			cli.SetTime(time.Now())
			cli.FastForward(time.Millisecond * 100)
		}
	}()
}

func NewClient() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr:         cli.Addr(),
		DB:           0,
		MinIdleConns: 100,
	})
}
