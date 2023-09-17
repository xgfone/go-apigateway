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

// Package allow provides a client-allowed middleware.
package allow

import (
	"fmt"

	"github.com/xgfone/go-apigateway/http/core"
	"github.com/xgfone/go-apigateway/http/middleware"
	"github.com/xgfone/go-apigateway/http/statuscode"
	"github.com/xgfone/go-apigateway/nets"
	"github.com/xgfone/go-binder"
)

func init() {
	middleware.DefaultRegistry.Register("allow", func(name string, conf any) (middleware.Middleware, error) {
		var cidrs []string
		switch vs := conf.(type) {
		case string:
			cidrs = []string{vs}

		case []string:
			cidrs = vs

		case []interface{}:
			var ok bool
			cidrs = make([]string, len(vs))
			for i, v := range vs {
				cidrs[i], ok = v.(string)
				if !ok {
					return nil, fmt.Errorf("Middleware<%s>: expect a string, but got %T", name, v)
				}
			}

		case map[string]any:
			var arg struct {
				Cidrs []string `json:"cidrs"`
			}
			if err := binder.BindStructToMap(&arg, "json", vs); err != nil {
				return nil, fmt.Errorf("Middleware<%s>: %w", name, err)
			}
			cidrs = arg.Cidrs

		default:
			return nil, fmt.Errorf("Middleware<%s>: unsupported config type %T", name, conf)
		}

		return Allow(cidrs...)
	})
}

// Allowlist returns a new middleware named "allow" that only allows the request
// that the client ip is contained in the given cidrs.
func Allow(cidrs ...string) (middleware.Middleware, error) {
	ipchecker, err := nets.NewIPCheckers(cidrs...)
	if err != nil {
		return nil, err
	}

	return middleware.New("allow", cidrs, func(next core.Handler) core.Handler {
		return func(c *core.Context) {
			ip := c.ClientIP()
			if !ip.IsValid() || ipchecker.ContainsAddr(ip) {
				next(c)
			} else {
				err := fmt.Errorf("ip '%s' is not allowed", ip.String())
				c.Abort(statuscode.ErrForbidden.WithError(err))
			}
		}
	}), nil
}
