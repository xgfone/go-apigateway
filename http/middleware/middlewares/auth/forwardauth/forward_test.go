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
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/xgfone/go-apigateway/http/core"
	"github.com/xgfone/go-apigateway/http/internal/httpx"
	"github.com/xgfone/go-apigateway/upstream"
	"github.com/xgfone/go-loadbalancer"
	"github.com/xgfone/go-loadbalancer/balancer"
	"github.com/xgfone/go-loadbalancer/endpoint"
	"github.com/xgfone/go-loadbalancer/forwarder"
)

func TestForwardAuthByRequest(t *testing.T) {
	client := &http.Client{Transport: httpx.RoundTripFunc(func(r *http.Request) (*http.Response, error) {
		ttype := r.Header.Get("X-Auth-Type")
		if ttype != "header" {
			return nil, fmt.Errorf("invalid header 'X-Auth-Type'")
		}

		return &http.Response{
			Body:   io.NopCloser(strings.NewReader("")),
			Header: http.Header{"X-User-Id": []string{"1000"}},
		}, nil
	})}

	// -------------------------------------------------------------------- //

	m, err := ForwardAuth(Config{
		URL:             "http://127.0.0.1/auth",
		Headers:         []string{"X-Auth-Type"},
		UpstreamHeaders: []string{"X-User-*"},
		Client:          client,
	})
	if err != nil {
		t.Fatal(err)
	}

	handler := m.Handler(func(c *core.Context) {
		c.UpstreamRequest = c.ClientRequest.Clone(context.Background())
		c.CallbackOnForward()
	})

	c := core.AcquireContext(context.Background())
	c.ClientRequest = &http.Request{
		Host:       "localhost",
		Method:     "GET",
		RequestURI: "/",
		RemoteAddr: "127.0.0.1:1234",
		Header:     http.Header{"X-Auth-Type": []string{"header"}},
	}
	handler(c)

	if c.Error != nil {
		t.Error(c.Error)
	} else if c.UpstreamRequest == nil {
		t.Error("expect upstream request, but got nil")
	} else if id := c.UpstreamRequest.Header.Get("X-User-Id"); id != "1000" {
		t.Errorf("expect user id '%s', but got '%s'", "1000", id)
	}
}

func TestForwardAuthByUpstream(t *testing.T) {
	static := &loadbalancer.Static{Endpoints: loadbalancer.Endpoints{
		endpoint.New("test", func(ctx context.Context, req any) (resp any, err error) {
			ttype := req.(*core.Context).ClientRequest.Header.Get("X-Auth-Type")
			if ttype != "header" {
				return nil, fmt.Errorf("invalid header 'X-Auth-Type'")
			}

			return &http.Response{
				Body:   io.NopCloser(strings.NewReader("")),
				Header: http.Header{"X-User-Id": []string{"1000"}},
			}, nil
		}),
	}}

	discovery := loadbalancer.DiscoveryFunc(func() *loadbalancer.Static { return static })
	forwarder := forwarder.New("forwardauth", balancer.DefaultBalancer, discovery)
	upstream.Manager.Add("forwardauth", upstream.New(forwarder))

	// -------------------------------------------------------------------- //

	m, err := ForwardAuth(Config{
		URL:             "http://127.0.0.1/auth",
		Upstream:        "forwardauth",
		Headers:         []string{"X-Auth-Type"},
		UpstreamHeaders: []string{"X-User-*"},
	})
	if err != nil {
		t.Fatal(err)
	}

	handler := m.Handler(func(c *core.Context) {
		c.UpstreamRequest = c.ClientRequest.Clone(context.Background())
		c.CallbackOnForward()
	})

	c := core.AcquireContext(context.Background())
	c.ClientRequest = &http.Request{
		Host:       "localhost",
		Method:     "GET",
		RequestURI: "/",
		RemoteAddr: "127.0.0.1:1234",
		Header:     http.Header{"X-Auth-Type": []string{"header"}},
	}
	handler(c)

	if c.Error != nil {
		t.Error(c.Error)
	} else if c.UpstreamRequest == nil {
		t.Error("expect upstream request, but got nil")
	} else if id := c.UpstreamRequest.Header.Get("X-User-Id"); id != "1000" {
		t.Errorf("expect user id '%s', but got '%s'", "1000", id)
	}
}
