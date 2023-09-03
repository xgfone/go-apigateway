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

import "testing"

func TestMergeStrings(t *testing.T) {
	teststrings(t, "test1", mergestrings("a", nil), "a")
	teststrings(t, "test2", mergestrings("", []string{"a", "b"}), "a", "b")
	teststrings(t, "test3", mergestrings("a", []string{"b", "c"}), "b", "c", "a")
}

func teststrings(t *testing.T, prefix string, results []string, expects ...string) {
	if len(results) != len(expects) {
		t.Errorf("%s: expect %d strings, but got %d: %v", prefix, len(expects), len(results), results)
	} else {
		for i, s := range results {
			if s != expects[i] {
				t.Errorf("%s: %d: expect '%s', but got '%s'", prefix, i, expects[i], s)
			}
		}
	}
}
