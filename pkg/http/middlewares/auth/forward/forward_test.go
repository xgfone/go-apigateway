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

package forward

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/xgfone/go-apigateway/pkg/http/discovery"
	"github.com/xgfone/go-apigateway/pkg/http/dynamicconfig"
	"github.com/xgfone/go-apigateway/pkg/http/runtime"
	"github.com/xgfone/go-apigateway/pkg/internal/httpx"
	"github.com/xgfone/go-loadbalancer"
	"github.com/xgfone/go-loadbalancer/endpoint"
)

func TestForwardAuth(t *testing.T) {
	var server *http.Server
	defer func() {
		if server != nil {
			_ = server.Shutdown(context.Background())
		}
	}()

	go func() {
		server = &http.Server{
			Addr: ":10000",
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("X-Auth-Type") == "" {
					w.WriteHeader(401)
					_, _ = io.WriteString(w, "missing header X-Auth-Type")
				} else {
					w.Header().Set("X-User-Id", "1000")
					w.WriteHeader(204)
				}
			}),
		}

		_ = server.ListenAndServe()
	}()
	time.Sleep(time.Millisecond * 100)

	m, err := ForwardAuth("forwardauth", 1, Config{
		URL:             "http://127.0.0.1:10000/auth",
		Headers:         []string{"X-Auth-Type"},
		UpstreamHeaders: []string{"X-User-*"},
	})
	if err != nil {
		t.Fatal(err)
	}

	mgroup := runtime.NewMiddlewareGroup(m)
	runtime.DefaultMiddlewareGroupManager.AddGroup("forwardauth", mgroup)

	discovery.BuildStaticServer = func(server dynamicconfig.Server) (loadbalancer.Endpoint, error) {
		return endpoint.New("127.0.0.1", func(ctx context.Context, req any) (resp any, err error) {
			c := req.(*runtime.Context)
			if id := c.UpstreamRequest().Header.Get("X-User-Id"); id != "1000" {
				t.Errorf("expect user id '%s', but got '%s'", "1000", id)
			}
			return nil, nil
		}), nil
	}

	up, err := runtime.NewUpstream(dynamicconfig.Upstream{
		Id:              "forwardauth",
		MiddlewareGroup: "forwardauth",
		Discovery: dynamicconfig.Discovery{
			Static: &dynamicconfig.StaticDiscovery{
				Servers:     []dynamicconfig.Server{{Host: "127.0.0.1"}},
				HealthCheck: &dynamicconfig.HealthCheck{Disable: true},
			},
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	runtime.AddUpstream(up)

	c := runtime.AcquireContext()
	defer runtime.ReleaseContext(c)

	c.SetModeForward()
	c.Route.Route.Upstream = "forwardauth"
	c.ClientRequest = httptest.NewRequest(http.MethodGet, "http://127.0.0.1", nil)
	c.ClientRequest.Header.Set("X-Auth-Type", "header")
	c.Context = c.ClientRequest.Context()

	rec := httptest.NewRecorder()
	c.ClientResponse = httpx.NewResponseWriter(rec)
	runtime.UpstreamForward(c)
	if rec.Code != 200 {
		t.Errorf("expect status code 200, but got %d: %s", rec.Code, rec.Body.String())
	}

	rec = httptest.NewRecorder()
	c.ClientRequest = c.ClientRequest.Clone(c.Context)
	c.ClientRequest.Header.Del("X-Auth-Type")
	c.ClientResponse = httpx.NewResponseWriter(rec)
	runtime.UpstreamForward(c)
	if rec.Code != 401 {
		t.Errorf("expect status code 401, but got %d: %s", rec.Code, rec.Body.String())
	}
}
