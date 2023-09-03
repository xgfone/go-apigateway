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
	"fmt"
	"net/http"

	"github.com/xgfone/go-apigateway/pkg/http/runtime"
	"github.com/xgfone/go-loadbalancer/http/processor"
)

var directives = make(map[string][]string, 16)

type Builder func(directive string, args ...string) (processor.Processor, error)

// GetAllRegisteredDirectives returns all the registered processor directive.
func GetAllRegisteredDirectives() map[string][]string { return directives }

// ClearAllRegisteredDirectives clears all the registered directives.
func ClearAllRegisteredDirectives() {
	processor.DefalutBuilderManager.Reset()
	clear(directives)
}

// Build builds a new processor by the directive and arguments.
func Build(directive string, args ...any) (processor.Processor, error) {
	return processor.DefalutBuilderManager.Build(directive, args...)
}

// RegisterDirective registers the processor builder based on the directive.
//
// If the directive has existed, override it.
func RegisterDirective(directive string, argsDesc []string, builder Builder) {
	build := func(directive string, args ...any) (processor.Processor, error) {
		ss := make([]string, len(args))
		for i, s := range args {
			ss[i] = s.(string)
		}
		return builder(directive, ss...)
	}

	processor.DefalutBuilderManager.Register(directive, build)
	directives[directive] = argsDesc
}

/// --------------------------------------------------------------------- ///

// RegisterRequestDirective is the same as RegisterDirective,
// but only handling the request.
func registerRequestDirective(directive string, argsDesc []string,
	checkArgs func(string, []string) error,
	handle func(*http.Request, []string) error,
) {
	RegisterDirective(directive, argsDesc, func(directive string, args ...string) (processor.Processor, error) {
		if err := checkArgs(directive, args); err != nil {
			return nil, err
		}

		return processor.ProcessorFunc(func(ctx context.Context, pc processor.Context) error {
			return handle(pc.DstReq, args)
		}), nil
	})
}

// RegisterRequestDirectiveOne is the same as RegisterRequestDirective,
// but only for one directive argument.
func registerRequestDirectiveOne(directive string, argDesc string, handle func(*http.Request, string)) {
	registerRequestDirective(directive, []string{argDesc}, checkOneArgs, func(r *http.Request, args []string) error {
		handle(r, args[0])
		return nil
	})
}

// RegisterRequestDirectiveTwo is the same as RegisterRequestDirective,
// but only for two directive arguments.
func registerRequestDirectiveTwo(directive string, arg1Desc, arg2Desc string, handle func(r *http.Request, arg1, arg2 string)) {
	registerRequestDirective(directive, []string{arg1Desc, arg2Desc}, checkTwoArgs, func(r *http.Request, args []string) error {
		handle(r, args[0], args[1])
		return nil
	})
}

/// --------------------------------------------------------------------- ///

// RegisterContextDirective is the same as RegisterDirective,
// but using the request context.
func registerContextDirective(directive string, argsDesc []string,
	checkArgs func(string, []string) error,
	handle func(c *runtime.Context, args []string) error,
) {
	RegisterDirective(directive, argsDesc, func(directive string, args ...string) (processor.Processor, error) {
		if err := checkArgs(directive, args); err != nil {
			return nil, err
		}

		return processor.ProcessorFunc(func(ctx context.Context, pc processor.Context) error {
			return handle(pc.CtxData.(*runtime.Context), args)
		}), nil
	})
}

// RegisterContextDirectiveOne is the same as RegisterContextDirective,
// but only for one directive argument.
func registerContextDirectiveOne(directive string, argDesc string, handle func(c *runtime.Context, arg string)) {
	registerContextDirective(directive, []string{argDesc}, checkOneArgs, func(c *runtime.Context, args []string) error {
		handle(c, args[0])
		return nil
	})
}

// RegisterContextDirectiveTwo is the same as RegisterContextDirective,
// but only for two directive arguments.
func registerContextDirectiveTwo(directive string, arg1Desc, arg2Desc string, handle func(c *runtime.Context, arg1, arg2 string)) {
	registerContextDirective(directive, []string{arg1Desc, arg2Desc}, checkTwoArgs, func(c *runtime.Context, args []string) error {
		handle(c, args[0], args[1])
		return nil
	})
}

/// --------------------------------------------------------------------- ///

func checkOneArgs(directive string, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("directive '%s' expect one argument, but got none", directive)
	} else if args[0] == "" {
		return fmt.Errorf("directive '%s' got an empty argument", directive)
	}
	return nil
}

func checkTwoArgs(directive string, args []string) error {
	switch len(args) {
	case 0:
		return fmt.Errorf("directive '%s' expect two arguments, but got none", directive)

	case 1:
		return fmt.Errorf("directive '%s' expect two arguments, but got only one", directive)

	default:
		if args[0] == "" {
			return fmt.Errorf("directive '%s' got an empty argument #1", directive)
		}
		if args[1] == "" {
			return fmt.Errorf("directive '%s' got an empty argument #2", directive)
		}
		return nil
	}
}
