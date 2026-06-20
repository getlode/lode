package cli

import (
	"os"
	"path/filepath"

	"github.com/jtorchia/lode/internal/cache"
	"github.com/jtorchia/lode/internal/dvcfile"
	"github.com/jtorchia/lode/internal/hashfile"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "status [target]...",
		Short: "Reporta cambios sin modificar el repo",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStatus(args, jsonOut)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "salida estructurada en JSON")
	return cmd
}

// statusEntry is one tracked output's state.
type statusEntry struct {
	Target string `json:"target"`
	State  string `json:"state"`
}

func runStatus(targets []string, jsonOut bool) error {
	r, err := findRepo()
	if err != nil {
		return err
	}
	st, err := hashfile.OpenState(r.StatePath())
	if err != nil {
		return err
	}
	defer st.Close()

	c := cache.New(r.CacheDir())
	dvcFiles, err := dvcFilesFor(r, targets)
	if err != nil {
		return err
	}

	var results []statusEntry
	for _, df := range dvcFiles {
		f, err := dvcfile.Load(df)
		if err != nil {
			return err
		}
		for _, out := range f.Outs {
			wsPath := filepath.Join(filepath.Dir(df), out.Path)
			results = append(results, statusEntry{
				Target: wsPath,
				State:  outState(c, st, out, wsPath),
			})
		}
	}

	if jsonOut {
		return printJSON(results)
	}
	clean := true
	for _, e := range results {
		if e.State != "up to date" {
			clean = false
			infof("%-24s %s", e.Target, e.State)
		}
	}
	if clean {
		infof("Data and pipelines are up to date.")
	}
	return nil
}

func outState(c *cache.Cache, st *hashfile.State, out dvcfile.Out, wsPath string) string {
	if !c.Has(out.MD5) {
		return "not in cache"
	}
	info, err := os.Stat(wsPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "deleted"
		}
		return "error: " + err.Error()
	}

	var current string
	if info.IsDir() {
		tree, err := hashfile.HashTree(wsPath, st)
		if err != nil {
			return "error: " + err.Error()
		}
		current = hashfile.DirOID(tree.Entries)
	} else {
		current, _, err = hashfile.HashFileCached(wsPath, st)
		if err != nil {
			return "error: " + err.Error()
		}
	}
	if current == out.MD5 {
		return "up to date"
	}
	return "modified"
}
