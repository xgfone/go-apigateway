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

// Config used to return the configuration information of an object.
type Config interface {
	Config() any
}

// GetConfig calls the Config method if i has implemented the interface Config.
// Or, return nil instead.
func GetConfig(i any) any {
	if c, ok := i.(Config); ok {
		return c.Config()
	}
	return nil
}
