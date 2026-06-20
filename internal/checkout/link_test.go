// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package checkout

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/getlode/lode/internal/cache"
	"github.com/getlode/lode/internal/dvcfile"
	"github.com/getlode/lode/internal/hashfile"
)

func TestParseCacheTypes(t *testing.T) {
	cases := map[string][]LinkType{
		"":                      {Reflink, Copy},
		"copy":                  {Copy},
		"reflink,hardlink,copy": {Reflink, Hardlink, Copy},
		"  symlink , copy ":     {Symlink, Copy},
	}
	for in, want := range cases {
		got := ParseCacheTypes(in)
		if len(got) != len(want) {
			t.Fatalf("%q: got %v want %v", in, got, want)
		}
		for i := range got {
			if got[i] != want[i] {
				t.Fatalf("%q: got %v want %v", in, got, want)
			}
		}
	}
}

// stageObject hashes content, stores it in the cache, and returns its oid.
func stageObject(t *testing.T, c *cache.Cache, content string) string {
	t.Helper()
	dir := t.TempDir()
	src := filepath.Join(dir, "src")
	if err := os.WriteFile(src, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	sum, _, err := hashfile.HashFile(src)
	if err != nil {
		t.Fatal(err)
	}
	if err := c.AddFile(src, sum); err != nil {
		t.Fatal(err)
	}
	return sum
}

func TestMaterialize_CopyFallback(t *testing.T) {
	root := t.TempDir()
	c := cache.New(filepath.Join(root, "cache"))
	oid := stageObject(t, c, "payload-data")

	dst := filepath.Join(root, "out.bin")
	out := dvcfile.Out{MD5: oid, Hash: "md5", Path: "out.bin"}

	// reflink may or may not be available depending on the FS; either way the
	// fallback guarantees a correct, readable result.
	if err := Materialize(c, out, dst, []LinkType{Reflink, Copy}); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(dst)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "payload-data" {
		t.Fatalf("unexpected content: %q", got)
	}

	// Idempotent: a second materialize over an up-to-date file is a no-op.
	if err := Materialize(c, out, dst, []LinkType{Copy}); err != nil {
		t.Fatal(err)
	}
}

func TestMaterialize_MissingObject(t *testing.T) {
	root := t.TempDir()
	c := cache.New(filepath.Join(root, "cache"))
	out := dvcfile.Out{MD5: "ffffffffffffffffffffffffffffffff", Hash: "md5", Path: "x"}
	err := Materialize(c, out, filepath.Join(root, "x"), []LinkType{Copy})
	if !os.IsNotExist(err) {
		t.Fatalf("expected ErrNotExist, got %v", err)
	}
}

func TestMaterialize_HardlinkSharesInode(t *testing.T) {
	root := t.TempDir()
	c := cache.New(filepath.Join(root, "cache"))
	oid := stageObject(t, c, "hl-content")
	dst := filepath.Join(root, "hl.bin")
	out := dvcfile.Out{MD5: oid, Hash: "md5", Path: "hl.bin"}

	if err := Materialize(c, out, dst, []LinkType{Hardlink}); err != nil {
		t.Fatal(err)
	}
	si, _ := os.Stat(c.ObjectPath(oid))
	di, _ := os.Stat(dst)
	if !os.SameFile(si, di) {
		t.Fatal("hardlink does not share an inode with the cache object")
	}
	// relink detection should skip on a second pass.
	if err := Materialize(c, out, dst, []LinkType{Hardlink}); err != nil {
		t.Fatal(err)
	}
}
