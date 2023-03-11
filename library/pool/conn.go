package pool

import (
	"context"
)

type Conn interface {
	GetID() string
	Close(ctx context.Context) error
}
