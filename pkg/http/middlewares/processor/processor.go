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
	"errors"

	"github.com/xgfone/go-apigateway/pkg/http/processors"
	"github.com/xgfone/go-apigateway/pkg/http/runtime"
	"github.com/xgfone/go-binder"
	"github.com/xgfone/go-loadbalancer/http/processor"
)

// Priority is the default priority of the middleware.
const Priority = 900

func init() {
	runtime.RegisterMiddlewareBuilder("processor", func(name string, conf map[string]any) (runtime.Middleware, error) {
		var arg struct {
			Directives [][]string `json:"directives"`
		}
		err := binder.BindStructToMap(&arg, "json", conf)
		if err != nil {
			return nil, err
		}

		_processors := make(processor.Processors, 0, len(arg.Directives))
		for _, ds := range arg.Directives {
			if len(ds) == 0 {
				continue
			}

			args := make([]any, len(ds)-1)
			for i, arg := range ds[1:] {
				args[i] = arg
			}

			p, err := processors.Build(ds[0], args...)
			if err != nil {
				return nil, err
			}

			_processors = append(_processors, p)
		}

		return Processor(name, Priority, _processors)
	})
}

// Processor returns a new middleware that executes the processor on the request,
// in general, which is the request forwarded to the upstream server.
func Processor(name string, priority int, p processor.Processor) (runtime.Middleware, error) {
	if p == nil {
		return nil, errors.New("processor must not be nil")
	}

	return runtime.NewMiddleware(name, priority, nil, func(h runtime.Handler) runtime.Handler {
		return func(c *runtime.Context) {
			pc := processor.NewContext(c.ClientResponse, c.ClientRequest, c.UpstreamRequest()).WithContext(c)
			if err := p.Process(c.Context, pc); err != nil {
				c.Abort(runtime.ErrInternalServerError.WithError(err))
			} else {
				h(c)
			}
		}
	}), nil
}
