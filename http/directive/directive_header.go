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

package directive

import (
	"net/http"
	"strings"

	"github.com/xgfone/go-apigateway/http/core"
)

// For Request Header
func init() {
	DefaultRegistry.RegisterHeaderKeyValue("setheader",
		"string: the key of the request header argument",
		"string: the value of the request header argument, which is a variable if starting with '$', '@' or '#'",
		func(c *core.Context, key, value string) {
			if value, _ = QueryVariable(c, value); value != "" {
				c.UpstreamRequest.Header.Set(key, value)
			}
		},
	)

	DefaultRegistry.RegisterHeaderKeyValue("addheader",
		"string: the key of the request header argument",
		"string: the value of the request header argument, which is a variable if starting with '$', '@' or '#'",
		func(c *core.Context, key, value string) {
			if value, _ = QueryVariable(c, value); value != "" {
				c.UpstreamRequest.Header.Add(key, value)
			}
		},
	)

	DefaultRegistry.RegisterHeaderKey("delheader",
		"string: the key of the request header argument",
		func(c *core.Context, key string) { c.UpstreamRequest.Header.Del(key) },
	)
}

// For Response Header
func init() {
	DefaultRegistry.RegisterHeaderKeyValue("setrespheader",
		"string: the key of the response header argument",
		"string: the value of the response header argument, which is a variable if starting with '$', '@' or '#'",
		func(c *core.Context, key, value string) {
			if value, _ = QueryVariable(c, value); value != "" {
				c.OnResponseHeader(func() { c.ClientResponse.Header().Set(key, value) })
			}
		},
	)

	DefaultRegistry.RegisterHeaderKeyValue("addrespheader",
		"string: the key of the response header argument",
		"string: the value of the response header argument, which is a variable if starting with '$', '@' or '#'",
		func(c *core.Context, key, value string) {
			if value, _ = QueryVariable(c, value); value != "" {
				c.OnResponseHeader(func() { c.ClientResponse.Header().Add(key, value) })
			}
		},
	)

	DefaultRegistry.RegisterHeaderKey("delrespheader",
		"string: the key of the response header argument, which supports the prefix pattern ending with '*'",
		func(c *core.Context, key string) {
			if last := len(key) - 1; key[last] == '*' {
				key := key[:last]
				c.OnResponseHeader(func() {
					header := c.ClientResponse.Header()
					for _key := range header {
						if strings.HasPrefix(_key, key) {
							header.Del(_key)
						}
					}
				})
			} else {
				c.OnResponseHeader(func() { c.ClientResponse.Header().Del(key) })
			}
		},
	)
}

// RegisterHeaderKey is the same as RegisterOneArg,
// but formats the header key argument beforehead.
func (r *Registry) RegisterHeaderKey(directive string, argDesc string, handle func(c *core.Context, key string)) {
	r.Register(directive, []string{argDesc}, func(directive string, args ...string) (Processor, error) {
		if err := checkOneArgs(directive, args); err != nil {
			return nil, err
		}

		key := http.CanonicalHeaderKey(args[0])
		return ProcessorFunc(func(c *core.Context) {
			handle(c, key)
		}), nil
	})
}

// RegisterHeaderKeyValue is the same as RegisterTwoArgs,
// but formats the header key argument beforehead.
func (r *Registry) RegisterHeaderKeyValue(directive string, keyDesc, valueDesc string, handle func(c *core.Context, key, value string)) {
	r.Register(directive, []string{keyDesc, valueDesc}, func(directive string, args ...string) (Processor, error) {
		if err := checkTwoArgs(directive, args); err != nil {
			return nil, err
		}

		key, value := http.CanonicalHeaderKey(args[0]), args[1]
		return ProcessorFunc(func(c *core.Context) {
			handle(c, key, value)
		}), nil
	})
}
