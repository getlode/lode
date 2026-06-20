package hashfile

import (
	"fmt"
	"strings"
	"testing"
)

// Reference bytes taken verbatim from DVC's documentation / source (research §2).
func TestSerializeDir_KnownVector(t *testing.T) {
	entries := []DirEntry{
		{MD5: "402e97968614f583ece3b35555971f64", RelPath: "index.jpeg"},
		{MD5: "de7371b0119f4f75f9de703c7c3bac16", RelPath: "cat.jpeg"},
	}
	want := `[{"md5": "de7371b0119f4f75f9de703c7c3bac16", "relpath": "cat.jpeg"}, {"md5": "402e97968614f583ece3b35555971f64", "relpath": "index.jpeg"}]`
	got := string(SerializeDir(entries))
	if got != want {
		t.Fatalf("serialize mismatch:\n got: %s\nwant: %s", got, want)
	}
}

func TestSerializeDir_UnicodeEscape(t *testing.T) {
	// Python json.dumps(ensure_ascii=True) escapes non-ASCII as \uXXXX.
	relpath := "café.txt" // café.txt
	entries := []DirEntry{{MD5: "00000000000000000000000000000000", RelPath: relpath}}
	want := fmt.Sprintf(`[{"md5": "00000000000000000000000000000000", "relpath": "caf\u%04x.txt"}]`, 'é')
	got := string(SerializeDir(entries))
	if got != want {
		t.Fatalf("unicode escape mismatch:\n got: %s\nwant: %s", got, want)
	}
	if strings.ContainsRune(got, 'é') {
		t.Fatalf("output must not contain raw non-ASCII rune: %s", got)
	}
}

func TestDirOID_SuffixAndLength(t *testing.T) {
	oid := DirOID([]DirEntry{{MD5: "de7371b0119f4f75f9de703c7c3bac16", RelPath: "cat.jpeg"}})
	if len(oid) != 36 || oid[32:] != ".dir" {
		t.Fatalf("unexpected dir oid: %q", oid)
	}
}
