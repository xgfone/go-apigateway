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
	"fmt"
	"net/http"

	"github.com/xgfone/go-loadbalancer/upstream"
)

// UpstreamCall calls the request to the upstream server,
// which must run on the Call mode.
func UpstreamCall(c *Context) {
	c.MustCall("runtime.UpstreamCall")

	upstream := upstream.DefaultManager.Get(c.Route.Route.Upstream)
	if upstream == nil {
		c.Error = fmt.Errorf("no upstream '%s'", c.Route.Route.Upstream)
	} else {
		c.Upstream = upstream.ContextData().(*Upstream)
		c.Upstream.handle(c)
	}
}

func upstreamcall(c *Context) {
	c.callbackOnForward()
	resp, err := c.Upstream.upstream.Forwader().Serve(c.ClientRequest.Context(), c)
	if err != nil {
		c.Error = err
	} else if resp != nil {
		c.UpstreamResponse = resp.(*http.Response)
	}
}
