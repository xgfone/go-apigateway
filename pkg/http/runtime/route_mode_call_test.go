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

package runtime_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	_ "github.com/xgfone/go-apigateway/pkg/http/middlewares"

	"github.com/xgfone/go-apigateway/pkg/http/discovery"
	"github.com/xgfone/go-apigateway/pkg/http/dynamicconfig"
	"github.com/xgfone/go-apigateway/pkg/http/runtime"
	"github.com/xgfone/go-loadbalancer"
)

func TestRouteCall(t *testing.T) {
	routegroupmw, err := runtime.BuildMiddleware("processor", map[string]any{
		"directives": [][]string{{"addprefix", "/routemw"}},
	})
	if err != nil {
		t.Fatal(err)
	}

	upstreamgroupmw, err := runtime.BuildMiddleware("processor", map[string]any{
		"directives": [][]string{{"addprefix", "/upstreammw"}},
	})
	if err != nil {
		t.Fatal(err)
	}

	rmgroup := runtime.NewMiddlewareGroup(routegroupmw)
	umgroup := runtime.NewMiddlewareGroup(upstreamgroupmw)
	runtime.DefaultMiddlewareGroupManager.AddGroup("rmgroup", rmgroup)
	runtime.DefaultMiddlewareGroupManager.AddGroup("umgroup", umgroup)

	route, err := runtime.NewRoute(dynamicconfig.Route{
		Id:              "route",
		Upstream:        "route_call",
		Matcher:         dynamicconfig.Matcher{Method: "GET"}, // TODO:
		MiddlewareGroup: "rmgroup",
		Middlewares: dynamicconfig.Middlewares{
			"processor": map[string]any{"directives": [][]string{{"addprefix", "/routegroup"}}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	discovery.BuildStaticServer = func(s dynamicconfig.Server) (loadbalancer.Endpoint, error) {
		return newRouteCallEndpoint(net.JoinHostPort(s.Host, fmt.Sprint(s.Port)), t), nil
	}

	up, err := runtime.NewUpstream(dynamicconfig.Upstream{
		Id:     "route_call",
		Policy: "round_robin",
		Discovery: dynamicconfig.Discovery{
			Static: &dynamicconfig.StaticDiscovery{
				HealthCheck: &dynamicconfig.HealthCheck{Disable: true},
				Servers: []dynamicconfig.Server{
					{
						Host: "127.0.0.1",
						Port: 8001,
					},
					{
						Host: "127.0.0.1",
						Port: 8002,
					},
				},
			},
		},

		MiddlewareGroup: "umgroup",
		Middlewares: dynamicconfig.Middlewares{
			"processor": map[string]any{"directives": [][]string{{"addprefix", "/upstreamgroup"}}},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	runtime.ClearUpstreams()
	runtime.AddUpstream(up)
	if _up := runtime.GetUpstream("route_call"); _up == nil {
		t.Error("not found upstream")
	}

	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/route", nil)
	req.RemoteAddr = "127.0.0.1"

	c := runtime.AcquireContext()
	defer runtime.ReleaseContext(c)

	c.ClientRequest = req
	c.Context = req.Context()
	c.Route = route

	route.Call(c)
	if c.Error != nil {
		t.Error(c.Error)
	} else if c.UpstreamResponse == nil {
		t.Error("expect got an upstream response, but got nil")
	} else if c.UpstreamResponse.StatusCode != 204 {
		t.Errorf("expect status code 204, but got %d", c.UpstreamResponse.StatusCode)
	}

	// TODO:
}

type rcallep struct {
	id string
	t  *testing.T
}

func newRouteCallEndpoint(id string, t *testing.T) rcallep {
	return rcallep{id: id, t: t}
}

func (ep rcallep) ID() string { return ep.id }
func (ep rcallep) Serve(c context.Context, i any) (any, error) {
	const path = "/upstreammw/upstreamgroup/routemw/routegroup/route"
	if req := i.(*runtime.Context).UpstreamRequest(); req.URL.Path != path {
		ep.t.Errorf("expect path '%s', but got '%s'", path, req.URL.Path)
	}
	return &http.Response{StatusCode: 204}, nil
}
