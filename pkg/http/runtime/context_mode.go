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

package runtime

import (
	"fmt"
	"log/slog"
)

// Pre-define the running modes.
const (
	ModeNone Mode = iota
	ModeCall
	ModeForward
)

// Mode represents the running mode.
type Mode int8

// String returns the string description of the running mode.
func (m Mode) String() string {
	switch m {
	case ModeNone:
		return "ModeNone"
	case ModeCall:
		return "ModeCall"
	case ModeForward:
		return "ModeForward"
	default:
		return fmt.Sprintf("Mode(%d)", m)
	}
}

// IsValid reports whether the mode is valid.
func (m Mode) IsValid() bool { return m != ModeNone }

// IsCall reports whether the mode is call.
func (m Mode) IsCall() bool { return m == ModeCall }

// IsForward reports whether the mode is forward.
func (m Mode) IsForward() bool { return m == ModeForward }

// SetModeCall sets the context running mode to ModeCall.
func (c *Context) SetModeCall() { c.Mode = ModeCall }

// SetModeForward sets the context running mode to ModeForward.
func (c *Context) SetModeForward() { c.Mode = ModeForward }

// NeedModeCall checks the running mode is on ModeCall.
// If not, call the not handler if not nil, and return false.
// If yes, do nothing and return true.
func (c *Context) NeedModeCall(source string, not Handler) (ok bool) {
	if ok = c.IsCall(); !ok {
		slog.Warn("the running mode is invalid",
			slog.String("reqid", c.RequestID()),
			slog.String("requester", source),
			slog.String("need", ModeCall.String()),
			slog.String("got", c.Mode.String()),
		)

		if not != nil {
			not(c)
		}
	}
	return
}

// NeedModeForward checks the running mode is on ModeForward.
// If not, call the not handler if not nil, and return false.
// If yes, do nothing and return true.
func (c *Context) NeedModeForward(source string, not Handler) (ok bool) {
	if ok = c.IsForward(); !ok {
		slog.Warn("the running mode is invalid",
			slog.String("reqid", c.RequestID()),
			slog.String("requester", source),
			slog.String("need", ModeForward.String()),
			slog.String("got", c.Mode.String()),
		)

		if not != nil {
			not(c)
		}
	}
	return
}

// MustCall checks whether the running mode is at ModeCall.
// If not, panic with the source.
func (c *Context) MustCall(source string) {
	if !c.IsCall() {
		panic(fmt.Sprintf("%s: the running mode is not at ModeCall", source))
	}
}

// MustForward checks whether the running mode is at ModeForward.
// If not, panic with the source.
func (c *Context) MustForward(source string) {
	if !c.IsForward() {
		panic(fmt.Sprintf("%s: the running mode is not at ModeForward", source))
	}
}

func (c *Context) setmode(mode Mode) {
	switch c.Mode {
	case mode:
	case ModeNone:
		c.Mode = mode
	default:
		const f = "runtime.setmode: the running mode has been set to %s, not %s"
		panic(fmt.Errorf(f, c.Mode, mode))
	}
}
