package cli

import (
	"context"

	"github.com/jtorchia/lode/internal/cache"
	"github.com/jtorchia/lode/internal/checkout"
	"github.com/jtorchia/lode/internal/lock"
	"github.com/jtorchia/lode/internal/repo"
	"github.com/jtorchia/lode/internal/transfer"
	"github.com/spf13/cobra"
)

func newPushCmd() *cobra.Command {
	var remoteName string
	cmd := &cobra.Command{
		Use:   "push [target]...",
		Short: "Sube objetos al remote",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPush(cmd.Context(), args, remoteName)
		},
	}
	cmd.Flags().StringVarP(&remoteName, "remote", "r", "", "remote a usar")
	return cmd
}

func runPush(ctx context.Context, targets []string, remoteName string) error {
	r, err := findRepo()
	if err != nil {
		return err
	}
	gl, err := lock.Acquire(r.LockPath())
	if err != nil {
		return err
	}
	defer gl.Release()

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
	infof("%d archivos subidos, %d ya presentes, %d fallidos", res.Transferred, res.Skipped, res.Failed)
	return nil
}

func newFetchCmd() *cobra.Command {
	var remoteName string
	cmd := &cobra.Command{
		Use:   "fetch [target]...",
		Short: "Descarga objetos del remote al cache (sin tocar el workspace)",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := runFetch(cmd.Context(), args, remoteName)
			return err
		},
	}
	cmd.Flags().StringVarP(&remoteName, "remote", "r", "", "remote a usar")
	return cmd
}

func runFetch(ctx context.Context, targets []string, remoteName string) (*cache.Cache, error) {
	r, err := findRepo()
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
	infof("%d objetos descargados, %d ya en cache", res.Transferred, res.Skipped)
	return c, nil
}

func newPullCmd() *cobra.Command {
	var remoteName string
	cmd := &cobra.Command{
		Use:   "pull [target]...",
		Short: "Descarga del remote y materializa el workspace (fetch + checkout)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPull(cmd.Context(), args, remoteName)
		},
	}
	cmd.Flags().StringVarP(&remoteName, "remote", "r", "", "remote a usar")
	return cmd
}

func runPull(ctx context.Context, targets []string, remoteName string) error {
	r, err := findRepo()
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
	infof("workspace actualizado (%d salidas)", len(outs))
	return nil
}
