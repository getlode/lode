//go:build !linux

package checkout

import "errors"

// reflinkFile is unsupported off Linux; callers fall back to the next strategy
// (typically copy). macOS clonefile could be added here later.
func reflinkFile(src, dst string) error {
	return errors.New("reflink no soportado en esta plataforma")
}
