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
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/xgfone/go-apigateway/pkg/http/dynamicconfig"
)

func TestUpstreamProvider(t *testing.T) {
	up1 := dynamicconfig.Upstream{Id: "up1"}
	up2 := dynamicconfig.Upstream{Id: "up2"}
	up3 := dynamicconfig.Upstream{Id: "up2"}
	ups := []dynamicconfig.Upstream{up1, up2, up3}
	dynamicconfig.SortUpstreams(ups)

	dir := "testdata"
	_ = os.MkdirAll(dir, 0700)
	defer os.RemoveAll(dir)

	file1 := filepath.Join(dir, "up1.json")
	file2 := filepath.Join(dir, "up2.json")
	file3 := filepath.Join(dir, "up3.json")
	defer os.Remove(file1)
	defer os.Remove(file2)
	defer os.Remove(file3)

	dumpfile(file1, []any{up1})
	dumpfile(file2, []any{up2})
	dumpfile(file3, []any{up3})

	var lastetag string
	p := UpstreamProvider(dir)

	newups, newetag, err := p.Upstreams(lastetag)
	if err != nil {
		t.Fatal(err)
	} else if lastetag == newetag {
		t.Errorf("unexpect the same etag")
	} else {
		lastetag = newetag
		dynamicconfig.SortUpstreams(newups)
		if !reflect.DeepEqual(ups, newups) {
			t.Errorf("expect upstreams %+v, but got %+v", ups, newups)
		}
	}

	_, newetag, err = p.Upstreams(lastetag)
	if err != nil {
		t.Fatal(err)
	} else if lastetag != newetag {
		t.Errorf("expect etag '%s', but got '%s'", lastetag, newetag)
		lastetag = newetag
	}

	up1.Id = "up"
	ups[0] = up1
	dumpfile(file1, []any{up1})

	newups, newetag, err = p.Upstreams(lastetag)
	if err != nil {
		t.Fatal(err)
	} else if lastetag == newetag {
		t.Errorf("unexpect the same etag")
	} else {
		dynamicconfig.SortUpstreams(newups)
		if !reflect.DeepEqual(ups, newups) {
			t.Errorf("expect upstreams %+v, but got %+v", ups, newups)
		}
	}
}
