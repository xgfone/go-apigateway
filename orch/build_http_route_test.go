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
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/xgfone/go-apigateway/http/core"
	"github.com/xgfone/go-apigateway/http/middleware"
	"github.com/xgfone/go-apigateway/http/router"
)

func addPathSuffixBuilder(name string, conf any) (middleware.Middleware, error) {
	suffix := conf.(string)
	return middleware.New(name, conf, func(next core.Handler) core.Handler {
		return func(c *core.Context) {
			c.ClientRequest.URL.Path += suffix
			next(c)
		}
	}), nil
}

func TestRouteMiddleware(t *testing.T) {
	router.AfterRoute = func(c *core.Context) { c.ClientRequest.URL.Path += "/upstream" }

	register := middleware.DefaultRegistry
	register.Register("addpathsuffix", addPathSuffixBuilder)

	addPathSuffix, _ := register.Build("addpathsuffix", "/groupmw")
	group := middleware.NewGroup("test_route_mw_group", addPathSuffix)
	middleware.DefaultGroupManager.Add(group.Name(), group)

	route, err := HttpRoute{
		Id:       "route",
		Upstream: "test_route_mw",
		Matchers: []HttpMatcher{{Paths: []string{"/"}}},

		MiddlewareGroup: "test_route_mw_group",
		Middlewares: Middlewares{
			{
				Name: "addpathsuffix",
				Conf: "/routemw",
			},
		},
	}.Build()
	if err != nil {
		t.Error(err)
	}

	c := core.AcquireContext(context.Background())
	c.ClientRequest = &http.Request{URL: &url.URL{Path: "/path"}}
	route.Handler(c)

	if c.Error != nil {
		t.Error(c.Error)
	}

	const expect = "/path/routemw/groupmw/upstream"
	if path := c.ClientRequest.URL.Path; path != expect {
		t.Errorf("expect path '%s', but got '%s'", expect, path)
	}
}
