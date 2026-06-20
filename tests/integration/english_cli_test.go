// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package integration

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// spanishMarkers are tokens that must never appear in user-facing output. They
// are Spanish words/characters that do not collide with English help text.
var spanishMarkers = []string{
	"á", "é", "í", "ó", "ú", "ñ", "¿", "¡",
	"archivo", "remoto", "objetos", "salida", "materializa", "subidos",
	"fallidos", "agregado", "eliminar", "configurado", "desconocida",
	"continuar", "cancelado", "directorio", "manifiesto", "versionado",
	"reporta", "gestiona", "trackea", "presentes",
}

func assertEnglish(t *testing.T, where, out string) {
	t.Helper()
	low := strings.ToLower(out)
	for _, m := range spanishMarkers {
		if strings.Contains(low, m) {
			t.Errorf("%s: Spanish marker %q found in output:\n%s", where, m, out)
		}
	}
}

// TestEnglishCLI_Help sweeps --help for every command and asserts no Spanish.
func TestEnglishCLI_Help(t *testing.T) {
	bin := os.Getenv("LODE_BIN")
	if bin == "" {
		t.Skip("LODE_BIN not set")
	}
	cmds := [][]string{
		{"--help"},
		{"init", "--help"},
		{"add", "--help"},
		{"status", "--help"},
		{"push", "--help"},
		{"fetch", "--help"},
		{"pull", "--help"},
		{"checkout", "--help"},
		{"gc", "--help"},
		{"doctor", "--help"},
		{"remote", "--help"},
		{"remote", "add", "--help"},
		{"remote", "modify", "--help"},
	}
	for _, args := range cmds {
		out, _ := exec.Command(bin, args...).CombinedOutput()
		assertEnglish(t, strings.Join(args, " "), string(out))
	}
}

// TestEnglishCLI_Errors checks that common error/operation paths are English.
func TestEnglishCLI_Errors(t *testing.T) {
	bin := os.Getenv("LODE_BIN")
	if bin == "" {
		t.Skip("LODE_BIN not set")
	}
	empty := t.TempDir()
	out, _ := exec.Command(bin, "add", "x").CombinedOutput() // run from cwd, but use --cd
	_ = out
	// no-repo error (use --cd to point at an empty dir)
	o1, _ := exec.Command(bin, "--cd", empty, "status").CombinedOutput()
	assertEnglish(t, "no-repo error", string(o1))

	// init + no-remote push error
	repoDir := t.TempDir()
	runTool(t, repoDir, bin, "init", "--no-scm")
	if err := os.WriteFile(repoDir+"/f", []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	runTool(t, repoDir, bin, "add", "f")
	o2, _ := exec.Command(bin, "--cd", repoDir, "push").CombinedOutput()
	assertEnglish(t, "no-remote error", string(o2))
}
