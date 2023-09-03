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

package runtime

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/xgfone/go-apigateway/pkg/discovery"
	"github.com/xgfone/go-apigateway/pkg/http/dynamicconfig"
	"github.com/xgfone/go-loadbalancer"
	"github.com/xgfone/go-loadbalancer/balancer"
	"github.com/xgfone/go-loadbalancer/balancer/retry"
	"github.com/xgfone/go-loadbalancer/forwarder"
	"github.com/xgfone/go-loadbalancer/upstream"
)

// Upstream represents the runtime upstream.
type Upstream struct {
	UpConfig   dynamicconfig.Upstream
	HttpClient *http.Client

	upstream  *upstream.Upstream
	mwhandler Handler
	mwgroup   string
}

func (u *Upstream) handle(c *Context) {
	switch {
	case u.mwhandler != nil:
		u.mwhandler(c)

	case u.mwgroup != "":
		DefaultMiddlewareGroupManager.Handle(c, u.mwgroup, forward)

	default:
		forward(c)
	}
}

func handleUpstreamMiddlewareGroup(c *Context) {
	switch {
	case c.Upstream.mwgroup != "":
		DefaultMiddlewareGroupManager.Handle(c, c.Upstream.mwgroup, forward)

	default:
		forward(c)
	}
}

func buildUpstreamMiddlewares(up dynamicconfig.Upstream) (Handler, error) {
	if len(up.Middlewares) == 0 {
		return nil, nil
	}
	return buildMiddlewaresHandler(handleUpstreamMiddlewareGroup, up.Middlewares)
}

func buildUpstreamBalancer(up dynamicconfig.Upstream) (balancer.Balancer, error) {
	_balancer, err := balancer.Build(up.ForwardPolicy(), nil)
	if err != nil || up.Retry.Number < 0 {
		return nil, err
	}

	if up.Retry.Interval > 0 {
		up.Retry.Interval *= time.Millisecond
	}
	return retry.New(_balancer, up.Retry.Interval, up.Retry.Number), nil
}

// BuildUpstreamOption is used to build the underlying upstream option with the configration.
func buildUpstreamOption(up dynamicconfig.Upstream) (option upstream.Option, err error) {
	// Balancer: forwarding policy
	balancer, err := buildUpstreamBalancer(up)
	if err != nil {
		return
	}

	// Middlewares
	mwhandler, err := buildUpstreamMiddlewares(up)
	if err != nil {
		return
	}

	// Endpoints Discovery
	discovery, err := BuildUpstreamDiscovery(up)
	if err != nil {
		return
	}

	if up.Timeout < 0 {
		up.Timeout = 0
	} else {
		up.Timeout *= time.Second
	}

	return func(u *upstream.Upstream) {
		forwarder := u.Forwader()
		forwarder.SetTimeout(up.Timeout)
		forwarder.SetBalancer(balancer)
		forwarder.SetDiscovery(discovery)
		u.SetContextData(&Upstream{
			UpConfig: up,
			upstream: u,

			mwhandler: mwhandler,
			mwgroup:   up.MiddlewareGroup,
		})
	}, nil
}

// UpdateOption returns the update option of the underlying upstream,
// which is just to return the UpdateTo method as the update option.
func (up *Upstream) UpdateOption() upstream.Option { return up.UpdateTo }

// Update udpates the underlying upstream with the current upstream configuration.
func (up *Upstream) UpdateTo(u *upstream.Upstream) {
	_forwarder := up.upstream.Forwader()
	forwarder := u.Forwader()
	forwarder.SetTimeout(_forwarder.GetTimeout())
	forwarder.SetBalancer(_forwarder.GetBalancer())
	forwarder.SetDiscovery(_forwarder.GetDiscovery())
}

// Forwarder exports the underlying the forwarder.
func (u *Upstream) Forwarder() *forwarder.Forwarder {
	return u.upstream.Forwader()
}

// NewUpstream builds the runtime upstream by the config.
func NewUpstream(config dynamicconfig.Upstream) (*Upstream, error) {
	option, err := buildUpstreamOption(config)
	if err != nil {
		return nil, err
	}

	up := upstream.New(config.Id, option)
	return up.ContextData().(*Upstream), nil
}

// ------------------------------------------------------------------------- //

var uplock = new(sync.Mutex)

func init() {
	upstream.DefaultManager.OnAdd(func(u *upstream.Upstream) {
		u.Discovery().(discovery.Discovery).Start()
	})

	upstream.DefaultManager.OnDel(func(u *upstream.Upstream) {
		u.Discovery().(discovery.Discovery).Stop()
	})
}

// AddUpstream adds the upstream into the global manager.
//
// If exist, update it.
func AddUpstream(up *Upstream) {
	uplock.Lock()
	defer uplock.Unlock()

	_up := upstream.DefaultManager.Get(up.UpConfig.Id)
	if _up == nil {
		upstream.DefaultManager.Add(up.upstream)
	} else {
		up.UpdateTo(_up)
	}
}

// DelUpstream deletes the upstream by the group id from the global manager.
//
// If not exist, do nothing.
func DelUpstream(id string) {
	uplock.Lock()
	defer uplock.Unlock()
	upstream.DefaultManager.Delete(id)
}

// GetUpstream returns the upstream by the group id from the global manager.
//
// If not exist, return nil.
func GetUpstream(id string) *Upstream {
	if up := upstream.DefaultManager.Get(id); up != nil {
		return up.ContextData().(*Upstream)
	}
	return nil
}

// GetUpstreams returns all the upstreams.
func GetUpstreams() []*Upstream {
	ups := upstream.DefaultManager.Gets()
	_ups := make([]*Upstream, 0, len(ups))
	for _, up := range ups {
		_ups = append(_ups, up.ContextData().(*Upstream))
	}
	return _ups
}

// ------------------------------------------------------------------------- //

// UpstreamForward forwards the request to one of the upstream servers by the upstream.
func UpstreamForward(c *Context) {
	upstream := upstream.DefaultManager.Get(c.Route.Route.Upstream)
	if upstream == nil {
		err := fmt.Errorf("no upstream '%s'", c.Route.Route.Upstream)
		c.SendResponse(nil, ErrInternalServerError.WithError(err))
		return
	}

	c.Upstream = upstream.ContextData().(*Upstream)
	if req := c.upRequest(); req != nil { // the request has been created.
		updateUpstreamRequest(c, req)
	} // We delay to the endpoint forwarding to creating the request.

	c.Upstream.handle(c)
}

func updateUpstreamRequest(c *Context, req *http.Request) {
	// Set the url scheme.
	if c.Upstream.UpConfig.Scheme == "https" {
		req.URL.Scheme = "https"
	} else {
		req.URL.Scheme = "http"
	}

	// Set the Host header.
	switch host := c.Upstream.UpConfig.Host; host {
	case "", dynamicconfig.HostClient:
		req.Host = c.ClientRequest.Host
	case dynamicconfig.HostServer:
		req.Host = "" // Clear the host and let the endpoint reset it.
	default:
		req.Host = host
	}
}

func forward(c *Context) {
	c.callbackOnForward()
	resp, err := c.Upstream.upstream.Forwader().Serve(c.ClientRequest.Context(), c)
	if err != nil {
		handleForwardError(c, err)
	} else if resp != nil {
		resp := resp.(*http.Response)
		defer resp.Body.Close()
		c.SendResponse(resp, nil)
	}
}

func handleForwardError(c *Context, err error) {
	switch err {
	case loadbalancer.ErrNoAvailableEndpoints:
		c.SendResponse(nil, ErrServiceUnavailable.WithError(err))

	case context.Canceled, context.DeadlineExceeded:
		c.SendResponse(nil, ErrStatusGatewayTimeout.WithError(err))

	default:
		c.SendResponse(nil, ErrInternalServerError.WithError(err))
	}
}
