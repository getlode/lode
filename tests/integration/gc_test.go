// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/getlode/lode/internal/cache"
	"github.com/getlode/lode/internal/repo"
)

// TestGC_SafetyAndRestorability runs the real lode binary: gc must remove only
// unreferenced objects and leave tracked data restorable (FR-019/FR-020). Set
// LODE_BIN to the built binary to run it.
func TestGC_SafetyAndRestorability(t *testing.T) {
	bin := os.Getenv("LODE_BIN")
	if bin == "" {
		t.Skip("LODE_BIN not set; skipping the gc binary test")
	}

	root := t.TempDir()
	if _, err := repo.Init(root); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "keep.bin"), []byte("keep-me"), 0o644); err != nil {
		t.Fatal(err)
	}
	runBin(t, root, bin, "add", "keep.bin")

	c := cache.New(filepath.Join(root, ".dvc", "cache"))

	// Inject an unreferenced object.
	junk := "deadbeefdeadbeefdeadbeefdeadbeef"
	if err := c.AddBytes([]byte("junk-data"), junk); err != nil {
		t.Fatal(err)
	}
	if !c.Has(junk) {
		t.Fatal("could not inject the junk object")
	}

	runBin(t, root, bin, "gc", "-f")

	if c.Has(junk) {
		t.Fatal("gc did not remove the unreferenced object")
	}

	// Tracked data must remain restorable: wipe workspace, checkout, verify.
	if err := os.Remove(filepath.Join(root, "keep.bin")); err != nil {
		t.Fatal(err)
	}
	runBin(t, root, bin, "checkout")
	assertContent(t, filepath.Join(root, "keep.bin"), "keep-me")
}

func runBin(t *testing.T, dir, bin string, args ...string) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("lode %v: %v\n%s", args, err, stderr.String())
	}
}
