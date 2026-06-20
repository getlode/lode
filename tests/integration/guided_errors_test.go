package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// runForOutput runs the binary and returns combined output + whether it failed.
func runForOutput(t *testing.T, dir, bin string, args ...string) (string, bool) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	return buf.String(), err != nil
}

// TestGuided_NoRepo: a command outside a repo names `lode init` (SC-003).
func TestGuided_NoRepo(t *testing.T) {
	lode := os.Getenv("LODE_BIN")
	if lode == "" {
		t.Skip("LODE_BIN not set")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "f"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	out, failed := runForOutput(t, dir, lode, "add", "f")
	if !failed {
		t.Fatal("expected `add` to fail outside a repo")
	}
	if !strings.Contains(out, "lode init") {
		t.Fatalf("error should suggest `lode init`, got:\n%s", out)
	}
}

// TestGuided_NoRemote: push without a remote names `lode remote add` (SC-003).
func TestGuided_NoRemote(t *testing.T) {
	lode := os.Getenv("LODE_BIN")
	if lode == "" {
		t.Skip("LODE_BIN not set")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "f"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	runTool(t, dir, lode, "init", "--no-scm")
	runTool(t, dir, lode, "add", "f")
	out, failed := runForOutput(t, dir, lode, "push")
	if !failed {
		t.Fatal("expected `push` to fail without a remote")
	}
	if !strings.Contains(out, "lode remote add") {
		t.Fatalf("error should suggest `lode remote add`, got:\n%s", out)
	}
}
