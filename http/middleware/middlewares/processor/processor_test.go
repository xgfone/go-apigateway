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

package processor

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/xgfone/go-apigateway/http/core"
	"github.com/xgfone/go-apigateway/http/middleware"
)

func TestProcessor(t *testing.T) {
	_, err := middleware.DefaultRegistry.Build("processor", nil)
	if err == nil {
		t.Errorf("expect an error, but got nil")
	}

	_, err = middleware.DefaultRegistry.Build("processor", map[string]any{
		"directives": [][]string{[]string{}},
	})
	if err != nil {
		t.Errorf("unexpect an error, but got one: %v", err)
	}

	_, err = middleware.DefaultRegistry.Build("processor", map[string]any{
		"directives": []string{"addquery"},
	})
	if err == nil {
		t.Errorf("expect an error, but got nil")
	}

	_, err = middleware.DefaultRegistry.Build("processor", map[string]any{
		"directives": [][]string{[]string{"addquery"}},
	})
	if err == nil {
		t.Errorf("expect an error, but got nil")
	}

	// --------------------------------------------------------------------- //

	mw, err := middleware.DefaultRegistry.Build("processor", map[string]any{
		"directives": [][]string{[]string{"addquery", "key", "value"}},
	})
	if err != nil {
		t.Fatal(err)
	}

	c := core.AcquireContext(context.Background())
	c.UpstreamRequest = &http.Request{URL: new(url.URL)}
	mw.Handler(func(c *core.Context) {})(c)
	c.CallbackOnForward()

	expect := "key=value"
	if query := c.UpstreamRequest.URL.RawQuery; query != expect {
		t.Errorf("expect query '%s', but got '%s'", expect, query)
	}
}
