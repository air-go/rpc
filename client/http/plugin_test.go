package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestJaegerBeforePlugin_Handle(t *testing.T) {
	p := &OpentracingBeforePlugin{}
	convey.Convey("TestJaegerBeforePlugin_Handle", t, func() {
		convey.Convey("success", func() {
			_, err := p.Handle(context.Background(), httptest.NewRequest(http.MethodGet, "/", strings.NewReader(``)))
			assert.Nil(t, err)
		})
	})
}
