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

// Package auth provides some common auth functions.
package auth

import (
	"errors"
	"fmt"
	"maps"
	"net/http"
	"strings"
	"sync"

	"github.com/xgfone/go-apigateway/pkg/http/runtime"
)

// Priority is the default priority of the auth middleware.
const Priority = 700

// Authorization returns a common auth middleware that extracts
// the auth value of type "_type" from the request header "Authorization",
// then calls the handle function to update the upstream header.
func Authorization(name string, priority int, _type string, handle func(value string, header http.Header) error) (runtime.Middleware, error) {
	if _type == "" {
		return nil, fmt.Errorf("Authorization: the auth type must not be empty")
	} else if handle == nil {
		panic("Authorization: the handle function must not be empty")
	}

	config := map[string]string{"type": _type}
	return runtime.NewMiddleware(name, Priority, config, func(next runtime.Handler) runtime.Handler {
		return func(c *runtime.Context) {
			auth := strings.TrimSpace(c.ClientRequest.Header.Get("Authorization"))
			if auth == "" {
				c.Abort(runtime.ErrUnauthorized.WithError(errMissingAuthorization))
				return
			}

			index := strings.IndexByte(auth, ' ')
			if index < 0 || auth[:index] != _type {
				c.Abort(runtime.ErrUnauthorized.WithError(errInvalidAuthorization))
				return
			}

			header := getheader()
			defer putheader(header)
			if err := handle(strings.TrimSpace(auth[index+1:]), header); err != nil {
				c.Abort(runtime.ErrUnauthorized.WithError(err))
			} else {
				c.OnForward(func() { maps.Copy(c.UpstreamRequest().Header, header) })
				next(c)
			}
		}
	}), nil
}

var (
	errMissingAuthorization = errors.New("missing header 'Authorization'")
	errInvalidAuthorization = errors.New("invalid header 'Authorization")
)

var headerpool = &sync.Pool{New: func() any { return make(http.Header, 4) }}

func getheader() http.Header  { return headerpool.Get().(http.Header) }
func putheader(h http.Header) { clear(h); headerpool.Put(h) }
