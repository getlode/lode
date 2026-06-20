// Package checkout handles workspace materialization and .gitignore management.
package checkout

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// AddToGitignore ensures the directory containing target has a .gitignore entry
// that ignores target, matching DVC: one .gitignore per directory, entry
// "/<basename>" with POSIX separators, idempotent.
func AddToGitignore(target string) error {
	dir := filepath.Dir(target)
	entry := "/" + filepath.ToSlash(filepath.Base(target))
	giPath := filepath.Join(dir, ".gitignore")

	lines, trailingNewline, err := readLines(giPath)
	if err != nil {
		return err
	}
	for _, l := range lines {
		if l == entry {
			return nil // already ignored
		}
	}

	f, err := os.OpenFile(giPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()

	var b strings.Builder
	if len(lines) > 0 && !trailingNewline {
		b.WriteByte('\n')
	}
	b.WriteString(entry)
	b.WriteByte('\n')
	_, err = f.WriteString(b.String())
	return err
}

func readLines(path string) (lines []string, trailingNewline bool, err error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, true, nil
		}
		return nil, false, err
	}
	if len(data) > 0 {
		trailingNewline = data[len(data)-1] == '\n'
	} else {
		trailingNewline = true
	}
	sc := bufio.NewScanner(strings.NewReader(string(data)))
	for sc.Scan() {
		lines = append(lines, sc.Text())
	}
	return lines, trailingNewline, sc.Err()
}
