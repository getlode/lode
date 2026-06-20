// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

// Package integration exercises push/pull/gc against a live S3-compatible
// backend. It connects to the MinIO described by these env vars and skips when
// they are absent (so `go test -short` and CI-without-S3 stay green):
//
//	MINIO_ENDPOINT     e.g. http://localhost:9000
//	MINIO_ACCESS_KEY   e.g. minioadmin
//	MINIO_SECRET_KEY   e.g. minioadmin
//
// A unique bucket is created per run.
package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/getlode/lode/internal/cache"
	"github.com/getlode/lode/internal/checkout"
	"github.com/getlode/lode/internal/dvcfile"
	"github.com/getlode/lode/internal/hashfile"
	"github.com/getlode/lode/internal/remote"
	"github.com/getlode/lode/internal/repo"
	"github.com/getlode/lode/internal/transfer"
)

func newStore(t *testing.T, bucket string) *remote.S3 {
	t.Helper()
	endpoint := os.Getenv("MINIO_ENDPOINT")
	if endpoint == "" || testing.Short() {
		t.Skip("MINIO_ENDPOINT no seteado (o -short); se omite la integración S3")
	}
	s, err := remote.NewS3(repo.Remote{
		Name:            bucket,
		URL:             "s3://" + bucket + "/store",
		EndpointURL:     endpoint,
		AccessKeyID:     os.Getenv("MINIO_ACCESS_KEY"),
		SecretAccessKey: os.Getenv("MINIO_SECRET_KEY"),
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := s.EnsureBucket(context.Background()); err != nil {
		t.Fatal(err)
	}
	return s
}

// cacheDir builds a fresh dataset (a directory + a single file), tracks it into
// a cache, and returns the cache plus the push/fetch items and outs.
func buildDataset(t *testing.T, root string) (*cache.Cache, []transfer.Item, []transfer.FetchItem, []dvcfile.Out, []string) {
	t.Helper()
	data := filepath.Join(root, "data")
	if err := os.MkdirAll(filepath.Join(data, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	files := map[string]string{"a.txt": "alpha", "sub/b.txt": "beta"}
	for rel, c := range files {
		if err := os.WriteFile(filepath.Join(data, filepath.FromSlash(rel)), []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	lonely := filepath.Join(root, "f.bin")
	if err := os.WriteFile(lonely, []byte("lonely-content"), 0o644); err != nil {
		t.Fatal(err)
	}

	c := cache.New(filepath.Join(root, ".dvc", "cache"))

	tree, err := hashfile.HashTree(data, nil)
	if err != nil {
		t.Fatal(err)
	}
	for i, e := range tree.Entries {
		if err := c.AddFile(tree.AbsPaths[i], e.MD5); err != nil {
			t.Fatal(err)
		}
	}
	dirOID := hashfile.DirOID(tree.Entries)
	if err := c.AddBytes(hashfile.SerializeDir(tree.Entries), dirOID); err != nil {
		t.Fatal(err)
	}
	fileSum, fileSize, _ := hashfile.HashFile(lonely)
	if err := c.AddFile(lonely, fileSum); err != nil {
		t.Fatal(err)
	}

	contents := make([]string, len(tree.Entries))
	for i, e := range tree.Entries {
		contents[i] = e.MD5
	}
	nfiles := int64(tree.NFiles)
	items := []transfer.Item{
		{OID: dirOID, Contents: contents},
		{OID: fileSum},
	}
	fitems := []transfer.FetchItem{{OID: dirOID, IsDir: true}, {OID: fileSum, IsDir: false}}
	outs := []dvcfile.Out{
		{MD5: dirOID, Size: tree.TotalSize, NFiles: &nfiles, Hash: "md5", Path: "data"},
		{MD5: fileSum, Size: fileSize, Hash: "md5", Path: "f.bin"},
	}
	wsPaths := []string{data, lonely}
	return c, items, fitems, outs, wsPaths
}

func TestRoundTrip(t *testing.T) {
	store := newStore(t, "rt-"+randSuffix(t))
	ctx := context.Background()
	root := t.TempDir()
	if _, err := repo.Init(root); err != nil {
		t.Fatal(err)
	}

	c, items, fitems, outs, wsPaths := buildDataset(t, root)

	res, err := transfer.Push(ctx, store, c, items, 8)
	if err != nil {
		t.Fatal(err)
	}
	if res.Failed != 0 {
		t.Fatalf("push con fallos: %+v", res)
	}

	// Simulate a clean clone: wipe cache and workspace data, keep nothing local.
	os.RemoveAll(filepath.Join(root, ".dvc", "cache"))
	os.RemoveAll(filepath.Join(root, "data"))
	os.Remove(filepath.Join(root, "f.bin"))

	c2 := cache.New(filepath.Join(root, ".dvc", "cache"))
	if _, err := transfer.Fetch(ctx, store, c2, fitems, 8); err != nil {
		t.Fatal(err)
	}
	types := checkout.ParseCacheTypes("")
	for i, out := range outs {
		if err := checkout.Materialize(c2, out, wsPaths[i], types); err != nil {
			t.Fatalf("materialize %s: %v", wsPaths[i], err)
		}
	}

	assertContent(t, filepath.Join(root, "data", "a.txt"), "alpha")
	assertContent(t, filepath.Join(root, "data", "sub", "b.txt"), "beta")
	assertContent(t, filepath.Join(root, "f.bin"), "lonely-content")
}

func assertContent(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != want {
		t.Fatalf("%s: got %q want %q", path, got, want)
	}
}
