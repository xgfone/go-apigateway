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

package localfiles

import (
	"encoding/json"
	"fmt"
	"maps"
	"sort"

	"github.com/xgfone/go-apigateway/pkg/http/dynamicconfig"
	"github.com/xgfone/go-apigateway/pkg/http/provider"
	gmaps "github.com/xgfone/go-apigateway/pkg/internal/maps"
)

// MiddlewareGroupProvider returns a middleware group provider
// based on all the local files in a directory.
func MiddlewareGroupProvider(dir string) provider.MiddlewareGroupProvider {
	p := newDirProvider(dir)
	return provider.MiddlewareGroupProviderFunc(func(etag string) (dynamicconfig.MiddlewareGroups, string, error) {
		datas, changed, err := p.Do()
		if err != nil || !changed {
			return nil, etag, err
		}

		groups := make(dynamicconfig.MiddlewareGroups, max(len(datas), 64))

		paths := gmaps.Keys(datas)
		sort.Strings(paths)
		for _, path := range paths {
			var _groups dynamicconfig.MiddlewareGroups
			if err = json.Unmarshal(datas[path], &_groups); err != nil {
				err = fmt.Errorf("fail to json decode the middleware group file '%s': %w", path, err)
				return nil, "", err
			}
			maps.Copy(groups, _groups)
		}

		etag = p.Etag()
		return groups, etag, nil
	})
}
