// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build windows

package hashfile

import "os"

// Windows has no stable inode via os.FileInfo; fall back to 0 so detection
// relies on (mtime, size). This is sufficient for change detection in practice.
func inodeOf(fi os.FileInfo) uint64 { return 0 }
