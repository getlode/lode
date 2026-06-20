// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cache

import (
	"os"
	"path/filepath"
)

// LegacyPath returns the DVC 2.x cache path for an object id
// (.dvc/cache/<2>/<rest>, the layout used before the files/md5 split).
func (c *Cache) LegacyPath(oid string) string {
	return filepath.Join(c.dir, oid[:2], oid[2:])
}

// ResolveRead returns the path of an existing object, preferring the modern
// 3.x layout and falling back to the legacy 2.x layout. This is read-only:
// lode always writes new objects in the modern layout.
func (c *Cache) ResolveRead(oid string) (string, bool) {
	if p := c.ObjectPath(oid); exists(p) {
		return p, true
	}
	if p := c.LegacyPath(oid); exists(p) {
		return p, true
	}
	return "", false
}

func exists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}
