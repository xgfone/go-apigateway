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

// Package middleware provides a common handler middleware.
package middleware

import (
	"encoding/json"
	"fmt"

	"github.com/xgfone/go-apigateway/http/core"
	"github.com/xgfone/go-apigateway/registry"
	"github.com/xgfone/go-binder"
)

// DefaultRegistry is the global default registry of the middleware builder.
var DefaultRegistry = registry.New[Middleware]()

// Unwrapper is used to unwrap an inner middleware.
type Unwrapper interface {
	Unwrap() Middleware
}

// Middleware is the middleare to wrap the handler to return a new.
type Middleware interface {
	Handler(core.Handler) core.Handler
	Name() string
}

// MiddlewareFunc is the middleware wrapping function.
type MiddlewareFunc func(next core.Handler) core.Handler

// Handler implements the interface Middleware.
func (f MiddlewareFunc) Handler(next core.Handler) core.Handler { return f(next) }

// ------------------------------------------------------------------------ //

type middleware struct {
	MiddlewareFunc
	name string
	conf any
}

func (m middleware) Name() string { return m.name }
func (m middleware) Config() any  { return m.conf }

// New returns a new middleware.
func New(name string, config any, wrap MiddlewareFunc) Middleware {
	if wrap == nil {
		panic("middleware.New: the wrap handler function must not be nil")
	}
	return middleware{name: name, conf: config, MiddlewareFunc: wrap}
}

// ------------------------------------------------------------------------ //

// Middlewares represents a set of middlewares.
type Middlewares []Middleware

// Handler wraps the handler with the middlewares and returns a new handler.
func (ms Middlewares) Handler(handler core.Handler) core.Handler {
	for _len := len(ms) - 1; _len >= 0; _len-- {
		handler = ms[_len].Handler(handler)
	}
	return handler
}

// ------------------------------------------------------------------------ //

// BindConf builds the config dstconf of the middleware named name from srcconf.
//
// scrconf may be one of types as follow:
//   - map[string]any
//   - json.RawMessage
//   - []byte
func BindConf(name string, dstconf, srcconf any) (err error) {
	switch v := srcconf.(type) {
	case map[string]any:
		err = binder.BindStructToMap(dstconf, "json", v)

	case []byte:
		err = json.Unmarshal(v, dstconf)

	case json.RawMessage:
		err = json.Unmarshal(v, dstconf)

	default:
		return fmt.Errorf("Middleware<%s>: expect a map type, but got %T", name, srcconf)
	}

	if err != nil {
		err = fmt.Errorf("Middleware<%s>: %w", name, err)
	}

	return
}
