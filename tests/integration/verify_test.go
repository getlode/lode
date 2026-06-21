// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package integration

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// stageVerifyRepo creates a lode repo with a dir + a file tracked.
func stageVerifyRepo(t *testing.T, bin string) string {
	t.Helper()
	dir := t.TempDir()
	runTool(t, dir, bin, "init", "--no-scm")
	if err := os.MkdirAll(filepath.Join(dir, "data"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, f := range []struct{ p, c string }{
		{"data/a", "alpha"}, {"data/b", "beta"}, {"single.bin", "gamma"},
	} {
		if err := os.WriteFile(filepath.Join(dir, f.p), []byte(f.c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	runTool(t, dir, bin, "add", "data", "single.bin")
	return dir
}

func TestVerify_Healthy(t *testing.T) {
	bin := lodeBin(t)
	dir := stageVerifyRepo(t, bin)
	out, failed := runForOutput(t, dir, bin, "verify")
	if failed {
		t.Fatalf("healthy repo should verify clean, got:\n%s", out)
	}
	if !strings.Contains(out, "intact") {
		t.Fatalf("expected intact report, got:\n%s", out)
	}
}

func TestVerify_Corrupted(t *testing.T) {
	bin := lodeBin(t)
	dir := stageVerifyRepo(t, bin)
	// Corrupt one non-.dir cache object.
	var obj string
	if err := filepath.WalkDir(filepath.Join(dir, ".dvc", "cache", "files"), func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && !strings.HasSuffix(p, ".dir") && obj == "" {
			obj = p
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if obj == "" {
		t.Fatal("no cache object found")
	}
	if err := os.Chmod(obj, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(obj, []byte("CORRUPTED"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, failed := runForOutput(t, dir, bin, "verify")
	if !failed {
		t.Fatalf("corrupted object should fail verify, got:\n%s", out)
	}
	if !strings.Contains(out, "corrupted") {
		t.Fatalf("should report corruption, got:\n%s", out)
	}
}

func TestVerify_Missing(t *testing.T) {
	bin := lodeBin(t)
	dir := stageVerifyRepo(t, bin)
	var obj string
	if err := filepath.WalkDir(filepath.Join(dir, ".dvc", "cache", "files"), func(p string, d os.DirEntry, err error) error {
		if err == nil && !d.IsDir() && !strings.HasSuffix(p, ".dir") && obj == "" {
			obj = p
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(obj, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(obj); err != nil {
		t.Fatal(err)
	}
	out, failed := runForOutput(t, dir, bin, "verify")
	if !failed {
		t.Fatalf("missing object should fail verify, got:\n%s", out)
	}
	if !strings.Contains(out, "missing") {
		t.Fatalf("should report missing, got:\n%s", out)
	}
}

// TestVerify_DVCRepo: lode verify passes on a repo DVC created — proving lode
// computes the same hashes DVC recorded (compatibility proof).
func TestVerify_DVCRepo(t *testing.T) {
	bin := lodeBin(t)
	dvc := os.Getenv("DVC_BIN")
	if dvc == "" {
		t.Skip("DVC_BIN not set")
	}
	dir := t.TempDir()
	runToolEnv(t, dir, dvc, []string{"DVC_NO_ANALYTICS=1"}, "init", "--no-scm", "-q")
	if err := os.MkdirAll(filepath.Join(dir, "data"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "data", "a"), []byte("alpha"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "data", "b"), []byte("beta"), 0o644); err != nil {
		t.Fatal(err)
	}
	runToolEnv(t, dir, dvc, []string{"DVC_NO_ANALYTICS=1"}, "add", "data")

	out, failed := runForOutput(t, dir, bin, "verify")
	if failed {
		t.Fatalf("lode verify should pass on a DVC-created repo, got:\n%s", out)
	}
}
