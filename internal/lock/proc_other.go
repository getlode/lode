// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build !unix

package lock

// On non-unix platforms we cannot cheaply probe liveness; assume alive so we
// never purge a real holder's entry.
func pidAlive(pid int) bool { return true }
