package integration

import (
	"context"
	"testing"
	"time"

	"github.com/getlode/lode/internal/remote"
	"github.com/getlode/lode/internal/repo"
)

// TestAuth_NoCredsFastFailNoPanic: off-cloud, with no credentials available, the
// credential chain (which now includes the IAM provider) must not panic and must
// not hang — the metadata probe is timeout-bounded. This is the regression test
// for the bug that originally forced IAM to be dropped.
func TestAuth_NoCredsFastFailNoPanic(t *testing.T) {
	s, err := remote.NewS3(repo.Remote{
		Name:        "x",
		URL:         "s3://bucket/prefix",
		EndpointURL: "http://127.0.0.1:1", // refused
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- s.Reachable(ctx) }()
	select {
	case e := <-done:
		if e == nil {
			t.Fatal("expected an error reaching a dead endpoint with no credentials")
		}
	case <-ctx.Done():
		t.Fatal("Reachable hung > 20s — the IAM metadata probe is not timeout-bounded")
	}
}
