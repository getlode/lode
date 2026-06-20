// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package checkout

import (
	"os"
	"strings"
)

// LinkType is a cache.type strategy.
type LinkType string

const (
	Reflink  LinkType = "reflink"
	Hardlink LinkType = "hardlink"
	Symlink  LinkType = "symlink"
	Copy     LinkType = "copy"
)

// ParseCacheTypes parses a cache.type value ("reflink,copy") into an ordered
// list of strategies, defaulting to DVC's default (reflink, copy).
func ParseCacheTypes(s string) []LinkType {
	if strings.TrimSpace(s) == "" {
		return []LinkType{Reflink, Copy}
	}
	var out []LinkType
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, LinkType(p))
		}
	}
	if len(out) == 0 {
		return []LinkType{Reflink, Copy}
	}
	return out
}

// link materializes the cache object src to dst using the first strategy that
// succeeds, falling back through the list and finally to a plain copy.
func link(src, dst string, types []LinkType) error {
	var lastErr error
	for _, t := range types {
		switch t {
		case Reflink:
			lastErr = reflinkFile(src, dst)
		case Hardlink:
			_ = os.Remove(dst)
			lastErr = os.Link(src, dst)
		case Symlink:
			_ = os.Remove(dst)
			lastErr = os.Symlink(src, dst)
		case Copy:
			lastErr = copyFile(src, dst)
		default:
			lastErr = copyFile(src, dst)
		}
		if lastErr == nil {
			return nil
		}
	}
	// Last resort: a plain writable copy always works.
	return copyFile(src, dst)
}
