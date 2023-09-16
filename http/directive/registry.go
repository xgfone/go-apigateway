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
	"fmt"

	"github.com/xgfone/go-apigateway/http/core"
)

// DefaultRegistry is the default directive registry.
var DefaultRegistry = NewRegistry()

type builder struct {
	directive string
	arguments []string
	Builder
}

// Registry is used to collect the registered directives.
type Registry struct {
	builders map[string]builder
}

// NewRegistry returns a new directive registry.
func NewRegistry() *Registry {
	return &Registry{builders: make(map[string]builder, 16)}
}

// Directives returns all the registered directives.
func (r *Registry) Directives() map[string][]string {
	directives := make(map[string][]string, len(r.builders))
	for _, b := range r.builders {
		directives[b.directive] = b.arguments
	}
	return directives
}

// Get returns the builder by the directive.
func (r *Registry) Get(directive string) Builder {
	if b, ok := r.builders[directive]; ok {
		return b.Builder
	}
	return nil
}

// Clear clears all the registered directives.
func (r *Registry) Clear() { clear(r.builders) }

// Build builds a directive processor by the directive and arguments.
func (r *Registry) Build(directive string, args ...string) (Processor, error) {
	if b, ok := r.builders[directive]; ok {
		return b.Builder(directive, args...)
	}
	return nil, NoDirectiveError{Directive: directive}
}

// Register registers a directive with the argument description and builder.
//
// If exists, override it.
func (r *Registry) Register(directive string, argsDesc []string, build Builder) {
	if directive == "" {
		panic("Registry: directive must not be empty")
	}
	if build == nil {
		panic("Registry: directive builder must not be nil")
	}
	r.builders[directive] = builder{
		directive: directive,
		arguments: argsDesc,
		Builder:   build,
	}
}

// RegisterOneArg is the same as Register, but only supports the exact one argument.
func (r *Registry) RegisterOneArg(directive string, argDesc string, handle func(*core.Context, string)) {
	r.Register(directive, []string{argDesc}, func(directive string, args ...string) (Processor, error) {
		if err := checkOneArgs(directive, args); err != nil {
			return nil, err
		}

		arg := args[0]
		return ProcessorFunc(func(c *core.Context) {
			handle(c, arg)
		}), nil
	})
}

// RegisterTwoArgs is the same as Register, but only supports the exact two arguments.
func (r *Registry) RegisterTwoArgs(directive string, arg1Desc, arg2Desc string, handle func(c *core.Context, arg1, arg2 string)) {
	r.Register(directive, []string{arg1Desc, arg2Desc}, func(directive string, args ...string) (Processor, error) {
		if err := checkTwoArgs(directive, args); err != nil {
			return nil, err
		}

		arg1, arg2 := args[0], args[1]
		return ProcessorFunc(func(c *core.Context) {
			handle(c, arg1, arg2)
		}), nil
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
