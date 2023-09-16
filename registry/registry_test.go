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

package registry

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestRegistry(t *testing.T) {
	r := New[error]()
	r.Register("fix", func(name string, conf any) (error, error) {
		return errors.New(conf.(string)), nil
	})
	r.Register("fmt", func(name string, conf any) (error, error) {
		args := conf.([]interface{})
		return fmt.Errorf(args[0].(string), args[1:]...), nil
	})

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Error("expect a panic, but got not")
			}
		}()
		r.Register("", nil)
	}()

	if n := r.Len(); n != 2 {
		t.Errorf("expect %d builders, but got %d", 2, n)
	}

	if _, ok := r.Get("fix"); !ok {
		t.Error("expect a builder, but got not")
	}
	if _, ok := r.Get("none"); ok {
		t.Error("unexpect a builder, but got one")
	}

	if e, err := r.Build("fix", "test"); err != nil {
		t.Error(err)
	} else if s := e.Error(); s != "test" {
		t.Errorf("expect '%s', but got '%s'", "test", s)
	}

	if e, err := r.Build("fmt", []interface{}{"fmt: %s", "test"}); err != nil {
		t.Error(err)
	} else if s := e.Error(); s != "fmt: test" {
		t.Errorf("expect '%s', but got '%s'", "fmt: test", s)
	}

	if _, err := r.Build("none", nil); err == nil {
		t.Error("expect a build error, but got not")
	} else if e, ok := err.(NotFoundError); !ok {
		t.Errorf("expect a NotFoundError, but got %T", err)
	} else if s := e.Error(); !strings.HasPrefix(s, "not found") {
		t.Errorf("unexpect notfound error: %s", s)
	}

	r.Unregister("fix")
	if names := r.Names(); !reflect.DeepEqual(names, []string{"fmt"}) {
		t.Errorf("expect builder names %v, but got %v", []string{"fmt"}, names)
	}

	r.Clear()
	if names := r.Names(); !reflect.DeepEqual(names, []string{}) {
		t.Errorf("expect builder names %v, but got %v", []string{}, names)
	}
}
