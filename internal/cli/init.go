package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/getlode/lode/internal/repo"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	var noSCM bool
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize a lode (DVC-compatible) repository",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(noSCM)
		},
	}
	cmd.Flags().BoolVar(&noSCM, "no-scm", false, "initialize without git (standalone, no Python or DVC required)")
	return cmd
}

func runInit(noSCM bool) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	mode := repo.ModeSCM
	switch {
	case noSCM:
		mode = repo.ModeNoSCM
	case !repo.InGitWorkTree(cwd):
		return fmt.Errorf("no git repository found here\n" +
			"  - run `lode init --no-scm` for a standalone repository (no git), or\n" +
			"  - run `git init` first, then `lode init`")
	}

	r, outcome, err := repo.InitRepo(cwd, mode)
	if err != nil {
		return err
	}
	switch outcome {
	case repo.AlreadyInitialized:
		infof("Already a repository (%s). Nothing to do.", r.DvcDir)
		return nil
	case repo.InsideExistingRepo:
		infof("Already inside a repository rooted at %s; not creating a nested one.", r.Root)
		return nil
	}

	if mode == repo.ModeSCM {
		_ = repo.GitAdd(r.Root,
			filepath.Join(".dvc", ".gitignore"),
			filepath.Join(".dvc", "config"),
			".dvcignore",
		)
	}
	infof("Initialized lode repository in %s", r.DvcDir)
	infof("Next: lode add <target>")
	return nil
}
