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

// Package slogx provides some slog functions for test.
package slogx

import (
	"context"
	"io"
	"log/slog"

	"github.com/xgfone/go-defaults"
)

// DisableSLog sets the log level to ERROR and output to io.Discard.
func DisableSLog() {
	level := new(slog.LevelVar)
	level.Set(slog.LevelError + 10)
	slog.SetDefault(slog.New(slog.NewJSONHandler(io.Discard, &slog.HandlerOptions{Level: level})))
}

// WrapPanic is used to wrap and log the panic.
func WrapPanic(ctx context.Context) {
	if r := recover(); r != nil {
		defaults.HandlePanicContext(ctx, r)
	}
}
