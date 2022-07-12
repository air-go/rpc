package response

import (
	"fmt"

	"github.com/pkg/errors"
)

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
func WrapToast(err error, toast string) *ResponseError {
	if err == nil {
		return &ResponseError{
			err:   errors.New(toast),
			toast: toast,
		}
	}

	return &ResponseError{
		err:   err,
		toast: toast,
	}
}

// WrapToastf return a new format ResponseError
func WrapToastf(err error, toast string, args ...interface{}) *ResponseError {
	if err == nil {
		return &ResponseError{
			err:   errors.Errorf(toast, args...),
			toast: fmt.Sprintf(toast, args...),
		}
	}

	return &ResponseError{
		err:   errors.Wrapf(err, toast, args...),
		toast: fmt.Sprintf(toast, args...),
	}
}
