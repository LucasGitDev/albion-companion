// Package loot aggregates OtherGrabbedLoot events into a session log with
// per-player and per-item views, Party Mode filtering and CSV export.
package loot

import (
	"encoding/csv"
	"io"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/LucasGitDev/albion-companion/internal/events"
	"github.com/LucasGitDev/albion-companion/internal/game"
	"github.com/LucasGitDev/albion-companion/internal/itemdb"
)

type ItemResolver interface {
	Resolve(index int) (itemdb.Item, bool)
}

type Entry struct {
	At             time.Time `json:"at"`
	Looter         string    `json:"looter"`
	LooterGuild    string    `json:"looterGuild"`
	LooterAlliance string    `json:"looterAlliance"`
	LootedFrom     string    `json:"lootedFrom"`
	FromGuild      string    `json:"fromGuild"`
	FromAlliance   string    `json:"fromAlliance"`
	FromMob        bool      `json:"fromMob"`
	IsSilver       bool      `json:"isSilver"`
	Silver         float64   `json:"silver"`
	ItemIndex      int       `json:"itemIndex"`
	ItemID         string    `json:"itemId"`
	ItemName       string    `json:"itemName"`      // English name (items.txt), used for CSV export
	ItemLocalName  string    `json:"itemLocalName"` // localized display name (PT-BR -> EN-US -> humanized)
	IsTrash        bool      `json:"isTrash"`
	Quantity       int64     `json:"quantity"`
	Location       string    `json:"location"`
	InParty        bool      `json:"inParty"` // evaluated when the loot happened
}

type Filter struct {
	PartyOnly    bool
	Player       string
	IncludeTrash bool
}

func (f Filter) match(e Entry) bool {
	if f.PartyOnly && !e.InParty {
		return false
	}
	if f.Player != "" && e.Looter != f.Player {
		return false
	}
	if e.IsTrash && !f.IncludeTrash {
		return false
	}
	return true
}

// Tracker is safe for concurrent use.
type Tracker struct {
	mu        sync.RWMutex
	items     ItemResolver
	entries   []Entry
	startedAt time.Time
}

func NewTracker(items ItemResolver) *Tracker {
	return &Tracker{items: items, startedAt: time.Now()}
}

// SetItems swaps the item resolver (itemdb loads asynchronously).
func (t *Tracker) SetItems(items ItemResolver) {
	t.mu.Lock()
	t.items = items
	t.mu.Unlock()
}

// Record converts a GrabbedLoot event into an Entry, enriched with guild/alliance
// from the world state, and appends it to the session.
func (t *Tracker) Record(ev events.GrabbedLoot, at time.Time, st *game.State) Entry {
	e := Entry{
		At:         at,
		Looter:     ev.Looter,
		LootedFrom: ev.LootedFrom,
		FromMob:    ev.IsFromMob(),
		IsSilver:   ev.IsSilver,
		ItemIndex:  ev.ItemIndex,
		Quantity:   ev.Quantity,
		Location:   st.Location(),
		InParty:    st.InParty(ev.Looter),
	}
	if ev.IsSilver {
		e.Silver = ev.SilverAmount()
		e.ItemID, e.ItemName, e.ItemLocalName = "SILVER", "Silver", "Silver"
	}
	if p, ok := st.EntityByName(ev.Looter); ok {
		e.LooterGuild, e.LooterAlliance = p.Guild, p.Alliance
	}
	if p, ok := st.EntityByName(ev.LootedFrom); ok {
		e.FromGuild, e.FromAlliance = p.Guild, p.Alliance
	}

	t.mu.Lock()
	defer t.mu.Unlock()
	if !ev.IsSilver && t.items != nil {
		if it, ok := t.items.Resolve(ev.ItemIndex); ok {
			e.ItemID, e.ItemName, e.IsTrash = it.UniqueName, it.Name, it.IsTrash()
			e.ItemLocalName = it.LocalName
			if e.ItemLocalName == "" {
				e.ItemLocalName = itemdb.LocalizedName(it.UniqueName)
			}
		}
	}
	if e.ItemID == "" {
		e.ItemID = "#" + strconv.Itoa(ev.ItemIndex)
		e.ItemName = e.ItemID
	}
	if e.ItemLocalName == "" {
		e.ItemLocalName = e.ItemName
	}
	t.entries = append(t.entries, e)
	return e
}

func (t *Tracker) Entries(f Filter) []Entry {
	t.mu.RLock()
	defer t.mu.RUnlock()
	out := make([]Entry, 0, len(t.entries))
	for _, e := range t.entries {
		if f.match(e) {
			out = append(out, e)
		}
	}
	return out
}

// StartedAt is when the current session began (set on New and reset by Reset).
// It backs the silver/hour rate so the rate is reset-safe.
func (t *Tracker) StartedAt() time.Time {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.startedAt
}

func (t *Tracker) Reset() {
	t.mu.Lock()
	t.entries = nil
	t.startedAt = time.Now()
	t.mu.Unlock()
}

type PlayerSummary struct {
	Name     string    `json:"name"`
	Guild    string    `json:"guild"`
	Items    int64     `json:"items"` // total quantity of non-silver items
	Pickups  int       `json:"pickups"`
	Silver   float64   `json:"silver"`
	InParty  bool      `json:"inParty"`
	LastLoot time.Time `json:"lastLoot"`
}

// ByPlayer returns looters sorted by item quantity (desc).
func (t *Tracker) ByPlayer(f Filter) []PlayerSummary {
	m := map[string]*PlayerSummary{}
	for _, e := range t.Entries(f) {
		s := m[e.Looter]
		if s == nil {
			s = &PlayerSummary{Name: e.Looter, Guild: e.LooterGuild}
			m[e.Looter] = s
		}
		s.Pickups++
		s.InParty = s.InParty || e.InParty
		if e.IsSilver {
			s.Silver += e.Silver
		} else {
			s.Items += e.Quantity
		}
		if e.At.After(s.LastLoot) {
			s.LastLoot = e.At
		}
	}
	out := make([]PlayerSummary, 0, len(m))
	for _, s := range m {
		out = append(out, *s)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Items != out[j].Items {
			return out[i].Items > out[j].Items
		}
		return out[i].Name < out[j].Name
	})
	return out
}

type ItemSummary struct {
	ItemID        string           `json:"itemId"`
	ItemName      string           `json:"itemName"`
	ItemLocalName string           `json:"itemLocalName"`
	Quantity      int64            `json:"quantity"`
	Looters       map[string]int64 `json:"looters"`
}

// ByItem returns non-silver items sorted by quantity (desc).
func (t *Tracker) ByItem(f Filter) []ItemSummary {
	m := map[string]*ItemSummary{}
	for _, e := range t.Entries(f) {
		if e.IsSilver {
			continue
		}
		s := m[e.ItemID]
		if s == nil {
			s = &ItemSummary{ItemID: e.ItemID, ItemName: e.ItemName, ItemLocalName: e.ItemLocalName, Looters: map[string]int64{}}
			m[e.ItemID] = s
		}
		s.Quantity += e.Quantity
		s.Looters[e.Looter] += e.Quantity
	}
	out := make([]ItemSummary, 0, len(m))
	for _, s := range m {
		out = append(out, *s)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Quantity != out[j].Quantity {
			return out[i].Quantity > out[j].Quantity
		}
		return out[i].ItemID < out[j].ItemID
	})
	return out
}

// csvHeader matches ao-loot-logger so exports open in AO Loot Logger Viewer.
var csvHeader = []string{
	"timestamp_utc", "looted_by__alliance", "looted_by__guild", "looted_by__name",
	"item_id", "item_name", "quantity",
	"looted_from__alliance", "looted_from__guild", "looted_from__name",
}

// WriteCSV exports item loot (silver excluded) in ao-loot-logger's format.
func (t *Tracker) WriteCSV(w io.Writer, f Filter) error {
	return WriteEntriesCSV(w, t.Entries(f))
}

// WriteEntriesCSV writes the given entries (silver excluded) in ao-loot-logger's
// format. Shared by the live export and the persisted-session export.
func WriteEntriesCSV(w io.Writer, entries []Entry) error {
	cw := csv.NewWriter(w)
	cw.Comma = ';'
	if err := cw.Write(csvHeader); err != nil {
		return err
	}
	for _, e := range entries {
		if e.IsSilver {
			continue
		}
		rec := []string{
			e.At.UTC().Format(time.RFC3339), e.LooterAlliance, e.LooterGuild, e.Looter,
			e.ItemID, e.ItemName, strconv.FormatInt(e.Quantity, 10),
			e.FromAlliance, e.FromGuild, e.LootedFrom,
		}
		if err := cw.Write(rec); err != nil {
			return err
		}
	}
	cw.Flush()
	return cw.Error()
}
