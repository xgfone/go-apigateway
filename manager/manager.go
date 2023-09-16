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

// Package manager provides a common object manager.
package manager

import (
	"maps"
	"sync"
	"sync/atomic"
)

// Manager is used to manage a kind of objects.
type Manager[V any] struct {
	lock sync.Mutex
	objm map[string]V
	objv atomic.Value // map[string]V
}

// New returns a new manager.
func New[V any]() *Manager[V] {
	m := &Manager[V]{objm: make(map[string]V, 8)}
	m.objv.Store(map[string]V(nil))
	return m
}

func (m *Manager[V]) update() { m.objv.Store(maps.Clone(m.objm)) }

// Gets returns all the values, which is read-only and must not be modified.
func (m *Manager[V]) Gets() map[string]V { return m.objv.Load().(map[string]V) }

// Get returns the value by the id.
//
// If not exist, return (ZERO, false).
func (m *Manager[V]) Get(id string) (v V, ok bool) {
	v, ok = m.Gets()[id]
	return v, ok
}

// Add adds the value v with the id.
//
// If exists, override it.
func (m *Manager[V]) Add(id string, v V) {
	if id == "" {
		panic("Manager.Add: the id must not be empty")
	}

	m.lock.Lock()
	m.objm[id] = v
	m.update()
	m.lock.Unlock()
}

// Add adds the value v with the id.
//
// If exists, override it.
func (m *Manager[V]) Adds(vs map[string]V) {
	if len(vs) == 0 {
		return
	}

	m.lock.Lock()
	maps.Copy(m.objm, vs)
	m.update()
	m.lock.Unlock()
}

// Del deletes and returns the value by the id.
//
// If not exist, do nothing and return (ZERO, false).
func (m *Manager[V]) Del(id string) (v V, ok bool) {
	m.lock.Lock()
	if v, ok = m.objm[id]; ok {
		delete(m.objm, id)
		m.update()
	}
	m.lock.Unlock()
	return
}

// Dels deletes and returns the value by the id.
func (m *Manager[V]) Dels(ids ...string) {
	if len(ids) == 0 {
		return
	}

	var exist bool
	m.lock.Lock()
	for _, id := range ids {
		_, ok := m.objm[id]
		if ok {
			delete(m.objm, id)
			exist = true
		}
	}
	if exist {
		m.update()
	}
	m.lock.Unlock()
}

// Clear clears all the values.
func (m *Manager[V]) Clear() {
	m.lock.Lock()
	clear(m.objm)
	m.objv.Store(map[string]V(nil))
	m.lock.Unlock()
}
