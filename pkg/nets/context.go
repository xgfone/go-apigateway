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

package nets

import (
	"context"
	"net"
)

type contextKey struct{ name string }

var (
	// ConnContextKey is a context key, which is used to access the connection.
	ConnContextKey = &contextKey{"conn"}

	// ListenerContextKey is a context key, which is used to access the listener.
	ListenerContextKey = &contextKey{"listener"}
)

// GetConnFromContext extracts the net.Conn from the context
// with the key ConnContextKey.
//
// Return nil if the value does not exist or is not of type net.Conn.
func GetConnFromContext(ctx context.Context) net.Conn {
	conn, _ := ctx.Value(ConnContextKey).(net.Conn)
	return conn
}

// SetConnIntoContext returns a new context with the key ConnContextKey
// and value conn.
func SetConnIntoContext(parent context.Context, conn net.Conn) context.Context {
	if conn == nil {
		panic("SetConnIntoContext: net.Conn must not be nil")
	}
	return context.WithValue(parent, ConnContextKey, conn)
}

// GetListenerFromContext extracts the net.Listener from the context
// with the key ListenerContextKey.
//
// Return nil if the value does not exist or is not of type net.Listener.
func GetListenerFromContext(ctx context.Context) net.Listener {
	ln, _ := ctx.Value(ListenerContextKey).(net.Listener)
	return ln
}

// SetListenerIntoContext returns a new context with the key ListenerContextKey
// and value ln.
func SetListenerIntoContext(parent context.Context, ln net.Listener) context.Context {
	if ln == nil {
		panic("SetListenerIntoContext: net.Listener must not be nil")
	}
	return context.WithValue(parent, ListenerContextKey, ln)
}
