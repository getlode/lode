// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/getlode/lode/internal/cache"
	"github.com/getlode/lode/internal/checkout"
	"github.com/getlode/lode/internal/dvcfile"
	"github.com/getlode/lode/internal/lock"
	"github.com/getlode/lode/internal/repo"
	"github.com/spf13/cobra"
)

func newCheckoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "checkout [target]...",
		Short: "Materialize the workspace from the cache",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runCheckout(args)
		},
	}
}

func runCheckout(targets []string) error {
	r, err := requireRepo()
	if err != nil {
		return err
	}
	gl, err := lock.Acquire(r.LockPath())
	if err != nil {
		return err
	}
	defer func() { _ = gl.Release() }()

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
		infof("%s", missingObjectsHint(len(missing)))
	}
	infof("materialized %s", plural(n, "output", "outputs"))
	return nil
}
