// Package oracle validates byte-for-byte compatibility with the reference DVC
// implementation. Each test generates artifacts with the real `dvc` binary and
// compares them against what lode's primitives produce.
//
// The tests are skipped when DVC is not available (set DVC_BIN to the dvc
// executable; PYTHONPATH may be required for a --target install).
package oracle

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/getlode/lode/internal/dvcfile"
	"github.com/getlode/lode/internal/hashfile"
)

func dvcBin(t *testing.T) string {
	t.Helper()
	if b := os.Getenv("DVC_BIN"); b != "" {
		return b
	}
	if b, err := exec.LookPath("dvc"); err == nil {
		return b
	}
	t.Skip("dvc no disponible; set DVC_BIN para correr el oráculo")
	return ""
}

// initDvcRepo creates a fresh DVC repo (no SCM) in a temp dir.
func initDvcRepo(t *testing.T, dvc string) string {
	t.Helper()
	dir := t.TempDir()
	run(t, dir, dvc, "init", "--no-scm", "-q")
	return dir
}

func run(t *testing.T, dir, name string, args ...string) {
	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "DVC_NO_ANALYTICS=1")
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("%s %s: %v\n%s", name, strings.Join(args, " "), err, stderr.String())
	}
}

func findDir(t *testing.T, cacheRoot string) []byte {
	t.Helper()
	var data []byte
	err := filepath.WalkDir(cacheRoot, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(p, ".dir") {
			data, err = os.ReadFile(p)
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if data == nil {
		t.Fatal("no se encontró objeto .dir en el cache de DVC")
	}
	return data
}

// TestOracle_SingleFile: lode's .dvc bytes match DVC's for a single file.
func TestOracle_SingleFile(t *testing.T) {
	dvc := dvcBin(t)
	dir := initDvcRepo(t, dvc)

	if err := os.WriteFile(filepath.Join(dir, "data.csv"), []byte("col\n1\n2\n3\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, dir, dvc, "add", "data.csv")

	want, err := os.ReadFile(filepath.Join(dir, "data.csv.dvc"))
	if err != nil {
		t.Fatal(err)
	}

	sum, size, err := hashfile.HashFile(filepath.Join(dir, "data.csv"))
	if err != nil {
		t.Fatal(err)
	}
	got := dvcfile.Marshal(&dvcfile.File{Outs: []dvcfile.Out{
		{MD5: sum, Size: size, Hash: "md5", Path: "data.csv"},
	}})

	if !bytes.Equal(got, want) {
		t.Fatalf("mismatch .dvc archivo:\n--- lode ---\n%s\n--- dvc ---\n%s", got, want)
	}
}

// TestOracle_Directory: lode's .dir object and .dvc bytes match DVC's for a
// directory (incl. a subdirectory).
func TestOracle_Directory(t *testing.T) {
	dvc := dvcBin(t)
	dir := initDvcRepo(t, dvc)

	data := filepath.Join(dir, "data")
	if err := os.MkdirAll(filepath.Join(data, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{
		"a.txt":        "alpha",
		"b.txt":        "beta",
		"sub/c.txt":    "gamma",
		"sub/zeta.bin": "\x00\x01\x02data",
	}
	for rel, content := range files {
		if err := os.WriteFile(filepath.Join(data, filepath.FromSlash(rel)), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	run(t, dir, dvc, "add", "data")

	wantDvc, err := os.ReadFile(filepath.Join(dir, "data.dvc"))
	if err != nil {
		t.Fatal(err)
	}
	wantDir := findDir(t, filepath.Join(dir, ".dvc", "cache"))

	tree, err := hashfile.HashTree(data, nil)
	if err != nil {
		t.Fatal(err)
	}
	gotDir := hashfile.SerializeDir(tree.Entries)
	if !bytes.Equal(gotDir, wantDir) {
		t.Fatalf("mismatch objeto .dir:\n--- lode ---\n%s\n--- dvc ---\n%s", gotDir, wantDir)
	}

	nfiles := int64(tree.NFiles)
	gotDvc := dvcfile.Marshal(&dvcfile.File{Outs: []dvcfile.Out{
		{MD5: hashfile.DirOID(tree.Entries), Size: tree.TotalSize, NFiles: &nfiles, Hash: "md5", Path: "data"},
	}})
	if !bytes.Equal(gotDvc, wantDvc) {
		t.Fatalf("mismatch .dvc dir:\n--- lode ---\n%s\n--- dvc ---\n%s", gotDvc, wantDvc)
	}
}

// TestOracle_CachePaths: lode lays out cache objects exactly where DVC does.
func TestOracle_CachePaths(t *testing.T) {
	dvc := dvcBin(t)
	dir := initDvcRepo(t, dvc)

	if err := os.WriteFile(filepath.Join(dir, "f.bin"), []byte("payload-123"), 0o644); err != nil {
		t.Fatal(err)
	}
	run(t, dir, dvc, "add", "f.bin")

	sum, _, err := hashfile.HashFile(filepath.Join(dir, "f.bin"))
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(dir, ".dvc", "cache", "files", "md5", sum[:2], sum[2:])
	if _, err := os.Stat(want); err != nil {
		t.Fatalf("DVC no dejó el objeto en %s (hash lode=%s): %v", want, sum, err)
	}
}
