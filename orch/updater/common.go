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

// Package updater provides a common mechanism to updates the runtime
// from the synchronized configurations.
package updater

import (
	"context"

	"github.com/xgfone/go-defaults"
)

type callback[T any] func([]T)

func _sync[T any](ctx context.Context, ch <-chan []T, f callback[T]) {
	for {
		select {
		case <-ctx.Done():
			return

		case configs := <-ch:
			_run(configs, f)
		}
	}
}

func _run[T any](configs []T, f callback[T]) {
	defer _wrappanic()
	f(configs)
}

func _wrappanic() {
	if r := recover(); r != nil {
		defaults.HandlePanicContext(context.Background(), r)
	}
}
