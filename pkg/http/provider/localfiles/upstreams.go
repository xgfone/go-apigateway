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

	"github.com/xgfone/go-apigateway/pkg/http/dynamicconfig"
	"github.com/xgfone/go-apigateway/pkg/http/provider"
)

// UpstreamProvider returns a upstream provider
// based on all the local files in a directory.
func UpstreamProvider(dir string) provider.UpstreamProvider {
	p := newDirProvider(dir)
	return provider.UpstreamProviderFunc(func(etag string) ([]dynamicconfig.Upstream, string, error) {
		datas, changed, err := p.Do()
		if err != nil || !changed {
			return nil, etag, err
		}

		ups := make([]dynamicconfig.Upstream, 0, max(len(datas), 64))
		for path, data := range datas {
			var _ups []dynamicconfig.Upstream
			if err = json.Unmarshal(data, &_ups); err != nil {
				err = fmt.Errorf("fail to json decode the upstream file '%s': %w", path, err)
				return nil, "", err
			}
			ups = append(ups, _ups...)
		}

		etag = p.Etag()
		return ups, etag, nil
	})
}
