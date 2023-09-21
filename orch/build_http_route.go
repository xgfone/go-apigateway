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

	"github.com/xgfone/go-apigateway/http/core"
	"github.com/xgfone/go-apigateway/http/middleware"
	"github.com/xgfone/go-apigateway/http/router"
)

func buildMiddlewaresHandler(ms Middlewares, next core.Handler) (core.Handler, error) {
	if len(ms) == 0 {
		return next, nil
	}

	var err error
	_ms := make(middleware.Middlewares, len(ms))
	for i, m := range ms {
		_ms[i], err = middleware.DefaultRegistry.Build(m.Name, m.Conf)
		if err != nil {
			return nil, err
		}
	}
	return _ms.Handler(next), err
}

func buildMiddlewareGroupHandler(group string, next core.Handler) (core.Handler, error) {
	if group == "" {
		return next, nil
	}
	return func(c *core.Context) { middleware.HandleGroup(c, group, next) }, nil
}

// Build builds the runtime route by the route config.
func (r HttpRoute) Build() (router.Route, error) {
	if r.Id == "" {
		return router.Route{}, errors.New("missing route id")
	} else if r.Upstream == "" {
		return router.Route{}, fmt.Errorf("route '%s' has no upstream", r.Id)
	}

	matcher, err := r.Matchers.Build()
	if err != nil {
		return router.Route{}, err
	}

	handler := router.AfterRoute
	for _len := len(r.MiddlewareGroups) - 1; _len >= 0; _len-- {
		handler, err = buildMiddlewareGroupHandler(r.MiddlewareGroups[_len], handler)
		if err != nil {
			return router.Route{}, err
		}
	}

	handler, err = buildMiddlewaresHandler(r.Middlewares, handler)
	if err != nil {
		return router.Route{}, err
	}

	extra := r.Extra
	r.Extra = nil

	priority := r.Priority + matcher.Priority()
	return router.Route{
		Priority:   priority,
		UpstreamId: r.Upstream,
		Protect:    r.Protect,
		RouteId:    r.Id,
		Config:     r,
		Extra:      extra,

		RequestTimeout: r.RequestTimeout,
		ForwardTimeout: r.ForwardTimeout,

		Desc:      matcher.String(),
		Matcher:   matcher,
		Handler:   handler,
		Responser: core.StdResponse,
	}, nil
}
