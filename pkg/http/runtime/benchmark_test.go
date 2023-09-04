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
	"github.com/xgfone/go-apigateway/pkg/http/middlewares/block"
	"github.com/xgfone/go-apigateway/pkg/http/runtime"
	"github.com/xgfone/go-apigateway/pkg/internal/httpx"
	"github.com/xgfone/go-apigateway/pkg/internal/slogx"
	"github.com/xgfone/go-loadbalancer"
)

func BenchmarkRouterNoop(b *testing.B) {
	slogx.DisableSLog()

	route, err := runtime.NewRoute(dynamicconfig.Route{
		Id:       "r1",
		Matcher:  dynamicconfig.Matcher{Path: "/"},
		Upstream: "testnoop",
	})
	if err != nil {
		panic(err)
	}

	discovery.BuildStaticServer = func(s dynamicconfig.Server) (loadbalancer.Endpoint, error) {
		return newNoopEndpoint(net.JoinHostPort(s.Host, fmt.Sprint(s.Port))), nil
	}

	up, err := runtime.NewUpstream(dynamicconfig.Upstream{
		Id: "testnoop",

		Policy: "round_robin",
		Discovery: dynamicconfig.Discovery{
			Static: &dynamicconfig.StaticDiscovery{
				HealthCheck: &dynamicconfig.HealthCheck{Disable: true},
				Servers: []dynamicconfig.Server{
					{
						Host: "127.0.0.1",
						Port: 80,
					},
				},
			},
		},
	})
	if err != nil {
		panic(err)
	}

	runtime.ClearUpstreams()
	runtime.AddUpstream(up)

	router := runtime.NewRouter()
	router.AddRoutes(route)

	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/", nil)

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			servehttp(router, req)
		}
	})
}

func BenchmarkRouterFull(b *testing.B) {
	slogx.DisableSLog()

	b192168, err := block.Block("block", 10, "192.168.0.0/24")
	if err != nil {
		panic(err)
	}

	b192169, err := block.Block("block", 10, "192.169.0.0/24")
	if err != nil {
		panic(err)
	}

	rmgroup := runtime.NewMiddlewareGroup(b192168)
	umgroup := runtime.NewMiddlewareGroup(b192169)
	runtime.DefaultMiddlewareGroupManager.AddGroup("rmg", rmgroup)
	runtime.DefaultMiddlewareGroupManager.AddGroup("umg", umgroup)

	route, err := runtime.NewRoute(dynamicconfig.Route{
		Id:              "r1",
		Matcher:         dynamicconfig.Matcher{Path: "/"},
		MiddlewareGroup: "rmg",
		Middlewares: dynamicconfig.Middlewares{
			"allow": map[string]any{"cidrs": []string{"127.0.0.0/8"}},
		},

		Upstream: "testfull",
	})
	if err != nil {
		panic(err)
	}

	discovery.BuildStaticServer = func(s dynamicconfig.Server) (loadbalancer.Endpoint, error) {
		return newNoopEndpoint(net.JoinHostPort(s.Host, fmt.Sprint(s.Port))), nil
	}

	up, err := runtime.NewUpstream(dynamicconfig.Upstream{
		Id: "testfull",

		Policy: "round_robin",
		Discovery: dynamicconfig.Discovery{
			Static: &dynamicconfig.StaticDiscovery{
				HealthCheck: &dynamicconfig.HealthCheck{Disable: true},
				Servers: []dynamicconfig.Server{
					{
						Host: "127.0.0.1",
						Port: 80,
					},
					{
						Host: "127.0.0.2",
						Port: 80,
					},
				},
			},
		},

		MiddlewareGroup: "umg",
		Middlewares: dynamicconfig.Middlewares{
			"allow": map[string]any{"cidrs": []string{"127.0.0.0/16"}},
		},
	})
	if err != nil {
		panic(err)
	}

	runtime.ClearUpstreams()
	runtime.AddUpstream(up)

	router := runtime.NewRouter()
	router.DisableLog(true)
	router.AddRoutes(route)

	req := httptest.NewRequest(http.MethodGet, "http://127.0.0.1/", nil)
	req.RemoteAddr = "127.0.0.1"

	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(p *testing.PB) {
		for p.Next() {
			servehttp(router, req)
		}
	})
}

func servehttp(r *runtime.Router, req *http.Request) {
	c := runtime.AcquireContext(req.Context())
	defer runtime.ReleaseContext(c)
	rec := httptest.NewRecorder()

	c.ClientRequest = req
	c.ClientResponse = httpx.NewResponseWriter(rec)
	r.HandleHTTP(c)
}

type noopep struct{ addr string }

func newNoopEndpoint(addr string) noopep { return noopep{addr: addr} }

func (ep noopep) ID() string                              { return ep.addr }
func (ep noopep) Serve(context.Context, any) (any, error) { return nil, nil }
