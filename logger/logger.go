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

// Package logger provides some log assistants.
package logger

import (
	"fmt"
	"io"
	"log/slog"

	"github.com/xgfone/go-apigateway/osx"
	"github.com/xgfone/go-defaults"
)

var (
	// DebugLogLogger is a log.Logger to log the DEBUG message.
	DebugLogLogger = slog.NewLogLogger(slog.Default().Handler(), slog.LevelDebug)

	// ErrorLogLogger is a log.Logger to log the ERROR message.
	ErrorLogLogger = slog.NewLogLogger(slog.Default().Handler(), slog.LevelError)
)

// Level is the log level, which can be changed to adjust the level
// of the logger that uses it.
var Level = new(slog.LevelVar)

// NewJSONHandler returns a new log handler based on JSON,
// which will use Level as the handler level.
func NewJSONHandler(w io.Writer) slog.Handler {
	o := slog.HandlerOptions{ReplaceAttr: replace, AddSource: true, Level: Level}
	return slog.NewJSONHandler(w, &o)
}

func replace(groups []string, a slog.Attr) slog.Attr {
	switch {
	case a.Key == slog.SourceKey:
		if src, ok := a.Value.Any().(*slog.Source); ok {
			a.Value = slog.StringValue(fmt.Sprintf("%s:%d", defaults.TrimPkgFile(src.File), src.Line))
		}
	}
	return a
}

// Fatal emits the log message with the ERROR level, and call osx.Exit(1).
func Fatal(msg string, args ...any) { slog.Error(msg, args...); osx.Exit(1) }
