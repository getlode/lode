package integration

import (
	"crypto/rand"
	"encoding/hex"
	"testing"
)

// randSuffix returns a short random hex string for unique bucket names.
func randSuffix(t *testing.T) string {
	t.Helper()
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return hex.EncodeToString(b)
}
