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

package localfiles

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/xgfone/go-apigateway/pkg/http/dynamicconfig"
)

func dumpfile(filename string, v any) {
	file, _ := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	_ = json.NewEncoder(file).Encode(v)
	_ = file.Close()
}

func TestRouteProvider(t *testing.T) {
	route1 := dynamicconfig.Route{Id: "route1", Upstream: "up1"}
	route2 := dynamicconfig.Route{Id: "route2", Upstream: "up2"}
	route3 := dynamicconfig.Route{Id: "route3", Upstream: "up3"}
	routes := []dynamicconfig.Route{route1, route2, route3}
	dynamicconfig.SortRoutes(routes)

	dir := "testdata"
	_ = os.MkdirAll(dir, 0700)
	defer os.RemoveAll(dir)

	file1 := filepath.Join(dir, "route1.json")
	file2 := filepath.Join(dir, "route2.json")
	file3 := filepath.Join(dir, "route3.json")
	defer os.Remove(file1)
	defer os.Remove(file2)
	defer os.Remove(file3)

	dumpfile(file1, []any{route1})
	dumpfile(file2, []any{route2})
	dumpfile(file3, []any{route3})

	var lastetag string
	p := RouteProvider(dir)
	newroutes, newetag, err := p.Routes(lastetag)
	if err != nil {
		t.Fatal(err)
	} else if lastetag == newetag {
		t.Errorf("unexpect the same etag")
	} else {
		lastetag = newetag
		dynamicconfig.SortRoutes(newroutes)
		if !reflect.DeepEqual(routes, newroutes) {
			t.Errorf("expect routes %+v, but got %+v", routes, newroutes)
		}
	}

	_, newetag, err = p.Routes(lastetag)
	if err != nil {
		t.Fatal(err)
	} else if lastetag != newetag {
		t.Errorf("expect etag '%s', but got '%s'", lastetag, newetag)
		lastetag = newetag
	}

	route1.Id = "route"
	routes[0] = route1
	dumpfile(file1, []any{route1})

	newroutes, newetag, err = p.Routes(lastetag)
	if err != nil {
		t.Fatal(err)
	} else if lastetag == newetag {
		t.Errorf("unexpect the same etag")
	} else {
		dynamicconfig.SortRoutes(newroutes)
		if !reflect.DeepEqual(routes, newroutes) {
			t.Errorf("expect routes %+v, but got %+v", routes, newroutes)
		}
	}
}
