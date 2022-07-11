package redis

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	convey.Convey("TestNewClient", t, func() {
		convey.Convey("success", func() {
			assert.NotNil(t, NewClient(&Config{}))
		})
	})
}
