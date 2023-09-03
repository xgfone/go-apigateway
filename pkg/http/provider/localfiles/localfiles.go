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

// Package localfiles provides a set of providers based on the local files.
package localfiles

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xgfone/go-apigateway/pkg/internal/bytex"
)

type fileinfo struct {
	modtime time.Time
	size    int64
}

func (fi fileinfo) Equal(other fileinfo) bool {
	return fi.size == other.size && fi.modtime.Equal(other.modtime)
}

type file struct {
	data *bytes.Buffer
	last fileinfo
	now  fileinfo
}

type dirProvider struct {
	directory string
	files     map[string]*file
	lock      sync.Mutex

	epoch uint64
}

func (p *dirProvider) updateEpoch() { atomic.AddUint64(&p.epoch, 1) }
func (p *dirProvider) Etag() string {
	return strconv.FormatUint(atomic.LoadUint64(&p.epoch), 10)
}

func newDirProvider(dir string) *dirProvider {
	return &dirProvider{directory: dir, files: make(map[string]*file, 8)}
}

func (p *dirProvider) Do() (datas map[string][]byte, changed bool, err error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	if err = p.scanfiles(); err != nil {
		return
	}

	if changed, err = p.checkfiles(); err != nil || !changed {
		return
	}

	datas = make(map[string][]byte, len(p.files))
	for path, file := range p.files {
		datas[path] = bytex.RemoveComments(file.data.Bytes())
	}
	return
}

func (p *dirProvider) checkfiles() (changed bool, err error) {
	for path, file := range p.files {
		if file.last.Equal(file.now) {
			continue
		}

		changed = true
		if err = p._readfile(file.data, path); err != nil {
			err = fmt.Errorf("fail to read the file '%s': %w", path, err)
			return
		}

		file.last = file.now
	}

	if changed {
		p.updateEpoch()
	}

	return
}

func (p *dirProvider) _readfile(buf *bytes.Buffer, path string) (err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	buf.Reset()
	_, err = io.CopyBuffer(buf, file, make([]byte, 1024))
	return
}

func (p *dirProvider) scanfiles() (err error) {
	files := make(map[string]struct{}, max(8, len(p.files)))
	err = filepath.WalkDir(p.directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}

		if name := d.Name(); name[0] == '_' || !strings.HasSuffix(name, ".json") {
			return nil
		}

		fi, err := d.Info()
		if err != nil {
			return err
		}

		f, ok := p.files[path]
		if !ok {
			f = &file{data: bytes.NewBuffer(make([]byte, 0, fi.Size()))}
			p.files[path] = f
		}

		f.now = fileinfo{modtime: fi.ModTime(), size: fi.Size()}
		files[path] = struct{}{}
		return nil
	})

	if err != nil {
		return
	}

	// Clean the non-exist files.
	for path := range p.files {
		if _, ok := files[path]; !ok {
			delete(p.files, path)
		}
	}

	return
}
