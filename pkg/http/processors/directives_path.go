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

package processors

import (
	"context"
	"net/http"
	"regexp"
	"strings"

	"github.com/xgfone/go-loadbalancer/http/processor"
)

func init() {
	fix := func(s string) string { return strings.TrimSuffix(s, "/") }

	registerRequestDirectiveOne("setpath", "string: change the path of the request",
		func(r *http.Request, s string) { setpath(r, s) },
	)

	registerRequestDirectiveOne("addprefix", "string: add a prefix to the path of the request",
		func(r *http.Request, s string) { setpath(r, fix(s)+r.URL.Path) },
	)

	registerRequestDirectiveOne("addsuffix", "string: add a suffix to the path of the request",
		func(r *http.Request, s string) { setpath(r, r.URL.Path+fix(s)) },
	)

	registerRequestDirectiveOne("delprefix", "string: remove a prefix from the path of the request",
		func(r *http.Request, s string) { setpath(r, strings.TrimPrefix(r.URL.Path, fix(s))) },
	)

	registerRequestDirectiveOne("delsuffix", "string: remove a suffix from the path of the request",
		func(r *http.Request, s string) { setpath(r, strings.TrimSuffix(r.URL.Path, fix(s))) },
	)

	registerRequestDirectiveTwo("replaceprefix",
		"string: the replaced original prefix of the path of the request",
		"string: the new prefix of the path of the request",
		func(r *http.Request, old, new string) {
			setpath(r, fix(new)+strings.TrimPrefix(r.URL.Path, fix(old)))
		},
	)
}

func init() {
	RegisterDirective("rewrite", []string{
		"string: the regular expression, based on go stdlib regexp, of the path of the request",
		"string: the replacement of the path of the request",
	}, func(directive string, args ...string) (processor.Processor, error) {
		if err := checkTwoArgs(directive, args); err != nil {
			return nil, err
		}

		re, err := regexp.Compile(args[0])
		if err != nil {
			return nil, err
		}

		replacement := args[1]
		return processor.ProcessorFunc(func(ctx context.Context, pc processor.Context) error {
			setpath(pc.DstReq, re.ReplaceAllString(pc.DstReq.URL.Path, replacement))
			return nil
		}), nil
	})
}

func setpath(r *http.Request, s string) {
	r.URL.Path = s
	r.URL.RawPath = ""
}
