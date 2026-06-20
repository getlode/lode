// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package integration

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestRemoteModify_SecretNotEchoed: setting a secret via stdin must not echo the
// value, and must not require it on argv (where ps/history would capture it).
func TestRemoteModify_SecretNotEchoed(t *testing.T) {
	bin := lodeBin(t)
	dir := t.TempDir()
	runTool(t, dir, bin, "init", "--no-scm")
	runTool(t, dir, bin, "remote", "add", "r", "s3://b/p")

	const secret = "TOPSECRETVALUE123"
	cmd := exec.Command(bin, "remote", "modify", "r", "secret_access_key")
	cmd.Dir = dir
	cmd.Stdin = strings.NewReader(secret)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("modify failed: %v\n%s", err, out)
	}
	if strings.Contains(string(out), secret) {
		t.Fatalf("secret value was echoed to output:\n%s", out)
	}
	// It must actually have been stored.
	cfg, _ := os.ReadFile(filepath.Join(dir, ".dvc", "config"))
	if !strings.Contains(string(cfg), secret) {
		t.Fatalf("secret was not written to config")
	}
}
