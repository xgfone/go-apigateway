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

// Processor is used to execute the directive and apply it to the request context.
type Processor interface {
	Process(*core.Context)
}

// ProcessorFunc is a processor function.
type ProcessorFunc func(c *core.Context)

// Process implements the interface Processor.
func (f ProcessorFunc) Process(c *core.Context) { f(c) }

// Processors is a set of processors.
type Processors []Processor

// Process implements the interface Processor.
func (ps Processors) Process(c *core.Context) {
	for _, p := range ps {
		p.Process(c)
	}
}

// Builder is used to build a directive with the arguments to a processor.
type Builder func(directive string, args ...string) (Processor, error)

// NoDirectiveError represents that a directive does not exist.
type NoDirectiveError struct {
	Directive string
}

// Error implements the interface error.
func (e NoDirectiveError) Error() string {
	return fmt.Sprintf("not found the processor directive '%s'", e.Directive)
}
