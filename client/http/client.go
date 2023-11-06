package http

import (
	"context"
)

type Client interface {
	Send(ctx context.Context, request Request, response Response) (err error)
}
