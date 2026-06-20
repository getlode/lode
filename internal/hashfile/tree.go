package hashfile

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// DirEntry is one file inside a tracked directory's .dir manifest.
type DirEntry struct {
	MD5     string
	RelPath string
}

// SerializeDir reproduces DVC's exact byte serialization of a directory
// manifest. DVC produces it with Python's json.dumps(entries, sort_keys=True):
//
//   - entries sorted ascending by relpath
//   - keys alphabetical inside each object (md5 before relpath)
//   - default separators ", " (items) and ": " (key/value), i.e. WITH spaces
//   - ensure_ascii=True: every rune > 0x7e escaped as \uXXXX (surrogate pairs
//     for runes > 0xFFFF)
//   - no trailing newline
//
// Go's encoding/json does NOT match this (no spaces, different escaping), so we
// emit the bytes by hand. A single divergent byte changes the directory hash
// and breaks DVC compatibility.
func SerializeDir(entries []DirEntry) []byte {
	sorted := make([]DirEntry, len(entries))
	copy(sorted, entries)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].RelPath < sorted[j].RelPath
	})

	var b strings.Builder
	b.WriteByte('[')
	for i, e := range sorted {
		if i > 0 {
			b.WriteString(", ")
		}
		b.WriteString(`{"md5": `)
		writePyJSONString(&b, e.MD5)
		b.WriteString(`, "relpath": `)
		writePyJSONString(&b, e.RelPath)
		b.WriteByte('}')
	}
	b.WriteByte(']')
	return []byte(b.String())
}

// DirOID returns the object id of a directory manifest: the lowercase hex MD5
// of the serialized bytes with a ".dir" suffix, matching DVC.
func DirOID(entries []DirEntry) string {
	sum := md5.Sum(SerializeDir(entries))
	return hex.EncodeToString(sum[:]) + ".dir"
}

// ParseDir parses a serialized .dir object back into entries. The serialized
// form is valid JSON, so standard decoding suffices for reading.
func ParseDir(data []byte) ([]DirEntry, error) {
	var raw []struct {
		MD5     string `json:"md5"`
		RelPath string `json:"relpath"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	entries := make([]DirEntry, len(raw))
	for i, r := range raw {
		entries[i] = DirEntry{MD5: r.MD5, RelPath: r.RelPath}
	}
	return entries, nil
}

// writePyJSONString writes s as a JSON string escaped exactly like Python's
// json.encoder with ensure_ascii=True.
func writePyJSONString(b *strings.Builder, s string) {
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			switch {
			case r < 0x20 || r > 0x7e:
				if r > 0xffff {
					v := r - 0x10000
					hi := 0xd800 | ((v >> 10) & 0x3ff)
					lo := 0xdc00 | (v & 0x3ff)
					fmt.Fprintf(b, `\u%04x\u%04x`, hi, lo)
				} else {
					fmt.Fprintf(b, `\u%04x`, r)
				}
			default:
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
}
