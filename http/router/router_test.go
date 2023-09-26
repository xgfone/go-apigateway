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

package router

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/xgfone/go-apigateway/http/core"
	"github.com/xgfone/go-apigateway/http/middleware"
)

func TestSortRoutes(t *testing.T) {
	routes := []Route{
		{RouteId: "r6", Priority: 1},
		{RouteId: "r5", Priority: 1},
		{RouteId: "r4", Priority: 2},
		{RouteId: "r3", Priority: 1, Protect: true},
		{RouteId: "r2", Priority: 1, Protect: true},
		{RouteId: "r1", Priority: 2, Protect: true},
	}
	sortroutes(routes)

	expects := []string{"r4", "r5", "r6", "r1", "r2", "r3"}
	for i, r := range routes {
		if id := expects[i]; id != r.RouteId {
			t.Errorf("%d: expect route '%s', but got '%s'", i, id, r.RouteId)
		}
	}
}

func TestRouter(t *testing.T) {
	AfterRoute = func(c *core.Context) { c.ClientResponse.WriteHeader(201) }
	DefaultRouter.DisableLog(false)

	DefaultRouter.AddRoutes()
	DefaultRouter.DelRoutes()

	if routes := DefaultRouter.Routes(); len(routes) > 0 {
		t.Errorf("unexpect routes, but got %+v", routes)
	}

	if route, ok := DefaultRouter.GetRoute("id"); ok {
		t.Errorf("unexpect route, but got %+v", route)
	}

	newMatcher := func(path string) MatcherFunc {
		return func(r *http.Request) bool { return r.URL.Path == path }
	}

	route1 := Route{
		RouteId:    "route1",
		UpstreamId: "router_test",
		Priority:   1,
		Matcher:    newMatcher("/path1"),
	}
	route2 := Route{
		RouteId:    "route2",
		UpstreamId: "router_test",
		Priority:   2,
		Matcher:    newMatcher("/path2"),
	}

	DefaultRouter.AddRoutes(route1, route2)
	if routes := DefaultRouter.Routes(); len(routes) != 2 {
		t.Errorf("expect 2 routes, but got %d: %+v", len(routes), routes)
	} else if id := routes[0].RouteId; id != "route2" {
		t.Errorf("expect route id '%s', but got '%s'", "route2", id)
		t.Error(routes)
	} else if id := routes[1].RouteId; id != "route1" {
		t.Errorf("expect route id '%s', but got '%s'", "route1", id)
	}

	req := &http.Request{URL: &url.URL{Path: "/path2"}}
	rec := httptest.NewRecorder()
	DefaultRouter.ServeHTTP(rec, req)
	if rec.Code != 201 {
		t.Errorf("expect status code %d, but got %d", 201, rec.Code)
	}

	DefaultRouter.DelRoutes(route1)
	if routes := DefaultRouter.Routes(); len(routes) != 1 {
		t.Errorf("expect 1 route, but got %d: %+v", len(routes), routes)
	} else if id := routes[0].RouteId; id != "route2" {
		t.Errorf("expect route id '%s', but got '%s'", "route2", id)
	}
}

func TestRouterMiddlewares(t *testing.T) {
	router := New()
	router.Use(middleware.New("test", nil, func(next core.Handler) core.Handler {
		return func(c *core.Context) {
			c.ClientResponse.Header().Set("X-Test", "1")
			next(c)
		}
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/", nil)
	router.ServeHTTP(rec, req)
	if rec.Code != 404 {
		t.Errorf("expect status code %d, but got %d", 404, rec.Code)
	}
	if v := rec.Header().Get("X-Test"); v != "1" {
		t.Errorf("expect 'X-Test' value '%s', but got '%s'", "1", v)
	}
}
