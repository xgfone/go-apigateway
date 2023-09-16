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

package middleware

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/xgfone/go-apigateway/http/core"
)

func TestGroup(t *testing.T) {
	DefaultRegistry.Register("m1", func(name string, conf any) (Middleware, error) {
		return New(name, conf, func(next core.Handler) core.Handler {
			return func(c *core.Context) {
				c.ClientRequest.URL.Path += "/m1"
				next(c)
			}
		}), nil
	})

	DefaultRegistry.Register("m2", func(name string, conf any) (Middleware, error) {
		return New(name, conf, func(next core.Handler) core.Handler {
			return func(c *core.Context) {
				c.ClientRequest.URL.Path += "/m2"
				next(c)
			}
		}), nil
	})

	m1, err := DefaultRegistry.Build("m1", nil)
	if err != nil {
		t.Fatal(err)
	}
	m2, err := DefaultRegistry.Build("m2", nil)
	if err != nil {
		t.Fatal(err)
	}

	group := NewGroup("g1", m1, m2)

	if name := group.Name(); name != "g1" {
		t.Errorf("expect group name '%s', but got '%s'", "g1", name)
	}

	if mws := group.Middlewares(); len(mws) != 2 {
		t.Errorf("expect %d middlewares, but got %d", 2, len(mws))
	}

	DefaultGroupManager.Add(group.Name(), group)

	c := core.AcquireContext(context.Background())
	defer core.ReleaseContext(c)

	c.ClientRequest = &http.Request{URL: &url.URL{Path: "/path"}}
	HandleGroup(c, "g1", func(c *core.Context) {
		c.ClientRequest.URL.Path += "/handler"
	})

	expect := "/path/m1/m2/handler"
	if path := c.ClientRequest.URL.Path; path != expect {
		t.Errorf("expect path '%s', but got '%s'", expect, path)
	}

	if c.Error != nil {
		t.Error(err)
	}

	c.Error = nil
	HandleGroup(c, "none", func(c *core.Context) { panic("unreachable") })
	if c.Error == nil {
		t.Error("expect an error, but got nil")
	}
}
