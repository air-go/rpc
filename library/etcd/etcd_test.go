package etcd

import (
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestNewClient(t *testing.T) {
	convey.Convey("TestNewClient", t, func() {
		convey.Convey("success", func() {
			cli, err := NewClient([]string{
				"127.0.0.1:23790",
				"127.0.0.1:23791",
				"127.0.0.1:23792",
			}, WithDialTimeout(time.Second))

			assert.NotNil(t, cli)
			assert.Nil(t, err)
		})
	})
}
