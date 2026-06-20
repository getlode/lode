//go:build linux

package checkout

import (
	"os"

	"golang.org/x/sys/unix"
)

// reflinkFile creates a copy-on-write clone of src at dst using FICLONE. Returns
// an error (so the caller falls back) on filesystems without reflink support.
func reflinkFile(src, dst string) error {
	s, err := os.Open(src)
	if err != nil {
		return err
	}
	defer s.Close()

	_ = os.Remove(dst)
	d, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fileMode)
	if err != nil {
		return err
	}
	if err := unix.IoctlFileClone(int(d.Fd()), int(s.Fd())); err != nil {
		d.Close()
		os.Remove(dst)
		return err
	}
	return d.Close()
}
