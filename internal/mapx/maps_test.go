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

package mapx

import (
	"sort"
	"testing"
)

func TestKeys(t *testing.T) {
	teststrings(t, Keys(map[string]string{"a": "1", "b": "2", "c": "3"}), "a", "b", "c")
	teststrings(t, Keys(map[string]interface{}{"1": "a", "2": "b", "3": "b"}), "1", "2", "3")
}

func teststrings(t *testing.T, results []string, expects ...string) {
	sort.Strings(results)
	if len(expects) != len(results) {
		t.Errorf("expect %d strings, but got %d: %v", len(expects), len(results), results)
	} else {
		for i, s := range expects {
			if results[i] != s {
				t.Errorf("%d: expect '%s', but got '%s'", i, s, results[i])
			}
		}
	}
}
