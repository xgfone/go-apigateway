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

package allow

import (
	"context"
	"net/http"
	"testing"

	"github.com/xgfone/go-apigateway/pkg/http/runtime"
)

func TestAllow(t *testing.T) {
	mw, err := Allow("allow", 0, "127.0.0.0/8", "192.168.0.0/16")
	if err != nil {
		t.Fatal(err)
	}

	var merr error
	c := runtime.AcquireContext()
	c.Context = context.Background()
	c.SetModeForward()
	c.SetResponseHandler(func(ctx *runtime.Context, r *http.Response, err error) {
		merr = err
	})

	handler := mw.Handler(func(c *runtime.Context) {})
	c.ClientRequest = &http.Request{RemoteAddr: "127.0.0.1"}
	handler(c)
	if merr != nil {
		t.Errorf("expect nil, but got an error: %v", merr)
	}

	merr = nil
	c.ClientRequest = &http.Request{RemoteAddr: "192.168.1.1"}
	handler(c)
	if merr != nil {
		t.Errorf("expect nil, but got an error: %v", merr)
	}

	merr = nil
	c.ClientRequest = &http.Request{RemoteAddr: "1.2.3.4"}
	handler(c)
	if merr == nil {
		t.Errorf("expect an error, but got nil")
	}
}
