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

package runtime

import (
	"io"
	"net/http"
	"slices"

	"github.com/xgfone/go-loadbalancer/http/processor"
)

// ResponseHandler is used to handle the response from the upstream server to the client.
type ResponseHandler func(*Context, *http.Response, error)

// StdResponse is a standard response handler.
func StdResponse(c *Context, resp *http.Response, err error) {
	switch e := err.(type) {
	case nil:
		if resp != nil {
			CopyResponse(c, resp)
		}

	case Error:
		sendtext(c.ClientResponse, e.Code, e.Error())

	default:
		sendtext(c.ClientResponse, 500, e.Error())
	}
}

func sendtext(w http.ResponseWriter, code int, msg string) {
	if len(msg) > 0 {
		w.Header().Set("Content-Type", "text/plain")
	}

	w.WriteHeader(code)

	if len(msg) > 0 {
		_, _ = io.WriteString(w, msg)
	}
}

// CopyResponseHeader copies the response header from the upstream server
// to the client.
func CopyResponseHeader(c *Context, resp *http.Response) {
	header := c.ClientResponse.Header()
	for k, vs := range resp.Header {
		switch {
		case len(vs) == 0:
		case len(vs) == 1 && vs[0] == "":
		case slices.Contains(processor.DefaultFilterededHeaders, k):
		default:
			header[k] = vs
		}
	}
	c.callbackOnResponseHeader()
}

// CopyResponse copies the response header and body from the upstream server
// to the client.
func CopyResponse(c *Context, resp *http.Response) {
	CopyResponseHeader(c, resp)
	_ = processor.HandleResponseBody(c.ClientResponse, resp)
	c.callbackOnResponseBody()
}
