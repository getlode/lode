// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

//go:build !linux

package checkout

import "errors"

// reflinkFile is unsupported off Linux; callers fall back to the next strategy
// (typically copy). macOS clonefile could be added here later.
func reflinkFile(src, dst string) error {
	return errors.New("reflink no soportado en esta plataforma")
}
