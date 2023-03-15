package limiter

import (
	"context"
	"time"
)

type Resource struct {
	Name   string
	Limit  int
	Burst  int
	Window time.Duration
}

type Entry interface {
	Allow() bool
	Finish()
	Error() error
}

type Limiter interface {
	// Check check can process
	Check(ctx context.Context, r Resource) Entry
	// SetLimit set a new limit
	SetLimit(ctx context.Context, r Resource)
	// SetBurst set a new burst
	SetBurst(ctx context.Context, r Resource)
	// SetWindow set a new window
	SetWindow(ctx context.Context, r Resource)
}
