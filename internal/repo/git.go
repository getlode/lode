package repo

import (
	"os/exec"
	"strings"
)

// InGitWorkTree reports whether dir is inside a git work tree. It shells out to
// git, which is only consulted in scm mode (the standalone --no-scm path never
// calls this).
func InGitWorkTree(dir string) bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = dir
	out, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(out)) == "true"
}

// GitAdd stages the given paths in the git repo at dir, mirroring what DVC does
// after init.
func GitAdd(dir string, paths ...string) error {
	cmd := exec.Command("git", append([]string{"add", "--"}, paths...)...)
	cmd.Dir = dir
	return cmd.Run()
}
