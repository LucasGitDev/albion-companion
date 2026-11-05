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

// Content is one item seen inside a container, with whether it was taken. It is
// recorded regardless of collection, so "what dropped" survives even when the
// local player never grabbed it.
type Content struct {
	ItemIndex int   `json:"itemIndex"`
	Quantity  int64 `json:"quantity"`
	Taken     bool  `json:"taken"`
}

// Container is a chest / loot bag and the items it held.
type Container struct {
	Name  string    `json:"name"` // unique name (chest) or body name (corpse bag)
	Items []Content `json:"items"`
}

// containerAgg accumulates a container's contents (keyed by item objectId so a
// re-announce updates in place). Not cleared on take — only marked.
type containerAgg struct {
	name  string
	items map[int64]*Content
	order []int64 // item objectId insertion order for stable output
}

// Tracker stitches chest/item/container events. Safe for concurrent use, though
// the engine drives it from the single decode goroutine.
type Tracker struct {
	mu         sync.Mutex
	chests     map[int64]string            // chest objectId → unique name
	items      map[int64]events.ItemObject // item objectId → {index, qty}
	chestItems map[int64]string            // item objectId → chest name (only items known to be in a chest)
	bySlot     map[string][]int64          // tracked-container GUID → item objectId per slot (for slot-based takes)

	contents  map[int64]*containerAgg // chest objectId → observed contents (non-consuming)
	order     []int64                 // chest objectId insertion order
	itemChest map[int64]int64         // item objectId → chest objectId (for marking takes)
}

func New() *Tracker {
	return &Tracker{
		chests:     map[int64]string{},
		items:      map[int64]events.ItemObject{},
		chestItems: map[int64]string{},
		bySlot:     map[string][]int64{},
		contents:   map[int64]*containerAgg{},
		itemChest:  map[int64]int64{},
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
	// If this item was already attached to a chest before being declared, fill
	// its recorded contents now.
	if objID, ok := t.itemChest[it.ObjectID]; ok {
		if agg := t.contents[objID]; agg != nil {
			if c := agg.items[it.ObjectID]; c != nil {
				c.ItemIndex = it.ItemIndex
				c.Quantity = it.Quantity
			}
		}
	}
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
	agg := t.contents[c.ObjectID]
	if agg == nil {
		agg = &containerAgg{name: name, items: map[int64]*Content{}}
		t.contents[c.ObjectID] = agg
		t.order = append(t.order, c.ObjectID)
	}
	for _, itemID := range c.Slots {
		if itemID == 0 {
			continue
		}
		t.chestItems[itemID] = name
		// Record the item's contents (what dropped), independent of any take.
		if _, seen := agg.items[itemID]; !seen {
			ct := &Content{}
			if it, ok := t.items[itemID]; ok {
				ct.ItemIndex = it.ItemIndex
				ct.Quantity = it.Quantity
			}
			agg.items[itemID] = ct
			agg.order = append(agg.order, itemID)
			t.itemChest[itemID] = c.ObjectID
		}
	}
	if c.GUID != "" {
		slots := make([]int64, len(c.Slots))
		copy(slots, c.Slots)
		t.bySlot[c.GUID] = slots
	}
}

// ContentsSnapshot returns every container observed this zone and the items it
// held, in observation order, marking which were taken. Items are reported even
// if nobody collected them.
func (t *Tracker) ContentsSnapshot() []Container {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]Container, 0, len(t.order))
	for _, objID := range t.order {
		agg := t.contents[objID]
		if agg == nil {
			continue
		}
		items := make([]Content, 0, len(agg.order))
		for _, itemID := range agg.order {
			if c := agg.items[itemID]; c != nil {
				items = append(items, *c)
			}
		}
		out = append(out, Container{Name: agg.name, Items: items})
	}
	return out
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
	// Mark the item taken in the contents record (do not delete — "what dropped"
	// must survive the take).
	if objID, ok := t.itemChest[itemObjectID]; ok {
		if agg := t.contents[objID]; agg != nil {
			if c := agg.items[itemObjectID]; c != nil {
				c.Taken = true
			}
		}
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
	t.contents = map[int64]*containerAgg{}
	t.order = nil
	t.itemChest = map[int64]int64{}
}
