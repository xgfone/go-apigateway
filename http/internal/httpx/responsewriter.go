// Copyright 2023 xgfone
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package httpx

import (
	"fmt"
	"io"
	"net/http"
	"sync"
)

var rwpool = &sync.Pool{New: func() any { return NewResponseWriter(nil) }}

// AcquireResponseWriter acquires a response writer from the pool,
// and sets the http response writer to rw.
func AcquireResponseWriter(rw http.ResponseWriter) *ResponseWriter {
	r := rwpool.Get().(*ResponseWriter)
	r.Reset(rw)
	return r
}

// ReleaseResponseWriter releases the response writer into the pool.
func ReleaseResponseWriter(r *ResponseWriter) {
	r.Reset(nil)
	rwpool.Put(r)
}

// ResponseWriter is used to wrap a http.ResponseWriter.
type ResponseWriter struct {
	http.ResponseWriter

	wroten int
	code   int
}

// NewResponseWriter returns a new ResponseWriter.
func NewResponseWriter(rw http.ResponseWriter) *ResponseWriter {
	return &ResponseWriter{ResponseWriter: rw}
}

// Reset resets the response writer to rw.
func (r *ResponseWriter) Reset(rw http.ResponseWriter) {
	*r = ResponseWriter{ResponseWriter: rw, wroten: 0, code: 0}
}

// Write implements the interface http.ResponseWriter#Write.
func (r *ResponseWriter) Write(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}

	if r.code == 0 {
		r.WriteHeader(http.StatusOK)
	}

	n, err := r.ResponseWriter.Write(p)
	r.wroten += n
	return n, err
}

// WriteString implements the interface io.StringWriter.
func (r *ResponseWriter) WriteString(s string) (int, error) {
	if len(s) == 0 {
		return 0, nil
	}

	if r.code == 0 {
		r.WriteHeader(http.StatusOK)
	}

	n, err := io.WriteString(r.ResponseWriter, s)
	r.wroten += n
	return n, err
}

// Write implements the interface http.ResponseWriter#WriteHeader.
func (r *ResponseWriter) WriteHeader(statusCode int) {
	if statusCode < 100 || statusCode > 999 {
		panic(fmt.Errorf("invalid status code %d", statusCode))
	}

	if r.code == 0 {
		r.code = statusCode
		r.ResponseWriter.WriteHeader(statusCode)
	}
}

// StatusCode returns the response status code.
func (r *ResponseWriter) StatusCode() int {
	if r.code == 0 {
		return http.StatusOK
	}
	return r.code
}

// WroteHeader reports whether the response wrote the header.
func (r *ResponseWriter) WroteHeader() bool { return r.code > 0 }

// Written returns the byte number of the data written into the response.
func (r *ResponseWriter) Written() int { return r.wroten }

// Unwrap unwraps the wrapped original http.ResponseWriter.
func (r *ResponseWriter) Unwrap() http.ResponseWriter { return r.ResponseWriter }
