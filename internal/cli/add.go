package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/getlode/lode/internal/cache"
	"github.com/getlode/lode/internal/checkout"
	"github.com/getlode/lode/internal/dvcfile"
	"github.com/getlode/lode/internal/hashfile"
	"github.com/getlode/lode/internal/lock"
	"github.com/getlode/lode/internal/repo"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

func newAddCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "add <target>...",
		Short: "Track files or directories (drop-in compatible with DVC)",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAdd(args)
		},
	}
}

func runAdd(targets []string) error {
	r, err := requireRepo()
	if err != nil {
		return err
	}
	gl, err := lock.Acquire(r.LockPath())
	if err != nil {
		return err
	}
	defer func() { _ = gl.Release() }()

	st, err := hashfile.OpenState(r.StatePath())
	if err != nil {
		return err
	}
	defer st.Close()

	c := cache.New(r.CacheDir())
	for _, t := range targets {
		if err := addTarget(c, st, t); err != nil {
			return fmt.Errorf("add %s: %w", t, err)
		}
	}
	return nil
}

func addTarget(c *cache.Cache, st *hashfile.State, target string) error {
	target = filepath.Clean(target)
	info, err := os.Stat(target)
	if err != nil {
		return err
	}

	out := dvcfile.Out{Hash: "md5", Path: filepath.Base(target)}

	if info.IsDir() {
		tree, err := hashfile.HashTree(target, st)
		if err != nil {
			return err
		}
		g := new(errgroup.Group)
		g.SetLimit(runtime.NumCPU())
		for i := range tree.Entries {
			i := i
			g.Go(func() error { return c.AddFile(tree.AbsPaths[i], tree.Entries[i].MD5) })
		}
		if err := g.Wait(); err != nil {
			return err
		}
		oid := hashfile.DirOID(tree.Entries)
		if err := c.AddBytes(hashfile.SerializeDir(tree.Entries), oid); err != nil {
			return err
		}
		nfiles := int64(tree.NFiles)
		out.MD5, out.Size, out.NFiles = oid, tree.TotalSize, &nfiles
	} else {
		sum, size, err := hashfile.HashFileCached(target, st)
		if err != nil {
			return err
		}
		// Safe-failure: reject if the file changed while we were hashing it.
		if fi, err := os.Stat(target); err != nil || fi.Size() != size {
			return fmt.Errorf("the file changed during add; aborted")
		}
		if err := c.AddFile(target, sum); err != nil {
			return err
		}
		out.MD5, out.Size = sum, size
	}

	dvcPath := target + ".dvc"
	if err := dvcfile.Save(dvcPath, &dvcfile.File{Outs: []dvcfile.Out{out}}); err != nil {
		return err
	}
	if err := checkout.AddToGitignore(target); err != nil {
		return err
	}
	infof("%-20s tracked -> %s", target, filepath.Base(dvcPath))
	return nil
}

// dvcFilesFor resolves the .dvc files to operate on: explicit targets map to
// <target>.dvc; with none, every *.dvc under the repo (excluding .dvc/).
func dvcFilesFor(r *repo.Repo, targets []string) ([]string, error) {
	if len(targets) > 0 {
		out := make([]string, 0, len(targets))
		for _, t := range targets {
			t = filepath.Clean(t)
			if filepath.Ext(t) == ".dvc" {
				out = append(out, t)
			} else {
				out = append(out, t+".dvc")
			}
		}
		return out, nil
	}
	var found []string
	err := filepath.WalkDir(r.Root, func(p string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && d.Name() == ".dvc" {
			return filepath.SkipDir
		}
		if !d.IsDir() && filepath.Ext(p) == ".dvc" {
			found = append(found, p)
		}
		return nil
	})
	return found, err
}
