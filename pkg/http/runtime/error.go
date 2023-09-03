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

import "net/http"

// Pre-define some errors with the status code.
var (
	ErrBadRequest           = NewError(http.StatusBadRequest)          // 400
	ErrUnauthorized         = NewError(http.StatusUnauthorized)        // 401
	ErrForbidden            = NewError(http.StatusForbidden)           // 403
	ErrNotFound             = NewError(http.StatusNotFound)            // 404
	ErrTooManyRequests      = NewError(http.StatusTooManyRequests)     // 429
	ErrInternalServerError  = NewError(http.StatusInternalServerError) // 500
	ErrBadGateway           = NewError(http.StatusBadGateway)          // 502
	ErrServiceUnavailable   = NewError(http.StatusServiceUnavailable)  // 503
	ErrStatusGatewayTimeout = NewError(http.StatusGatewayTimeout)      // 504
)

var _ error = Error{}

// Error represents an error with the status code.
type Error struct {
	Code int
	Err  error
}

// NewError returns a new Error with the status code and without error.
func NewError(statusCode int) Error { return Error{Code: statusCode} }

// Error returns the error message.
func (e Error) Error() string {
	if e.Err == nil {
		return ""
	}
	return e.Err.Error()
}

// StatusCode returns the error status code.
func (e Error) StatusCode() int { return e.Code }

// WithError returns a new Error with the new error.
func (e Error) WithError(err error) Error {
	e.Err = err
	return e
}
