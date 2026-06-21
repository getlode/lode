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

func lodeBin(t *testing.T) string {
	t.Helper()
	bin := os.Getenv("LODE_BIN")
	if bin == "" {
		t.Skip("LODE_BIN not set")
	}
	return bin
}

func initRepo(t *testing.T, bin string) string {
	t.Helper()
	dir := t.TempDir()
	runTool(t, dir, bin, "init", "--no-scm")
	return dir
}

func TestDoctor_NoRepo(t *testing.T) {
	bin := lodeBin(t)
	dir := t.TempDir()
	out, failed := runForOutput(t, dir, bin, "doctor")
	if !failed {
		t.Fatal("doctor should exit non-zero with no repo")
	}
	if !strings.Contains(out, "lode init") {
		t.Fatalf("should suggest init, got:\n%s", out)
	}
}

func TestDoctor_Healthy(t *testing.T) {
	bin := lodeBin(t)
	dir := initRepo(t, bin)
	out, failed := runForOutput(t, dir, bin, "doctor")
	if failed {
		t.Fatalf("healthy repo should exit zero, got:\n%s", out)
	}
	if !strings.Contains(out, "no remote configured") {
		t.Fatalf("should warn about missing remote, got:\n%s", out)
	}
}

func TestDoctor_UnreachableRemote(t *testing.T) {
	bin := lodeBin(t)
	dir := initRepo(t, bin)
	runTool(t, dir, bin, "remote", "add", "-d", "dead", "s3://bucket/store")
	// Port 1 is refused immediately.
	runTool(t, dir, bin, "remote", "modify", "dead", "endpointurl", "http://127.0.0.1:1")

	out, failed := runForOutput(t, dir, bin, "doctor")
	if !failed {
		t.Fatalf("unreachable remote should exit non-zero, got:\n%s", out)
	}
	if !strings.Contains(out, "unreachable") {
		t.Fatalf("should report unreachable remote, got:\n%s", out)
	}
}

func TestDoctor_LegacyFormat(t *testing.T) {
	bin := lodeBin(t)
	dir := initRepo(t, bin)
	// Seed a legacy 2.x cache object: .dvc/cache/ab/<rest> (no files/md5).
	legacy := filepath.Join(dir, ".dvc", "cache", "ab")
	if err := os.MkdirAll(legacy, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(legacy, "cdef0123456789"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, _ := runForOutput(t, dir, bin, "doctor")
	if !strings.Contains(out, "legacy") {
		t.Fatalf("should warn about legacy 2.x layout, got:\n%s", out)
	}
}

func TestDoctor_CacheUnwritable(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("running as root; chmod write-protection is ineffective")
	}
	bin := lodeBin(t)
	dir := initRepo(t, bin)
	dvcDir := filepath.Join(dir, ".dvc")
	if err := os.Chmod(dvcDir, 0o555); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chmod(dvcDir, 0o755) }() // restore so t.TempDir cleanup works

	out, failed := runForOutput(t, dir, bin, "doctor")
	if !failed {
		t.Fatalf("unwritable cache should exit non-zero, got:\n%s", out)
	}
	if !strings.Contains(out, "cannot write") {
		t.Fatalf("should report cache not writable, got:\n%s", out)
	}
}
