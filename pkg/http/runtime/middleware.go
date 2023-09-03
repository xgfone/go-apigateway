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
	"sort"

	"github.com/xgfone/go-apigateway/pkg/http/dynamicconfig"
)

// Middleware is the middleare to wrap the handler to a new.
type Middleware interface {
	Name() string
	Priority() int
	Handler(Handler) Handler
}

// Middlewares represents a set of middlewares.
type Middlewares []Middleware

// Handler wraps the handler with the middlewares and returns a new handler.
func (ms Middlewares) Handler(handler Handler) Handler {
	if len(ms) == 0 {
		return handler
	}

	sortmiddlewares(ms)
	for _, mw := range ms {
		handler = mw.Handler(handler)
	}
	return handler
}

func sortmiddlewares(mws Middlewares) {
	less := func(i, j int) bool { return mws[i].Priority() < mws[j].Priority() }
	if !sort.SliceIsSorted(mws, less) {
		sort.Slice(mws, less)
	}
}

type middleware struct {
	name string
	prio int
	conf any
	wrap func(Handler) Handler
}

func (m middleware) Name() string              { return m.name }
func (m middleware) Config() any               { return m.conf }
func (m middleware) Priority() int             { return m.prio }
func (m middleware) Handler(h Handler) Handler { return m.wrap(h) }

// NewMiddleware returns a new middleware.
func NewMiddleware(name string, prio int, config any, wrap func(next Handler) Handler) Middleware {
	if name == "" {
		panic("NewMiddleware: the middleware name must not be empty")
	} else if wrap == nil {
		panic("NewMiddleware: the wrap handler function must not be nil")
	}
	return middleware{name: name, prio: prio, conf: config, wrap: wrap}
}

func buildMiddlewaresHandler(next Handler, mws dynamicconfig.Middlewares) (Handler, error) {
	_mws, err := BuildMiddlewares(mws)
	if err == nil {
		next = _mws.Handler(next)
	}
	return next, err
}
