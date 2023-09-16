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

package directive

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/xgfone/go-apigateway/http/core"
)

func TestRewrite(t *testing.T) {
	origin := "/prefix/path/suffix"
	expect := "/prefix/suffix/path"

	p, err := DefaultRegistry.Build("rewrite", "/prefix(.*)/suffix", "/prefix/suffix$1")
	if err != nil {
		t.Fatal(err)
	}

	req := &http.Request{URL: &url.URL{Path: origin}}
	c := core.AcquireContext(req.Context())
	c.UpstreamRequest = req
	p.Process(c)

	if req.URL.Path != expect {
		t.Errorf("expect path '%s', but got '%s'", expect, req.URL.Path)
	}
}
