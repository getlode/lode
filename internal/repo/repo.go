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

// Init creates the minimal .dvc layout (used by tests and to operate on a fresh
// directory). It mirrors what `dvc init --no-scm` produces for the parts we use.
func Init(root string) (*Repo, error) {
	dvc := filepath.Join(root, ".dvc")
	for _, d := range []string{dvc, filepath.Join(dvc, "cache"), filepath.Join(dvc, "tmp")} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			return nil, err
		}
	}
	cfg := filepath.Join(dvc, "config")
	if _, err := os.Stat(cfg); os.IsNotExist(err) {
		if err := os.WriteFile(cfg, nil, 0o644); err != nil {
			return nil, err
		}
	}
	return &Repo{Root: root, DvcDir: dvc}, nil
}
