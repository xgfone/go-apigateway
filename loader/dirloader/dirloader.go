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

// Package dirloader provides a loader based on the local directory
// to load some resources.
package dirloader

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/xgfone/go-apigateway/internal/jsonx"
	"github.com/xgfone/go-apigateway/internal/slogx"
	"github.com/xgfone/go-apigateway/loader"
)

type info struct {
	modtime time.Time
	size    int64
}

func (i info) Equal(other info) bool {
	return i.size == other.size && i.modtime.Equal(other.modtime)
}

type file struct {
	buf  *bytes.Buffer
	data []byte

	last info
	now  info
}

// DirLoader is used to load the resources from the files in a directory.
type DirLoader[T any] struct {
	*loader.ResourceManager[[]T]

	dir   string
	lock  sync.Mutex
	files map[string]*file
	epoch uint64
}

// New returns a new DirLoader with the directory.
func New[T any](dir string) *DirLoader[T] {
	return &DirLoader[T]{
		dir:   dir,
		files: make(map[string]*file, 8),

		ResourceManager: loader.NewResourceManager[[]T](),
	}
}

// Sync is used to synchronize the resources to the chan ch periodically.
func (l *DirLoader[T]) Sync(ctx context.Context, rsctype string, interval time.Duration, reload <-chan struct{}, cb func([]T) (changed bool)) {
	if interval <= 0 {
		interval = time.Minute
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	var lastEtag string
	load := func() {
		defer slogx.WrapPanic(ctx)
		slog.LogAttrs(ctx, slog.LevelDebug-4, "start to load the resource", slog.String("type", rsctype))

		resources, etag, err := l.Load()
		if err != nil {
			slog.Error("fail to load the resources from the local files", "err", err)
			return
		}

		if etag == lastEtag {
			return
		}

		if cb(resources) {
			l.SetResource(resources)
		}

		lastEtag = etag
	}

	// first laod
	load()
	for {
		select {
		case <-ctx.Done():
			select {
			case <-ticker.C:
			default:
			}
			return

		case <-reload:
			load()

		case <-ticker.C:
			load()
		}
	}
}

func (l *DirLoader[T]) updateEpoch() {
	epoch := atomic.AddUint64(&l.epoch, 1)
	l.SetEtag(strconv.FormatUint(epoch, 10))
}

// Load scans the files in the directory, loads and returns them if changed.
func (l *DirLoader[T]) Load() (resources []T, etag string, err error) {
	l.lock.Lock()
	defer l.lock.Unlock()

	if err = l.scanfiles(); err != nil {
		return
	}

	changed, err := l.checkfiles()
	if err != nil {
		return
	} else if !changed {
		return l.Resource(), l.Etag(), nil
	}

	resources = make([]T, 0, len(l.files))
	for path, file := range l.files {
		var resource []T
		if err = l.decode(&resource, file.data); err != nil {
			err = fmt.Errorf("fail to decode resource file '%s': %w", path, err)
			return
		}
		resources = append(resources, resource...)
	}

	l.SetResource(resources)
	etag = l.Etag()
	return
}

func (l *DirLoader[T]) decode(dst *[]T, data []byte) error {
	return json.Unmarshal(data, dst)
}

func (l *DirLoader[T]) checkfiles() (changed bool, err error) {
	for path, file := range l.files {
		if file.last.Equal(file.now) {
			continue
		}

		changed = true
		if err = l._readfile(file.buf, path); err != nil {
			err = fmt.Errorf("fail to read the file '%s': %w", path, err)
			return
		}

		file.data = jsonx.RemoveComments(file.buf.Bytes())
		file.last = file.now
	}

	if changed {
		l.updateEpoch()
	}

	return
}

func (l *DirLoader[T]) _readfile(buf *bytes.Buffer, path string) (err error) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	buf.Reset()
	_, err = io.CopyBuffer(buf, file, make([]byte, 1024))
	return
}

func (l *DirLoader[T]) scanfiles() (err error) {
	files := make(map[string]struct{}, max(8, len(l.files)))
	err = filepath.WalkDir(l.dir, func(path string, d fs.DirEntry, err error) error {
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

		f, ok := l.files[path]
		if !ok {
			f = &file{buf: bytes.NewBuffer(make([]byte, 0, fi.Size()))}
			l.files[path] = f
		}

		f.now = info{modtime: fi.ModTime(), size: fi.Size()}
		files[path] = struct{}{}
		return nil
	})

	if err != nil {
		return
	}

	// Clean the non-exist files.
	for path := range l.files {
		if _, ok := files[path]; !ok {
			delete(l.files, path)
		}
	}

	return
}
