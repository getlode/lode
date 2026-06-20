package transfer

import (
	"context"
	"sync"

	"github.com/jtorchia/lode/internal/cache"
	"golang.org/x/sync/errgroup"
)

// Item is one tracked output to transfer. Contents is non-nil for a directory
// (.dir) and lists the object ids of its files; the directory's own .dir object
// id is OID.
type Item struct {
	OID      string
	Contents []string
}

// Result summarizes a transfer.
type Result struct {
	Transferred int
	Skipped     int
	Failed      int
}

// Push uploads the objects backing items to the store. Directory (.dir) objects
// are uploaded only after all their contents succeed, so the remote never has a
// dangling .dir. Already-present objects are skipped; a re-run after an
// interruption resumes from what is missing.
func Push(ctx context.Context, store Store, c *cache.Cache, items []Item, jobs int) (Result, error) {
	var res Result

	// Phase 1: data objects (files + directory contents).
	var dataOIDs []string
	for _, it := range items {
		if it.Contents == nil {
			dataOIDs = append(dataOIDs, it.OID)
		} else {
			dataOIDs = append(dataOIDs, it.Contents...)
		}
	}
	dataOIDs = uniq(dataOIDs)

	missing, err := MissingOnRemote(ctx, store, dataOIDs, jobs)
	if err != nil {
		return res, err
	}
	failed := uploadSet(ctx, store, c, missing, jobs)
	res.Transferred += len(missing) - len(failed)
	res.Skipped += len(dataOIDs) - len(missing)
	res.Failed += len(failed)

	// Phase 2: .dir objects whose contents are fully present.
	var dirOIDs []string
	for _, it := range items {
		if it.Contents == nil {
			continue
		}
		if anyIn(it.Contents, failed) {
			res.Failed++
			continue
		}
		dirOIDs = append(dirOIDs, it.OID)
	}
	dirOIDs = uniq(dirOIDs)

	dirMissing, err := MissingOnRemote(ctx, store, dirOIDs, jobs)
	if err != nil {
		return res, err
	}
	dirFailed := uploadSet(ctx, store, c, dirMissing, jobs)
	res.Transferred += len(dirMissing) - len(dirFailed)
	res.Skipped += len(dirOIDs) - len(dirMissing)
	res.Failed += len(dirFailed)

	return res, nil
}

// uploadSet uploads each oid from the cache concurrently, returning the set of
// oids that failed.
func uploadSet(ctx context.Context, store Store, c *cache.Cache, oids []string, jobs int) map[string]struct{} {
	failed := make(map[string]struct{})
	var mu sync.Mutex
	g := new(errgroup.Group)
	g.SetLimit(jobs)
	for _, o := range oids {
		o := o
		g.Go(func() error {
			if err := store.Put(ctx, o, c.ObjectPath(o)); err != nil {
				mu.Lock()
				failed[o] = struct{}{}
				mu.Unlock()
			}
			return nil
		})
	}
	_ = g.Wait()
	return failed
}

func anyIn(oids []string, set map[string]struct{}) bool {
	for _, o := range oids {
		if _, ok := set[o]; ok {
			return true
		}
	}
	return false
}
