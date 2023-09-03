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

package processors

import (
	"net/http"

	"github.com/xgfone/go-apigateway/pkg/http/runtime"
)

// For Request Header
func init() {
	registerContextDirectiveTwo("setheader",
		"string: the key of the request header argument",
		"string: the value of the request header argument, which is a variable if starting with '$', '@' or '#'",
		func(c *runtime.Context, key, value string) {
			if value, _ = QueryVariable(c, value); value != "" {
				c.UpstreamRequest().Header.Set(key, value)
			}
		},
	)

	registerContextDirectiveTwo("addheader",
		"string: the key of the request header argument",
		"string: the value of the request header argument, which is a variable if starting with '$', '@' or '#'",
		func(c *runtime.Context, key, value string) {
			if value, _ = QueryVariable(c, value); value != "" {
				c.UpstreamRequest().Header.Add(key, value)
			}
		},
	)

	registerRequestDirectiveOne("delheader",
		"string: the key of the request header argument",
		func(r *http.Request, key string) { r.Header.Del(key) },
	)
}

// For Response Header
func init() {
	registerContextDirectiveTwo("setrespheader",
		"string: the key of the response header argument",
		"string: the value of the response header argument, which is a variable if starting with '$', '@' or '#'",
		func(c *runtime.Context, key, value string) {
			if value, _ = QueryVariable(c, value); value != "" {
				c.OnResponseHeader(func() { c.ClientResponse.Header().Set(key, value) })
			}
		},
	)

	registerContextDirectiveTwo("addrespheader",
		"string: the key of the response header argument",
		"string: the value of the response header argument, which is a variable if starting with '$', '@' or '#'",
		func(c *runtime.Context, key, value string) {
			if value, _ = QueryVariable(c, value); value != "" {
				c.OnResponseHeader(func() { c.ClientResponse.Header().Add(key, value) })
			}
		},
	)

	registerContextDirectiveOne("delrespheader",
		"string: the key of the response header argument",
		func(c *runtime.Context, key string) {
			c.OnResponseHeader(func() { c.ClientResponse.Header().Del(key) })
		},
	)
}
