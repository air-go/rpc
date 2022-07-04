package logger

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
)

func TestExtractLogID(t *testing.T) {
	convey.Convey("TestExtractLogID", t, func() {
		convey.Convey("header exists", func() {
			request := httptest.NewRequest(http.MethodGet, "/", bytes.NewBufferString(""))
			request.Header.Set(LogHeader, "id")

			id := ExtractLogID(request)

			assert.Equal(t, id, "id")
		})
		convey.Convey("header non-exists", func() {
			request := httptest.NewRequest(http.MethodGet, "/", bytes.NewBufferString(""))
			id := ExtractLogID(request)

			assert.NotEmpty(t, id)
		})
	})
}

func TestGetRequestBody(t *testing.T) {
	convey.Convey("TestExtractLogID", t, func() {
		convey.Convey("success", func() {
			body := []byte(`{"a":"a"}`)
			request := httptest.NewRequest(http.MethodGet, "/", bytes.NewBuffer(body))
			res := GetRequestBody(request)

			assert.Equal(t, res, body)

			b, _ := ioutil.ReadAll(request.Body)
			assert.Equal(t, b, body)
		})
	})
}
