package transfer

import (
	"context"
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/getlode/lode/internal/cache"
	"github.com/getlode/lode/internal/hashfile"
	"golang.org/x/sync/errgroup"
)

// FetchItem is an output to fetch: OID is the file or .dir object id; IsDir
// marks directory outputs whose .dir must be resolved to discover contents.
type FetchItem struct {
	OID   string
	IsDir bool
}

// Fetch downloads the objects backing items from the store into the cache,
// verifying each object's integrity. For directories the .dir object is fetched
// first so its contents can be resolved.
func Fetch(ctx context.Context, store Store, c *cache.Cache, items []FetchItem, jobs int) (Result, error) {
	var res Result

	// Resolve directory manifests, collecting all data oids to fetch.
	var dataOIDs []string
	for _, it := range items {
		if !it.IsDir {
			dataOIDs = append(dataOIDs, it.OID)
			continue
		}
		if !c.Has(it.OID) {
			if err := downloadVerify(ctx, store, c, it.OID, jobs); err != nil {
				return res, fmt.Errorf("fetch .dir %s: %w", it.OID, err)
			}
			res.Transferred++
		} else {
			res.Skipped++
		}
		entries, err := readManifest(c, it.OID)
		if err != nil {
			return res, err
		}
		for _, e := range entries {
			dataOIDs = append(dataOIDs, e.MD5)
		}
	}
	dataOIDs = uniq(dataOIDs)

	var mu sync.Mutex
	g, gctx := errgroup.WithContext(ctx)
	g.SetLimit(jobs)
	for _, o := range dataOIDs {
		o := o
		g.Go(func() error {
			if c.Has(o) {
				mu.Lock()
				res.Skipped++
				mu.Unlock()
				return nil
			}
			if err := downloadVerify(gctx, store, c, o, jobs); err != nil {
				mu.Lock()
				res.Failed++
				mu.Unlock()
				return nil
			}
			mu.Lock()
			res.Transferred++
			mu.Unlock()
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return res, err
	}
	return res, nil
}

// downloadVerify downloads oid to a temp file, checks its hash, and atomically
// adopts it into the cache. A corrupted object is discarded and reported.
func downloadVerify(ctx context.Context, store Store, c *cache.Cache, oid string, jobs int) error {
	tmpDir := c.TempDir()
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(tmpDir, "dl-*")
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	tmp.Close()
	defer os.Remove(tmpName)

	if err := retry(ctx, DefaultRetry, func() error {
		return store.Get(ctx, oid, tmpName)
	}); err != nil {
		return err
	}

	got, _, err := hashfile.HashFile(tmpName)
	if err != nil {
		return err
	}
	// For a .dir object the expected content hash is the oid without its suffix.
	want := strings.TrimSuffix(oid, ".dir")
	if got != want {
		return fmt.Errorf("corrupted object %s: hash %s does not match", oid, got)
	}
	return c.Adopt(tmpName, oid)
}

func readManifest(c *cache.Cache, dirOID string) ([]hashfile.DirEntry, error) {
	data, err := os.ReadFile(c.ObjectPath(dirOID))
	if err != nil {
		return nil, err
	}
	return hashfile.ParseDir(data)
}
