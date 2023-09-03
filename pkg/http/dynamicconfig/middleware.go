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

package dynamicconfig

import "reflect"

// Middlewares is the configuration of a set of middlewares.
type Middlewares map[string]map[string]any

// MiddlewareGroup is the configuration of a middleare group.
type MiddlewareGroups map[string]Middlewares

// ------------------------------------------------------------------------ //

// DiffMiddlewares compares the difference between new and old middlewares,
// and returns the added and deleted middlewares.
func DiffMiddlewares(news, olds Middlewares) (adds, dels Middlewares) {
	// add
	for name, mw := range news {
		if _mw, ok := olds[name]; !ok || !middlewareequal(mw, _mw) {
			if adds == nil {
				adds = make(Middlewares, len(news)/2)
			}
			adds[name] = mw
		}
	}

	// del
	for name, mw := range olds {
		if _, ok := news[name]; !ok {
			if dels == nil {
				dels = make(Middlewares, len(olds)/2)
			}
			dels[name] = mw
		}
	}

	return
}

// MiddlewaresDiff is used to represent the changed middlewares.
type MiddlewaresDiff struct {
	Adds Middlewares
	Dels Middlewares
}

// DiffMiddlewareGroups compares the difference between new and old middleware groups,
// and returns the changed middleware groups.
func DiffMiddlewareGroups(news, olds MiddlewareGroups) (changes map[string]MiddlewaresDiff) {
	changes = make(map[string]MiddlewaresDiff, len(news)/2)

	for id, mws := range news {
		if _mws, ok := olds[id]; !ok {
			changes[id] = MiddlewaresDiff{Adds: mws}
		} else if adds, dels := DiffMiddlewares(mws, _mws); len(adds) > 0 || len(dels) > 0 {
			changes[id] = MiddlewaresDiff{Adds: adds, Dels: dels}
		}
	}

	for id, mws := range olds {
		if _, ok := news[id]; !ok {
			diff := changes[id]
			diff.Dels = mws
			changes[id] = diff
		}
	}

	return
}

func middlewareequal(m1, m2 map[string]any) bool {
	return reflect.DeepEqual(m1, m2)
}
