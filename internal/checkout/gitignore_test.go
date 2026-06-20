// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package checkout

import (
	"os"
	"path/filepath"
	"testing"
)

func TestAddToGitignore_FormatAndIdempotency(t *testing.T) {
	dir := t.TempDir()
	data := filepath.Join(dir, "data")
	if err := os.MkdirAll(data, 0o755); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(data, "foo.csv")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := AddToGitignore(target); err != nil {
		t.Fatal(err)
	}
	gi := filepath.Join(data, ".gitignore")
	got, err := os.ReadFile(gi)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "/foo.csv\n" {
		t.Fatalf("unexpected gitignore: %q", got)
	}

	// Idempotent: second call must not duplicate.
	if err := AddToGitignore(target); err != nil {
		t.Fatal(err)
	}
	got2, _ := os.ReadFile(gi)
	if string(got2) != "/foo.csv\n" {
		t.Fatalf("entrada duplicada: %q", got2)
	}
}

func TestAddToGitignore_PreservesExisting(t *testing.T) {
	dir := t.TempDir()
	gi := filepath.Join(dir, ".gitignore")
	if err := os.WriteFile(gi, []byte("/other\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(dir, "data")
	if err := os.WriteFile(target, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := AddToGitignore(target); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(gi)
	if string(got) != "/other\n/data\n" {
		t.Fatalf("did not preserve previous entries: %q", got)
	}
}
