package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/jtorchia/lode/internal/cache"
	"github.com/jtorchia/lode/internal/checkout"
	"github.com/jtorchia/lode/internal/dvcfile"
	"github.com/jtorchia/lode/internal/lock"
	"github.com/jtorchia/lode/internal/repo"
	"github.com/spf13/cobra"
)

func newCheckoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "checkout [target]...",
		Short: "Materializa el workspace según los .dvc (desde el cache)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheckout(args)
		},
	}
}

func runCheckout(targets []string) error {
	r, err := findRepo()
	if err != nil {
		return err
	}
	gl, err := lock.Acquire(r.LockPath())
	if err != nil {
		return err
	}
	defer gl.Release()

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

	var missing []string
	n := 0
	for _, df := range dvcFiles {
		f, err := dvcfile.Load(df)
		if err != nil {
			return err
		}
		dir := filepath.Dir(df)
		for _, out := range f.Outs {
			wsPath := filepath.Join(dir, out.Path)
			err := checkout.Materialize(c, out, wsPath, types)
			if errors.Is(err, os.ErrNotExist) {
				missing = append(missing, out.MD5)
				continue
			}
			if err != nil {
				return fmt.Errorf("checkout %s: %w", wsPath, err)
			}
			n++
		}
	}

	if len(missing) > 0 {
		infof("%d objetos no están en cache; ejecutá `lode pull` para traerlos", len(missing))
	}
	infof("%d salidas materializadas", n)
	return nil
}
