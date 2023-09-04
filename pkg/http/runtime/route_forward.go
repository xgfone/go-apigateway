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
	"time"

	"github.com/xgfone/go-defaults"
)

// Match reports whether the route matches the http request or not.
func (r *Route) Match(c *Context) bool {
	return r.matcher.Match(c)
}

// Handle handles and forwards the http request by the route.
func (r *Route) Handle(c *Context) {
	defer r.recover(c)

	if r.Route.Timeout > 0 {
		var cancel context.CancelFunc
		c.Context, cancel = context.WithTimeout(c.Context, r.Route.Timeout*time.Second)
		defer cancel()
	}

	switch {
	case r.mwhandler != nil:
		r.mwhandler(c)

	case r.mwgroup != "":
		DefaultMiddlewareGroupManager.Handle(c, r.mwgroup, UpstreamForward)

	default:
		UpstreamForward(c)
	}
}

func (r *Route) recover(c *Context) {
	if r := recover(); r != nil {
		defaults.HandlePanicContext(c.Context, r)
		if e, ok := r.(error); ok {
			c.Abort(fmt.Errorf("panic: %w", e))
		} else {
			c.Abort(fmt.Errorf("panic: %v", r))
		}
	}
}
