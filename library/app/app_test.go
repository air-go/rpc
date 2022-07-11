package app

import (
	"testing"
	"time"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestApp(t *testing.T) {
	convey.Convey("TestApp", t, func() {
		assert.Equal(t, "", Name())

		assert.Equal(t, "", RegistryName())

		assert.Equal(t, 0, Port())

		assert.Equal(t, false, Pprof())

		assert.Equal(t, false, Debug())

		assert.Equal(t, time.Millisecond*1000, ContextTimeout())
		app.ContextTimeout = 2000
		assert.Equal(t, time.Millisecond*2000, ContextTimeout())

		assert.Equal(t, time.Millisecond*1000, ConnectTimeout())
		app.ConnectTimeout = 2000
		assert.Equal(t, time.Millisecond*2000, ConnectTimeout())

		assert.Equal(t, time.Millisecond*1000, WriteTimeout())
		app.WriteTimeout = 2000
		assert.Equal(t, time.Millisecond*2000, WriteTimeout())

		assert.Equal(t, time.Millisecond*1000, ReadTimeout())
		app.ReadTimeout = 2000
		assert.Equal(t, time.Millisecond*2000, ReadTimeout())
	})
}
