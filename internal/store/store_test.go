package store

import (
	"encoding/json"
	"testing"
	"time"
)

func TestSaveListLoadDelete(t *testing.T) {
	s := New(t.TempDir())

	// Empty store: no error, no sessions.
	if ms, err := s.List(); err != nil || len(ms) != 0 {
		t.Fatalf("empty List = %v, %v", ms, err)
	}

	t0 := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	a := Meta{ID: NewID(t0), StartedAt: t0, EndedAt: t0.Add(time.Hour), Zone: "Lymhurst", Me: "L0rdCap", LootCount: 12, Silver: 5000, Fame: 99000, Damage: 42000}
	b := Meta{ID: NewID(t0.Add(2 * time.Hour)), StartedAt: t0.Add(2 * time.Hour), EndedAt: t0.Add(3 * time.Hour), Zone: "Avalon", Me: "L0rdCap"}

	if err := s.Save(a, json.RawMessage(`{"loot":[1,2,3]}`)); err != nil {
		t.Fatal(err)
	}
	if err := s.Save(b, json.RawMessage(`{"loot":[]}`)); err != nil {
		t.Fatal(err)
	}

	// List is newest-first.
	ms, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(ms) != 2 || ms[0].ID != b.ID || ms[1].ID != a.ID {
		t.Fatalf("List order wrong: %+v", ms)
	}

	// Load returns meta + detail.
	m, detail, err := s.Load(a.ID)
	if err != nil {
		t.Fatal(err)
	}
	if m.Zone != "Lymhurst" || m.LootCount != 12 || string(detail) != `{"loot":[1,2,3]}` {
		t.Fatalf("Load mismatch: %+v / %s", m, detail)
	}

	// Overwrite is idempotent on id.
	if err := s.Save(a, json.RawMessage(`{"loot":[9]}`)); err != nil {
		t.Fatal(err)
	}
	if ms, _ := s.List(); len(ms) != 2 {
		t.Fatalf("overwrite created a duplicate: %d", len(ms))
	}

	// Delete.
	if err := s.Delete(a.ID); err != nil {
		t.Fatal(err)
	}
	if _, _, err := s.Load(a.ID); err == nil {
		t.Fatal("Load after delete should fail")
	}
	if err := s.Delete(a.ID); err != nil {
		t.Fatalf("delete of missing should be nil: %v", err)
	}
}

func TestSafeID(t *testing.T) {
	s := New(t.TempDir())
	for _, bad := range []string{"", "../x", "a/b", `a\b`, "foo..bar"} {
		if err := s.Save(Meta{ID: bad}, nil); err == nil {
			t.Errorf("Save accepted unsafe id %q", bad)
		}
		if _, _, err := s.Load(bad); err == nil {
			t.Errorf("Load accepted unsafe id %q", bad)
		}
	}
}
