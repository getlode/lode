//go:build !unix

package lock

// On non-unix platforms we cannot cheaply probe liveness; assume alive so we
// never purge a real holder's entry.
func pidAlive(pid int) bool { return true }
