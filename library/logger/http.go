package logger

import (
	"bytes"
	"io/ioutil"
	"net/http"

	"github.com/why444216978/go-util/snowflake"
)

// ExtractLogID init log id
func ExtractLogID(req *http.Request) string {
	logID := req.Header.Get(LogHeader)

	if logID == "" {
		logID = snowflake.Generate().String()
	}

	req.Header.Add(LogHeader, logID)

	return logID
}

// GetRequestBody get http request body
func GetRequestBody(req *http.Request) []byte {
	reqBody := []byte{}
	if req.Body != nil {
		reqBody, _ = ioutil.ReadAll(req.Body)
	}
	req.Body = ioutil.NopCloser(bytes.NewBuffer(reqBody)) // Reset

	return reqBody
}
