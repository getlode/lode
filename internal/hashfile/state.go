package hashfile

import (
	"encoding/binary"
	"os"
	"time"

	bolt "go.etcd.io/bbolt"
)

var stateBucket = []byte("hashes")

// State is a local cache mapping a file path to its content hash, keyed by
// (inode, mtime, size). It lets status/add skip re-hashing unchanged files —
// the single largest performance win. It is private to lode (it does not
// interoperate with DVC-Python's diskcache) and lives under .dvc/tmp/lode/.
type State struct {
	db *bolt.DB
}

// OpenState opens (creating if needed) the state DB at path.
func OpenState(path string) (*State, error) {
	if err := os.MkdirAll(dirOf(path), 0o755); err != nil {
		return nil, err
	}
	db, err := bolt.Open(path, 0o600, &bolt.Options{Timeout: 5 * time.Second})
	if err != nil {
		return nil, err
	}
	if err := db.Update(func(tx *bolt.Tx) error {
		_, e := tx.CreateBucketIfNotExists(stateBucket)
		return e
	}); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &State{db: db}, nil
}

func (s *State) Close() error { return s.db.Close() }

// Get returns the cached hash for path if its (inode, mtime, size) match the
// stored entry.
func (s *State) Get(path string) (md5sum string, size int64, ok bool) {
	fi, err := os.Stat(path)
	if err != nil {
		return "", 0, false
	}
	ino := inodeOf(fi)
	mtime := fi.ModTime().UnixNano()
	sz := fi.Size()

	_ = s.db.View(func(tx *bolt.Tx) error {
		v := tx.Bucket(stateBucket).Get([]byte(path))
		if len(v) < 24 {
			return nil
		}
		if binary.LittleEndian.Uint64(v[0:8]) == ino &&
			int64(binary.LittleEndian.Uint64(v[8:16])) == mtime &&
			int64(binary.LittleEndian.Uint64(v[16:24])) == sz {
			md5sum = string(v[24:])
			size = sz
			ok = true
		}
		return nil
	})
	return md5sum, size, ok
}

// Put records the hash for path along with its current (inode, mtime, size).
func (s *State) Put(path, md5sum string, size int64) error {
	return s.PutMany([]Entry{{Path: path, MD5: md5sum, Size: size}})
}

// Entry is a single state record to persist.
type Entry struct {
	Path string
	MD5  string
	Size int64
}

// PutMany records many entries in a single transaction. Crucial for
// performance: one bbolt commit (one fsync) instead of one per file.
func (s *State) PutMany(entries []Entry) error {
	if len(entries) == 0 {
		return nil
	}
	return s.db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(stateBucket)
		for _, e := range entries {
			fi, err := os.Stat(e.Path)
			if err != nil {
				continue
			}
			buf := make([]byte, 24+len(e.MD5))
			binary.LittleEndian.PutUint64(buf[0:8], inodeOf(fi))
			binary.LittleEndian.PutUint64(buf[8:16], uint64(fi.ModTime().UnixNano()))
			binary.LittleEndian.PutUint64(buf[16:24], uint64(e.Size))
			copy(buf[24:], e.MD5)
			if err := b.Put([]byte(e.Path), buf); err != nil {
				return err
			}
		}
		return nil
	})
}

func dirOf(p string) string {
	for i := len(p) - 1; i >= 0; i-- {
		if p[i] == '/' {
			return p[:i]
		}
	}
	return "."
}
