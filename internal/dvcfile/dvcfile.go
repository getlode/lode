// Package dvcfile reads and writes .dvc files with byte-exact compatibility
// with DVC 3.x.
package dvcfile

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Out is a single tracked output inside a .dvc file.
type Out struct {
	MD5    string // 32 hex (file) or "<32hex>.dir" (directory)
	Size   int64
	NFiles *int64 // present only for directories
	Hash   string // "md5" for 3.x outputs
	Path   string
}

// IsDir reports whether the output is a directory.
func (o Out) IsDir() bool { return strings.HasSuffix(o.MD5, ".dir") }

// File is the parsed content of a .dvc file.
type File struct {
	Outs []Out
}

// Marshal serializes f to the exact byte layout DVC 3.x emits: key order
// md5, size, [nfiles], hash, path; 2-space indentation; a single trailing
// newline.
func Marshal(f *File) []byte {
	var b strings.Builder
	b.WriteString("outs:\n")
	for _, o := range f.Outs {
		fmt.Fprintf(&b, "- md5: %s\n", o.MD5)
		fmt.Fprintf(&b, "  size: %d\n", o.Size)
		if o.NFiles != nil {
			fmt.Fprintf(&b, "  nfiles: %d\n", *o.NFiles)
		}
		fmt.Fprintf(&b, "  hash: %s\n", o.Hash)
		fmt.Fprintf(&b, "  path: %s\n", o.Path)
	}
	return []byte(b.String())
}

type yamlOut struct {
	MD5    string `yaml:"md5"`
	Size   int64  `yaml:"size"`
	NFiles *int64 `yaml:"nfiles,omitempty"`
	Hash   string `yaml:"hash,omitempty"`
	Path   string `yaml:"path"`
}

type yamlFile struct {
	Outs []yamlOut `yaml:"outs"`
}

// Parse reads a .dvc file's bytes into a File. It accepts any valid YAML that
// DVC may have written (robust reading, exact writing).
func Parse(data []byte) (*File, error) {
	var yf yamlFile
	if err := yaml.Unmarshal(data, &yf); err != nil {
		return nil, err
	}
	f := &File{Outs: make([]Out, 0, len(yf.Outs))}
	for _, o := range yf.Outs {
		f.Outs = append(f.Outs, Out(o))
	}
	return f, nil
}

// Load reads and parses a .dvc file from disk.
func Load(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(data)
}

// Save writes f to path with byte-exact formatting.
func Save(path string, f *File) error {
	return os.WriteFile(path, Marshal(f), 0o644)
}
