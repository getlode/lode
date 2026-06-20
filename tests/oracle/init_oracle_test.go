package oracle

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/getlode/lode/internal/repo"
)

// gitInit makes dir a git repo (needed for DVC's scm-mode init).
func gitInit(t *testing.T, dir string) {
	t.Helper()
	for _, args := range [][]string{
		{"init", "-q"},
		{"config", "user.email", "t@t.co"},
		{"config", "user.name", "t"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		if err := cmd.Run(); err != nil {
			t.Fatalf("git %v: %v", args, err)
		}
	}
}

func sameBytes(t *testing.T, a, b, label string) {
	t.Helper()
	ba, err := os.ReadFile(a)
	if err != nil {
		t.Fatalf("%s: read lode %s: %v", label, a, err)
	}
	bb, err := os.ReadFile(b)
	if err != nil {
		t.Fatalf("%s: read dvc %s: %v", label, b, err)
	}
	if !bytes.Equal(ba, bb) {
		t.Fatalf("%s mismatch:\n--- lode ---\n%q\n--- dvc ---\n%q", label, ba, bb)
	}
}

// TestOracle_InitNoSCM: lode's --no-scm init is byte-identical to dvc init --no-scm.
func TestOracle_InitNoSCM(t *testing.T) {
	dvc := dvcBin(t)

	lodeDir := t.TempDir()
	if _, _, err := repo.InitRepo(lodeDir, repo.ModeNoSCM); err != nil {
		t.Fatal(err)
	}
	ref := t.TempDir()
	run(t, ref, dvc, "init", "--no-scm", "-q")

	sameBytes(t, filepath.Join(lodeDir, ".dvc", "config"), filepath.Join(ref, ".dvc", "config"), "config")
	sameBytes(t, filepath.Join(lodeDir, ".dvcignore"), filepath.Join(ref, ".dvcignore"), ".dvcignore")

	// btime present and empty; no cache; no .dvc/.gitignore in no-scm.
	assertEmptyFile(t, filepath.Join(lodeDir, ".dvc", "tmp", "btime"))
	assertAbsent(t, filepath.Join(lodeDir, ".dvc", "cache"))
	assertAbsent(t, filepath.Join(lodeDir, ".dvc", ".gitignore"))
}

// TestOracle_InitSCM: lode's git-mode init is byte-identical to dvc init.
func TestOracle_InitSCM(t *testing.T) {
	dvc := dvcBin(t)

	lodeDir := t.TempDir()
	gitInit(t, lodeDir)
	if _, _, err := repo.InitRepo(lodeDir, repo.ModeSCM); err != nil {
		t.Fatal(err)
	}
	ref := t.TempDir()
	gitInit(t, ref)
	run(t, ref, dvc, "init", "-q")

	sameBytes(t, filepath.Join(lodeDir, ".dvc", "config"), filepath.Join(ref, ".dvc", "config"), "config")
	sameBytes(t, filepath.Join(lodeDir, ".dvc", ".gitignore"), filepath.Join(ref, ".dvc", ".gitignore"), ".dvc/.gitignore")
	sameBytes(t, filepath.Join(lodeDir, ".dvcignore"), filepath.Join(ref, ".dvcignore"), ".dvcignore")
	assertEmptyFile(t, filepath.Join(lodeDir, ".dvc", "tmp", "btime"))
	assertAbsent(t, filepath.Join(lodeDir, ".dvc", "cache"))
}

func assertEmptyFile(t *testing.T, path string) {
	t.Helper()
	fi, err := os.Stat(path)
	if err != nil {
		t.Fatalf("expected file %s: %v", path, err)
	}
	if fi.Size() != 0 {
		t.Fatalf("expected %s empty, got %d bytes", path, fi.Size())
	}
}

func assertAbsent(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected %s to be absent", path)
	}
}
