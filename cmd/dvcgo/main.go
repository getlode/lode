package main

import "github.com/jtorchia/dvcgo/internal/cli"

// Injected by the linker at release time (see .goreleaser.yaml).
var (
	version = "dev"
	commit  = ""
)

func main() {
	v := version
	if commit != "" {
		v += " (" + commit + ")"
	}
	cli.SetVersion(v)
	cli.Execute()
}
