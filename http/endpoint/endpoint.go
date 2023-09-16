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

package endpoint

import (
	"context"
	"net"
	"strconv"

	"github.com/xgfone/go-apigateway/http/core"
	"github.com/xgfone/go-apigateway/http/upstream"
	"github.com/xgfone/go-loadbalancer/endpoint"
)

// New returns a new endpoint by the static server.
//
// host must not be empty, or it will panic.
func New(host string, port uint16, weight int) *endpoint.Endpoint {
	if host == "" {
		panic("NewEndpoint: host must not be empty")
	}

	addr := host
	if port > 0 {
		addr = net.JoinHostPort(addr, strconv.FormatInt(int64(port), 10))
	}

	ep := endpoint.New(addr, nil)
	ep.SetWeight(weight)
	ep.SetServeFunc(proxy{Endpoint: ep, addr: addr}.Serve)
	ep.SetConfig(map[string]any{"addr": addr, "weight": ep.Weight()})
	return ep
}

type proxy struct {
	*endpoint.Endpoint
	addr string
}

func (p proxy) Serve(_ context.Context, req any) (any, error) {
	c := req.(*core.Context)
	c.Endpoint = p.Endpoint

	r := c.UpstreamRequest
	r.URL.Host = p.addr
	if r.Host == "" {
		r.Host = p.addr
	}

	resp, err := upstream.DefaultHttpClient.Do(r)
	if err != nil && resp != nil {
		resp.Body.Close() // For status code 3xx
	}
	return resp, err
}
