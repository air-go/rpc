package limiter

import (
	"context"
)

type Resource struct {
	Name  string
	Limit int
	Burst int
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
}
