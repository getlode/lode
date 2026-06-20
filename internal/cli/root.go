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
		Short:         "Fast, drop-in compatible data versioning (DVC 3.x)",
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

	root.AddCommand(
		newInitCmd(),
		newAddCmd(),
		newStatusCmd(),
		newPushCmd(),
		newFetchCmd(),
		newPullCmd(),
		newCheckoutCmd(),
		newGCCmd(),
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
