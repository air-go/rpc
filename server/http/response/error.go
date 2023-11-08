package response

import (
	"fmt"

	"github.com/pkg/errors"
)

var EmptyError = &ResponseError{}

// ResponseError is an response error
type ResponseError struct {
	toast string
	err   error
}

// Toast return toast
func (r *ResponseError) Toast() string { return r.toast }

// SetToast set toast
func (r *ResponseError) SetToast(toast string) { r.toast = toast }

// Error return err string
func (r *ResponseError) Error() string { return r.err.Error() }

// SetError set err
func (r *ResponseError) SetError(err error) { r.err = err }

// Unwrap return err
func (r *ResponseError) Unwrap() error { return r.err }

// Cause return err
func (r *ResponseError) Cause() error { return r.err }

// WrapToast return a new ResponseError
func WrapToast(toast string) *ResponseError {
	return &ResponseError{
		toast: toast,
		err:   errors.New(toast),
	}
}

// WrapToastf return a new format ResponseError
func WrapToastf(toast string, args ...interface{}) *ResponseError {
	return &ResponseError{
		toast: fmt.Sprintf(toast, args...),
		err:   errors.Errorf(toast, args...),
	}
}

func WrapError(err error, toast string) *ResponseError {
	return &ResponseError{
		toast: toast,
		err:   err,
	}
}
