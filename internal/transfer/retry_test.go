package transfer

import (
	"context"
	"errors"
	"testing"
	"time"
)

func fastPolicy(maxAttempts int) RetryPolicy {
	return RetryPolicy{MaxAttempts: maxAttempts, BaseDelay: time.Millisecond, MaxDelay: 4 * time.Millisecond}
}

func TestRetry_TransientThenSuccess(t *testing.T) {
	calls := 0
	err := retry(context.Background(), fastPolicy(4), func() error {
		calls++
		if calls < 3 {
			return errors.New("SlowDown: reduce your request rate")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("expected success after transient retries, got %v", err)
	}
	if calls != 3 {
		t.Fatalf("expected 3 attempts, got %d", calls)
	}
}

func TestRetry_PermanentFailsFast(t *testing.T) {
	calls := 0
	err := retry(context.Background(), fastPolicy(5), func() error {
		calls++
		return errors.New("AccessDenied")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Fatalf("permanent error must not be retried; got %d calls", calls)
	}
}

func TestRetry_ExhaustsOnPersistentTransient(t *testing.T) {
	calls := 0
	err := retry(context.Background(), fastPolicy(3), func() error {
		calls++
		return errors.New("503 ServiceUnavailable")
	})
	if err == nil {
		t.Fatal("expected error after exhausting attempts")
	}
	if calls != 3 {
		t.Fatalf("expected 3 attempts, got %d", calls)
	}
}

func TestRetry_ContextCancelStops(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	calls := 0
	err := retry(ctx, DefaultRetry, func() error {
		calls++
		return errors.New("SlowDown")
	})
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 1 {
		t.Fatalf("cancelled context must stop after the first attempt; got %d", calls)
	}
}

func TestIsTransient(t *testing.T) {
	transient := []string{"SlowDown", "503 boom", "InternalError", "connection reset by peer", "unexpected EOF", "i/o timeout"}
	permanent := []string{"AccessDenied", "NoSuchKey", "connection refused", "no such host", "SignatureDoesNotMatch"}
	for _, m := range transient {
		if !isTransient(errors.New(m)) {
			t.Errorf("%q should be transient", m)
		}
	}
	for _, m := range permanent {
		if isTransient(errors.New(m)) {
			t.Errorf("%q should be permanent", m)
		}
	}
}

type tempErr struct{}

func (tempErr) Error() string   { return "throttled (503)" }
func (tempErr) Temporary() bool { return true }

func TestRetry_TemporaryInterface(t *testing.T) {
	calls := 0
	err := retry(context.Background(), fastPolicy(3), func() error {
		calls++
		if calls < 2 {
			return tempErr{}
		}
		return nil
	})
	if err != nil || calls != 2 {
		t.Fatalf("Temporary() error should be retried: err=%v calls=%d", err, calls)
	}
}
