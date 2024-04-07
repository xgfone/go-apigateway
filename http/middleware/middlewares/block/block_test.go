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

package block

import (
	"context"
	"net/http"
	"testing"

	"github.com/xgfone/go-apigateway/http/core"
	"github.com/xgfone/go-apigateway/http/middleware"
)

func TestBlock(t *testing.T) {
	_, err := middleware.DefaultRegistry.Build("block", "127.0.0.0/8")
	if err != nil {
		t.Error(err)
	}

	_, err = middleware.DefaultRegistry.Build("block", map[string]any{"cidrs": []string{"127.0.0.0/8"}})
	if err != nil {
		t.Error(err)
	}

	_, err = middleware.DefaultRegistry.Build("block", nil)
	if err == nil {
		t.Error("expect an error, but got nil")
	}

	_, err = middleware.DefaultRegistry.Build("block", "127.0.0.1")
	if err == nil {
		t.Error("expect an error, but got nil")
	}

	mw, err := middleware.DefaultRegistry.Build("block", []string{"127.0.0.0/8", "192.168.0.0/16"})
	if err != nil {
		t.Fatal(err)
	}

	c := core.AcquireContext(context.Background())

	handler := mw.Handler(func(c *core.Context) {})
	c.ClientRequest = &http.Request{RemoteAddr: "127.0.0.1"}
	handler(c)
	if c.Error == nil {
		t.Errorf("expect an error, but got nil")
	}

	c.Error = nil
	c.IsAborted = false
	c.ClientRequest = &http.Request{RemoteAddr: "192.168.1.1"}
	handler(c)
	if c.Error == nil {
		t.Errorf("expect an error, but got nil")
	}

	c.Error = nil
	c.IsAborted = false
	c.ClientRequest = &http.Request{RemoteAddr: "1.2.3.4"}
	handler(c)
	if c.Error != nil {
		t.Errorf("expect nil, but got an error: %v", c.Error)
	}

}
