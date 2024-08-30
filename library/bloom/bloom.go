package bloom

import (
	"context"
	"time"
)

type Bloom interface {
	Add(ctx context.Context, key string, data []byte, ttl time.Duration) error
	// Check is return whether it exists or not
	Check(ctx context.Context, key string, data []byte, ttl time.Duration) (bool, error)
	// CheckAndAdd is return whether it exists, if not exists add.
	CheckAndAdd(ctx context.Context, key string, data []byte, ttl time.Duration) (bool, error)
}
