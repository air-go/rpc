package job

import (
	"context"
	"log"

	"github.com/pkg/errors"
	"github.com/why444216978/go-util/assert"
	"github.com/why444216978/go-util/snowflake"

	"github.com/air-go/rpc/library/logger"
)

type HandleFunc func(ctx context.Context) error

var Handlers = map[string]HandleFunc{}

func Handle(job string, l logger.Logger) {
	ctx := logger.WithLogID(context.Background(), snowflake.Generate().String())

	log.Println("start job by " + job)

	handle, ok := Handlers[job]
	if !ok {
		log.Println("job " + job + " not found")
		return
	}

	err := handle(ctx)
	if err == nil {
		return
	}

	if !assert.IsNil(l) {
		l.Error(ctx, errors.Wrap(err, "handle job "+job).Error())
	}
}
