// Package chest reconstructs "items taken from a loot chest" from the four
// separate events the game sends (NewLootChest, NewSimpleItem/NewEquipmentItem,
// AttachItemContainer, InventoryPutItem), which carry no single "grabbed from
// chest" event. It stitches them by object id. All object ids are per-zone, so
// the engine calls Reset on every Join.
package chest

import (
	"sync"

	"github.com/LucasGitDev/albion-companion/internal/events"
)

// Take is a resolved chest pickup, ready to be recorded as loot.
type Take struct {
	ChestName string // the chest's unique name (lootedFrom)
	ItemIndex int    // items.txt line index
	Quantity  int64
}

// Tracker stitches chest/item/container events. Safe for concurrent use, though
// the engine drives it from the single decode goroutine.
type Tracker struct {
	mu         sync.Mutex
	chests     map[int64]string            // chest objectId → unique name
	items      map[int64]events.ItemObject // item objectId → {index, qty}
	chestItems map[int64]string            // item objectId → chest name (only items known to be in a chest)
	bySlot     map[string][]int64          // tracked-container GUID → item objectId per slot (for slot-based takes)
}

func New() *Tracker {
	return &Tracker{
		chests:     map[int64]string{},
		items:      map[int64]events.ItemObject{},
		chestItems: map[int64]string{},
		bySlot:     map[string][]int64{},
	}
}

// NewContainer records a lootable container object and the name to credit as
// "looted from" — a dungeon/treasure chest (NewLootChest → unique name) or a
// corpse bag / loot body (NewLoot → body name). Upsert; the game re-announces.
func (t *Tracker) NewContainer(objectID int64, name string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.chests[objectID] = name
}

// NewItem records an item instance declared by NewSimpleItem/NewEquipmentItem.
func (t *Tracker) NewItem(it events.ItemObject) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.items[it.ObjectID] = it
}

// Attach links a container's slots to item objects. Only containers that were
// announced as chests / loot bodies mark their items as chest-sourced, so
// ordinary inventory or equipment attaches are ignored. The slot→item mapping is
// kept (keyed by the container GUID) so a later slot-based InventoryMoveItem take
// can resolve which item left which slot.
func (t *Tracker) Attach(c events.ItemContainer) {
	t.mu.Lock()
	defer t.mu.Unlock()
	name, isChest := t.chests[c.ObjectID]
	if !isChest {
		return
	}
	for _, itemID := range c.Slots {
		if itemID != 0 {
			t.chestItems[itemID] = name
		}
	}
	if c.GUID != "" {
		slots := make([]int64, len(c.Slots))
		copy(slots, c.Slots)
		t.bySlot[c.GUID] = slots
	}
}

// TakeSlot reports the pickup when the local player moves the item in slot
// `slot` of the tracked container `guid` (a slot-based InventoryMoveItem take —
// e.g. from a loot body / chest opened as a container). Returns ok=false for an
// unknown container/slot or an empty slot. The item is consumed so a later event
// does not double-count it.
func (t *Tracker) TakeSlot(guid string, slot int) (Take, bool) {
	t.mu.Lock()
	slots, ok := t.bySlot[guid]
	if !ok || slot < 0 || slot >= len(slots) {
		t.mu.Unlock()
		return Take{}, false
	}
	itemObjID := slots[slot]
	t.mu.Unlock()
	if itemObjID == 0 {
		return Take{}, false
	}
	return t.Take(itemObjID)
}

// Take reports the pickup when an item that belongs to a chest enters the
// inventory. It returns ok=false for any item not sourced from a tracked chest
// (ordinary inventory inserts: crafting, buying, mob-body loot). The item is
// consumed so a later move does not double-count it.
func (t *Tracker) Take(itemObjectID int64) (Take, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	chestName, fromChest := t.chestItems[itemObjectID]
	if !fromChest {
		return Take{}, false
	}
	it, known := t.items[itemObjectID]
	if !known {
		return Take{}, false
	}
	delete(t.chestItems, itemObjectID)
	delete(t.items, itemObjectID)
	return Take{ChestName: chestName, ItemIndex: it.ItemIndex, Quantity: it.Quantity}, true
}

func (t *Tracker) Reset() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.chests = map[int64]string{}
	t.items = map[int64]events.ItemObject{}
	t.chestItems = map[int64]string{}
	t.bySlot = map[string][]int64{}
}
