// Package transfer moves objects between the local cache and a remote store,
// preserving DVC's ordering and integrity guarantees.
package transfer

import "context"

// Store is a remote object store addressed by DVC object ids. Implemented by
// internal/remote for S3-compatible backends.
type Store interface {
	// Has reports whether the object exists in the remote.
	Has(ctx context.Context, oid string) (bool, error)
	// ListOIDs returns the set of object ids present in the remote (used for
	// the bulk traverse status strategy).
	ListOIDs(ctx context.Context) (map[string]struct{}, error)
	// Put uploads localPath as the object oid.
	Put(ctx context.Context, oid, localPath string) error
	// Get downloads the object oid into localPath.
	Get(ctx context.Context, oid, localPath string) error
	// Delete removes the object oid from the remote.
	Delete(ctx context.Context, oid string) error
}
