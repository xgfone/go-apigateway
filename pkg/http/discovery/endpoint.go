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

package discovery

import (
	"context"
	"errors"
	"net"
	"strconv"

	"github.com/xgfone/go-apigateway/pkg/http/dynamicconfig"
	"github.com/xgfone/go-apigateway/pkg/http/runtime"
	"github.com/xgfone/go-loadbalancer"
	"github.com/xgfone/go-loadbalancer/endpoint"
)

func init() {
	BuildStaticServer = func(server dynamicconfig.Server) (loadbalancer.Endpoint, error) {
		return NewEndpoint(server)
	}
}

// NewEndpoint returns a new endpoint by the static server.
func NewEndpoint(server dynamicconfig.Server) (*endpoint.Endpoint, error) {
	if server.Host == "" {
		return nil, errors.New("the server host must not be empty")
	}

	addr := server.Host
	if server.Port > 0 {
		addr = net.JoinHostPort(addr, strconv.FormatInt(int64(server.Port), 10))
	}

	ep := endpoint.New(addr, nil)
	ep.SetWeight(server.Weight)
	ep.SetServe(proxy{Endpoint: ep, addr: addr}.Serve)
	ep.SetConfig(map[string]any{"addr": addr, "weight": ep.Weight()})
	return ep, nil
}

type proxy struct {
	*endpoint.Endpoint
	addr string
}

func (p proxy) Serve(_ context.Context, req any) (any, error) {
	c := req.(*runtime.Context)
	c.Endpoint = p.Endpoint

	ureq := c.UpstreamRequest()
	ureq.URL.Host = p.addr
	if ureq.Host == "" {
		ureq.Host = p.addr
	}

	resp, err := c.SendRequest()
	if err != nil && resp != nil {
		resp.Body.Close() // For status code 3xx
	}
	return resp, err
}
