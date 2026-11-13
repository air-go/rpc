package logger

import (
	"bytes"
	"io"
	"path/filepath"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestRotateWriter(t *testing.T) {
	convey.Convey("TestRotateWriter", t, func() {
		convey.Convey("success", func() {
			_, _, err := RotateWriter("info.log", "error.log")
			assert.Nil(t, err)
		})
	})
}

func TestResolveWriters(t *testing.T) {
	convey.Convey("TestResolveWriters", t, func() {
		convey.Convey("explicit writers are kept as they are", func() {
			infoBuf := &bytes.Buffer{}
			errBuf := &bytes.Buffer{}

			infoWriter, errWriter, err := ResolveWriters(infoBuf, errBuf, "unused-info.log", "unused-error.log")

			assert.Nil(t, err)
			assert.Same(t, infoBuf, infoWriter)
			assert.Same(t, errBuf, errWriter)
		})

		convey.Convey("nil writers fall back to rotating files", func() {
			dir := t.TempDir()

			infoWriter, errWriter, err := ResolveWriters(nil, nil, filepath.Join(dir, "info.log"), filepath.Join(dir, "error.log"))

			assert.Nil(t, err)
			assert.NotNil(t, infoWriter)
			assert.NotNil(t, errWriter)
		})

		convey.Convey("only the info writer returns an error", func() {
			infoWriter, errWriter, err := ResolveWriters(&bytes.Buffer{}, nil, "info.log", "error.log")

			assert.NotNil(t, err)
			assert.Nil(t, infoWriter)
			assert.Nil(t, errWriter)
		})

		convey.Convey("only the error writer returns an error", func() {
			infoWriter, errWriter, err := ResolveWriters(nil, io.Discard, "info.log", "error.log")

			assert.NotNil(t, err)
			assert.Nil(t, infoWriter)
			assert.Nil(t, errWriter)
		})

		convey.Convey("empty paths keep RotateWriter semantics", func() {
			infoWriter, errWriter, err := ResolveWriters(nil, nil, "", "")

			assert.Nil(t, err)
			assert.NotNil(t, infoWriter)
			assert.NotNil(t, errWriter)
		})

		convey.Convey("invalid file pattern is returned", func() {
			infoWriter, errWriter, err := ResolveWriters(nil, nil, "bad%q.log", "error.log")

			assert.NotNil(t, err)
			assert.Nil(t, infoWriter)
			assert.Nil(t, errWriter)
		})

		convey.Convey("invalid error file pattern is returned", func() {
			infoWriter, errWriter, err := ResolveWriters(nil, nil, "info.log", "bad%q.log")

			assert.NotNil(t, err)
			assert.Nil(t, infoWriter)
			assert.Nil(t, errWriter)
		})
	})
}

func Test_rotateWriter(t *testing.T) {
	convey.Convey("Test_rotateWriter", t, func() {
		convey.Convey("success", func() {
			_, err := rotateWriter("log")
			assert.Nil(t, err)
		})
	})
}
