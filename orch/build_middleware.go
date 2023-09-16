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

package orch

import (
	"errors"
	"fmt"

	"github.com/xgfone/go-apigateway/http/middleware"
)

// HttpBuild builds the middleware based on http.
func (m Middleware) HttpBuild() (middleware.Middleware, error) {
	if m.Name == "" {
		return nil, errors.New("HttpMiddleware: missing name")
	}
	return middleware.DefaultRegistry.Build(m.Name, m.Conf)
}

// HttpBuild builds a middleware group based on http.
func (g MiddlewareGroup) HttpBuild() (*middleware.Group, error) {
	if g.Name == "" {
		return nil, errors.New("HttpMiddlewareGroup: missing name")
	}

	var err error
	ms := make(middleware.Middlewares, len(g.Middlewares))
	for i, m := range g.Middlewares {
		if ms[i], err = m.HttpBuild(); err != nil {
			return nil, fmt.Errorf("HttpMiddlewareGroup<%s>: %w", g.Name, err)
		}
	}

	return middleware.NewGroup(g.Name, ms...), nil
}
