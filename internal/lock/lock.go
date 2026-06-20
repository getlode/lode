// Package lock implements DVC-compatible repository locking so lode can
// coexist with DVC-Python on the same repo.
package lock

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/gofrs/flock"
)

// ErrLocked is returned when the global lock cannot be acquired in time.
var ErrLocked = errors.New("unable to acquire lock; another DVC process is likely running")

// timeout and retry mirror DVC's defaults (3s, 0.5s between attempts).
const (
	timeout    = 3 * time.Second
	retryDelay = 500 * time.Millisecond
)

// Global is the exclusive repository lock held over .dvc/tmp/lock. It uses
// flock(2), the same primitive as DVC's zc.lockfile, so the two tools mutually
// exclude each other.
type Global struct {
	fl *flock.Flock
}

// Acquire takes the exclusive lock at lockPath, retrying until the timeout.
func Acquire(lockPath string) (*Global, error) {
	if err := os.MkdirAll(filepath.Dir(lockPath), 0o755); err != nil {
		return nil, err
	}
	fl := flock.New(lockPath)
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	locked, err := fl.TryLockContext(ctx, retryDelay)
	if err != nil {
		return nil, ErrLocked
	}
	if !locked {
		return nil, ErrLocked
	}
	return &Global{fl: fl}, nil
}

// Release frees the lock.
func (g *Global) Release() error { return g.fl.Unlock() }
