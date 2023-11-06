package http

import (
	"bytes"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/why444216978/codec"
)

type Request interface {
	GetServiceName() string
	GetHeader() http.Header
	SetHeader(h http.Header)
	GetMethod() string
	GetPath() string
	GetQuery() url.Values
	GetBody() interface{}
	GetCodec() codec.Codec
}

var _ Request = (*DefaultRequest)(nil)

type DefaultRequest struct {
	ServiceName string
	Path        string
	Query       url.Values
	Method      string
	Header      http.Header
	Body        interface{}
	Codec       codec.Codec
}

func (r *DefaultRequest) GetServiceName() string {
	return r.ServiceName
}

func (r *DefaultRequest) GetHeader() http.Header {
	return r.Header
}

func (r *DefaultRequest) SetHeader(h http.Header) {
	r.Header = h
}

func (r *DefaultRequest) GetMethod() string {
	return r.Method
}

func (r *DefaultRequest) GetPath() string {
	return r.Path
}

func (r *DefaultRequest) GetQuery() url.Values {
	if r.Query == nil {
		r.Query = url.Values{}
	}
	return r.Query
}

func (r *DefaultRequest) GetBody() interface{} {
	return r.Body
}

func (r *DefaultRequest) GetCodec() codec.Codec {
	return r.Codec
}

type MultiFormFile struct {
	Content io.ReadCloser
	Name    string
}

type MultiRequest struct {
	ServiceName string
	Path        string
	Query       url.Values
	Method      string
	Header      http.Header
	Values      url.Values
	Files       map[string]*MultiFormFile
}

var _ Request = (*MultiRequest)(nil)

func (r *MultiRequest) GetServiceName() string {
	return r.ServiceName
}

func (r *MultiRequest) GetHeader() http.Header {
	return r.Header
}

func (r *MultiRequest) SetHeader(h http.Header) {
	r.Header = h
}

func (r *MultiRequest) GetMethod() string {
	return r.Method
}

func (r *MultiRequest) GetPath() string {
	return r.Path
}

func (r *MultiRequest) GetQuery() url.Values {
	if r.Query == nil {
		r.Query = url.Values{}
	}
	return r.Query
}

func (r *MultiRequest) GetValues() url.Values {
	return r.Values
}

func (r *MultiRequest) GetBody() interface{} {
	return nil
}

func (r *MultiRequest) GetCodec() codec.Codec {
	return r
}

func (r *MultiRequest) Encode(_ interface{}) (io.Reader, error) {
	body := bytes.NewBuffer(nil)
	w := multipart.NewWriter(body)

	if r.Values != nil {
		for k := range r.Values {
			if err := w.WriteField(k, r.Values.Get(k)); err != nil {
				return nil, err
			}
		}
	}

	if r.Files != nil {
		for k, f := range r.Files {
			pw, err := w.CreateFormFile(k, f.Name)
			if err != nil {
				return nil, err
			}
			_, err = io.Copy(pw, f.Content)
			if err != nil {
				return nil, err
			}
		}
		if err := w.Close(); err != nil {
			return nil, err
		}
	}

	if r.Header == nil {
		r.Header = http.Header{}
	}
	r.Header.Set("Content-Type", w.FormDataContentType())

	return body, nil
}

func (r *MultiRequest) Decode(in io.Reader, dst interface{}) error {
	return errors.New("not implement")
}
