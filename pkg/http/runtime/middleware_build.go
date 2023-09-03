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

import "fmt"

var mwbuilders = make(map[string]builder, 8)

type builder func(name string, conf map[string]any) (Middleware, error)

// ResetMiddlewareBuilder clears and resets the middleware builders to ZERO.
func ResetMiddlewareBuilder() { clear(mwbuilders) }

// RegisterMiddlewareBuilder registers the middleware builder with the name.
func RegisterMiddlewareBuilder(name string, builder func(name string, conf map[string]any) (Middleware, error)) {
	mwbuilders[name] = builder
}

// BuildMiddleware is a convenient function to find the builder
// and build the middleware with the config.
func BuildMiddleware(name string, conf map[string]any) (Middleware, error) {
	if builder, ok := mwbuilders[name]; ok {
		return builder(name, conf)
	}
	return nil, fmt.Errorf("not found the middleware builder named '%s'", name)
}

// BuildMiddlewares is a convenient function to builds a set of middlewares.
func BuildMiddlewares(name2conf map[string]map[string]any) (Middlewares, error) {
	if len(name2conf) == 0 {
		return nil, nil
	}

	mws := make(Middlewares, 0, len(name2conf))
	for name, conf := range name2conf {
		mw, err := BuildMiddleware(name, conf)
		if err != nil {
			return nil, fmt.Errorf("fail to build the middleware '%s': %w", name, err)
		}
		mws = append(mws, mw)
	}

	sortmiddlewares(mws)
	return mws, nil
}
