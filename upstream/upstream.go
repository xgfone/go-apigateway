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

// Package upstream provides some upstream functions.
package upstream

import (
	"github.com/xgfone/go-apigateway/manager"
	"github.com/xgfone/go-atomicvalue"
	"github.com/xgfone/go-loadbalancer/forwarder"
)

// Manager is used to manage a set of the http upstreams.
var Manager = manager.New[*Upstream]()

// Upstream represents a http upstream.
type Upstream struct {
	*forwarder.Forwarder

	scheme atomicvalue.Value[string]
	host   atomicvalue.Value[string]
}

// New returns a new upstream based on the forwarder.
func New(forwarder *forwarder.Forwarder) *Upstream {
	return &Upstream{Forwarder: forwarder}
}

// Host returns the host of the upstream.
func (u *Upstream) Host() string { return u.host.Load() }

// Scheme returns the scheme of the upstream.
func (u *Upstream) Scheme() string { return u.scheme.Load() }

// SetHost sets the host of the upstream.
func (u *Upstream) SetHost(host string) { u.host.Store(host) }

// SetScheme sets the scheme of the upstream.
func (u *Upstream) SetScheme(scheme string) { u.scheme.Store(scheme) }
