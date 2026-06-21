// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

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
	defer func() { _ = s.Close() }()

	_ = os.Remove(dst)
	d, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fileMode)
	if err != nil {
		return err
	}
	if err := unix.IoctlFileClone(int(d.Fd()), int(s.Fd())); err != nil {
		_ = d.Close()
		_ = os.Remove(dst)
		return err
	}
	return d.Close()
}
