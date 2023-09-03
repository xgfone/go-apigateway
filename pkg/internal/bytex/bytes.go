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

// Package bytex provides some helpful bytes functions.
package bytex

import "bytes"

// RemoveComments removes the whole line comment that the first non-white character
// starts with "//", and the similar line tail comment.
func RemoveComments(data []byte) []byte {
	result := make([]byte, 0, len(data))
	for len(data) > 0 {
		var line []byte
		if index := bytes.IndexByte(data, '\n'); index == -1 {
			line = data
			data = nil
		} else {
			line = data[:index]
			data = data[index+1:]
		}

		orig := line
		line = bytes.TrimLeft(line, " \t")
		if _len := len(line); _len == 0 { // Empty Line
			continue
		} else if _len >= 2 && line[0] == '/' && line[1] == '/' { // Comment Line
			continue
		}

		// Line Suffix Comment
		if index := bytes.Index(orig, []byte{'/', '/'}); index == -1 {
			result = append(result, orig...)
		} else if bytes.IndexByte(orig[index:], '"') > -1 { // Don't support the case: ... // the comment containing "
			result = append(result, orig...)
		} else {
			result = append(result, bytes.TrimRight(orig[:index], " \t")...)
		}
		result = append(result, '\n')
	}
	return result
}
