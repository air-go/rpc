package job

import (
	"context"
	"testing"

	"github.com/air-go/rpc/library/logger/zap"
	"github.com/pkg/errors"
	"github.com/smartystreets/goconvey/convey"
)

func handle(ctx context.Context) error { return nil }

func handleErr(ctx context.Context) error { return errors.New("error") }

func TestHandle(t *testing.T) {
	Handlers = map[string]HandleFunc{
		"success": handle,
		"err":     handleErr,
	}
	convey.Convey("TestHandle", t, func() {
		convey.Convey("not found", func() {
			Handle("test", zap.StdLogger)
		})
		convey.Convey("err", func() {
			Handle("err", zap.StdLogger)
		})
		convey.Convey("success", func() {
			Handle("success", zap.StdLogger)
		})
	})
}
