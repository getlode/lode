// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package integration

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/getlode/lode/internal/cache"
	"github.com/getlode/lode/internal/checkout"
	"github.com/getlode/lode/internal/dvcfile"
	"github.com/getlode/lode/internal/transfer"
)

// TestInterop_DVCPushDvcgoPull verifies lode can fetch+materialize objects that
// the reference DVC implementation pushed to the same S3 remote (SC-002). It
// needs both a MinIO (MINIO_* env) and a working `dvc` with the s3 plugin
// (DVC_BIN; PYTHONPATH may be required).
func TestInterop_DVCPushDvcgoPull(t *testing.T) {
	dvc := os.Getenv("DVC_BIN")
	if dvc == "" {
		t.Skip("DVC_BIN not set; skipping interop with real DVC")
	}
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" || testing.Short() {
		t.Skip("MINIO_ENDPOINT not set (or -short)")
	}
	bucket := "io-" + randSuffix(t)
	store := newStore(t, bucket) // creates the bucket
	ctx := context.Background()

	dir := t.TempDir()
	runDVC(t, dir, dvc, "init", "--no-scm", "-q")
	if err := os.WriteFile(filepath.Join(dir, "payload.bin"), []byte("interop-payload-xyz"), 0o644); err != nil {
		t.Fatal(err)
	}
	runDVC(t, dir, dvc, "remote", "add", "-d", "r", "s3://"+bucket+"/store")
	runDVC(t, dir, dvc, "remote", "modify", "r", "endpointurl", endpoint)
	runDVC(t, dir, dvc, "add", "payload.bin")
	runDVC(t, dir, dvc, "push")

	// Read the .dvc DVC produced and fetch via lode from the same remote.
	f, err := dvcfile.Load(filepath.Join(dir, "payload.bin.dvc"))
	if err != nil {
		t.Fatal(err)
	}
	out := f.Outs[0]

	c := cache.New(filepath.Join(t.TempDir(), "cache"))
	if _, err := transfer.Fetch(ctx, store, c, []transfer.FetchItem{{OID: out.MD5, IsDir: out.IsDir()}}, 8); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "restored.bin")
	if err := checkout.Materialize(c, dvcfile.Out{MD5: out.MD5, Hash: "md5", Path: "restored.bin"}, dst, checkout.ParseCacheTypes("")); err != nil {
		t.Fatal(err)
	}
	assertContent(t, dst, "interop-payload-xyz")
}

func runDVC(t *testing.T, dir, name string, args ...string) { //nolint

	t.Helper()
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"DVC_NO_ANALYTICS=1",
		"AWS_ACCESS_KEY_ID="+os.Getenv("MINIO_ACCESS_KEY"),
		"AWS_SECRET_ACCESS_KEY="+os.Getenv("MINIO_SECRET_KEY"),
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("dvc %v: %v\n%s", args, err, stderr.String())
	}
}
