// Package dungeon segments a session into per-dungeon records. The game has no
// "entered dungeon" event: instead the local player re-Joins into a new cluster
// (internal/game.State.Location) each time they cross a zone boundary. Albion's
// instanced content uses descriptive cluster ids (e.g. "@RANDOMDUNGEON@...",
// "...HELLGATE...", "...AVALON..."), while open-world zones, cities and
// hideouts use plain numeric map ids. The tracker opens a record on entering a
// dungeon cluster and closes it on leaving (any non-dungeon zone, or a move to
// a different dungeon), accumulating the fame/silver/loot/chests the engine
// already decodes and snapshotting the combat meter at close.
//
// It is UI-agnostic and fed entirely by the engine; it never decodes packets.
package dungeon

import (
	"strings"
	"sync"
	"time"
)

// CombatSnap is one player's combat contribution captured when a dungeon
// record closes (a projection of combat.PlayerStats, so the package stays
// decoupled from the combat meter).
type CombatSnap struct {
	Name   string  `json:"name"`
	Damage int64   `json:"damage"`
	DPS    float64 `json:"dps"`
}

// Type is a coarse dungeon category derived from the cluster id. Values are
// stable (JSON) so the UI can offer a type filter and a label/icon per type.
type Type string

const (
	TypeNone      Type = ""          // not a dungeon
	TypeRandom    Type = "random"    // solo/group random "Expedition"-style dungeons
	TypeGroup     Type = "group"     // group dungeons
	TypeCorrupted Type = "corrupted" // corrupted (1v1) dungeons
	TypeHellgate  Type = "hellgate"  // hellgates
	TypeAvalon    Type = "avalon"    // avalonian roads / dungeons
	TypeMists     Type = "mists"     // the mists
	TypeOther     Type = "other"     // matched a dungeon marker but no specific type
)

// Record is one visited dungeon. Zero ExitedAt means it is still open (the
// player is inside). Name defaults from the type and is user-renamable; Tier
// is a free user tag (e.g. "T6"). Totals are the local player's.
type Record struct {
	ID        int          `json:"id"`
	Name      string       `json:"name"`
	Cluster   string       `json:"cluster"`
	Type      Type         `json:"type"`
	Tier      string       `json:"tier"`
	EnteredAt time.Time    `json:"enteredAt"`
	ExitedAt  time.Time    `json:"exitedAt"`
	Fame      float64      `json:"fame"`
	Silver    float64      `json:"silver"`
	Loot      int64        `json:"loot"`
	Chests    int          `json:"chests"`
	Combat    []CombatSnap `json:"combat"`

	renamed bool // user set the name: keep it across re-detection
}

// Open reports whether the player is still inside this dungeon.
func (r Record) Open() bool { return r.ExitedAt.IsZero() }

// Tracker is safe for concurrent use. The engine drives the mutating calls from
// its single decode goroutine; the UI reads via Records.
type Tracker struct {
	mu      sync.Mutex
	records []*Record
	current *Record // the open dungeon, or nil
	nextID  int
}

// New returns an empty tracker.
func New() *Tracker { return &Tracker{} }

// OnZone is called on every zone change (Join) with the new cluster id. It
// closes the current dungeon when the player leaves it and opens a new record
// when they enter one. snap is called (while the record is still open) to
// capture the combat meter at close; it may be nil. It returns true when the
// set of records changed (a record opened or closed), so the engine can signal
// the UI only on real transitions.
func (t *Tracker) OnZone(cluster string, at time.Time, snap func() []CombatSnap) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	typ := Classify(cluster)

	// Re-Join into the same dungeon cluster (e.g. a loading screen): keep the
	// current record open, do not segment.
	if t.current != nil && t.current.Cluster == cluster {
		return false
	}

	changed := false
	// Leaving the current dungeon (to anywhere else) closes it.
	if t.current != nil {
		t.closeLocked(at, snap)
		changed = true
	}

	// Entering a dungeon opens a fresh record.
	if typ != TypeNone {
		t.nextID++
		r := &Record{
			ID:        t.nextID,
			Name:      defaultName(typ),
			Cluster:   cluster,
			Type:      typ,
			EnteredAt: at,
		}
		t.records = append(t.records, r)
		t.current = r
		changed = true
	}
	return changed
}

func (t *Tracker) closeLocked(at time.Time, snap func() []CombatSnap) {
	if t.current == nil {
		return
	}
	t.current.ExitedAt = at
	if snap != nil {
		t.current.Combat = snap()
	}
	t.current = nil
}

// AddFame adds fame gained by the local player to the open dungeon (no-op when
// not in a dungeon).
func (t *Tracker) AddFame(f float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.current != nil {
		t.current.Fame += f
	}
}

// AddSilver adds silver picked up by the local player to the open dungeon.
func (t *Tracker) AddSilver(s float64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.current != nil {
		t.current.Silver += s
	}
}

// AddLoot adds an item pickup (quantity) by the local player to the open
// dungeon.
func (t *Tracker) AddLoot(qty int64) {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.current != nil {
		t.current.Loot += qty
	}
}

// AddChest counts one chest seen in the open dungeon.
func (t *Tracker) AddChest() {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.current != nil {
		t.current.Chests++
	}
}

// Rename sets a user-chosen display name. Returns false if no such record.
func (t *Tracker) Rename(id int, name string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if r := t.find(id); r != nil {
		r.Name = name
		r.renamed = true
		return true
	}
	return false
}

// SetTier tags a record's tier (free text, e.g. "T6"). Returns false if no
// such record.
func (t *Tracker) SetTier(id int, tier string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if r := t.find(id); r != nil {
		r.Tier = tier
		return true
	}
	return false
}

func (t *Tracker) find(id int) *Record {
	for _, r := range t.records {
		if r.ID == id {
			return r
		}
	}
	return nil
}

// Records returns a copy of every record in visit order (oldest first). Safe to
// hand to the UI.
func (t *Tracker) Records() []Record {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]Record, len(t.records))
	for i, r := range t.records {
		out[i] = *r
	}
	return out
}

// Reset clears every record.
func (t *Tracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.records = nil
	t.current = nil
	t.nextID = 0
}

// Classify maps a cluster id to a dungeon Type, or TypeNone when the cluster is
// not a dungeon. Open-world zones, cities and hideouts use plain numeric map
// ids (e.g. "4333"); instanced dungeons use descriptive ids carrying one of the
// markers below. Matching is case-insensitive and substring-based so it is
// resilient to the surrounding GUID/suffix the game appends.
func Classify(cluster string) Type {
	c := strings.ToUpper(strings.TrimSpace(cluster))
	if c == "" {
		return TypeNone
	}
	switch {
	case strings.Contains(c, "CORRUPTED"):
		return TypeCorrupted
	case strings.Contains(c, "HELLGATE"), strings.Contains(c, "@HG@"):
		return TypeHellgate
	case strings.Contains(c, "AVALON"):
		return TypeAvalon
	case strings.Contains(c, "MISTS"), strings.Contains(c, "MIST_"):
		return TypeMists
	case strings.Contains(c, "RANDOMDUNGEON"), strings.Contains(c, "EXPEDITION"):
		return TypeRandom
	case strings.Contains(c, "GROUPDUNGEON"), strings.Contains(c, "GROUP_DUNGEON"):
		return TypeGroup
	case strings.Contains(c, "DUNGEON"):
		return TypeOther
	}
	return TypeNone
}

// defaultName is the human label shown before the user renames a dungeon.
func defaultName(t Type) string {
	switch t {
	case TypeRandom:
		return "Random Dungeon"
	case TypeGroup:
		return "Group Dungeon"
	case TypeCorrupted:
		return "Corrupted Dungeon"
	case TypeHellgate:
		return "Hellgate"
	case TypeAvalon:
		return "Avalonian Dungeon"
	case TypeMists:
		return "The Mists"
	default:
		return "Dungeon"
	}
}
