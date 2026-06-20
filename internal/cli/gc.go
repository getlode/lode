package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/getlode/lode/internal/cache"
	"github.com/getlode/lode/internal/dvcfile"
	"github.com/getlode/lode/internal/hashfile"
	"github.com/getlode/lode/internal/lock"
	"github.com/getlode/lode/internal/repo"
	"github.com/spf13/cobra"
)

func newGCCmd() *cobra.Command {
	var (
		force      bool
		cloud      bool
		remoteName string
	)
	cmd := &cobra.Command{
		Use:   "gc",
		Short: "Remove unreferenced objects from the cache (and the remote with -c)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runGC(cmd.Context(), force, cloud, remoteName)
		},
	}
	cmd.Flags().BoolVarP(&force, "force", "f", false, "delete without asking for confirmation")
	cmd.Flags().BoolVarP(&cloud, "cloud", "c", false, "also clean the remote")
	cmd.Flags().StringVarP(&remoteName, "remote", "r", "", "remote to use with -c")
	cmd.Flags().BoolP("workspace", "w", true, "take references from the workspace")
	return cmd
}

func runGC(ctx context.Context, force, cloud bool, remoteName string) error {
	r, err := requireRepo()
	if err != nil {
		return err
	}
	gl, err := lock.Acquire(r.LockPath())
	if err != nil {
		return err
	}
	defer func() { _ = gl.Release() }()

	c := cache.New(r.CacheDir())
	reachable, err := reachableOIDs(r, c)
	if err != nil {
		return err
	}

	all, err := c.AllObjects()
	if err != nil {
		return err
	}
	var unreferenced []string
	var freed int64
	for _, oid := range all {
		if _, ok := reachable[oid]; !ok {
			unreferenced = append(unreferenced, oid)
			freed += c.Size(oid)
		}
	}

	if len(unreferenced) == 0 {
		infof("No unreferenced objects to remove.")
		return nil
	}

	infof("Will remove %s from the cache (%s).", plural(len(unreferenced), "object", "objects"), humanBytes(freed))
	if !force && !confirm() {
		infof("Cancelled.")
		return nil
	}

	for _, oid := range unreferenced {
		if err := c.Remove(oid); err != nil {
			return err
		}
	}
	infof("Freed %s from the local cache.", humanBytes(freed))

	if cloud {
		store, err := openStore(r, remoteName)
		if err != nil {
			return err
		}
		present, err := store.ListOIDs(ctx)
		if err != nil {
			return err
		}
		n := 0
		for oid := range present {
			if _, ok := reachable[oid]; !ok {
				if err := store.Delete(ctx, oid); err != nil {
					return err
				}
				n++
			}
		}
		infof("Removed %s from the remote.", plural(n, "unreferenced object", "unreferenced objects"))
	}
	return nil
}

// reachableOIDs collects every object id referenced by the workspace's .dvc
// files: each output's id plus, for directories, every file in its manifest.
func reachableOIDs(r *repo.Repo, c *cache.Cache) (map[string]struct{}, error) {
	reachable := make(map[string]struct{})
	dvcFiles, err := dvcFilesFor(r, nil)
	if err != nil {
		return nil, err
	}
	for _, df := range dvcFiles {
		f, err := dvcfile.Load(df)
		if err != nil {
			return nil, err
		}
		for _, out := range f.Outs {
			reachable[out.MD5] = struct{}{}
			if out.IsDir() {
				data, err := os.ReadFile(c.ObjectPath(out.MD5))
				if err != nil {
					continue // manifest not local; its contents stay unreachable from here
				}
				entries, err := hashfile.ParseDir(data)
				if err != nil {
					return nil, err
				}
				for _, e := range entries {
					reachable[e.MD5] = struct{}{}
				}
			}
		}
	}
	return reachable, nil
}

func confirm() bool {
	fmt.Fprint(os.Stderr, "Continue? (yes/no): ")
	sc := bufio.NewScanner(os.Stdin)
	if !sc.Scan() {
		return false
	}
	ans := strings.ToLower(strings.TrimSpace(sc.Text()))
	return ans == "yes" || ans == "y"
}

func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for x := n / unit; x >= unit; x /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}
