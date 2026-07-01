// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package cli wires the lode command-line interface.
package cli

import (
	"fmt"
	"os"
	"runtime"

	"github.com/spf13/cobra"
)

var version = "dev"

// SetVersion is called from main to inject the build version.
func SetVersion(v string) { version = v }

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:           "lode",
		Short:         "Fast DVC-compatible data versioning for add/status and S3 sync",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if flagChdir != "" {
				if err := os.Chdir(flagChdir); err != nil {
					return err
				}
			}
			return nil
		},
	}
	pf := root.PersistentFlags()
	pf.BoolVarP(&flagVerbose, "verbose", "v", false, "verbose output")
	pf.BoolVarP(&flagQuiet, "quiet", "q", false, "errors only")
	pf.IntVarP(&flagJobs, "jobs", "j", runtime.NumCPU(), "concurrency")
	pf.StringVar(&flagChdir, "cd", "", "run as if the current directory were this path")
	pf.BoolVar(&flagRehash, "rehash", false, "ignore the state cache and re-hash every file (use on NFS / restored backups where mtime is unreliable)")

	root.AddCommand(
		newInitCmd(),
		newAddCmd(),
		newStatusCmd(),
		newPushCmd(),
		newFetchCmd(),
		newPullCmd(),
		newCheckoutCmd(),
		newGCCmd(),
		newVerifyCmd(),
		newDoctorCmd(),
		newRemoteCmd(),
	)
	return root
}

// Execute runs the root command.
func Execute() {
	if err := newRootCmd().Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}
