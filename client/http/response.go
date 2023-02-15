package http

import "github.com/why444216978/codec"

type Response struct {
	HTTPCode int
	Body     interface{}
	Codec    codec.Codec
}
