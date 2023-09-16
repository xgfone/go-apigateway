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
	"fmt"
	"sync"

	"github.com/xgfone/go-apigateway/http/core"
	"github.com/xgfone/go-apigateway/manager"
	"github.com/xgfone/go-atomicvalue"
)

// DefaultGroupManager is the default global manager of middleware groups.
var DefaultGroupManager = manager.New[*Group]()

// GroupGetter is used to get the middleware group by the name.
type GroupGetter interface {
	Get(name string) (g *Group, ok bool)
}

// HandleGroupWithGetter is a convenient function, which is equal to
//
//	HandleGroupWithGetter(c, DefaultGroupManager, group, next)
func HandleGroup(c *core.Context, group string, next core.Handler) {
	HandleGroupWithGetter(c, DefaultGroupManager, group, next)
}

// HandleGroupWithGetter forwards the request to the middleware group,
// which is got from the middleware group getter g by the group name, to handle.
func HandleGroupWithGetter(c *core.Context, g GroupGetter, group string, next core.Handler) {
	if c.IsAborted {
		return
	}

	if g, ok := g.Get(group); ok {
		g.Handle(c, next)
	} else {
		c.Abort(fmt.Errorf("not found the middleware group '%s'", group))
	}
}

// Group is used to manage a set of middlewares.
type Group struct {
	name    string
	lock    sync.Mutex
	mdws    atomicvalue.Value[Middlewares]
	handler atomicvalue.Value[core.Handler]
}

// NewGroup returns a new middleware group, which also adds the middlewares if exists.
func NewGroup(name string, mws ...Middleware) *Group {
	if name == "" {
		panic("middleware.Group.New: name must not be empty")
	}

	g := &Group{name: name}
	g.Reset(mws...)
	return g
}

// Name returns the name of the group.
func (g *Group) Name() string { return g.name }

// Middlewares returns all the middlewares.
func (g *Group) Middlewares() Middlewares { return g.mdws.Load() }

// Handle handles the request with the middlewares, and forwards it to next.
func (g *Group) Handle(c *core.Context, next core.Handler) {
	c.Next = next
	g.handler.Load()(c)
}

// Reset resets the middlewares to mws.
func (g *Group) Reset(mws ...Middleware) {
	g.lock.Lock()
	defer g.lock.Unlock()

	g.mdws.Store(mws)
	g.handler.Store(Middlewares(mws).Handler(handlenext))
}

func handlenext(c *core.Context) { c.Next(c) }
