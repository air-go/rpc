//go:generate go run -mod=mod github.com/golang/mock/mockgen -package mock -source ./lock.go -destination ./mock/lock.go Locker
package lock

import (
	"context"
	"errors"
	"time"
)

var (
	// ErrClientNil 客户端nil
	ErrClientNil = errors.New("client is nil")
	// ErrUnLock 解锁失败
	ErrUnLock = errors.New("unlock fail")
)

type Locker interface {
	Lock(ctx context.Context, key string, random interface{}, duration time.Duration, try int) (ok bool, err error)
	Unlock(ctx context.Context, key string, random interface{}) (err error)
}
