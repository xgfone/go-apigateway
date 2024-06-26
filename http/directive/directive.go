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

// Package directive provides a request or response processor based on the directive.
package directive

import (
	"fmt"
	"net/http"

	"github.com/xgfone/go-apigateway/http/core"
)

var (
	// GetHeaderValue is used to get the value of the header by the key.
	GetHeaderValue = getHeaderValue

	// GetCookieName is used by GetHeaderValue to get the cookie name.
	GetCookieName = getCookieName
)

func getCookieName(r *http.Request) string {
	return r.Header.Get("X-Cookie")
}

func getHeaderValue(c *core.Context, key string) string {
	switch key {
	case "cookie", "Cookie":
		if cname := GetCookieName(c.ClientRequest); cname != "" {
			key = cname
		}
		return c.Cookie(key)

	default:
		return c.ClientRequest.Header.Get(key)
	}
}

// QueryVariable queries the value of the variable
// if it starting with "@", "#" or "$".
//
// Return "" if not found the variable.
func QueryVariable(c *core.Context, variable string) (value string, isvar bool) {
	isvar = true
	switch variable[0] {
	case '@': // For Header
		value = GetHeaderValue(c, variable[1:])

	case '#': // For Query
		value = c.Queries().Get(variable[1:])

	case '$': // May be one of Query, Header, or others,
		variable = variable[1:]

		// 1. Context Kvs
		if v, ok := c.Kvs[variable]; ok {
			if value, ok = v.(string); !ok {
				value = fmt.Sprint(v)
			}
			return
		}

		// 2. Query
		if value = c.Queries().Get(variable); value != "" {
			break
		}

		// 3. Header
		if value = GetHeaderValue(c, variable); value != "" {
			break
		}

	default:
		return variable, false
	}

	return
}
