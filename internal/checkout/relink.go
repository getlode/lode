// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package checkout

import (
	"os"

	"github.com/getlode/lode/internal/hashfile"
)

// relinkUpToDate reports whether dst already holds the object oid, avoiding a
// relink. It uses cheap checks first (symlink target, shared inode for
// hardlinks) and falls back to a content hash for copies/reflinks.
func relinkUpToDate(dst, src, oid string) bool {
	fi, err := os.Lstat(dst)
	if err != nil {
		return false
	}

	// Symlink: up to date iff it points at the cache object.
	if fi.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(dst)
		return err == nil && target == src
	}

	// Hardlink: up to date iff dst and the cache object share an inode.
	if si, err := os.Stat(src); err == nil {
		if di, err := os.Stat(dst); err == nil && os.SameFile(si, di) {
			return true
		}
	}

	// Copy/reflink: compare content hash. The oid may carry a ".dir" suffix for
	// directory manifests, but only regular files reach materializeOne.
	got, _, err := hashfile.HashFile(dst)
	return err == nil && got == oid
}
