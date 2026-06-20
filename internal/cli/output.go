package cli

import (
	"encoding/json"
	"fmt"
	"os"
)

// Global flag state shared by subcommands.
var (
	flagVerbose bool
	flagQuiet   bool
	flagJobs    int
	flagChdir   string
)

// infof prints a normal user-facing message (suppressed by --quiet).
func infof(format string, args ...any) {
	if flagQuiet {
		return
	}
	fmt.Fprintf(os.Stdout, format+"\n", args...)
}

// verbosef prints only when --verbose is set.
func verbosef(format string, args ...any) {
	if flagVerbose {
		fmt.Fprintf(os.Stderr, format+"\n", args...)
	}
}

// printJSON writes v as indented JSON to stdout.
func printJSON(v any) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}
