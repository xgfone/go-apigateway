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
	"maps"
	"sync"

	"github.com/xgfone/go-atomicvalue"
)

// MiddlewareGroup is used to manage a set of middlewares.
type MiddlewareGroup struct {
	lock sync.Mutex
	mdwm map[string]Middleware

	mdws    atomicvalue.Value[Middlewares]
	handler atomicvalue.Value[Handler]
}

// NewMiddlewareGroup returns a new middleware group, and adds the middlewares.
func NewMiddlewareGroup(mws ...Middleware) *MiddlewareGroup {
	g := &MiddlewareGroup{mdwm: make(map[string]Middleware, 4)}
	if len(mws) == 0 {
		g.handler.Store(g.handle)
	} else {
		g.AddMiddlewares(mws...)
	}
	return g
}

// AddMiddlewares adds a set of middlewares in bulk.
func (g *MiddlewareGroup) AddMiddlewares(mws ...Middleware) {
	if len(mws) == 0 {
		return
	}

	g.lock.Lock()
	defer g.lock.Unlock()
	for _, mw := range mws {
		g.mdwm[mw.Name()] = mw
	}
	g.update()
}

// DelMiddlewares deletes a set of middlewares by the names in bulk.
func (g *MiddlewareGroup) DelMiddlewares(names ...string) {
	if len(names) == 0 {
		return
	}

	g.lock.Lock()
	defer g.lock.Unlock()
	for _, name := range names {
		delete(g.mdwm, name)
	}
	g.update()
}

// Middlewares returns all the middlewares.
func (g *MiddlewareGroup) Middlewares() Middlewares {
	return g.mdws.Load()
}

func (g *MiddlewareGroup) update() {
	mdws := make(Middlewares, 0, len(g.mdwm))
	for _, mw := range g.mdwm {
		mdws = append(mdws, mw)
	}

	g.mdws.Store(mdws)
	g.handler.Store(mdws.Handler(g.handle))
}

func (g *MiddlewareGroup) handle(c *Context) { c.next(c) }

// Handle handles the request with the middlewares, and forwards it to next.
func (g *MiddlewareGroup) Handle(c *Context) { g.handler.Load()(c) }

// ------------------------------------------------------------------------- //

// DefaultMiddlewareGroupManager is the default global manager of middleware groups.
var DefaultMiddlewareGroupManager = NewMiddlewareGroupManager()

// MiddlewareGroupManager is used to manage a set of middleware groups.
type MiddlewareGroupManager struct {
	lock   sync.Mutex
	groupm map[string]*MiddlewareGroup
	groups atomicvalue.Value[map[string]*MiddlewareGroup]
}

// NewMiddlewareGroupManager returns a new middleware group manager.
func NewMiddlewareGroupManager() *MiddlewareGroupManager {
	return &MiddlewareGroupManager{groupm: make(map[string]*MiddlewareGroup, 8)}
}

/*
// AddGroups adds a set of middleware groups in bulk.
func (m *MiddlewareGroupManager) AddGroups(groups map[string]*MiddlewareGroup) {
	if len(groups) == 0 {
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	for name, group := range groups {
		m.groupm[name] = group
	}
	m.update()
}

// DelGroups deletes a set of middleware groups in bulk.
func (m *MiddlewareGroupManager) DelGroups(names ...string) {
	if len(names) == 0 {
		return
	}

	m.lock.Lock()
	defer m.lock.Unlock()
	for _, name := range names {
		delete(m.groupm, name)
	}
	m.update()
}
*/

// AddGroup adds the middleware group with the name.
//
// Override it if the group name has existed.
func (m *MiddlewareGroupManager) AddGroup(name string, group *MiddlewareGroup) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.groupm[name] = group
	m.update()
}

// DelGroup deletes the middleware group by the group name.
func (m *MiddlewareGroupManager) DelGroup(name string) {
	m.lock.Lock()
	defer m.lock.Unlock()
	if _, ok := m.groupm[name]; ok {
		delete(m.groupm, name)
		m.update()
	}
}

// GetGroup returns the middleware group by the group name.
func (m *MiddlewareGroupManager) GetGroup(name string) *MiddlewareGroup {
	return m.groups.Load()[name]
}

// Groups returns all the middleware groups, which is read-only.
func (m *MiddlewareGroupManager) Groups() map[string]*MiddlewareGroup {
	return m.groups.Load()
}

func (m *MiddlewareGroupManager) update() {
	m.groups.Store(maps.Clone(m.groupm))
}

// Handle finds the middleware group and the group name,
// and forwards the request to it to handle.
func (m *MiddlewareGroupManager) Handle(c *Context, group string, next Handler) {
	if g := m.GetGroup(group); g != nil {
		c.next = next
		g.Handle(c)
	} else {
		err := fmt.Errorf("not found the middleware group '%s'", group)
		c.SendResponse(nil, ErrInternalServerError.WithError(err))
	}
}
