package service

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestNewServiceLogger(t *testing.T) {
	convey.Convey("TestNewServiceLogger", t, func() {
		convey.Convey("success", func() {
			l, err := NewServiceLogger("test", &Config{})
			assert.Nil(t, err)
			assert.NotNil(t, l)
		})
	})
}
