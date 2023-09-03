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
	"testing"

	"github.com/xgfone/go-apigateway/pkg/http/dynamicconfig"
)

func TestSortMiddlewares(t *testing.T) {
	RegisterMiddlewareBuilder("m1", func(name string, conf map[string]any) (Middleware, error) {
		return NewMiddleware(name, 1, nil, func(h Handler) Handler { return h }), nil
	})

	RegisterMiddlewareBuilder("m2", func(name string, conf map[string]any) (Middleware, error) {
		return NewMiddleware(name, 2, nil, func(h Handler) Handler { return h }), nil
	})

	RegisterMiddlewareBuilder("m3", func(name string, conf map[string]any) (Middleware, error) {
		return NewMiddleware(name, 3, nil, func(h Handler) Handler { return h }), nil
	})

	mws, err := BuildMiddlewares(dynamicconfig.Middlewares{"m2": nil, "m3": nil, "m1": nil})
	if err != nil {
		t.Fatal(err)
	}

	sortmiddlewares(mws)
	if len(mws) != 3 {
		t.Errorf("expect %d middlewares, but got %d", 3, len(mws))
	} else if name := mws[0].Name(); name != "m1" {
		t.Errorf("expect middleware named '%s', but got '%s'", "m1", name)
	} else if name := mws[1].Name(); name != "m2" {
		t.Errorf("expect middleware named '%s', but got '%s'", "m2", name)
	} else if name := mws[2].Name(); name != "m3" {
		t.Errorf("expect middleware named '%s', but got '%s'", "m3", name)
	}

	if mws.Handler(nil) != nil {
		t.Error("expect an nil handler, but got other")
	}
}
