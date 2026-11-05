package chest

import (
	"testing"

	"github.com/LucasGitDev/albion-companion/internal/events"
)

func TestTracker_StitchesChestPickup(t *testing.T) {
	tr := New()
	tr.NewContainer(46753, "TREASURE_COFFER_SOLO")
	tr.NewItem(events.ItemObject{ObjectID: 54194, ItemIndex: 2014, Quantity: 3})
	tr.Attach(events.ItemContainer{ObjectID: 46753, Slots: []int64{0, 54194, 0}})

	take, ok := tr.Take(54194)
	if !ok {
		t.Fatal("expected chest pickup")
	}
	if take.ChestName != "TREASURE_COFFER_SOLO" || take.ItemIndex != 2014 || take.Quantity != 3 {
		t.Fatalf("unexpected %+v", take)
	}
	// Consumed: a second take (e.g. a later move) must not double-count.
	if _, ok := tr.Take(54194); ok {
		t.Fatal("item should be consumed after the first take")
	}
}

func TestTracker_ContentsSurviveUntakenAndMarkTaken(t *testing.T) {
	tr := New()
	tr.NewContainer(100, "AVALON_ELITE_LOOTCHEST_SET_UNCOMMON")
	tr.NewItem(events.ItemObject{ObjectID: 2189, ItemIndex: 2013, Quantity: 4})
	tr.NewItem(events.ItemObject{ObjectID: 2190, ItemIndex: 2005, Quantity: 9})
	tr.Attach(events.ItemContainer{ObjectID: 100, Slots: []int64{2189, 2190, 0}})

	snap := tr.ContentsSnapshot()
	if len(snap) != 1 || snap[0].Name != "AVALON_ELITE_LOOTCHEST_SET_UNCOMMON" {
		t.Fatalf("contents = %+v", snap)
	}
	if len(snap[0].Items) != 2 {
		t.Fatalf("want 2 items, got %d", len(snap[0].Items))
	}
	// Nothing taken yet — contents report the full drop regardless.
	for _, it := range snap[0].Items {
		if it.Taken {
			t.Fatalf("no item should be taken yet: %+v", it)
		}
	}
	if snap[0].Items[0].ItemIndex != 2013 || snap[0].Items[0].Quantity != 4 {
		t.Fatalf("first item = %+v", snap[0].Items[0])
	}

	// Take one item: it stays listed but marked taken; the other stays untaken.
	if _, ok := tr.Take(2189); !ok {
		t.Fatal("expected take")
	}
	snap = tr.ContentsSnapshot()
	if len(snap[0].Items) != 2 {
		t.Fatalf("taken item must remain listed: %+v", snap[0].Items)
	}
	var taken, untaken int
	for _, it := range snap[0].Items {
		if it.Taken {
			taken++
		} else {
			untaken++
		}
	}
	if taken != 1 || untaken != 1 {
		t.Fatalf("want 1 taken / 1 untaken, got %d/%d", taken, untaken)
	}
}

// Item declared AFTER the attach still fills the recorded contents.
func TestTracker_ContentsFillWhenItemDeclaredAfterAttach(t *testing.T) {
	tr := New()
	tr.NewContainer(200, "CHEST")
	tr.Attach(events.ItemContainer{ObjectID: 200, Slots: []int64{777}})
	tr.NewItem(events.ItemObject{ObjectID: 777, ItemIndex: 2029, Quantity: 1})
	snap := tr.ContentsSnapshot()
	if len(snap) != 1 || len(snap[0].Items) != 1 || snap[0].Items[0].ItemIndex != 2029 {
		t.Fatalf("contents = %+v", snap)
	}
}

func TestTracker_ContentsClearedOnReset(t *testing.T) {
	tr := New()
	tr.NewContainer(1, "C")
	tr.NewItem(events.ItemObject{ObjectID: 9, ItemIndex: 5, Quantity: 1})
	tr.Attach(events.ItemContainer{ObjectID: 1, Slots: []int64{9}})
	tr.Reset()
	if snap := tr.ContentsSnapshot(); len(snap) != 0 {
		t.Fatalf("reset should clear contents, got %+v", snap)
	}
}

func TestTracker_IgnoresNonChestContainers(t *testing.T) {
	tr := New()
	// Container 5 is the player's own inventory — never announced as a chest.
	tr.NewItem(events.ItemObject{ObjectID: 900, ItemIndex: 1, Quantity: 1})
	tr.Attach(events.ItemContainer{ObjectID: 5, Slots: []int64{900}})
	if _, ok := tr.Take(900); ok {
		t.Fatal("inventory/crafting inserts must not be logged as chest loot")
	}
}

func TestTracker_TakeUnknownItem(t *testing.T) {
	tr := New()
	tr.NewContainer(1, "C")
	// Attach references an item we never saw declared.
	tr.Attach(events.ItemContainer{ObjectID: 1, Slots: []int64{42}})
	if _, ok := tr.Take(42); ok {
		t.Fatal("take without a declared item must return false")
	}
}

func TestTracker_Reset(t *testing.T) {
	tr := New()
	tr.NewContainer(1, "C")
	tr.NewItem(events.ItemObject{ObjectID: 2, ItemIndex: 5, Quantity: 1})
	tr.Attach(events.ItemContainer{ObjectID: 1, Slots: []int64{2}})
	tr.Reset()
	if _, ok := tr.Take(2); ok {
		t.Fatal("reset should clear chest/item state")
	}
}

func TestTakeSlot(t *testing.T) {
	tr := New()
	tr.NewContainer(269, "corpse")
	tr.NewItem(events.ItemObject{ObjectID: 247, ItemIndex: 566, Quantity: 5})
	tr.Attach(events.ItemContainer{ObjectID: 269, GUID: "g", Slots: []int64{247, 0}})

	// Empty slot, out-of-range, and unknown container → not ok.
	if _, ok := tr.TakeSlot("g", 1); ok {
		t.Fatal("empty slot should be ok=false")
	}
	if _, ok := tr.TakeSlot("g", 9); ok {
		t.Fatal("out-of-range slot should be ok=false")
	}
	if _, ok := tr.TakeSlot("unknown", 0); ok {
		t.Fatal("unknown container should be ok=false")
	}
	// Real take.
	take, ok := tr.TakeSlot("g", 0)
	if !ok || take.ChestName != "corpse" || take.ItemIndex != 566 || take.Quantity != 5 {
		t.Fatalf("take = %+v ok=%v", take, ok)
	}
	// Consumed: a second take of the same slot is empty.
	if _, ok := tr.TakeSlot("g", 0); ok {
		t.Fatal("item should be consumed after take")
	}
}
