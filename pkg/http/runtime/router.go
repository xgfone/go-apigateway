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
	"fmt"
	"maps"
	"net/http"
	"sort"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xgfone/go-atomicvalue"
	"github.com/xgfone/go-defaults"
)

// DefaultRouter is the default global http router.
var DefaultRouter = NewRouter()

var (
	routemappool   = &sync.Pool{New: func() any { return make(map[string]Route, 128) }}
	routeslicepool = &sync.Pool{New: func() any { return &routeswrapper{make([]Route, 0, 128)} }}
)

type routeswrapper struct{ Routes []Route }

// Router represents a router to match and forward the http request.
type Router struct {
	lock   sync.Mutex
	routem map[string]Route
	allmap atomicvalue.Value[map[string]Route]
	routes atomicvalue.Value[*routeswrapper]
	notlog atomic.Bool
}

// NewRouter returns a new router.
func NewRouter() *Router {
	r := &Router{routem: make(map[string]Route, 32)}
	r.allmap.Store(routemappool.Get().(map[string]Route))
	r.routes.Store(routeslicepool.Get().(*routeswrapper))
	return r
}

// AddRoutes adds the routes if they do not exist. Or. update them.
func (r *Router) AddRoutes(routes ...Route) {
	if len(routes) == 0 {
		return
	}

	r.lock.Lock()
	defer r.lock.Unlock()

	for i, _len := 0, len(routes); i < _len; i++ {
		route := routes[i]
		if route.Route.Priority == 0 {
			route.matcher.Priority()
		}
		r.routem[route.Route.Id] = route
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
		delete(r.routem, route.Route.Id)
	}
	r.updateroutes()
	r.lock.Unlock()
}

// GetRoute returns the route by the route id.
func (r *Router) GetRoute(id string) (Route, bool) {
	route, ok := r.allmap.Load()[id]
	return route, ok
}

// Routes returns the added the routes, which are read-only and must not be modified.
func (r *Router) Routes() []Route { return r.routes.Load().Routes }

func (r *Router) updateroutes() {
	routes := routeslicepool.Get().(*routeswrapper)
	for _, route := range r.routem {
		routes.Routes = append(routes.Routes, route)
	}
	sortroutes(routes.Routes)

	oldroutes := r.routes.Swap(routes)
	clear(oldroutes.Routes)
	oldroutes.Routes = oldroutes.Routes[:0]
	routeslicepool.Put(oldroutes)

	oldroutem := r.allmap.Swap(maps.Clone(r.routem))
	clear(oldroutem)
	routemappool.Put(oldroutem)
}

func sortroutes(routes []Route) {
	sort.Slice(routes, func(i, j int) bool {
		ri, rj := &routes[i].Route, &routes[j].Route
		switch {
		case ri.Protect && rj.Protect:
			return ri.Priority > rj.Priority

		case ri.Protect:
			return false

		case rj.Protect:
			return true

		default:
			return ri.Priority > rj.Priority
		}
	})
}

// HandleHTTP is the same as ServeHTTP, but use Context as the input argument.
func (r *Router) HandleHTTP(c *Context) {
	if c.Context == nil {
		c.Context = c.ClientRequest.Context()
	}
	r.serveHTTP(c, c.ClientRequest)
}

var _ http.Handler = new(Router)

// ServeHTTP implements the interface http.Handler.
func (r *Router) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	c := AcquireContext()
	defer ReleaseContext(c)

	if w, ok := rw.(ResponseWriter); ok {
		c.ClientResponse = w
	} else {
		c.ClientResponse = NewResponseWriter(rw)
	}

	c.ClientRequest = req
	c.Context = req.Context()
	r.serveHTTP(c, req)
}

func (r *Router) recover(c *Context) {
	if r := recover(); r != nil {
		defaults.HandlePanicContext(c.Context, r)
		c.SendResponse(nil, ErrBadGateway)
		c.Error = fmt.Errorf("panic: %v", r)
	}
}

func (r *Router) serveHTTP(c *Context, req *http.Request) {
	defer r.recover(c)
	start := time.Now()

	var matched bool
	for _, route := range r.Routes() {
		if route.Route.Protect {
			// (xgf): after sorting the routes, the protected routes
			// are at the end, so we have no need to match them.
			break
		}

		if matched = route.Match(c); matched {
			c.Route = route
			c.Route.Handle(c)
			break
		}
	}
	if !matched {
		c.SendResponse(nil, ErrNotFound)
	}

	r.log(c, start, matched)
}
