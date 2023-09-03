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

// Package loader provides some lodaders to load the dynamic configurations.
package loader

import (
	"context"
	"time"

	"github.com/xgfone/go-atomicvalue"
	"github.com/xgfone/go-defaults"
)

// Loader is used to load a certain kind of resource.
type Loader interface {
	Load(context.Context)
}

// LoaderFunc is a loader function.
type LoaderFunc func(context.Context)

// Load implements the interface Loader.
func (f LoaderFunc) Load(c context.Context) { f(c) }

// ResourceLoader is the extended loader to return the inner loaded resource.
type ResourceLoader interface {
	Resource() any
	Loader
}

// ResourceLoaderFunc converts a function to a resource loader.
func ResourceLoaderFunc(load func(c context.Context, cb func(any))) ResourceLoader {
	return &resourceLoader{load: load}
}

type resourceLoader struct {
	rsc  atomicvalue.Value[any]
	load func(context.Context, func(any))
}

func (l *resourceLoader) Load(c context.Context) { l.load(c, l.cb) }
func (l *resourceLoader) Resource() any          { return l.rsc.Load() }
func (l *resourceLoader) cb(rsc any)             { l.rsc.Store(rsc) }

func load(ctx context.Context, interval time.Duration, load func()) {
	saferun(load)
	if interval < time.Second {
		interval = time.Minute
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return

		case <-ticker.C:
			saferun(load)
		}
	}
}

func saferun(f func()) {
	defer func() {
		if r := recover(); r != nil {
			defaults.HandlePanic(r)
		}
	}()
	f()
}
