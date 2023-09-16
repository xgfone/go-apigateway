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

// Package loader provides some resource loading functions.
package loader

import "sync"

// ResourceManager is used to manage a kind of resource.
type ResourceManager[R any] struct {
	lock sync.RWMutex
	etag string
	rsrc R
}

// NewResourceManager returns a new resource manager.
func NewResourceManager[R any]() *ResourceManager[R] {
	return new(ResourceManager[R])
}

// Resource only returns the resource.
func (m *ResourceManager[R]) Resource() (rsc R) {
	m.lock.RLock()
	rsc = m.rsrc
	m.lock.RUnlock()
	return
}

// Etag only returns the etag.
func (m *ResourceManager[R]) Etag() string {
	m.lock.RLock()
	etag := m.etag
	m.lock.RUnlock()
	return etag
}

// Get returns the resource and etag.
func (m *ResourceManager[R]) Get() (rsc R, etag string) {
	m.lock.RLock()
	rsc, etag = m.rsrc, m.etag
	m.lock.RUnlock()
	return
}

// Set resets the resource and etag.
func (m *ResourceManager[R]) Set(rsc R, etag string) {
	m.lock.Lock()
	m.rsrc, m.etag = rsc, etag
	m.lock.Unlock()
}

// SetResource only sets the resource.
func (m *ResourceManager[R]) SetResource(rsc R) {
	m.lock.Lock()
	m.rsrc = rsc
	m.lock.Unlock()
}

// SetEtag only resets the etag.
func (m *ResourceManager[R]) SetEtag(etag string) {
	m.lock.Lock()
	m.etag = etag
	m.lock.Unlock()
}
