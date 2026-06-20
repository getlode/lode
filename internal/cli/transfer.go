package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/getlode/lode/internal/cache"
	"github.com/getlode/lode/internal/dvcfile"
	"github.com/getlode/lode/internal/hashfile"
	"github.com/getlode/lode/internal/remote"
	"github.com/getlode/lode/internal/repo"
	"github.com/getlode/lode/internal/transfer"
)

// openStore resolves the remote (explicit name or repo default) and builds its
// S3 store.
func openStore(r *repo.Repo, remoteName string) (transfer.Store, error) {
	cfg, err := repo.LoadConfig(r.ConfigPath())
	if err != nil {
		return nil, err
	}
	name := remoteName
	if name == "" {
		name = cfg.CoreRemote
	}
	if name == "" {
		return nil, errNoRemote
	}
	rm, ok := cfg.Remotes[name]
	if !ok {
		return nil, fmt.Errorf("remote %q is not configured", name)
	}
	return remote.NewS3(rm)
}

// pushItems builds transfer items from .dvc files, reading directory manifests
// from the cache to enumerate their contents.
func pushItems(c *cache.Cache, dvcFiles []string) ([]transfer.Item, error) {
	var items []transfer.Item
	for _, df := range dvcFiles {
		f, err := dvcfile.Load(df)
		if err != nil {
			return nil, err
		}
		for _, out := range f.Outs {
			if out.IsDir() {
				data, err := os.ReadFile(c.ObjectPath(out.MD5))
				if err != nil {
					return nil, fmt.Errorf("manifest %s is not in the cache (add the data first): %w", out.MD5, err)
				}
				entries, err := hashfile.ParseDir(data)
				if err != nil {
					return nil, err
				}
				contents := make([]string, len(entries))
				for i, e := range entries {
					contents[i] = e.MD5
				}
				items = append(items, transfer.Item{OID: out.MD5, Contents: contents})
			} else {
				items = append(items, transfer.Item{OID: out.MD5})
			}
		}
	}
	return items, nil
}

// fetchItems builds fetch items from .dvc files.
func fetchItems(dvcFiles []string) ([]transfer.FetchItem, []dvcfile.Out, []string, error) {
	var items []transfer.FetchItem
	var outs []dvcfile.Out
	var wsPaths []string
	for _, df := range dvcFiles {
		f, err := dvcfile.Load(df)
		if err != nil {
			return nil, nil, nil, err
		}
		dir := filepath.Dir(df)
		for _, out := range f.Outs {
			items = append(items, transfer.FetchItem{OID: out.MD5, IsDir: out.IsDir()})
			outs = append(outs, out)
			wsPaths = append(wsPaths, filepath.Join(dir, out.Path))
		}
	}
	return items, outs, wsPaths, nil
}
