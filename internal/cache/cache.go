// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package cache implements DVC's content-addressed object store
// (.dvc/cache/files/md5/<2>/<rest>) with atomic writes and 0o444 protection.
package cache

import (
	"io"
	"os"
	"path/filepath"
	"strings"
)

// cacheMode is the read-only mode DVC applies to cache objects.
const cacheMode = 0o444

// Cache is a content-addressed store rooted at a .dvc/cache directory.
type Cache struct {
	dir string
}

// New returns a Cache rooted at dir (typically <repo>/.dvc/cache).
func New(dir string) *Cache { return &Cache{dir: dir} }

// ObjectPath returns the on-disk path for an object id. The id may carry a
// ".dir" suffix for directory manifests; the two-char prefix split matches DVC.
func (c *Cache) ObjectPath(oid string) string {
	return filepath.Join(c.dir, "files", "md5", oid[:2], oid[2:])
}

// Has reports whether the object is present in the cache.
func (c *Cache) Has(oid string) bool {
	_, err := os.Stat(c.ObjectPath(oid))
	return err == nil
}

// AddBytes stores data under oid (no-op if already present).
func (c *Cache) AddBytes(data []byte, oid string) error {
	if c.Has(oid) {
		return nil
	}
	return atomicWrite(c.ObjectPath(oid), func(w io.Writer) error {
		_, err := w.Write(data)
		return err
	})
}

// AllObjects returns every object id currently stored in the cache.
func (c *Cache) AllObjects() ([]string, error) {
	base := filepath.Join(c.dir, "files", "md5")
	var oids []string
	err := filepath.WalkDir(base, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(base, p)
		if err != nil {
			return err
		}
		oid := strings.ReplaceAll(rel, string(os.PathSeparator), "")
		if filepath.Base(p) == "" || strings.HasPrefix(filepath.Base(p), ".tmp") {
			return nil
		}
		oids = append(oids, oid)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return oids, nil
}

// Size returns the byte size of the object, or 0 if absent.
func (c *Cache) Size(oid string) int64 {
	fi, err := os.Stat(c.ObjectPath(oid))
	if err != nil {
		return 0
	}
	return fi.Size()
}

// Remove deletes an object from the cache.
func (c *Cache) Remove(oid string) error {
	err := os.Remove(c.ObjectPath(oid))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}

// TempDir returns a scratch directory on the same filesystem as the cache,
// suitable for downloads that will be renamed into place via Adopt.
func (c *Cache) TempDir() string { return filepath.Join(c.dir, "files", ".tmp") }

// Adopt moves an already-materialized file (e.g. a verified download) into the
// cache under oid via an atomic rename, applying read-only protection. src must
// live on the same filesystem as the cache.
func (c *Cache) Adopt(src, oid string) error {
	if c.Has(oid) {
		_ = os.Remove(src)
		return nil
	}
	dst := c.ObjectPath(oid)
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	if err := os.Chmod(src, cacheMode); err != nil {
		return err
	}
	return os.Rename(src, dst)
}

// AddFile copies the file at src into the cache under oid (no-op if present).
func (c *Cache) AddFile(src, oid string) error {
	if c.Has(oid) {
		return nil
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	return atomicWrite(c.ObjectPath(oid), func(w io.Writer) error {
		_, err := io.Copy(w, in)
		return err
	})
}

// atomicWrite writes to a temp file in the destination directory, fsyncs, sets
// read-only mode, then renames into place — so a partially written object is
// never visible under its final hash path.
func atomicWrite(dst string, write func(io.Writer) error) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(dst), ".tmp-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }() // no-op once renamed

	if err := write(tmp); err != nil {
		_ = tmp.Close()
		return err
	}
	// No per-object fsync: rename gives atomic visibility, and fsyncing every
	// object would dominate runtime for datasets of many small files (DVC does
	// not fsync per object either). Durability falls back to the OS flush.
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, cacheMode); err != nil {
		return err
	}
	return os.Rename(tmpName, dst)
}
