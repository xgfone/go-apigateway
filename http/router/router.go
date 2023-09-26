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

// Package router provides an entrypoint router for the api gateway.
package router

import (
	"context"
	"fmt"
	"log/slog"
	"maps"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xgfone/go-apigateway/http/core"
	"github.com/xgfone/go-apigateway/http/middleware"
	"github.com/xgfone/go-apigateway/http/statuscode"
	"github.com/xgfone/go-defaults"
)

// DefaultRouter is the default global http router.
var DefaultRouter = New()

type routeswrapper struct{ Routes []Route }

// Router represents a http router handler
// to match and forward the http request to the upstream.
type Router struct {
	lock   sync.Mutex
	routem map[string]Route

	allmap atomic.Value // map[string]Route
	routes atomic.Pointer[routeswrapper]

	gmddlws middleware.Middlewares
	handler core.Handler
}

// New returns a new router.
func New() *Router {
	r := &Router{routem: make(map[string]Route, 32)}
	r.allmap.Store(map[string]Route(nil))
	r.routes.Store(new(routeswrapper))
	r.handler = r.serve
	return r
}

// Use appends the global middlewares that act on all the routes,
// which is not thread-safe and should be used only before running.
func (r *Router) Use(mws ...middleware.Middleware) {
	r.gmddlws = append(r.gmddlws, mws...)
	r.handler = r.gmddlws.Handler(r.serve)
}

// AddRoutes adds the routes if they do not exist. Or. update them.
func (r *Router) AddRoutes(routes ...Route) {
	if len(routes) == 0 {
		return
	}

	for _, r := range routes {
		if r.RouteId == "" {
			panic("Router.AddRoutes: the route id must not be empty")
		}
		if r.UpstreamId == "" {
			panic("Router.AddRoutes: the upstream id must not be empty")
		}
		if r.Matcher == nil {
			panic("Router.AddRoutes: the matcher must not be nil")
		}
	}

	r.lock.Lock()
	defer r.lock.Unlock()
	for _, route := range routes {
		if route.Handler == nil {
			route.Handler = AfterRoute
		}
		r.routem[route.RouteId] = route

		slog.Info("add or update the http route",
			"routeid", route.RouteId, "config", route.Config)
	}
	r.updateroutes()
}

// DelRoutes deletes the routes if they exist. Or, do nothing.
func (r *Router) DelRoutes(routes ...Route) {
	if len(routes) == 0 {
		return
	}

	r.lock.Lock()
	for _, route := range routes {
		delete(r.routem, route.RouteId)
		slog.Info("delete the http route", "routeid", route.RouteId)
	}
	r.updateroutes()
	r.lock.Unlock()
}

// DelRoutesByIds is the same as DelRoutes, but only uses the ids.
func (r *Router) DelRoutesByIds(ids ...string) {
	if len(ids) == 0 {
		return
	}

	r.lock.Lock()
	for _, id := range ids {
		delete(r.routem, id)
		slog.Info("delete the http route", "routeid", id)
	}
	r.updateroutes()
	r.lock.Unlock()
}

// GetRoute returns the route by the route id.
func (r *Router) GetRoute(id string) (Route, bool) {
	route, ok := r.allmap.Load().(map[string]Route)[id]
	return route, ok
}

// Routes returns the added the routes, which are read-only.
func (r *Router) Routes() []Route { return r.routes.Load().Routes }

func (r *Router) updateroutes() {
	routes := &routeswrapper{Routes: make([]Route, 0, len(r.routem))}
	for _, route := range r.routem {
		routes.Routes = append(routes.Routes, route)
	}
	sortroutes(routes.Routes)

	r.routes.Store(routes)
	r.allmap.Store(maps.Clone(r.routem))
}

func sortroutes(routes []Route) {
	sort.Slice(routes, func(i, j int) bool {
		ri, rj := &routes[i], &routes[j]
		switch {
		case ri.Protect && rj.Protect:
			return cmproute(ri, rj)

		case ri.Protect:
			return false

		case rj.Protect:
			return true

		default:
			return cmproute(ri, rj)
		}
	})
}

func cmproute(left, right *Route) bool {
	switch {
	case left.Priority > right.Priority:
		return true

	case left.Priority < right.Priority:
		return false

	default:
		return left.RouteId <= right.RouteId
	}
}

// Handle tries to match and handle the request.
//
// If matching successfully, return true. Or, return false.
func (r *Router) Handle(c *core.Context) (handled bool) { return r.serveRoute(c) }

// Serve is the same as ServeHTTP, but use Context as the input argument.
func (r *Router) Serve(c *core.Context) { r.handler(c) }

var _ http.Handler = new(Router)

// ServeHTTP implements the interface http.Handler.
func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	c := core.AcquireContext(req.Context())
	defer core.ReleaseContext(c)

	if w, ok := rw.(core.ResponseWriter); ok {
		c.ClientResponse = w
	} else {
		r := core.AcquireResponseWriter(rw)
		defer core.ReleaseResponseWriter(r)
		c.ClientResponse = r
	}

	c.ClientRequest = req
	r.handler(c)
}

func (r *Router) serve(c *core.Context) {
	defer closeResponse(c)
	matched := r.serveRoute(c)
	if !matched {
		c.Error = statuscode.ErrNotFound
	}

	if !c.ClientResponse.WroteHeader() {
		c.SendResponse()
	}
}

func (r *Router) serveRoute(c *core.Context) (matched bool) {
	routes := r.Routes()
	for i, _len := 0, len(routes); i < _len; i++ {
		route := &routes[i]
		if route.Protect {
			// (xgf): after sorting the routes, the protected routes
			// are at the end, so we have no need to match them.
			break
		}

		if matched = route.Match(c.ClientRequest); matched {
			c.RouteId = route.RouteId
			c.UpstreamId = route.UpstreamId
			c.Responser = route.Responser
			c.ForwardTimeout = route.ForwardTimeout
			serveRoute(c, route.Handler, route.RequestTimeout)
			break
		}
	}
	return
}

func serveRoute(c *core.Context, handler core.Handler, timeout time.Duration) {
	defer wrappanic(c)

	if timeout > 0 {
		var cancel context.CancelFunc
		c.Context, cancel = context.WithTimeout(c.Context, timeout)
		defer cancel()
	}

	handler(c)
}

func wrappanic(c *core.Context) {
	if r := recover(); r != nil {
		defaults.HandlePanicContext(c.Context, r)
		if e, ok := r.(error); ok {
			c.Abort(fmt.Errorf("panic: %w", e))
		} else {
			c.Abort(fmt.Errorf("panic: %v", r))
		}
	}
}

func closeResponse(c *core.Context) {
	if c.UpstreamResponse != nil {
		c.UpstreamResponse.Body.Close()
	}
}
