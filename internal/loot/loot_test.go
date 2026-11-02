package loot

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/LucasGitDev/albion-companion/internal/events"
	"github.com/LucasGitDev/albion-companion/internal/game"
	"github.com/LucasGitDev/albion-companion/internal/itemdb"
)

type fakeItems map[int]itemdb.Item

func (f fakeItems) Resolve(i int) (itemdb.Item, bool) { it, ok := f[i]; return it, ok }

func setup() (*Tracker, *game.State) {
	st := game.NewState()
	st.ApplyJoin(events.JoinResponse{ObjectID: 1, Name: "Me", GUID: "g-me", Location: "3008", Guild: "MyGuild", Alliance: "MyAlli"})
	st.ApplyNewCharacter(events.NewCharacter{ObjectID: 2, Name: "Rat", Guild: "Thieves", Alliance: "RatAlli"})
	items := fakeItems{
		2888: {Index: 2888, UniqueName: "T4_BAG", Name: "Adept's Bag"},
		2043: {Index: 2043, UniqueName: "T1_TRASH", Name: "Trash"},
	}
	return NewTracker(items), st
}

func TestTracker_RecordAndViews(t *testing.T) {
	tr, st := setup()
	at := time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)
	tr.Record(events.GrabbedLoot{Looter: "Me", LootedFrom: "@MOB_X", ItemIndex: 2888, Quantity: 2}, at, st)
	tr.Record(events.GrabbedLoot{Looter: "Rat", LootedFrom: "Me", ItemIndex: 2888, Quantity: 1}, at, st)
	tr.Record(events.GrabbedLoot{Looter: "Me", ItemIndex: 2043, Quantity: 5}, at, st)
	tr.Record(events.GrabbedLoot{Looter: "Me", IsSilver: true, Quantity: 1500000}, at, st)

	all := tr.Entries(Filter{})
	if len(all) != 3 { // trash hidden by default
		t.Fatalf("entries = %d, want 3", len(all))
	}
	// Rat (guild Thieves / alliance RatAlli) looted Me (guild MyGuild / alliance
	// MyAlli): both sides' guild AND alliance are attached from entity state.
	if all[1].LooterGuild != "Thieves" || all[1].LooterAlliance != "RatAlli" ||
		all[1].FromGuild != "MyGuild" || all[1].FromAlliance != "MyAlli" || all[1].InParty {
		t.Fatalf("enrichment wrong: %+v", all[1])
	}

	party := tr.ByPlayer(Filter{PartyOnly: true})
	if len(party) != 1 || party[0].Name != "Me" || party[0].Items != 2 || party[0].Silver != 150 {
		t.Fatalf("ByPlayer party = %+v", party)
	}

	items := tr.ByItem(Filter{})
	if len(items) != 1 || items[0].Quantity != 3 || items[0].Looters["Rat"] != 1 {
		t.Fatalf("ByItem = %+v", items)
	}
}

func TestTracker_UnknownItemAndCSV(t *testing.T) {
	tr, st := setup()
	e := tr.Record(events.GrabbedLoot{Looter: "Me", ItemIndex: 99999, Quantity: 1}, time.Unix(0, 0), st)
	if e.ItemID != "#99999" {
		t.Fatalf("ItemID = %q", e.ItemID)
	}
	var buf bytes.Buffer
	if err := tr.WriteCSV(&buf, Filter{}); err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) != 2 || !strings.HasPrefix(lines[0], "timestamp_utc;looted_by__alliance") {
		t.Fatalf("csv:\n%s", buf.String())
	}
}

// TestSilver_ExactSumThenFloor proves silver is accumulated exactly (matching
// SA's FixPoint accounting) and floored once for the total, not floored per
// pickup — so fractional drops don't lose silver (TASK-45).
func TestSilver_ExactSumThenFloor(t *testing.T) {
	tr, st := setup()
	at := time.Unix(0, 0)
	// 7.83 + 31.17 = 39.00 exactly; per-pickup flooring would give 7+31 = 38.
	tr.Record(events.GrabbedLoot{Looter: "Me", IsSilver: true, Quantity: 78300}, at, st)
	tr.Record(events.GrabbedLoot{Looter: "Me", IsSilver: true, Quantity: 311700}, at, st)

	p := tr.ByPlayer(Filter{})
	if len(p) != 1 || p[0].Silver != 39 {
		t.Fatalf("silver total = %v want 39 (exact sum, not 38 per-pickup floor)", p)
	}
	// Each entry keeps the exact fractional value; display floors it elsewhere.
	all := tr.Entries(Filter{})
	if all[0].Silver != 7.83 {
		t.Fatalf("entry silver = %v want exact 7.83", all[0].Silver)
	}
}

// TestCSV_GuildAllianceColumns proves the ao-loot-logger columns carry both
// sides' guild AND alliance for a player-on-player loot (TASK-31), so the export
// stays drop-in compatible with the AO Loot Logger Viewer.
func TestCSV_GuildAllianceColumns(t *testing.T) {
	tr, st := setup()
	at := time.Date(2026, 10, 31, 12, 0, 0, 0, time.UTC)
	tr.Record(events.GrabbedLoot{Looter: "Rat", LootedFrom: "Me", ItemIndex: 2888, Quantity: 1}, at, st)

	var buf bytes.Buffer
	if err := tr.WriteCSV(&buf, Filter{}); err != nil {
		t.Fatal(err)
	}
	row := strings.Split(strings.TrimSpace(buf.String()), "\n")[1]
	cols := strings.Split(row, ";")
	// header: timestamp, looted_by__alliance, looted_by__guild, looted_by__name,
	// item_id, item_name, quantity, looted_from__alliance, looted_from__guild, looted_from__name
	if cols[1] != "RatAlli" || cols[2] != "Thieves" || cols[3] != "Rat" {
		t.Fatalf("looter columns = %q", cols)
	}
	if cols[7] != "MyAlli" || cols[8] != "MyGuild" || cols[9] != "Me" {
		t.Fatalf("looted-from columns = %q", cols)
	}
}
