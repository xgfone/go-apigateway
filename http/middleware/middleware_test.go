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

package middleware

import (
	"context"
	"reflect"
	"testing"

	"github.com/xgfone/go-apigateway/http/core"
)

func TestMiddlewares(t *testing.T) {
	var ss []string

	newmw := func(name string) Middleware {
		return New(name, nil, func(next core.Handler) core.Handler {
			return func(c *core.Context) {
				ss = append(ss, name)
				next(c)
			}
		})
	}

	ms := Middlewares{newmw("m1"), newmw("m2"), newmw("m3")}
	h := ms.Handler(func(c *core.Context) { ss = append(ss, "handler") })

	c := core.AcquireContext(context.Background())
	defer core.ReleaseContext(c)

	h(c)

	expects := []string{"m1", "m2", "m3", "handler"}
	if !reflect.DeepEqual(expects, ss) {
		t.Errorf("expect %v, but got %v", expects, ss)
	}
}
