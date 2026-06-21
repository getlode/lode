// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package oracle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/getlode/lode/internal/hashfile"
)

// TestNoRehash proves the state DB short-circuits hashing for unchanged files,
// and re-hashes when content changes. We poison the cache with a sentinel hash:
// if HashFileCached returns the sentinel for an unchanged file, it provably did
// not read the file again (SC-005).
func TestNoRehash(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "f.bin")
	if err := os.WriteFile(path, []byte("original-content"), 0o644); err != nil {
		t.Fatal(err)
	}

	st, err := hashfile.OpenState(filepath.Join(dir, "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = st.Close() }()

	const sentinel = "deadbeefdeadbeefdeadbeefdeadbeef"
	if err := st.Put(path, sentinel, 16); err != nil {
		t.Fatal(err)
	}

	// Unchanged file -> cached (sentinel) value returned, NO rehash.
	got, _, err := hashfile.HashFileCached(path, st)
	if err != nil {
		t.Fatal(err)
	}
	if got != sentinel {
		t.Fatalf("re-hashed an unchanged file: got %s", got)
	}

	// Change content (and size) -> state invalidated -> real hash.
	if err := os.WriteFile(path, []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	got2, _, err := hashfile.HashFileCached(path, st)
	if err != nil {
		t.Fatal(err)
	}
	if got2 == sentinel {
		t.Fatal("state was not invalidated after the content changed")
	}
	want, _, _ := hashfile.HashFile(path)
	if got2 != want {
		t.Fatalf("hash recomputado incorrecto: got %s want %s", got2, want)
	}
}
