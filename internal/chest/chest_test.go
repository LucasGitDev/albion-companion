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
