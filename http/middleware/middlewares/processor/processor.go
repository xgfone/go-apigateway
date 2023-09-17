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

// Package processor provides a processor middleware to handle the forwarding
// request and response.
package processor

import (
	"github.com/xgfone/go-apigateway/http/core"
	"github.com/xgfone/go-apigateway/http/directive"
	"github.com/xgfone/go-apigateway/http/middleware"
)

func init() {
	middleware.DefaultRegistry.Register("processor", func(name string, conf any) (middleware.Middleware, error) {
		var arg struct {
			Directives [][]string `json:"directives"`
		}
		if err := middleware.BindConf(name, &arg, conf); err != nil {
			return nil, err
		}

		processors := make(directive.Processors, 0, len(arg.Directives))
		for _, ds := range arg.Directives {
			if len(ds) == 0 {
				continue
			}

			p, err := directive.DefaultRegistry.Build(ds[0], ds[1:]...)
			if err != nil {
				return nil, err
			}

			processors = append(processors, p)
		}

		if len(processors) == 0 {
			return Processor(nil, nil)
		}
		return Processor(processors, arg)
	})
}

// Processor returns a new middleware that executes the processor on the request,
// in general, which is the request forwarded to the upstream server.
func Processor(p directive.Processor, conf any) (middleware.Middleware, error) {
	return middleware.New("processor", conf, func(next core.Handler) core.Handler {
		if p == nil {
			return next
		}

		return func(c *core.Context) {
			c.OnForward(func() { p.Process(c) })
			next(c)
		}
	}), nil
}
