package pool

import "errors"

var (
	ErrNewFunc     = errors.New("new func is nil")
	ErrClosed      = errors.New("pool is closed")
	ErrGetTimeout  = errors.New("get connection timeout")
	ErrOverMaxSize = errors.New("over pool max size")
	ErrIDConflict  = errors.New("conn id conflict")
)
