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
	"maps"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/xgfone/go-apigateway/pkg/http/dynamicconfig"
)

func TestMiddlewareGroupProvider(t *testing.T) {
	group1 := dynamicconfig.MiddlewareGroups{"g1": {"mw1": map[string]any{"key1": "value1"}}}
	group2 := dynamicconfig.MiddlewareGroups{"g2": {"mw2": map[string]any{"key2": "value2"}}}
	group3 := dynamicconfig.MiddlewareGroups{"g3": {"mw3": map[string]any{"key3": "value3"}}}

	groups := make(dynamicconfig.MiddlewareGroups)
	maps.Copy(groups, group1)
	maps.Copy(groups, group2)
	maps.Copy(groups, group3)

	dir := "testdata"
	_ = os.MkdirAll(dir, 0700)
	defer os.RemoveAll(dir)

	file1 := filepath.Join(dir, "mg1.json")
	file2 := filepath.Join(dir, "mg2.json")
	file3 := filepath.Join(dir, "mg3.json")
	defer os.Remove(file1)
	defer os.Remove(file2)
	defer os.Remove(file3)

	dumpfile(file1, group1)
	dumpfile(file2, group2)
	dumpfile(file3, group3)

	var lastetag string
	p := MiddlewareGroupProvider(dir)
	newgroups, newetag, err := p.MiddlewareGroups(lastetag)
	if err != nil {
		t.Fatal(err)
	} else if lastetag == newetag {
		t.Errorf("unexpect the same etag")
	} else {
		lastetag = newetag
		if !reflect.DeepEqual(groups, newgroups) {
			t.Errorf("expect middleware groups %+v, but got %+v", groups, newgroups)
		}
	}

	_, newetag, err = p.MiddlewareGroups(lastetag)
	if err != nil {
		t.Fatal(err)
	} else if lastetag != newetag {
		t.Errorf("expect etag '%s', but got '%s'", lastetag, newetag)
		lastetag = newetag
	}

	mw1 := group1["g1"]["mw1"]
	delete(group1["g1"], "mw1")
	group1["g1"]["mw"] = mw1
	dumpfile(file1, group1)

	newgroups, newetag, err = p.MiddlewareGroups(lastetag)
	if err != nil {
		t.Fatal(err)
	} else if lastetag == newetag {
		t.Errorf("unexpect the same etag")
	} else {
		if !reflect.DeepEqual(groups, newgroups) {
			t.Errorf("expect middleware groups %+v, but got %+v", groups, newgroups)
		}
	}
}
