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

package core

import (
	"io"
	"net/http"
	"slices"

	innerhttpx "github.com/xgfone/go-apigateway/http/internal/httpx"
	"github.com/xgfone/go-loadbalancer/httpx"
)

var (
	// AcquireResponseWriter is used to customize the wrapper
	// from http.ResponseWriter to ResponseWriter.
	AcquireResponseWriter func(http.ResponseWriter) ResponseWriter = acquireResponseWriter

	// ReleaseResponseWriter is used to release the response writer
	// acquired by AcquireResponseWriter.
	ReleaseResponseWriter func(ResponseWriter) = releaseResponseWriter
)

func acquireResponseWriter(w http.ResponseWriter) ResponseWriter {
	return innerhttpx.AcquireResponseWriter(w)
}

func releaseResponseWriter(w ResponseWriter) {
	if r, ok := w.(*innerhttpx.ResponseWriter); ok {
		innerhttpx.ReleaseResponseWriter(r)
	}
}

// ResponseWriter is the extension of http.ResponseWriter.
type ResponseWriter interface {
	http.ResponseWriter
	WroteHeader() bool
	StatusCode() int
}

// Responser is used to handle the response from the upstream server to the client.
type Responser func(*Context, *http.Response, error)

// StdResponse is a standard response handler.
func StdResponse(c *Context, resp *http.Response, err error) {
	switch e := err.(type) {
	case nil:
		if resp != nil {
			CopyResponse(c, resp)
		}

	case http.Handler:
		e.ServeHTTP(c.ClientResponse, c.ClientRequest)

	default:
		sendtext(c.ClientResponse, 500, e.Error())
	}
}

func sendtext(w http.ResponseWriter, code int, msg string) {
	if len(msg) == 0 {
		w.WriteHeader(code)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(code)
	_, _ = io.WriteString(w, msg)
}

// CopyResponseHeader copies the response header from the upstream server
// to the client.
func CopyResponseHeader(c *Context, resp *http.Response) {
	header := c.ClientResponse.Header()
	for k, vs := range resp.Header {
		switch {
		case len(vs) == 0:
		case len(vs) == 1 && vs[0] == "":
		case slices.Contains(httpx.DefaultFilterededHeaders, k):
		default:
			header[k] = vs
		}
	}
	c.CallbackOnResponseHeader()
}

// CopyResponse copies the response header and body from the upstream server
// to the client.
func CopyResponse(c *Context, resp *http.Response) {
	CopyResponseHeader(c, resp)
	_ = httpx.HandleResponseBody(c.ClientResponse, resp)
}
