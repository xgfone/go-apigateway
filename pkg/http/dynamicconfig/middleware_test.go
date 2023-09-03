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
)

func TestDiffMiddlewares(t *testing.T) {
	mws1 := Middlewares{
		"allow": {"cidrs": []string{"127.0.0.0/8"}},
		"block": {"cidrs": []string{"0.0.0.0/0"}},
	}
	mws2 := Middlewares{
		"block": {"cidrs": []string{"0.0.0.0/0"}},
		"allow": {"cidrs": []string{"127.0.0.0/8"}},
	}

	adds, dels := DiffMiddlewares(mws1, mws2)
	if len(adds) > 0 {
		t.Errorf("unexpect added middlewares, but got %+v", adds)
	}
	if len(dels) > 0 {
		t.Errorf("unexpect deled middlewares, but got %+v", dels)
	}

	mws2 = Middlewares{
		"block":    {"cidrs": []string{"0.0.0.0/0"}},
		"redirect": {"code": 302, "location": "http://localhost"},
	}

	onlymw := func(mws Middlewares, name string) Middlewares {
		return Middlewares{name: mws[name]}
	}

	adds, dels = DiffMiddlewares(mws1, mws2)
	if len(adds) != 1 {
		t.Errorf("expect one added middlewares, but got %+v", adds)
	} else if expect := onlymw(mws1, "allow"); !reflect.DeepEqual(adds, expect) {
		t.Errorf("expect added middleware %+v, but got %+v", expect, adds)
	}

	if len(dels) != 1 {
		t.Errorf("expect one deled middlewares, but got %+v", dels)
	} else if expect := onlymw(mws2, "redirect"); !reflect.DeepEqual(dels, expect) {
		t.Errorf("expect deled middleware %+v, but got %+v", expect, dels)
	}
}

func TestDiffMiddlewareGroup(t *testing.T) {
	group1 := MiddlewareGroups{
		"g1": {
			"allow": {"cidrs": []string{"127.0.0.0/8"}},
			"block": {"cidrs": []string{"0.0.0.0/0"}},
		},
		"g2": {
			"allow": {"cidrs": []string{"127.0.0.0/8"}},
			"block": {"cidrs": []string{"0.0.0.0/0"}},
		},
	}

	group2 := MiddlewareGroups{
		"g1": {
			"allow": {"cidrs": []string{"127.0.0.0/8"}},
			"block": {"cidrs": []string{"0.0.0.0/0"}},
		},
		"g2": {
			"allow": {"cidrs": []string{"127.0.0.0/8"}},
			"block": {"cidrs": []string{"0.0.0.0/0"}},
		},
	}

	changes := DiffMiddlewareGroups(group1, group2)
	if len(changes) > 0 {
		t.Errorf("unexpect changed middleware group, but got %+v", changes)
	}

	group2["g2"] = Middlewares{
		"block":    {"cidrs": []string{"0.0.0.0/0"}},
		"redirect": {"code": 302, "location": "http://localhost"},
	}

	changes = DiffMiddlewareGroups(group1, group2)
	if len(changes) != 1 {
		t.Errorf("expect one changed middleware group, but got %+v", changes)
	} else if group, ok := changes["g2"]; !ok {
		t.Errorf("expect changed middleware group named '%s', but got %+v", "g2", changes)
	} else {
		onlymw := func(mws Middlewares, name string) Middlewares {
			return Middlewares{name: mws[name]}
		}

		adds, dels := group.Adds, group.Dels
		if len(adds) != 1 {
			t.Errorf("expect one added middlewares, but got %+v", adds)
		} else if expect := onlymw(group1["g2"], "allow"); !reflect.DeepEqual(adds, expect) {
			t.Errorf("expect added middleware %+v, but got %+v", expect, adds)
		}

		if len(dels) != 1 {
			t.Errorf("expect one deled middlewares, but got %+v", dels)
		} else if expect := onlymw(group2["g2"], "redirect"); !reflect.DeepEqual(dels, expect) {
			t.Errorf("expect deled middleware %+v, but got %+v", expect, dels)
		}
	}
}
