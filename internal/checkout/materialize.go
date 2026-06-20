package checkout

import (
	"io"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/jtorchia/dvcgo/internal/cache"
	"github.com/jtorchia/dvcgo/internal/dvcfile"
	"github.com/jtorchia/dvcgo/internal/hashfile"
)

// fileMode is the writable mode for materialized workspace files (cache objects
// are read-only 0o444).
const fileMode = 0o644

// Materialize reconstructs the workspace path for out from cache objects using
// the given link strategies (reflink/hardlink/symlink/copy). Files already
// matching are left untouched. For directories, workspace files absent from the
// manifest are pruned, matching `dvc checkout`.
func Materialize(c *cache.Cache, out dvcfile.Out, wsPath string, types []LinkType) error {
	if !out.IsDir() {
		return materializeOne(c, out.MD5, wsPath, types)
	}

	entries, err := readManifest(c, out.MD5)
	if err != nil {
		return err
	}
	keep := make(map[string]struct{}, len(entries))
	for _, e := range entries {
		rel := filepath.FromSlash(e.RelPath)
		keep[rel] = struct{}{}
		if err := materializeOne(c, e.MD5, filepath.Join(wsPath, rel), types); err != nil {
			return err
		}
	}
	return prune(wsPath, keep)
}

func materializeOne(c *cache.Cache, oid, dst string, types []LinkType) error {
	src, ok := c.ResolveRead(oid)
	if !ok {
		return os.ErrNotExist
	}
	if relinkUpToDate(dst, src, oid) {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	return link(src, dst, types)
}

// prune removes workspace files under root whose relative path is not in keep.
func prune(root string, keep map[string]struct{}) error {
	return filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		if _, ok := keep[rel]; !ok {
			return os.Remove(p)
		}
		return nil
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, fileMode)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func readManifest(c *cache.Cache, dirOID string) ([]hashfile.DirEntry, error) {
	p, ok := c.ResolveRead(dirOID)
	if !ok {
		return nil, os.ErrNotExist
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	return hashfile.ParseDir(data)
}
