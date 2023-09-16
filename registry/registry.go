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

// Package registry provides a common builder registry.
package registry

import "fmt"

// NotFoundError represents an error that the builder does not being found.
type NotFoundError struct{ Name string }

// Error implements the interface error.
func (e NotFoundError) Error() string {
	return fmt.Sprintf("not found the builder named '%s'", e.Name)
}

// Builder is used to build an instance named name by the config.
type Builder[T any] func(name string, conf any) (T, error)

// Registry is used to collect the builders.
type Registry[T any] struct{ builders map[string]Builder[T] }

// New returns a new common builder registry.
func New[T any]() *Registry[T] {
	return &Registry[T]{builders: make(map[string]Builder[T], 8)}
}

// Clear clears all the builders.
func (r *Registry[T]) Clear() { clear(r.builders) }

// Len returns the number of the builders.
func (r *Registry[T]) Len() int { return len(r.builders) }

// Names returns the names of all the builders.
func (r *Registry[T]) Names() []string {
	names := make([]string, 0, len(r.builders))
	for name := range r.builders {
		names = append(names, name)
	}
	return names
}

// Get returns the builder by the name.
func (r *Registry[T]) Get(name string) (Builder[T], bool) {
	b, ok := r.builders[name]
	return b, ok
}

// Unregister unregisters the builder by the name.
func (r *Registry[T]) Unregister(name string) { delete(r.builders, name) }

// Register registers the builder with the name.
func (r *Registry[T]) Register(name string, builder Builder[T]) {
	if name == "" {
		panic("Registry.Register: the builder name must not be empty")
	}
	r.builders[name] = builder
}

// Build builds the instance named name with the config.
func (r *Registry[T]) Build(name string, conf any) (T, error) {
	if builder, ok := r.builders[name]; ok {
		return builder(name, conf)
	}

	var v T
	return v, NotFoundError{Name: name}
}
