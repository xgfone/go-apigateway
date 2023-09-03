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

func init() {
	registerRequestDirectiveOne("setmethod", "string: change the method of the request",
		func(r *http.Request, s string) { r.Method = s },
	)
}

// GetHeaderValue is used to get the value of the header by the key.
var GetHeaderValue = getHeaderValue

func getHeaderValue(c *runtime.Context, key string) string {
	switch key {
	case "cookie", "Cookie":
		if cname := c.ClientRequest.Header.Get("X-Cookie"); cname != "" {
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
func QueryVariable(c *runtime.Context, variable string) (value string, isvar bool) {
	isvar = true
	switch variable[0] {
	case '@': // For Header
		value = GetHeaderValue(c, variable[1:])

	case '#': // For Query
		value = c.Queries().Get(variable[1:])

	case '$': // May be one of Query, Header, or others,
		variable = variable[1:]
		// 1. Path Argument
		// TODO:

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
