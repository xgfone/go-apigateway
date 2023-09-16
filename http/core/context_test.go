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

package core

import (
	"context"
	"testing"
)

func TestContext(t *testing.T) {
	c := AcquireContext(context.Background())
	c.OnResponseHeader(func() {})
	c.OnForward(func() {})
	ReleaseContext(c)

	c = AcquireContext(context.Background())

	if _len := len(c.forwards); _len != 0 {
		t.Errorf("expect %d len forward callback functions, but got %d", 0, _len)
	}
	if _cap := cap(c.forwards); _cap != 4 {
		t.Errorf("expect %d cap forward callback functions, but got %d", 4, _cap)
	}

	if _len := len(c.respheaders); _len != 0 {
		t.Errorf("expect %d len respheader callback functions, but got %d", 0, _len)
	}
	if _cap := cap(c.respheaders); _cap != 4 {
		t.Errorf("expect %d cap respheader callback functions, but got %d", 4, _cap)
	}
}
