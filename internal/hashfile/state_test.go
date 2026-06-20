package hashfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestState_ForceRehashForcesMiss(t *testing.T) {
	dir := t.TempDir()
	st, err := OpenState(filepath.Join(dir, "tmp", "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	f := filepath.Join(dir, "a.bin")
	if err := os.WriteFile(f, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := st.Put(f, "deadbeef", 5); err != nil {
		t.Fatal(err)
	}

	// Normal mode: a matching (inode, mtime, size) is a cache hit.
	if _, _, ok := st.Get(f); !ok {
		t.Fatal("expected a cache hit for an unchanged file")
	}
	// Strict mode: must always miss so the file is re-hashed.
	st.ForceRehash = true
	if _, _, ok := st.Get(f); ok {
		t.Fatal("ForceRehash must force a cache miss (no false up-to-date)")
	}
}

func TestState_DetectsSizeAndContentChange(t *testing.T) {
	dir := t.TempDir()
	st, err := OpenState(filepath.Join(dir, "tmp", "state.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	f := filepath.Join(dir, "a.bin")
	if err := os.WriteFile(f, []byte("hello"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := st.Put(f, "deadbeef", 5); err != nil {
		t.Fatal(err)
	}
	// Append changes size (and mtime) → the heuristic must report a miss.
	if err := os.WriteFile(f, []byte("hello world"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, _, ok := st.Get(f); ok {
		t.Fatal("a changed file must not be reported as up to date")
	}
}
