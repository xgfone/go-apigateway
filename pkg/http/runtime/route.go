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
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/xgfone/go-apigateway/pkg/http/dynamicconfig"
	"github.com/xgfone/go-defaults"
)

// Route represents a runtime route.
type Route struct {
	Route dynamicconfig.Route `json:"route" yaml:"route"`

	response  ResponseHandler
	matcher   matcher
	mwgroup   string
	mwhandler Handler
}

// Match reports whether the route matches the http request or not.
func (r *Route) Match(c *Context) bool {
	return r.matcher.Match(c)
}

// Handle handles and forwards the http request by the route.
func (r *Route) Handle(c *Context) {
	if r.Route.Timeout > 0 {
		var cancel context.CancelFunc
		c.Context, cancel = context.WithTimeout(c.Context, r.Route.Timeout*time.Second)
		defer cancel()
	}

	defer wrapPanic(c)

	switch {
	case r.mwhandler != nil:
		r.mwhandler(c)

	case r.mwgroup != "":
		DefaultMiddlewareGroupManager.Handle(c, r.mwgroup, UpstreamForward)

	default:
		UpstreamForward(c)
	}
}

func wrapPanic(c *Context) {
	r := recover()
	switch e := r.(type) {
	case nil:
		return
	case error:
		c.SendResponse(nil, ErrInternalServerError.WithError(e))
	default:
		c.SendResponse(nil, ErrInternalServerError.WithError(fmt.Errorf("%v", e)))
	}
	defaults.HandlePanic(r)
}

// ------------------------------------------------------------------------- //

func handleRouteMiddlewareGroup(c *Context) {
	switch {
	case c.Route.mwgroup != "":
		DefaultMiddlewareGroupManager.Handle(c, c.Route.mwgroup, UpstreamForward)

	default:
		UpstreamForward(c)
	}
}

func buildRouteMiddlewares(r dynamicconfig.Route) (Handler, error) {
	if len(r.Middlewares) == 0 {
		return nil, nil
	}
	return buildMiddlewaresHandler(handleRouteMiddlewareGroup, r.Middlewares)
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
