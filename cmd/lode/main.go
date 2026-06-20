// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package main

import (
	"runtime/debug"

	"github.com/getlode/lode/internal/cli"
)

// version/commit are injected by the linker at release time (see .goreleaser.yaml).
var (
	version = "dev"
	commit  = ""
)

func main() {
	cli.SetVersion(resolveVersion())
	cli.Execute()
}

// resolveVersion prefers the linker-injected version (release builds); falls back
// to the module version embedded by the Go toolchain so `go install ...@vX.Y.Z`
// reports the tag instead of "dev".
func resolveVersion() string {
	if version != "dev" {
		if commit != "" {
			return version + " (" + commit + ")"
		}
		return version
	}
	if bi, ok := debug.ReadBuildInfo(); ok {
		if v := bi.Main.Version; v != "" && v != "(devel)" {
			return v
		}
	}
	return "dev"
}
