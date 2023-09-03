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

func TestSortRoutes(t *testing.T) {
	routes := []Route{
		{Route: dynamicconfig.Route{Id: "route1", Priority: 1}},
		{Route: dynamicconfig.Route{Id: "route3", Priority: 3}},
		{Route: dynamicconfig.Route{Id: "route2", Priority: 2}},
	}
	sortroutes(routes)
	if id := routes[0].Route.Id; id != "route3" {
		t.Errorf("expect route '%s', but got '%s'", "route3", id)
	}
	if id := routes[1].Route.Id; id != "route2" {
		t.Errorf("expect route '%s', but got '%s'", "route2", id)
	}
	if id := routes[2].Route.Id; id != "route1" {
		t.Errorf("expect route '%s', but got '%s'", "route1", id)
	}

	routes[0].Route.Protect = true
	sortroutes(routes)
	if id := routes[0].Route.Id; id != "route2" {
		t.Errorf("expect route '%s', but got '%s'", "route2", id)
	}
	if id := routes[1].Route.Id; id != "route1" {
		t.Errorf("expect route '%s', but got '%s'", "route1", id)
	}
	if id := routes[2].Route.Id; id != "route3" {
		t.Errorf("expect route '%s', but got '%s'", "route3", id)
	}

	routes[0].Route.Protect = true
	sortroutes(routes)
	if id := routes[0].Route.Id; id != "route1" {
		t.Errorf("expect route '%s', but got '%s'", "route1", id)
	}
	if id := routes[1].Route.Id; id != "route3" {
		t.Errorf("expect route '%s', but got '%s'", "route3", id)
	}
	if id := routes[2].Route.Id; id != "route2" {
		t.Errorf("expect route '%s', but got '%s'", "route2", id)
	}
}
