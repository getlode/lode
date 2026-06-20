package cli

import (
	"context"

	"github.com/getlode/lode/internal/cache"
	"github.com/getlode/lode/internal/checkout"
	"github.com/getlode/lode/internal/lock"
	"github.com/getlode/lode/internal/repo"
	"github.com/getlode/lode/internal/transfer"
	"github.com/spf13/cobra"
)

func newPushCmd() *cobra.Command {
	var remoteName string
	cmd := &cobra.Command{
		Use:   "push [target]...",
		Short: "Upload objects to the remote",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPush(cmd.Context(), args, remoteName)
		},
	}
	cmd.Flags().StringVarP(&remoteName, "remote", "r", "", "remote to use")
	return cmd
}

func runPush(ctx context.Context, targets []string, remoteName string) error {
	r, err := requireRepo()
	if err != nil {
		return err
	}
	gl, err := lock.Acquire(r.LockPath())
	if err != nil {
		return err
	}
	defer func() { _ = gl.Release() }()

	store, err := openStore(r, remoteName)
	if err != nil {
		return err
	}
	c := cache.New(r.CacheDir())
	dvcFiles, err := dvcFilesFor(r, targets)
	if err != nil {
		return err
	}
	items, err := pushItems(c, dvcFiles)
	if err != nil {
		return err
	}
	res, err := transfer.Push(ctx, store, c, items, flagJobs)
	if err != nil {
		return err
	}
	infof("uploaded %s, %d already present, %d failed", plural(res.Transferred, "object", "objects"), res.Skipped, res.Failed)
	if res.Failed > 0 {
		return errPartialTransfer(res.Failed)
	}
	return nil
}

func newFetchCmd() *cobra.Command {
	var remoteName string
	cmd := &cobra.Command{
		Use:   "fetch [target]...",
		Short: "Download objects from the remote into the cache (no workspace changes)",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := runFetch(cmd.Context(), args, remoteName)
			return err
		},
	}
	cmd.Flags().StringVarP(&remoteName, "remote", "r", "", "remote to use")
	return cmd
}

func runFetch(ctx context.Context, targets []string, remoteName string) (*cache.Cache, error) {
	r, err := requireRepo()
	if err != nil {
		return nil, err
	}
	store, err := openStore(r, remoteName)
	if err != nil {
		return nil, err
	}
	c := cache.New(r.CacheDir())
	dvcFiles, err := dvcFilesFor(r, targets)
	if err != nil {
		return nil, err
	}
	items, _, _, err := fetchItems(dvcFiles)
	if err != nil {
		return nil, err
	}
	res, err := transfer.Fetch(ctx, store, c, items, flagJobs)
	if err != nil {
		return nil, err
	}
	if res.Failed > 0 {
		infof("downloaded %s, %d already in cache, %d failed", plural(res.Transferred, "object", "objects"), res.Skipped, res.Failed)
		return c, errPartialTransfer(res.Failed)
	}
	infof("downloaded %s, %d already in cache", plural(res.Transferred, "object", "objects"), res.Skipped)
	return c, nil
}

func newPullCmd() *cobra.Command {
	var remoteName string
	cmd := &cobra.Command{
		Use:   "pull [target]...",
		Short: "Download from the remote and materialize the workspace (fetch + checkout)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPull(cmd.Context(), args, remoteName)
		},
	}
	cmd.Flags().StringVarP(&remoteName, "remote", "r", "", "remote to use")
	return cmd
}

func runPull(ctx context.Context, targets []string, remoteName string) error {
	r, err := requireRepo()
	if err != nil {
		return err
	}
	store, err := openStore(r, remoteName)
	if err != nil {
		return err
	}
	cfg, err := repo.LoadConfig(r.ConfigPath())
	if err != nil {
		return err
	}
	types := checkout.ParseCacheTypes(cfg.CacheType)

	c := cache.New(r.CacheDir())
	dvcFiles, err := dvcFilesFor(r, targets)
	if err != nil {
		return err
	}
	items, outs, wsPaths, err := fetchItems(dvcFiles)
	if err != nil {
		return err
	}
	if _, err := transfer.Fetch(ctx, store, c, items, flagJobs); err != nil {
		return err
	}
	for i, out := range outs {
		if err := checkout.Materialize(c, out, wsPaths[i], types); err != nil {
			return err
		}
	}
	infof("updated workspace (%s)", plural(len(outs), "output", "outputs"))
	return nil
}
