package integration

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/getlode/lode/internal/cache"
	"github.com/getlode/lode/internal/hashfile"
)

// BenchmarkAddDir measures the hot path (hash a directory of many small files
// and stage it into the cache). This is the workload behind SC-001 — manual
// comparison against DVC-Python on 20k files showed lode ~13× faster
// (0.44s vs 5.79s).
func BenchmarkAddDir(b *testing.B) {
	root := b.TempDir()
	data := filepath.Join(root, "data")
	const nFiles = 5000
	if err := os.MkdirAll(data, 0o755); err != nil {
		b.Fatal(err)
	}
	for i := 0; i < nFiles; i++ {
		p := filepath.Join(data, fmt.Sprintf("f_%06d.bin", i))
		if err := os.WriteFile(p, []byte(fmt.Sprintf("content-%d", i)), 0o644); err != nil {
			b.Fatal(err)
		}
	}

	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		c := cache.New(filepath.Join(b.TempDir(), "cache"))
		tree, err := hashfile.HashTree(data, nil)
		if err != nil {
			b.Fatal(err)
		}
		for i, e := range tree.Entries {
			if err := c.AddFile(tree.AbsPaths[i], e.MD5); err != nil {
				b.Fatal(err)
			}
		}
		if err := c.AddBytes(hashfile.SerializeDir(tree.Entries), hashfile.DirOID(tree.Entries)); err != nil {
			b.Fatal(err)
		}
	}
}
