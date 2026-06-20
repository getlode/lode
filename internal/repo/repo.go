// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package repo handles DVC repository discovery and well-known paths.
package repo

import (
	"errors"
	"os"
	"path/filepath"
)

// ErrNotFound is returned when no .dvc directory is found.
var ErrNotFound = errors.New("not a DVC repository (no .dvc directory found)")

// Repo points at a discovered DVC repository.
type Repo struct {
	Root   string // directory containing .dvc
	DvcDir string // the .dvc directory
}

// Find walks up from start until it locates a .dvc directory.
func Find(start string) (*Repo, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return nil, err
	}
	for {
		dvc := filepath.Join(dir, ".dvc")
		if fi, err := os.Stat(dvc); err == nil && fi.IsDir() {
			return &Repo{Root: dir, DvcDir: dvc}, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return nil, ErrNotFound
		}
		dir = parent
	}
}

func (r *Repo) CacheDir() string   { return filepath.Join(r.DvcDir, "cache") }
func (r *Repo) TmpDir() string     { return filepath.Join(r.DvcDir, "tmp") }
func (r *Repo) StatePath() string  { return filepath.Join(r.TmpDir(), "lode", "state.db") }
func (r *Repo) LockPath() string   { return filepath.Join(r.TmpDir(), "lock") }
func (r *Repo) ConfigPath() string { return filepath.Join(r.DvcDir, "config") }

// InitMode selects the initialization flavor.
type InitMode int

const (
	ModeSCM   InitMode = iota // git-tracked repo (empty config + .dvc/.gitignore)
	ModeNoSCM                 // standalone repo (config has no_scm = True)
)

// InitOutcome reports what InitRepo did.
type InitOutcome int

const (
	Created            InitOutcome = iota // structure created
	AlreadyInitialized                    // .dvc/ already exists in root
	InsideExistingRepo                    // an ancestor already has a .dvc/
)

// InitRepo creates a .dvc repository structure byte-compatible with what
// `dvc init` (ModeSCM) or `dvc init --no-scm` (ModeNoSCM) produces. It does not
// create the cache directory (DVC creates it lazily on the first add). It is
// safe: if the directory already has a repo, or sits inside one, it reports the
// outcome without modifying anything.
func InitRepo(root string, mode InitMode) (*Repo, InitOutcome, error) {
	root, err := filepath.Abs(root)
	if err != nil {
		return nil, Created, err
	}
	dvc := filepath.Join(root, ".dvc")

	if fi, err := os.Stat(dvc); err == nil && fi.IsDir() {
		return &Repo{Root: root, DvcDir: dvc}, AlreadyInitialized, nil
	}
	if existing, err := Find(root); err == nil && existing.Root != root {
		return existing, InsideExistingRepo, nil
	}

	if err := os.MkdirAll(filepath.Join(dvc, "tmp"), 0o755); err != nil {
		return nil, Created, err
	}

	config := configNoSCM
	if mode == ModeSCM {
		config = configSCM
	}
	writes := map[string]string{
		filepath.Join(dvc, "config"):       config,
		filepath.Join(root, ".dvcignore"):  dvcignoreTemplate,
		filepath.Join(dvc, "tmp", "btime"): "",
	}
	if mode == ModeSCM {
		writes[filepath.Join(dvc, ".gitignore")] = dvcGitignore
	}
	for path, content := range writes {
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			return nil, Created, err
		}
	}
	return &Repo{Root: root, DvcDir: dvc}, Created, nil
}

// Init creates a standalone (no-scm) repository. Convenience wrapper kept for
// callers/tests that just need a working repo.
func Init(root string) (*Repo, error) {
	r, _, err := InitRepo(root, ModeNoSCM)
	return r, err
}
