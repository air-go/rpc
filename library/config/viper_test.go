package config

import (
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	Init("./")
	m.Run()
}

func TestPath(t *testing.T) {
	convey.Convey("TestPath", t, func() {
		assert.Equal(t, "./", Path())
	})
}

func TestDir(t *testing.T) {
	convey.Convey("TestDir", t, func() {
		d, err := Dir()
		assert.Nil(t, err)
		assert.NotEqual(t, "", d)
	})
}

func TestConfig(t *testing.T) {
	convey.Convey("TestConfig", t, func() {
		assert.Equal(t, defaultConf, Config())
	})
}
