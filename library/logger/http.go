package logger

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/why444216978/go-util/snowflake"

	lc "github.com/air-go/rpc/library/context"
)

// ExtractLogID init log id
func ExtractLogID(req *http.Request) string {
	logID := req.Header.Get(LogHeader)

	if logID == "" {
		logID = NewLogID()
	}

	req.Header.Set(LogHeader, logID)

	return logID
}

func SetLogID(ctx context.Context, header http.Header) (err error) {
	logID := lc.ValueLogID(ctx)
	if logID == "" {
		logID = snowflake.Generate().String()
	}
	header.Set(LogHeader, logID)
	return
}

// GetRequestBody get http request body
func GetRequestBody(req *http.Request) []byte {
	reqBody := []byte{}
	if req.Body != nil {
		reqBody, _ = io.ReadAll(req.Body)
	}
	req.Body = io.NopCloser(bytes.NewBuffer(reqBody)) // Reset

	return reqBody
}
