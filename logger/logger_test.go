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

package logger

import (
	"bytes"
	"context"
	"log/slog"
	"runtime"
	"strings"
	"testing"
)

func TestNewJSONHandler(t *testing.T) {
	buf := bytes.NewBuffer(nil)
	handler := NewJSONHandler(buf)

	if handler.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("expect DEBUG level is not enabled, but got enabled")
	}

	if !handler.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("expect INFO level is enabled, but got not")
	}

	Level.Set(slog.LevelError)
	if handler.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("expect ERROR level is enabled, but got not")
	}

	pc, _, _, _ := runtime.Caller(0)
	rcd := slog.Record{Message: "test", Level: slog.LevelError, PC: pc}
	if err := handler.Handle(context.Background(), rcd); err != nil {
		t.Error(err)
	}

	expect := `{"level":"ERROR","source":"github.com/xgfone/go-apigateway/logger/logger_test.go:43","msg":"test"}`
	if s := strings.TrimSpace(buf.String()); s != expect {
		t.Errorf("expect '%s', but got '%s'", expect, s)
	}
}
