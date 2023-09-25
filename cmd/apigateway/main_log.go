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

package main

import (
	"flag"
	"fmt"
	"io"
	"log/slog"
	"os"

	"github.com/xgfone/go-defaults"
)

var loglevel = flag.String("log.level", "info", "The log level, such as debug, info, warn, error.")

var level = new(slog.LevelVar)

func initlogging() {
	if err := level.UnmarshalText([]byte(*loglevel)); err != nil {
		slog.Error("fail to parse the log level", "level", *loglevel, "err", err)
		defaults.Exit(1)
	}
	slog.SetDefault(slog.New(newJSONHandler(os.Stderr)))
}

func newJSONHandler(w io.Writer) slog.Handler {
	o := slog.HandlerOptions{ReplaceAttr: replace, AddSource: true, Level: level}
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
