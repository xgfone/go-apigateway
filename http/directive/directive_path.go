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
	"regexp"
	"strings"

	"github.com/xgfone/go-apigateway/http/core"
)

func init() {
	fix := func(s string) string { return strings.TrimSuffix(s, "/") }

	DefaultRegistry.RegisterOneArg("setpath", "string: change the path of the request",
		func(c *core.Context, s string) {
			setpath(c.UpstreamRequest, s)
		},
	)

	DefaultRegistry.RegisterOneArg("addprefix", "string: add a prefix to the path of the request",
		func(c *core.Context, s string) {
			setpath(c.UpstreamRequest, fix(s)+c.UpstreamRequest.URL.Path)
		},
	)

	DefaultRegistry.RegisterOneArg("addsuffix", "string: add a suffix to the path of the request",
		func(c *core.Context, s string) {
			setpath(c.UpstreamRequest, c.UpstreamRequest.URL.Path+fix(s))
		},
	)

	DefaultRegistry.RegisterOneArg("delprefix", "string: remove a prefix from the path of the request",
		func(c *core.Context, s string) {
			setpath(c.UpstreamRequest, strings.TrimPrefix(c.UpstreamRequest.URL.Path, fix(s)))
		},
	)

	DefaultRegistry.RegisterOneArg("delsuffix", "string: remove a suffix from the path of the request",
		func(c *core.Context, s string) {
			setpath(c.UpstreamRequest, strings.TrimSuffix(c.UpstreamRequest.URL.Path, fix(s)))
		},
	)

	DefaultRegistry.RegisterTwoArgs("replaceprefix",
		"string: the replaced original prefix of the path of the request",
		"string: the new prefix of the path of the request",
		func(c *core.Context, old, new string) {
			setpath(c.UpstreamRequest, fix(new)+strings.TrimPrefix(c.UpstreamRequest.URL.Path, fix(old)))
		},
	)
}

func init() {
	DefaultRegistry.Register("rewrite", []string{
		"string: the regular expression, based on go stdlib regexp, of the path of the request",
		"string: the replacement of the path of the request",
	}, func(directive string, args ...string) (Processor, error) {
		if err := checkTwoArgs(directive, args); err != nil {
			return nil, err
		}

		re, err := regexp.Compile(args[0])
		if err != nil {
			return nil, err
		}

		replacement := args[1]
		return ProcessorFunc(func(c *core.Context) {
			setpath(c.UpstreamRequest, re.ReplaceAllString(c.UpstreamRequest.URL.Path, replacement))
		}), nil

	})
}

func setpath(r *http.Request, s string) {
	r.URL.Path = s
	r.URL.RawPath = ""
}
