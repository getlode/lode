package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// TestInterop_InitThenDVC: a repo created by `lode init` is operated by the real
// dvc without errors (SC-002). Needs LODE_BIN and DVC_BIN.
func TestInterop_InitThenDVC(t *testing.T) {
	lode := os.Getenv("LODE_BIN")
	dvc := os.Getenv("DVC_BIN")
	if lode == "" || dvc == "" {
		t.Skip("LODE_BIN/DVC_BIN not set; skipping init interop")
	}

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "data.txt"), []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}

	// lode init (standalone) + lode add, then dvc must understand the repo.
	runTool(t, dir, lode, "init", "--no-scm")
	runTool(t, dir, lode, "add", "data.txt")
	runToolEnv(t, dir, dvc, []string{"DVC_NO_ANALYTICS=1"}, "status")
}

// TestInterop_DVCInitThenLode: a repo created by `dvc init` is operated by lode.
func TestInterop_DVCInitThenLode(t *testing.T) {
	lode := os.Getenv("LODE_BIN")
	dvc := os.Getenv("DVC_BIN")
	if lode == "" || dvc == "" {
		t.Skip("LODE_BIN/DVC_BIN not set; skipping init interop")
	}

	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "data.txt"), []byte("payload"), 0o644); err != nil {
		t.Fatal(err)
	}
	runToolEnv(t, dir, dvc, []string{"DVC_NO_ANALYTICS=1"}, "init", "--no-scm", "-q")
	runTool(t, dir, lode, "add", "data.txt")
	runTool(t, dir, lode, "status")
}

func runTool(t *testing.T, dir, bin string, args ...string) {
	t.Helper()
	runToolEnv(t, dir, bin, nil, args...)
}

func runToolEnv(t *testing.T, dir, bin string, extraEnv []string, args ...string) {
	t.Helper()
	cmd := exec.Command(bin, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), extraEnv...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s %v: %v\n%s", filepath.Base(bin), args, err, stderr.String())
	}
}
