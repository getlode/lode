package integration

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/jtorchia/dvcgo/internal/cache"
	"github.com/jtorchia/dvcgo/internal/repo"
)

// TestGC_SafetyAndRestorability runs the real dvcgo binary: gc must remove only
// unreferenced objects and leave tracked data restorable (FR-019/FR-020). Set
// DVCGO_BIN to the built binary to run it.
func TestGC_SafetyAndRestorability(t *testing.T) {
	bin := os.Getenv("DVCGO_BIN")
	if bin == "" {
		t.Skip("DVCGO_BIN no seteado; se omite el test de gc por binario")
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
		t.Fatal("no se pudo inyectar el objeto junk")
	}

	runBin(t, root, bin, "gc", "-f")

	if c.Has(junk) {
		t.Fatal("gc no eliminó el objeto no referenciado")
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
		t.Fatalf("dvcgo %v: %v\n%s", args, err, stderr.String())
	}
}
