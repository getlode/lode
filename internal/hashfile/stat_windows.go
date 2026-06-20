//go:build windows

package hashfile

import "os"

// Windows has no stable inode via os.FileInfo; fall back to 0 so detection
// relies on (mtime, size). This is sufficient for change detection in practice.
func inodeOf(fi os.FileInfo) uint64 { return 0 }
