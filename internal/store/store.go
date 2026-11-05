// Package store persists finished capture sessions to disk as one JSON file per
// session. It is deliberately format-simple (no SQLite/CGO): a session is a
// small Meta header plus an opaque detail blob that the app layer owns, so the
// store stays decoupled from the view types it serializes.
package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Meta is the session header shown in the History list.
type Meta struct {
	ID            string    `json:"id"` // filename-safe start timestamp, e.g. 20060102-150405
	StartedAt     time.Time `json:"startedAt"`
	EndedAt       time.Time `json:"endedAt"`
	Zone          string    `json:"zone"`
	Me            string    `json:"me"`
	Label         string    `json:"label,omitempty"`         // user-set session name (optional)
	Type          string    `json:"type,omitempty"`          // user-set session type (avalon/group/…)
	RecordingPath string    `json:"recordingPath,omitempty"` // .aorec this session was recorded to / rebuilt from (enables replay)
	LootCount     int       `json:"lootCount"`
	Silver        float64   `json:"silver"`
	Fame          float64   `json:"fame"`
	Damage        int64     `json:"damage"`
}

// record is the on-disk shape: header + opaque detail the app layer defines.
type record struct {
	Meta   Meta            `json:"meta"`
	Detail json.RawMessage `json:"detail"`
}

// Store reads and writes session files under a single directory.
type Store struct{ dir string }

// New returns a store rooted at dir (created on first Save).
func New(dir string) *Store { return &Store{dir: dir} }

// NewID builds a filename-safe session id from a start time.
func NewID(start time.Time) string { return start.Format("20060102-150405") }

// safeID rejects ids that could escape the store directory.
func safeID(id string) error {
	if id == "" || strings.ContainsAny(id, `/\`) || strings.Contains(id, "..") {
		return fmt.Errorf("store: invalid session id %q", id)
	}
	return nil
}

func (s *Store) path(id string) string { return filepath.Join(s.dir, id+".json") }

// Save writes (or overwrites) the session file atomically.
func (s *Store) Save(m Meta, detail json.RawMessage) error {
	if err := safeID(m.ID); err != nil {
		return err
	}
	if err := os.MkdirAll(s.dir, 0o755); err != nil {
		return err
	}
	b, err := json.Marshal(record{Meta: m, Detail: detail})
	if err != nil {
		return err
	}
	tmp := s.path(m.ID) + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, s.path(m.ID))
}

// List returns every session's Meta, newest start first. A missing directory is
// not an error (no sessions yet).
func (s *Store) List() ([]Meta, error) {
	entries, err := os.ReadDir(s.dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var out []Meta
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		b, err := os.ReadFile(filepath.Join(s.dir, e.Name()))
		if err != nil {
			continue // skip unreadable/partial files rather than fail the whole list
		}
		var r record
		if json.Unmarshal(b, &r) != nil || r.Meta.ID == "" {
			continue
		}
		out = append(out, r.Meta)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].StartedAt.After(out[j].StartedAt) })
	return out, nil
}

// Load returns one session's Meta and detail blob.
func (s *Store) Load(id string) (Meta, json.RawMessage, error) {
	if err := safeID(id); err != nil {
		return Meta{}, nil, err
	}
	b, err := os.ReadFile(s.path(id))
	if err != nil {
		return Meta{}, nil, err
	}
	var r record
	if err := json.Unmarshal(b, &r); err != nil {
		return Meta{}, nil, err
	}
	return r.Meta, r.Detail, nil
}

// Delete removes one session file. A missing file is not an error.
func (s *Store) Delete(id string) error {
	if err := safeID(id); err != nil {
		return err
	}
	if err := os.Remove(s.path(id)); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
