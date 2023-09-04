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

	"github.com/xgfone/go-loadbalancer"
	"github.com/xgfone/go-loadbalancer/upstream"
)

// UpstreamForward forwards the request to one of the upstream servers
// by the upstream, which must run on the Forward mode.
func UpstreamForward(c *Context) {
	c.MustModeForward("runtime.UpstreamForward")

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

func upstreamforward(c *Context) {
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
