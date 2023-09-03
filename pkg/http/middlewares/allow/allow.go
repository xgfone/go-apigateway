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

	"github.com/xgfone/go-apigateway/pkg/http/runtime"
	"github.com/xgfone/go-apigateway/pkg/nets"
	"github.com/xgfone/go-binder"
)

// Priority is the default priority of the middleware.
const Priority = 200

func init() {
	runtime.RegisterMiddlewareBuilder("allow", func(name string, conf map[string]any) (runtime.Middleware, error) {
		var arg struct {
			Cidrs []string `json:"cidrs"`
		}
		if err := binder.BindStructToMap(&arg, "json", conf); err != nil {
			return nil, err
		}
		return Allow(name, Priority, arg.Cidrs...)
	})
}

// Allowlist returns a new middleware that only allows the request
// that the client ip is contained in the given cidrs.
func Allow(name string, priority int, cidrs ...string) (runtime.Middleware, error) {
	ipchecker, err := nets.NewIPCheckers(cidrs...)
	if err != nil {
		return nil, err
	}

	return runtime.NewMiddleware(name, priority, cidrs, func(h runtime.Handler) runtime.Handler {
		return func(c *runtime.Context) {
			ip := c.ClientIP()
			if ipchecker.ContainsAddr(ip) {
				h(c)
			} else {
				err := fmt.Errorf("ip '%s' is not allowed", ip.String())
				c.SendResponse(nil, runtime.ErrForbidden.WithError(err))
			}
		}
	}), nil
}
