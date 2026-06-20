package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/getlode/lode/internal/repo"
)

// requireRepo locates the repository from the current working directory and
// returns an actionable error when none is found (guiding the user to `init`).
func requireRepo() (*repo.Repo, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	r, err := repo.Find(cwd)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, errNoRepo
	}
	return r, err
}

// errNoRepo is the guided "not a repository" error.
var errNoRepo = fmt.Errorf("not a lode repository (no .dvc directory found)\n" +
	"  - run `lode init` inside a git repo, or `lode init --no-scm` for a standalone one\n" +
	"  - or pass `--cd <dir>` if your repository is elsewhere")

// errNoRemote is the guided "no remote configured" error.
var errNoRemote = fmt.Errorf("no remote configured\n" +
	"  - add one: `lode remote add -d myremote s3://bucket/path`\n" +
	"  - then set its endpoint if needed: `lode remote modify myremote endpointurl <url>`\n" +
	"  - or target an existing remote with `-r <name>`")

// errVerifyFailed is returned when verify finds missing or corrupted objects.
var errVerifyFailed = fmt.Errorf("verification failed: some objects are missing or corrupted")

// errPartialTransfer is returned when some objects could not be transferred even
// after retries. A re-run resumes — already-completed objects are skipped.
func errPartialTransfer(n int) error {
	return fmt.Errorf("%s failed to transfer after retries; re-run to resume (completed objects are skipped)", plural(n, "object", "objects"))
}

// errAddNoTarget is shown when `add` is run without a file or directory.
var errAddNoTarget = fmt.Errorf("specify at least one file or directory to track, e.g. `lode add data/`")

// missingObjectsHint formats the guidance shown when objects are missing locally.
func missingObjectsHint(n int) string {
	return fmt.Sprintf("%d object(s) not in cache; run `lode pull` to fetch them from the remote", n)
}
