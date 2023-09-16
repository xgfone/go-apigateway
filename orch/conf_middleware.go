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

package orch

import (
	"reflect"
	"slices"
)

// Middleware represents a middleware.
type Middleware struct {
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	Conf any    `json:"conf,omitempty" yaml:"conf,omitempty"`
}

// Middlewares represents a set of middlewares.
type Middlewares []Middleware

// MiddlewareGroup is the configuration of a middleare group.
type MiddlewareGroup struct {
	// Optional, One of "tcp", "http"
	//
	// Default: http
	Type string `json:"type,omitempty" yaml:"type,omitempty"`

	// Required
	Name        string      `json:"name,omitempty" yaml:"name,omitempty"`
	Middlewares Middlewares `json:"middlewares,omitempty" yaml:"middlewares,omitempty"`
}

// Equal reports whether it is equal to other middlewares.
func (g MiddlewareGroup) Equal(other MiddlewareGroup) bool {
	return reflect.DeepEqual(g, other)
}

// DiffMiddlewareGroups compares the difference
// between new and old middleware groups,
// and returns the changed middleware groups.
func DiffMiddlewareGroups(news, olds []MiddlewareGroup) (adds, dels []MiddlewareGroup) {
	adds = make([]MiddlewareGroup, 0, len(news)/2)
	dels = make([]MiddlewareGroup, 0, len(olds)/2)
	names := make(map[string]struct{}, len(news))

	// add
	for _, group := range news {
		names[group.Name] = struct{}{}
		index := findmwgroup(olds, group.Name)
		if index < 0 || !group.Equal(olds[index]) {
			adds = append(adds, group)
		}
	}

	// del
	for _, group := range olds {
		if _, ok := names[group.Name]; !ok {
			dels = append(dels, group)
		}
	}

	return
}

func findmwgroup(routes []MiddlewareGroup, name string) (index int) {
	return slices.IndexFunc(routes, func(g MiddlewareGroup) bool { return g.Name == name })
}
