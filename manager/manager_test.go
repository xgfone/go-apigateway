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

package manager

import (
	"errors"
	"testing"
)

func TestManagerPointer(t *testing.T) {
	m := New[*int]()

	v1 := new(int)
	v2 := new(int)
	v3 := new(int)

	*v1 = 1
	*v2 = 2
	*v3 = 3

	m.Add("1", v1)
	m.Adds(map[string]*int{"2": v2, "3": v3})

	if v, _ := m.Get("1"); v == nil {
		t.Errorf("expect a value for '%s', but got nil", "1")
	} else if *v != 1 {
		t.Errorf("expect a value %d, but got %d", 1, *v)
	}
	if v, _ := m.Get("0"); v != nil {
		t.Errorf("unexpect a value for '%s', but got %d", "0", *v)
	}

	if v, _ := m.Del("1"); v == nil {
		t.Errorf("expect to delete a value %d, but got not", 1)
	} else if *v != 1 {
		t.Errorf("expect a value %d, but got %d", 1, *v)
	}

	m.Dels("2")
	if vs := m.Gets(); len(vs) != 1 {
		t.Errorf("expect 1 value, but got %d", len(vs))
	} else {
		for id := range vs {
			if id != "3" {
				t.Errorf("unexpect value %s", id)
			}
		}
	}

	m.Clear()
	if vs := m.Gets(); len(vs) != 0 {
		t.Errorf("unexpect any value, but got %+v", vs)
	}
}

func TestManagerInerface(t *testing.T) {
	m := New[error]()
	m.Add("1", errors.New("1"))
	m.Adds(map[string]error{"2": errors.New("2"), "3": errors.New("3")})
	m.Adds(nil)

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expect a panic, but got not")
			}
		}()

		m.Add("", nil)
	}()

	_, _ = m.Del("0")
	m.Dels("0")
	m.Dels()

	if vs := m.Gets(); len(vs) != 3 {
		t.Errorf("expect %d values, but got %d: %v", 3, len(vs), vs)
	} else {
		for id, v := range vs {
			if id != v.Error() {
				t.Errorf("unexpect value '%s' -> '%s'", id, v.Error())
			}
			switch id {
			case "1", "2", "3":
			default:
				t.Errorf("expect value '%s'", id)
			}
		}
	}
}
