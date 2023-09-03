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
	"strings"

	"github.com/xgfone/go-apigateway/pkg/http/runtime"
)

func init() {
	registerContextDirectiveTwo("addquery",
		"string: the key of the query argument",
		"string: the value of the query argument, which is a variable if starting with '$', '@' or '#'",
		func(c *runtime.Context, key, value string) {
			if value, _ = QueryVariable(c, value); value == "" {
				return
			}

			kvs := strings.Join([]string{key, value}, "=")
			if req := c.UpstreamRequest(); req.URL.RawQuery == "" {
				req.URL.RawQuery = kvs
			} else {
				req.URL.RawQuery = strings.Join([]string{req.URL.RawQuery, kvs}, "&")
			}
		},
	)
}
