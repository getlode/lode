// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package hashfile

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	"golang.org/x/sync/errgroup"
)

// chunkSize matches DVC's read chunk (2**20). It does not affect the digest
// (MD5 is streaming) but keeps memory bounded for huge files.
const chunkSize = 1 << 20

var bufPool = sync.Pool{
	New: func() any { b := make([]byte, chunkSize); return &b },
}

// HashFile returns the lowercase hex MD5 of the file content and its size.
// Only the raw bytes are hashed (DVC 3.x: no CRLF normalization).
func HashFile(path string) (md5sum string, size int64, err error) {
	f, err := os.Open(path)
	if err != nil {
		return "", 0, err
	}
	defer func() { _ = f.Close() }()

	bp := bufPool.Get().(*[]byte)
	defer bufPool.Put(bp)

	h := md5.New()
	n, err := io.CopyBuffer(h, f, *bp)
	if err != nil {
		return "", 0, err
	}
	return hex.EncodeToString(h.Sum(nil)), n, nil
}

// Tree is the result of hashing a directory: the manifest entries plus the
// absolute path of each entry's content (absPaths[i] corresponds to
// entries[i]), and aggregate totals.
type Tree struct {
	Entries   []DirEntry
	AbsPaths  []string
	TotalSize int64
	NFiles    int
}

// HashTree walks root, hashes every file in parallel (bounded to NumCPU, since
// hashing is CPU-bound), and returns the directory manifest. When state is
// non-nil it is consulted to skip re-hashing unchanged files.
func HashTree(root string, state *State) (*Tree, error) {
	type file struct{ abs, rel string }
	var files []file
	walkErr := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		files = append(files, file{abs: p, rel: filepath.ToSlash(rel)})
		return nil
	})
	if walkErr != nil {
		return nil, walkErr
	}

	tree := &Tree{
		Entries:  make([]DirEntry, len(files)),
		AbsPaths: make([]string, len(files)),
		NFiles:   len(files),
	}
	sizes := make([]int64, len(files))

	var mu sync.Mutex
	var newEntries []Entry // freshly hashed files to persist to state in one tx

	g := new(errgroup.Group)
	g.SetLimit(runtime.NumCPU())
	for i := range files {
		i := i
		g.Go(func() error {
			if state != nil {
				if sum, size, ok := state.Get(files[i].abs); ok {
					tree.Entries[i] = DirEntry{MD5: sum, RelPath: files[i].rel}
					tree.AbsPaths[i] = files[i].abs
					sizes[i] = size
					return nil
				}
			}
			sum, size, err := HashFile(files[i].abs)
			if err != nil {
				return err
			}
			tree.Entries[i] = DirEntry{MD5: sum, RelPath: files[i].rel}
			tree.AbsPaths[i] = files[i].abs
			sizes[i] = size
			if state != nil {
				mu.Lock()
				newEntries = append(newEntries, Entry{Path: files[i].abs, MD5: sum, Size: size})
				mu.Unlock()
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	if state != nil {
		if err := state.PutMany(newEntries); err != nil {
			return nil, err
		}
	}

	for _, s := range sizes {
		tree.TotalSize += s
	}
	return tree, nil
}

// HashFileCached returns the file's MD5, using the state DB to avoid re-hashing
// files whose (inode, mtime, size) are unchanged.
func HashFileCached(path string, state *State) (md5sum string, size int64, err error) {
	if state != nil {
		if sum, sz, ok := state.Get(path); ok {
			return sum, sz, nil
		}
	}
	sum, sz, err := HashFile(path)
	if err != nil {
		return "", 0, err
	}
	if state != nil {
		_ = state.Put(path, sum, sz)
	}
	return sum, sz, nil
}
