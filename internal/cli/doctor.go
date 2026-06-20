// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/getlode/lode/internal/lock"
	"github.com/getlode/lode/internal/remote"
	"github.com/getlode/lode/internal/repo"
	"github.com/spf13/cobra"
)

const (
	statusOK      = "ok"
	statusWarn    = "warn"
	statusProblem = "problem"
)

// doctorCheck is one diagnostic result.
type doctorCheck struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Detail     string `json:"detail"`
	Suggestion string `json:"suggestion,omitempty"`
}

func newDoctorCmd() *cobra.Command {
	var (
		jsonOut    bool
		remoteName string
	)
	cmd := &cobra.Command{
		Use:   "doctor",
		Short: "Diagnose the repository, cache, remotes and DVC coexistence",
		RunE: func(cmd *cobra.Command, args []string) error {
			checks := runDoctorChecks(cmd.Context(), remoteName)
			if jsonOut {
				if err := printJSON(checks); err != nil {
					return err
				}
			} else {
				printChecks(checks)
			}
			for _, c := range checks {
				if c.Status == statusProblem {
					os.Exit(1)
				}
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "structured JSON output")
	cmd.Flags().StringVarP(&remoteName, "remote", "r", "", "only check this remote")
	return cmd
}

func runDoctorChecks(ctx context.Context, remoteName string) []doctorCheck {
	var checks []doctorCheck

	cwd, _ := os.Getwd()
	r, err := repo.Find(cwd)
	if errors.Is(err, repo.ErrNotFound) {
		return []doctorCheck{{
			Name: "repo", Status: statusProblem,
			Detail:     "no repository found from the current directory",
			Suggestion: "run `lode init` (git) or `lode init --no-scm` (standalone)",
		}}
	} else if err != nil {
		return []doctorCheck{{Name: "repo", Status: statusProblem, Detail: err.Error()}}
	}
	checks = append(checks, doctorCheck{Name: "repo", Status: statusOK, Detail: "repository at " + r.Root})

	checks = append(checks, checkCacheWritable(r))
	checks = append(checks, checkFormat(r))
	checks = append(checks, checkRemotes(ctx, r, remoteName)...)
	checks = append(checks, checkCoexistence(r))
	return checks
}

func checkCacheWritable(r *repo.Repo) doctorCheck {
	probe := filepath.Join(r.DvcDir, ".lode-doctor-probe")
	if err := os.WriteFile(probe, []byte("x"), 0o644); err != nil {
		return doctorCheck{
			Name: "cache", Status: statusProblem,
			Detail:     "cannot write under " + r.DvcDir + ": " + err.Error(),
			Suggestion: "check directory permissions / available space",
		}
	}
	_ = os.Remove(probe)
	return doctorCheck{Name: "cache", Status: statusOK, Detail: "writable (" + r.CacheDir() + ")"}
}

func checkFormat(r *repo.Repo) doctorCheck {
	modern := filepath.Join(r.CacheDir(), "files", "md5")
	if _, err := os.Stat(modern); err == nil {
		return doctorCheck{Name: "format", Status: statusOK, Detail: "DVC 3.x cache layout (files/md5)"}
	}
	// Heuristic for legacy 2.x: top-level 2-char hex dirs directly in cache.
	entries, _ := os.ReadDir(r.CacheDir())
	for _, e := range entries {
		if e.IsDir() && len(e.Name()) == 2 && isHex(e.Name()) {
			return doctorCheck{
				Name: "format", Status: statusWarn,
				Detail:     "legacy DVC 2.x cache layout detected",
				Suggestion: "read is supported; new objects are written in the 3.x layout",
			}
		}
	}
	return doctorCheck{Name: "format", Status: statusOK, Detail: "DVC 3.x (no cached objects yet)"}
}

func checkRemotes(ctx context.Context, r *repo.Repo, only string) []doctorCheck {
	cfg, err := repo.LoadConfig(r.ConfigPath())
	if err != nil {
		return []doctorCheck{{Name: "remote", Status: statusProblem, Detail: "cannot read config: " + err.Error()}}
	}
	if len(cfg.Remotes) == 0 {
		return []doctorCheck{{
			Name: "remote", Status: statusWarn,
			Detail:     "no remote configured",
			Suggestion: "add one: `lode remote add -d myremote s3://bucket/path`",
		}}
	}
	var out []doctorCheck
	for name, rm := range cfg.Remotes {
		if only != "" && name != only {
			continue
		}
		store, err := remote.NewS3(rm)
		if err != nil {
			out = append(out, doctorCheck{
				Name: "remote:" + name, Status: statusProblem,
				Detail: err.Error(), Suggestion: "check the remote url/config",
			})
			continue
		}
		cctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		err = store.Reachable(cctx)
		cancel()
		if err != nil {
			out = append(out, doctorCheck{
				Name: "remote:" + name, Status: statusProblem,
				Detail:     "unreachable or invalid credentials: " + err.Error(),
				Suggestion: "verify endpoint, network, and credentials",
			})
			continue
		}
		out = append(out, doctorCheck{Name: "remote:" + name, Status: statusOK, Detail: rm.URL + " reachable"})
	}
	return out
}

func checkCoexistence(r *repo.Repo) doctorCheck {
	gl, err := lock.Acquire(r.LockPath())
	if err != nil {
		return doctorCheck{
			Name: "coexistence", Status: statusWarn,
			Detail:     "the repository lock is held by another process",
			Suggestion: "another lode/DVC process may be running; retry later",
		}
	}
	_ = gl.Release()
	return doctorCheck{Name: "coexistence", Status: statusOK, Detail: "lock free; safe to operate alongside DVC"}
}

func printChecks(checks []doctorCheck) {
	mark := map[string]string{statusOK: "✓", statusWarn: "!", statusProblem: "✗"}
	for _, c := range checks {
		infof("%s %-14s %s", mark[c.Status], c.Name, c.Detail)
		if c.Suggestion != "" && c.Status != statusOK {
			infof("    → %s", c.Suggestion)
		}
	}
}

func isHex(s string) bool {
	for _, r := range s {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f')) {
			return false
		}
	}
	return true
}
