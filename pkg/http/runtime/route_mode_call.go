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
	"time"
)

// Call calls the request by the route as the client,
// which will set the running mode to ModeCall.
func (r *Route) Call(c *Context) {
	c.setmode(ModeCall)

	if r.Route.Timeout > 0 {
		var cancel context.CancelFunc
		c.Context, cancel = context.WithTimeout(c.Context, r.Route.Timeout*time.Second)
		defer cancel()
	}

	switch {
	case r.mwhandler != nil:
		r.mwhandler(c)

	case r.mwgroup != "":
		DefaultMiddlewareGroupManager.Handle(c, r.mwgroup, UpstreamCall)

	default:
		UpstreamCall(c)
	}
}
