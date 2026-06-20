// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/getlode/lode/internal/hashfile"
	"github.com/getlode/lode/internal/repo"
)

// Global flag state shared by subcommands.
var (
	flagVerbose bool
	flagQuiet   bool
	flagJobs    int
	flagChdir   string
	flagRehash  bool
)

// openState opens the state cache for r, honoring --rehash. If the DB is missing
// or corrupt it degrades to a full re-hash (returns nil) instead of failing —
// the state cache is only a performance optimization, never a source of truth,
// so it must never cause a wrong or aborted result.
func openState(r *repo.Repo) *hashfile.State {
	st, err := hashfile.OpenState(r.StatePath())
	if err != nil {
		infof("state cache unavailable (%v); re-hashing all files", err)
		return nil
	}
	st.ForceRehash = flagRehash
	return st
}

// infof prints a normal user-facing message (suppressed by --quiet).
func infof(format string, args ...any) {
	if flagQuiet {
		return
	}
	fmt.Fprintf(os.Stdout, format+"\n", args...)
}

// printJSON writes v as indented JSON to stdout.
func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
