//go:generate go run -mod=mod github.com/golang/mock/mockgen -package mock -source ./lock.go -destination ./mock/lock.go Locker
package cache

import (
	"context"
	"time"

	panicErr "github.com/why444216978/go-util/panic"
)

// CacheData is cache data struct
type CacheData struct {
	ExpireAt int64  // ExpireAt is virtual expire time
	Data     string // Data is cache data
}

// LoadFunc is define load data func
type LoadFunc func(ctx context.Context, target interface{}) (err error)

// Cacher is used to load cache
type Cacher interface {
	// GetData load data from cache
	// if cache not exist load data by LoadFunc
	// ttl is redis server ttl
	// virtualTTL is developer ttl
	GetData(ctx context.Context, key string, ttl time.Duration, virtualTTL time.Duration, f LoadFunc, data interface{}) (err error)

	// FlushCache flush cache
	// if cache not exist, load data and save cache
	FlushCache(ctx context.Context, key string, ttl time.Duration, virtualTTL time.Duration, f LoadFunc, data interface{}) (err error)
}

// HandleLoad is used load cache
func HandleLoad(ctx context.Context, f LoadFunc, data interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = panicErr.NewPanicError(r)
		}
	}()
	err = f(ctx, data)
	return
}
