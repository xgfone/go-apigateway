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

package runtime

import (
	"errors"
	"fmt"

	"github.com/xgfone/go-apigateway/pkg/http/dynamicconfig"
)

// Route represents a runtime route.
type Route struct {
	Route dynamicconfig.Route `json:"route" yaml:"route"`

	response  ResponseHandler
	matcher   matcher
	mwgroup   string
	mwhandler Handler
}

// NewRoute builds the runtime route by the config.
func NewRoute(r dynamicconfig.Route) (route Route, err error) {
	if r.Id == "" {
		err = errors.New("missing route id")
		return
	} else if r.Upstream == "" {
		err = fmt.Errorf("route '%s' has no upstream", r.Id)
		return
	}

	matcher, err := buildRouteMatcher(r)
	if err != nil {
		return
	}

	handler, err := buildRouteMiddlewares(r)
	if err != nil {
		return
	}

	return Route{
		Route: r,

		matcher:   matcher,
		mwgroup:   r.MiddlewareGroup,
		mwhandler: handler,
		response:  StdResponse,
	}, nil
}

func buildRouteMiddlewares(r dynamicconfig.Route) (Handler, error) {
	if len(r.Middlewares) == 0 {
		return nil, nil
	}
	return buildMiddlewaresHandler(handleRouteMiddlewareGroup, r.Middlewares)
}

func handleRouteMiddlewareGroup(c *Context) {
	switch {
	case c.Route.mwgroup != "":
		DefaultMiddlewareGroupManager.Handle(c, c.Route.mwgroup, UpstreamForward)

	default:
		UpstreamForward(c)
	}
}
