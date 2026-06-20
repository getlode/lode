package lock

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/gofrs/flock"
)

// ErrRWConflict is returned when a read/write request conflicts with another
// live process holding an overlapping path.
var ErrRWConflict = errors.New("path is being used by another DVC process")

type rwEntry struct {
	PID int    `json:"pid"`
	Cmd string `json:"cmd"`
}

type rwState struct {
	Read  map[string][]rwEntry `json:"read"`
	Write map[string]rwEntry   `json:"write"`
}

// RW is a fine-grained per-path lock compatible with DVC's .dvc/tmp/rwlock.
// It lets non-overlapping commands run concurrently while still blocking real
// conflicts, and matches the JSON DVC reads/writes.
type RW struct {
	jsonPath string
	read     []string
	write    []string
	pid      int
}

// AcquireRW registers read/write intents for the given paths, blocking on
// conflicts with other live processes. Editing of the JSON is guarded by
// rwlock.lock (flock), like DVC.
func AcquireRW(tmpDir, cmd string, read, write []string) (*RW, error) {
	jsonPath := filepath.Join(tmpDir, "rwlock")
	guard := flock.New(filepath.Join(tmpDir, "rwlock.lock"))
	if err := guard.Lock(); err != nil {
		return nil, err
	}
	defer guard.Unlock()

	st, err := readRW(jsonPath)
	if err != nil {
		return nil, err
	}
	purgeDead(st)

	pid := os.Getpid()
	for _, p := range write {
		if _, ok := st.Write[p]; ok {
			return nil, ErrRWConflict
		}
		if len(st.Read[p]) > 0 {
			return nil, ErrRWConflict
		}
	}
	for _, p := range read {
		if _, ok := st.Write[p]; ok {
			return nil, ErrRWConflict
		}
	}

	for _, p := range write {
		st.Write[p] = rwEntry{PID: pid, Cmd: cmd}
	}
	for _, p := range read {
		st.Read[p] = append(st.Read[p], rwEntry{PID: pid, Cmd: cmd})
	}
	if err := writeRW(jsonPath, st); err != nil {
		return nil, err
	}
	return &RW{jsonPath: jsonPath, read: read, write: write, pid: pid}, nil
}

// Release removes this process's entries from the rwlock JSON.
func (r *RW) Release() error {
	guard := flock.New(filepath.Join(filepath.Dir(r.jsonPath), "rwlock.lock"))
	if err := guard.Lock(); err != nil {
		return err
	}
	defer guard.Unlock()

	st, err := readRW(r.jsonPath)
	if err != nil {
		return err
	}
	for _, p := range r.write {
		if e, ok := st.Write[p]; ok && e.PID == r.pid {
			delete(st.Write, p)
		}
	}
	for _, p := range r.read {
		kept := st.Read[p][:0]
		for _, e := range st.Read[p] {
			if e.PID != r.pid {
				kept = append(kept, e)
			}
		}
		if len(kept) == 0 {
			delete(st.Read, p)
		} else {
			st.Read[p] = kept
		}
	}
	return writeRW(r.jsonPath, st)
}

func readRW(path string) (*rwState, error) {
	st := &rwState{Read: map[string][]rwEntry{}, Write: map[string]rwEntry{}}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return st, nil
		}
		return nil, err
	}
	if len(data) == 0 {
		return st, nil
	}
	if err := json.Unmarshal(data, st); err != nil {
		return nil, err
	}
	if st.Read == nil {
		st.Read = map[string][]rwEntry{}
	}
	if st.Write == nil {
		st.Write = map[string]rwEntry{}
	}
	return st, nil
}

func writeRW(path string, st *rwState) error {
	data, err := json.Marshal(st)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func purgeDead(st *rwState) {
	for p, e := range st.Write {
		if !pidAlive(e.PID) {
			delete(st.Write, p)
		}
	}
	for p, entries := range st.Read {
		kept := entries[:0]
		for _, e := range entries {
			if pidAlive(e.PID) {
				kept = append(kept, e)
			}
		}
		if len(kept) == 0 {
			delete(st.Read, p)
		} else {
			st.Read[p] = kept
		}
	}
}
