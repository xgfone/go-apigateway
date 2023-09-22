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
	"errors"
	"fmt"
	"time"

	"github.com/xgfone/go-apigateway/http/endpoint"
	"github.com/xgfone/go-apigateway/upstream"
	"github.com/xgfone/go-loadbalancer"
	"github.com/xgfone/go-loadbalancer/balancer"
	"github.com/xgfone/go-loadbalancer/balancer/retry"
	"github.com/xgfone/go-loadbalancer/forwarder"
)

var (
	// BuildStaticServer is used to customize the building of the static server,
	// which is used by the default discovery builder.
	//
	// Default: use http/endpoint.New to build it.
	BuildStaticServer func(server Server) (loadbalancer.Endpoint, error)

	// BuildDiscovery builds a discovery by the config.
	BuildDiscovery func(upid string, discovery Discovery) (loadbalancer.Discovery, error)
)

func init() {
	BuildDiscovery = buildDiscovery
	BuildStaticServer = func(s Server) (loadbalancer.Endpoint, error) {
		if s.Host == "" {
			return nil, errors.New("BuildStaticServer: host must not be empty")
		}
		return endpoint.New(s.Host, s.Port, s.Weight), nil
	}
}

func buildDiscovery(_ string, discovery Discovery) (loadbalancer.Discovery, error) {
	eps, err := BuildStaticServers(discovery.Static.Servers)
	if err != nil {
		return nil, err
	}
	return upstream.NewDiscovery(eps...), nil
}

// BuildStaticServers builds a set of servers,
// which use BuildStaticServer to build each server.
func BuildStaticServers(servers []Server) (loadbalancer.Endpoints, error) {
	if len(servers) == 0 {
		return nil, nil
	}

	endpoints := make(loadbalancer.Endpoints, len(servers))
	for i, s := range servers {
		ep, err := BuildStaticServer(s)
		if err != nil {
			return nil, err
		}
		endpoints[i] = ep
	}
	return endpoints, nil
}

// Build builds an upstream by the config.
func (up Upstream) Build() (*upstream.Upstream, error) {
	if up.Id == "" {
		return nil, errors.New("Upstream: missing Id")
	}

	discovery, err := BuildDiscovery(up.Id, up.Discovery)
	if err != nil {
		return nil, fmt.Errorf("Upstream<%s>: fail to build discovery: %w", up.Id, err)
	}

	policy := up.ForwardPolicy()
	balancer := balancer.Get(policy)
	if balancer == nil {
		return nil, fmt.Errorf("Upstream<%s>: invalid forwarding policy '%s'", up.Id, policy)
	}
	if up.Retry.Number >= 0 {
		balancer = retry.New(balancer, up.Retry.Interval*time.Millisecond, up.Retry.Number)
	}

	forwarder := forwarder.New(up.Id, balancer, discovery)
	forwarder.SetConfig(up)
	if up.Timeout > 0 {
		forwarder.SetTimeout(up.Timeout)
	}

	_up := upstream.New(forwarder)
	_up.SetScheme(up.Scheme)
	_up.SetHost(up.Host)
	return _up, nil
}
