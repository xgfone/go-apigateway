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

import (
	"reflect"
	"testing"
	"time"
)

var (
	route1 = Route{
		Id:       "route",
		Upstream: "upstream",

		Timeout:         time.Second,
		Priority:        123,
		MiddlewareGroup: "group",
		Middlewares: Middlewares{
			"allow": map[string]any{"cidrs": []string{"127.0.0.1/8"}},
			"block": map[string]any{"cidrs": []string{"0.0.0.0/0"}},
		},

		Matcher: Matcher{
			Method:  "GET",
			Paths:   []string{"/path1", "/path2"},
			Headers: map[string]string{"key1": "value1", "key2": "value2"},
		},
	}

	route2 = Route{
		Id:       "route",
		Upstream: "upstream",

		Timeout:         time.Second,
		Priority:        123,
		MiddlewareGroup: "group",
		Middlewares: Middlewares{
			"allow": map[string]any{"cidrs": []string{"127.0.0.1/8"}},
			"block": map[string]any{"cidrs": []string{"0.0.0.0/0"}},
		},

		Matcher: Matcher{
			Method:  "GET",
			Paths:   []string{"/path1", "/path2"},
			Headers: map[string]string{"key2": "value2", "key1": "value1"},
		},
	}
)

func TestRouteEqual(t *testing.T) {
	route1 := route1
	route2 := route2
	if !reflect.DeepEqual(route1, route2) {
		t.Error("expect true, but got false")
	}

	route2.Matcher.Paths = []string{"/path2", "/path1"}
	if reflect.DeepEqual(route1, route2) {
		t.Error("expect false, but got true")
	}
}

func TestDiffRoutes(t *testing.T) {
	route1 := route1
	route2 := route2
	route1.Id = "route1"
	route2.Id = "route2"

	routes1 := []Route{route1, route2}
	routes2 := []Route{route2, route1}

	adds, dels := DiffRoutes(routes1, routes2)
	if len(adds) != 0 {
		t.Errorf("unexpect added routes, but got %+v", adds)
	}
	if len(dels) != 0 {
		t.Errorf("unexpect deled routes, but got %+v", dels)
	}

	routes1[0].Matcher.Paths = []string{"/path2", "/path1"}
	adds, dels = DiffRoutes(routes1, routes2)

	if len(adds) != 1 {
		t.Errorf("expect one added routes, but got %+v", adds)
	} else if !routeequal(adds[0], routes1[0]) {
		t.Errorf("expect added route %+v, but got %+v", routes1[0], adds[0])
	}
	if len(dels) != 0 {
		t.Errorf("unexpect deled routes, but got %+v", dels)
	}

	routes2[1].Id = "route3"
	adds, dels = DiffRoutes(routes1, routes2)
	if !reflect.DeepEqual(adds, routes1) {
		t.Errorf("expect added routes %+v, but got %+v", routes1, adds)
	}

	if len(dels) != 1 {
		t.Errorf("expect one deled routes, but got %+v", dels)
	} else if !routeequal(dels[0], routes2[1]) {
		t.Errorf("expect deled route %+v, but got %+v", routes2[1], dels[0])
	}
}
