// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package repo

// Byte-exact templates matching what `dvc init` (DVC 3.x) writes. Captured from
// the reference implementation; see specs/002-init-onboarding/research.md. Any
// change here must keep the init oracle test green (Constitution I/II).
const (
	// dvcignoreTemplate is the root .dvcignore DVC writes (139 bytes).
	dvcignoreTemplate = "# Add patterns of files dvc should ignore, which could improve\n" +
		"# the performance. Learn more at\n" +
		"# https://dvc.org/doc/user-guide/dvcignore\n"

	// dvcGitignore is .dvc/.gitignore, written only in scm (git) mode.
	dvcGitignore = "/config.local\n/tmp\n/cache\n"

	// configNoSCM is the .dvc/config written in --no-scm mode (4-space indent).
	configNoSCM = "[core]\n    no_scm = True\n"
	// configSCM is the .dvc/config written in git mode: empty.
	configSCM = ""
)
