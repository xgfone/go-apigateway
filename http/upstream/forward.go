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
	"fmt"
	"net/http"

	"github.com/xgfone/go-apigateway/http/core"
	"github.com/xgfone/go-apigateway/upstream"
)

// Forward forwards the http request by the upstream.
func Forward(c *core.Context) {
	if c.IsAborted {
		return
	}

	up, ok := upstream.Manager.Get(c.UpstreamId)
	if !ok {
		c.Abort(fmt.Errorf("no upstream '%s'", c.UpstreamId))
		return
	}

	c.Upstream = up
	if c.UpstreamRequest == nil {
		c.UpstreamRequest = newRequest(c)
	}

	// Set the url scheme.
	switch scheme := up.Scheme(); scheme {
	case "https", "http":
		c.UpstreamRequest.URL.Scheme = scheme
	default:
		if c.UpstreamRequest.URL.Scheme == "" {
			c.UpstreamRequest.URL.Scheme = "http"
		}
	}

	// Set the Host header.
	switch host := up.Host(); host {
	case "", HostClient:
		c.UpstreamRequest.Host = c.ClientRequest.Host
	case HostServer:
		c.UpstreamRequest.Host = "" // Clear the host and let the endpoint reset it.
	default:
		c.UpstreamRequest.Host = host
	}

	c.CallbackOnForward()
	if c.IsAborted {
		return
	}

	var resp interface{}
	resp, c.Error = up.Serve(c.UpstreamRequest.Context(), c)
	if resp != nil {
		c.UpstreamResponse = resp.(*http.Response)
	}
}
