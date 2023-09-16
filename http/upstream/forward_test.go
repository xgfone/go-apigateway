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

package upstream

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/xgfone/go-apigateway/http/core"
	"github.com/xgfone/go-apigateway/upstream"
	"github.com/xgfone/go-loadbalancer"
	"github.com/xgfone/go-loadbalancer/balancer"
	"github.com/xgfone/go-loadbalancer/forwarder"
)

func TestForward(t *testing.T) {
	up := upstream.New(forwarder.New("http_forward_test", balancer.DefaultBalancer, loadbalancer.None))
	up.SetHost("localhost")
	up.SetScheme("http")
	upstream.Manager.Add(up.Name(), up)

	c := core.AcquireContext(context.Background())
	c.ClientRequest = &http.Request{URL: &url.URL{Path: "/"}}
	c.UpstreamId = "http_forward_test"
	Forward(c)

	if c.UpstreamRequest.Host != "localhost" {
		t.Errorf("expect the upstream request host '%s', but got '%s'",
			c.UpstreamRequest.Host, "localhost")
	}

	switch {
	case c.Error == nil:
		t.Error("expect an error, but got nil")
	case c.Error != loadbalancer.ErrNoAvailableEndpoints:
		t.Errorf("expect error '%s', but got '%s'", loadbalancer.ErrNoAvailableEndpoints, c.Error)
	}
}
