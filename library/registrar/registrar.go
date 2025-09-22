package registrar

import (
	"context"

	"github.com/air-go/rpc/library/logger"
)

// Registrar is service registrar
type Registrar interface {
	Register(ctx context.Context) error
	DeRegister(ctx context.Context) error
	SetLogger(l logger.Logger)
}
