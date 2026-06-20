package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/getlode/lode/internal/cache"
	"github.com/getlode/lode/internal/dvcfile"
	"github.com/getlode/lode/internal/hashfile"
	"github.com/getlode/lode/internal/transfer"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"
)

// verifyResult is the per-target outcome.
type verifyResult struct {
	Target        string `json:"target"`
	Checked       int    `json:"checked"`
	Missing       int    `json:"missing"`
	Corrupted     int    `json:"corrupted"`
	RemoteMissing int    `json:"remote_missing,omitempty"`
}

func (v verifyResult) ok() bool { return v.Missing == 0 && v.Corrupted == 0 && v.RemoteMissing == 0 }

func newVerifyCmd() *cobra.Command {
	var (
		jsonOut    bool
		remoteName string
	)
	cmd := &cobra.Command{
		Use:   "verify [target]...",
		Short: "Verify cache integrity and DVC compatibility",
		Long: "Re-hash every cached object and check it matches its recorded hash.\n" +
			"Because lode and DVC share the same format, a clean verify on a DVC repo\n" +
			"proves lode computes the same hashes DVC recorded — your data is intact and\n" +
			"compatible. With -r, also checks the remote has every referenced object.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runVerify(cmd.Context(), args, jsonOut, remoteName)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "structured JSON output")
	cmd.Flags().StringVarP(&remoteName, "remote", "r", "", "also verify this remote has every object")
	return cmd
}

func runVerify(ctx context.Context, targets []string, jsonOut bool, remoteName string) error {
	r, err := requireRepo()
	if err != nil {
		return err
	}
	c := cache.New(r.CacheDir())
	dvcFiles, err := dvcFilesFor(r, targets)
	if err != nil {
		return err
	}

	var store transfer.Store
	if remoteName != "" {
		store, err = openStore(r, remoteName)
		if err != nil {
			return err
		}
	}

	var results []verifyResult
	for _, df := range dvcFiles {
		f, err := dvcfile.Load(df)
		if err != nil {
			return err
		}
		dir := filepath.Dir(df)
		for _, out := range f.Outs {
			results = append(results, verifyOut(ctx, c, store, out, filepath.Join(dir, out.Path)))
		}
	}

	if jsonOut {
		if err := printJSON(results); err != nil {
			return err
		}
	} else {
		reportVerify(results)
	}
	for _, res := range results {
		if !res.ok() {
			return errVerifyFailed
		}
	}
	return nil
}

// verifyOut checks every object backing an output: presence + content integrity
// (re-hash and compare to the recorded id), and optionally remote presence.
func verifyOut(ctx context.Context, c *cache.Cache, store transfer.Store, out dvcfile.Out, target string) verifyResult {
	res := verifyResult{Target: target}

	// Collect every object id the output references.
	oids := []string{out.MD5}
	if out.IsDir() {
		if data, ok := readObject(c, out.MD5); ok {
			if entries, err := hashfile.ParseDir(data); err == nil {
				for _, e := range entries {
					oids = append(oids, e.MD5)
				}
			}
		}
	}

	var mu sync.Mutex
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(runtime.NumCPU())
	for _, oid := range oids {
		oid := oid
		g.Go(func() error {
			present, intact := objectIntact(c, oid)
			mu.Lock()
			res.Checked++
			switch {
			case !present:
				res.Missing++
			case !intact:
				res.Corrupted++
			}
			mu.Unlock()
			if store != nil {
				if ok, err := store.Has(gctx, oid); err == nil && !ok {
					mu.Lock()
					res.RemoteMissing++
					mu.Unlock()
				}
			}
			return nil
		})
	}
	_ = g.Wait()
	return res
}

// objectIntact reports whether the object is present and its content hashes to
// its id (the ".dir" suffix is stripped for directory manifests).
func objectIntact(c *cache.Cache, oid string) (present, intact bool) {
	p, ok := c.ResolveRead(oid)
	if !ok {
		return false, false
	}
	got, _, err := hashfile.HashFile(p)
	if err != nil {
		return true, false
	}
	return true, got == strings.TrimSuffix(oid, ".dir")
}

func readObject(c *cache.Cache, oid string) ([]byte, bool) {
	p, ok := c.ResolveRead(oid)
	if !ok {
		return nil, false
	}
	data, err := os.ReadFile(p)
	if err != nil {
		return nil, false
	}
	return data, true
}

func reportVerify(results []verifyResult) {
	allOK := true
	for _, res := range results {
		if res.ok() {
			infof("✓ %-24s %s intact", res.Target, plural(res.Checked, "object", "objects"))
			continue
		}
		allOK = false
		infof("✗ %-24s %d checked: %d missing, %d corrupted%s", res.Target, res.Checked,
			res.Missing, res.Corrupted, remoteMissingSuffix(res.RemoteMissing))
	}
	if allOK {
		infof("All objects intact and match their recorded hashes.")
	}
}

func remoteMissingSuffix(n int) string {
	if n == 0 {
		return ""
	}
	return fmt.Sprintf(", %d missing on remote", n)
}
