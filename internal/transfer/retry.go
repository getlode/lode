package transfer

import (
	"context"
	"errors"
	"math/rand"
	"net"
	"strings"
	"time"
)

// RetryPolicy controls retry/backoff for transient transfer failures.
type RetryPolicy struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
}

// DefaultRetry is the policy used by Push/Fetch. Tuned for object stores that
// throttle (S3 "SlowDown" / 503) under large transfers.
var DefaultRetry = RetryPolicy{
	MaxAttempts: 4,
	BaseDelay:   200 * time.Millisecond,
	MaxDelay:    10 * time.Second,
}

// retry runs op, retrying transient failures with exponential backoff plus full
// jitter up to the policy's attempt limit. Permanent errors and context
// cancellation stop retrying immediately.
func retry(ctx context.Context, p RetryPolicy, op func() error) error {
	attempts := p.MaxAttempts
	if attempts < 1 {
		attempts = 1
	}
	var err error
	for attempt := 1; attempt <= attempts; attempt++ {
		if err = op(); err == nil {
			return nil
		}
		if ctx.Err() != nil || !isTransient(err) || attempt == attempts {
			return err
		}
		// Exponential backoff with full jitter, capped at MaxDelay, so concurrent
		// retries do not synchronize and worsen throttling.
		backoff := p.BaseDelay << (attempt - 1)
		if p.MaxDelay > 0 && backoff > p.MaxDelay {
			backoff = p.MaxDelay
		}
		delay := time.Duration(rand.Int63n(int64(backoff) + 1))
		select {
		case <-ctx.Done():
			return err
		case <-time.After(delay):
		}
	}
	return err
}

// isTransient reports whether err is worth retrying (network blips, throttling,
// 5xx) versus a permanent failure (auth, not found, bad request, missing file).
func isTransient(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}
	var tmp interface{ Temporary() bool }
	if errors.As(err, &tmp) && tmp.Temporary() {
		return true
	}
	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return true
	}
	msg := err.Error()
	for _, s := range transientMarkers {
		if strings.Contains(msg, s) {
			return true
		}
	}
	return false
}

// transientMarkers are substrings of errors that object stores return for
// retriable conditions. Deliberately excludes "connection refused" / "no such
// host" (a dead endpoint or bad DNS is treated as permanent — fail fast).
var transientMarkers = []string{
	"SlowDown", "RequestTimeout", "InternalError", "ServiceUnavailable",
	"503", "500", "connection reset", "broken pipe", "unexpected EOF",
	"i/o timeout", "TLS handshake timeout",
}
