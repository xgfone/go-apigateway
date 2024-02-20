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

// Package requestid provides a request id middleware
// based on the request header "X-Request-Id".
package requestid

import (
	"net/http"
	"unsafe"

	"github.com/xgfone/go-apigateway/http/core"
	"github.com/xgfone/go-apigateway/http/middleware"
	"github.com/xgfone/go-apigateway/internal/rand"
)

// Generate is used to generate a request id for the http request.
//
// Default: a random string with 24 characters
var Generate func(*http.Request) string = generate

func generate(*http.Request) string {
	const charset = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	var buf [24]byte
	nlen := len(charset)
	for i, _len := 0, len(buf); i < _len; i++ {
		buf[i] = charset[rand.Intn(nlen)]
	}

	return unsafe.String(unsafe.SliceData(buf[:]), len(buf))
}

func _generate(r *http.Request) string { return Generate(r) }

// RequestID returns a http middleware to set the request header "X-Request-Id"
// if not set.
//
// If generate is nil, use Generate instead.
func RequestID(generate func(*http.Request) string) middleware.Middleware {
	generate = _generate
	return middleware.New("requestid", nil, func(next core.Handler) core.Handler {
		return func(c *core.Context) {
			if c.ClientRequest.Header.Get("X-Request-Id") == "" {
				c.ClientRequest.Header.Set("X-Request-Id", generate(c.ClientRequest))
			}
			next(c)
		}
	})
}
