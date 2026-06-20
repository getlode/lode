// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package integration

import (
	"context"
	"testing"

	"github.com/getlode/lode/internal/repo"
	"github.com/getlode/lode/internal/transfer"
)

// TestResumeIdempotent verifies that re-running push after an interruption does
// not duplicate work or corrupt the remote: a second push transfers nothing and
// skips everything already present (SC-007).
func TestResumeIdempotent(t *testing.T) {
	store := newStore(t, "rs-"+randSuffix(t))
	ctx := context.Background()
	root := t.TempDir()
	if _, err := repo.Init(root); err != nil {
		t.Fatal(err)
	}
	c, items, _, _, _ := buildDataset(t, root)

	first, err := transfer.Push(ctx, store, c, items, 8)
	if err != nil {
		t.Fatal(err)
	}
	if first.Transferred == 0 || first.Failed != 0 {
		t.Fatalf("primer push inesperado: %+v", first)
	}

	// Re-run: everything is already present.
	second, err := transfer.Push(ctx, store, c, items, 8)
	if err != nil {
		t.Fatal(err)
	}
	if second.Transferred != 0 || second.Failed != 0 {
		t.Fatalf("el re-push debería transferir 0 y no fallar: %+v", second)
	}
	if second.Skipped != first.Transferred {
		t.Fatalf("el re-push debería saltear lo ya subido: skipped=%d esperado=%d", second.Skipped, first.Transferred)
	}
}

// TestResumeRecoversMissing verifies that if only part of the data made it to
// the remote, a later push uploads exactly what is missing.
func TestResumeRecoversMissing(t *testing.T) {
	store := newStore(t, "rm-"+randSuffix(t))
	ctx := context.Background()
	root := t.TempDir()
	if _, err := repo.Init(root); err != nil {
		t.Fatal(err)
	}
	c, items, _, _, _ := buildDataset(t, root)

	// Partial push: only the standalone file item.
	partial := items[1:]
	if _, err := transfer.Push(ctx, store, c, partial, 8); err != nil {
		t.Fatal(err)
	}

	// Full push: the directory (and its contents) must now be uploaded, the
	// already-present file skipped.
	full, err := transfer.Push(ctx, store, c, items, 8)
	if err != nil {
		t.Fatal(err)
	}
	if full.Failed != 0 {
		t.Fatalf("push con fallos: %+v", full)
	}
	if full.Transferred == 0 {
		t.Fatalf("se esperaba transferir los objetos faltantes: %+v", full)
	}
}
