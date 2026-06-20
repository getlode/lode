package transfer

import (
	"context"
	"sync"

	"golang.org/x/sync/errgroup"
)

// traverseThreshold mirrors DVC's heuristic intent: with many oids to check, a
// single bulk LIST beats one HEAD per object; with few, HEAD per object wins.
const traverseThreshold = 100

// MissingOnRemote returns the subset of oids not present in the store.
func MissingOnRemote(ctx context.Context, store Store, oids []string, jobs int) ([]string, error) {
	if len(oids) == 0 {
		return nil, nil
	}
	if len(oids) >= traverseThreshold {
		present, err := store.ListOIDs(ctx)
		if err != nil {
			return nil, err
		}
		var missing []string
		for _, o := range oids {
			if _, ok := present[o]; !ok {
				missing = append(missing, o)
			}
		}
		return missing, nil
	}

	var mu sync.Mutex
	var missing []string
	g, ctx := errgroup.WithContext(ctx)
	g.SetLimit(jobs)
	for _, o := range oids {
		o := o
		g.Go(func() error {
			ok, err := store.Has(ctx, o)
			if err != nil {
				return err
			}
			if !ok {
				mu.Lock()
				missing = append(missing, o)
				mu.Unlock()
			}
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}
	return missing, nil
}

func uniq(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := in[:0]
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
