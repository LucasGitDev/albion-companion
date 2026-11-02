// Package events decodes protocol.Params into typed Albion events.
//
// Field indices come from AlbionOnline-StatisticsAnalysis (see docs/protocol.md
// for the source file of each struct). Decoders are pure functions: no state,
// no I/O — every one of them must have a table test in events_test.go.
package events

import (
	"fmt"
	"math"
	"strings"

	"github.com/LucasGitDev/albion-companion/internal/protocol"
)

// Message names as they appear in codes.json (generated from SA's enums).
const (
	OpJoin = "Join" // response

	// OpInventoryMoveGivenItems is the "take all / move these items" request
	// (client→server) used when looting a dungeon loot chest via the open-chest
	// UI, which sends no per-item InventoryPutItem.
	OpInventoryMoveGivenItems = "InventoryMoveGivenItems"

	// OpInventoryMoveItem is the single-item slot-based move request (taking one
	// item out of a loot body / chest opened as a container).
	OpInventoryMoveItem = "InventoryMoveItem"

	EvLeave             = "Leave"
	EvNewCharacter      = "NewCharacter"
	EvHealthUpdate      = "HealthUpdate"
	EvHealthUpdates     = "HealthUpdates"
	EvOtherGrabbedLoot  = "OtherGrabbedLoot"
	EvNewLoot           = "NewLoot"
	EvPartyJoined       = "PartyJoined"
	EvPartyPlayerJoined = "PartyPlayerJoined"
	EvPartyPlayerLeft   = "PartyPlayerLeft"
	EvPartyDisbanded    = "PartyDisbanded"
	EvUpdateFame        = "UpdateFame"

	EvUpdateReSpecPoints    = "UpdateReSpecPoints"
	EvMightAndFavorReceived = "MightAndFavorReceivedEvent"

	EvNewLootChest        = "NewLootChest"
	EvNewSimpleItem       = "NewSimpleItem"
	EvNewEquipmentItem    = "NewEquipmentItem"
	EvAttachItemContainer = "AttachItemContainer"
	EvInventoryPutItem    = "InventoryPutItem"

	EvDied         = "Died"
	EvKilledPlayer = "KilledPlayer"

	EvHarvestFinished = "HarvestFinished"
)

// fixPoint converts Albion's FixPoint internal integer to its real value
// (SA: FixPoint.DoubleValue = InternalValue / 10000).
func fixPoint(internal int64) float64 { return float64(internal) / 10000 }

// MissingFieldError means a required parameter was absent or had an unexpected
// type — usually a sign that a patch changed the event layout.
type MissingFieldError struct {
	Event string
	Field string
	Key   byte
}

func (e *MissingFieldError) Error() string {
	return fmt.Sprintf("%s: missing/invalid field %s (param %d)", e.Event, e.Field, e.Key)
}

func missing(ev, field string, key byte) error {
	return &MissingFieldError{Event: ev, Field: field, Key: key}
}

// ---------------------------------------------------------------------------
// Join (operation response) — the local player entering a zone.
// SA: Network/Operations/Responses/JoinResponse.cs

type JoinResponse struct {
	ObjectID      int64
	GUID          string
	Name          string
	Location      string // map index / cluster id
	Guild         string
	Alliance      string
	InventoryGUID string // param 54: the local player's inventory container GUID
}

func DecodeJoinResponse(p protocol.Params) (JoinResponse, error) {
	var r JoinResponse
	var ok bool
	if r.ObjectID, ok = p.Int64(0); !ok {
		return r, missing(OpJoin, "ObjectID", 0)
	}
	if r.Name, ok = p.String(2); !ok {
		return r, missing(OpJoin, "Name", 2)
	}
	r.GUID, _ = p.GUID(1)
	r.Location, _ = p.String(8)
	r.Guild, _ = p.String(58)
	r.Alliance, _ = p.String(79)
	r.InventoryGUID, _ = p.GUID(54)
	return r, nil
}

// ---------------------------------------------------------------------------
// NewCharacter — another player became visible.
// SA: Network/Events/NewCharacterEvent.cs

type NewCharacter struct {
	ObjectID int64
	Name     string
	GUID     string
	Guild    string
	Alliance string
}

func DecodeNewCharacter(p protocol.Params) (NewCharacter, error) {
	var r NewCharacter
	var ok bool
	if r.ObjectID, ok = p.Int64(0); !ok {
		return r, missing(EvNewCharacter, "ObjectID", 0)
	}
	if r.Name, ok = p.String(1); !ok {
		return r, missing(EvNewCharacter, "Name", 1)
	}
	r.GUID, _ = p.GUID(7)
	r.Guild, _ = p.String(8)
	r.Alliance, _ = p.String(51)
	return r, nil
}

// ---------------------------------------------------------------------------
// Leave — an object left the visible area.
// SA: Network/Events/LeaveEvent.cs

type Leave struct{ ObjectID int64 }

func DecodeLeave(p protocol.Params) (Leave, error) {
	id, ok := p.Int64(0)
	if !ok {
		return Leave{}, missing(EvLeave, "ObjectID", 0)
	}
	return Leave{ObjectID: id}, nil
}

// ---------------------------------------------------------------------------
// OtherGrabbedLoot — someone (incl. the local player) took loot from a body/chest.
// SA: Network/Events/GrabbedLootEvent.cs
//   item:   map[0:424 1:Triky313 2:Bruno313 4:1841 5:1]
//   silver: map[0:6436 2:Triky313 3:true 5:1550115]

type GrabbedLoot struct {
	LootedFrom string // body/chest owner; contains "@MOB" for mobs
	Looter     string
	IsSilver   bool
	ItemIndex  int // index into ao-bin-dumps formatted/items.txt
	Quantity   int64
}

// IsFromMob reports whether the loot came from a mob (SA: MobController.IsMob).
func (g GrabbedLoot) IsFromMob() bool { return strings.Contains(g.LootedFrom, "@MOB") }

// SilverAmount converts the silver quantity from Albion's FixPoint (÷10000).
// Kept EXACT (fractional), matching StatisticsAnalysis's FixPoint accounting: the
// game accumulates the fractional drops internally, so the true silver gained is
// the sum of the exact values, floored once for display (floor(7.83 + 31.17) =
// 39, not floor(7.83) + floor(31.17) = 38). Per-drop display still floors via the
// UI's fmtSilver / the CLI, matching the floored "+7" combat text.
func (g GrabbedLoot) SilverAmount() float64 { return float64(g.Quantity) / 10000 }

func DecodeGrabbedLoot(p protocol.Params) (GrabbedLoot, error) {
	var r GrabbedLoot
	var ok bool
	if r.Looter, ok = p.String(2); !ok {
		return r, missing(EvOtherGrabbedLoot, "Looter", 2)
	}
	r.LootedFrom, _ = p.String(1)
	r.IsSilver, _ = p.Bool(3)
	if !r.IsSilver {
		idx, ok := p.Int64(4)
		if !ok {
			return r, missing(EvOtherGrabbedLoot, "ItemIndex", 4)
		}
		r.ItemIndex = int(idx)
	}
	if r.Quantity, ok = p.Int64(5); !ok {
		r.Quantity = 1
	}
	return r, nil
}

// ---------------------------------------------------------------------------
// NewLoot — a lootable body/bag appeared.
// SA: Network/Events/NewLootEvent.cs

type NewLoot struct {
	ObjectID int64
	LootBody string
}

func DecodeNewLoot(p protocol.Params) (NewLoot, error) {
	id, ok := p.Int64(0)
	if !ok {
		return NewLoot{}, missing(EvNewLoot, "ObjectID", 0)
	}
	body, _ := p.String(3)
	return NewLoot{ObjectID: id, LootBody: body}, nil
}

// ---------------------------------------------------------------------------
// HealthUpdate(s) — damage/heal. Negative Change = damage, positive = heal.
// SA: Network/Events/HealthUpdateEvent.cs, HealthUpdatesEvent.cs

type EffectType uint8 // 0 None, 1 Physical, 2 Magic

type EffectOrigin uint8 // 0 Melee, 1 Ranged, 2 SpellDirect, 3 SpellArea, 4 SpellPassive, 5 OverTime, 6 Reflected, 7 SpellCost, 8 ServerAuthority, 9 Unknown

type HealthUpdate struct {
	AffectedID   int64
	CauserID     int64
	Change       float64
	NewHealth    float64
	HasNewHealth bool
	EffectType   EffectType
	EffectOrigin EffectOrigin
	SpellIndex   int
}

func (h HealthUpdate) IsDamage() bool { return h.Change < 0 }

// Amount is the rounded absolute value of Change (SA rounds away from zero).
func (h HealthUpdate) Amount() int64 { return int64(math.Round(math.Abs(h.Change))) }

func DecodeHealthUpdate(p protocol.Params) (HealthUpdate, error) {
	var r HealthUpdate
	var ok bool
	if r.AffectedID, ok = p.Int64(0); !ok {
		return r, missing(EvHealthUpdate, "AffectedID", 0)
	}
	// Change (2) is optional: when the delta is zero the serializer omits it
	// (health syncs, non-damaging effects). SA reads it via TryGetValue → 0.
	r.Change, _ = p.Float64(2)
	r.NewHealth, r.HasNewHealth = p.Float64(3)
	if v, ok := p.Int64(4); ok {
		r.EffectType = EffectType(v)
	}
	r.EffectOrigin = 9
	if v, ok := p.Int64(5); ok {
		r.EffectOrigin = EffectOrigin(v)
	}
	// CauserID (6) is optional: server/environment damage (EffectOrigin
	// ServerAuthority = 8) carries no causer. SA reads it via TryGetValue → 0.
	r.CauserID, _ = p.Int64(6)
	if v, ok := p.Int64(7); ok {
		r.SpellIndex = int(v)
	}
	return r, nil
}

// DecodeHealthUpdates expands the batched form (parallel arrays) into
// individual updates that all share AffectedID.
func DecodeHealthUpdates(p protocol.Params) ([]HealthUpdate, error) {
	affected, ok := p.Int64(0)
	if !ok {
		return nil, missing(EvHealthUpdates, "AffectedID", 0)
	}
	changes, ok := p.Float64s(2)
	if !ok {
		return nil, missing(EvHealthUpdates, "Changes", 2)
	}
	causers, ok := p.Int64s(6)
	if !ok || len(causers) != len(changes) {
		return nil, missing(EvHealthUpdates, "CauserIDs", 6)
	}
	newVals, _ := p.Float64s(3)
	types, _ := p.Int64s(4)
	origins, _ := p.Int64s(5)
	spells, _ := p.Int64s(7)

	out := make([]HealthUpdate, len(changes))
	for i := range changes {
		u := HealthUpdate{AffectedID: affected, CauserID: causers[i], Change: changes[i], EffectOrigin: 9}
		if i < len(newVals) {
			u.NewHealth, u.HasNewHealth = newVals[i], true
		}
		if i < len(types) {
			u.EffectType = EffectType(types[i])
		}
		if i < len(origins) {
			u.EffectOrigin = EffectOrigin(origins[i])
		}
		if i < len(spells) {
			u.SpellIndex = int(spells[i])
		}
		out[i] = u
	}
	return out, nil
}

// ---------------------------------------------------------------------------
// Party.
// SA: Network/Events/PartyJoinedEvent.cs, PartyPlayerJoinedEvent.cs,
//     PartyPlayerLeftEvent.cs, PartyDisbandedEvent.cs

type PartyMember struct {
	GUID string
	Name string
}

type PartyJoined struct{ Members []PartyMember }

func DecodePartyJoined(p protocol.Params) (PartyJoined, error) {
	// SA returns an empty party when 8/9 are absent (no error). Param 8 is a
	// flat concatenation of 16-byte GUIDs; param 9 the parallel name list.
	if !p.Has(8) || !p.Has(9) {
		return PartyJoined{}, nil
	}
	guids, ok := p.GUIDs(8)
	if !ok {
		return PartyJoined{}, missing(EvPartyJoined, "GUIDs", 8)
	}
	names, ok := p.Strings(9)
	if !ok {
		return PartyJoined{}, missing(EvPartyJoined, "Names", 9)
	}
	// SA pairs by the shorter of the two (Math.Min).
	n := min(len(guids), len(names))
	r := PartyJoined{Members: make([]PartyMember, 0, n)}
	for i := range n {
		r.Members = append(r.Members, PartyMember{GUID: guids[i], Name: names[i]})
	}
	return r, nil
}

func DecodePartyPlayerJoined(p protocol.Params) (PartyMember, error) {
	g, ok := p.GUID(1)
	if !ok {
		return PartyMember{}, missing(EvPartyPlayerJoined, "GUID", 1)
	}
	n, ok := p.String(2)
	if !ok || n == "" {
		return PartyMember{}, missing(EvPartyPlayerJoined, "Name", 2)
	}
	return PartyMember{GUID: g, Name: n}, nil
}

type PartyPlayerLeft struct{ GUID string }

func DecodePartyPlayerLeft(p protocol.Params) (PartyPlayerLeft, error) {
	g, ok := p.GUID(1)
	if !ok {
		return PartyPlayerLeft{}, missing(EvPartyPlayerLeft, "GUID", 1)
	}
	return PartyPlayerLeft{GUID: g}, nil
}

// ---------------------------------------------------------------------------
// UpdateFame — fame gained (progression). Albion revolves around fame, so this
// is a first-class metric alongside loot and damage.
// SA: Network/Events/UpdateFameEvent.cs
//   map[0:ObjectId 1:TotalPlayerFame 2:fameWithZoneMultiplier 3:ZoneFame
//       5:IsPremiumBonus 10:SatchelFame 17:bonusFactor(float)]
// All fame params are FixPoint (÷10000).

type FameUpdate struct {
	PlayerObjectID  int64   // param 0: subject of the fame gain
	TotalPlayerFame float64 // param 1: lifetime player fame
	Base            float64 // param 2: fame with zone multiplier, without premium
	Zone            float64 // param 3: zone fame
	IsPremium       bool    // param 5: premium bonus active
	Satchel         float64 // param 10: satchel (bag insight) fame
	BonusFactor     float64 // derived from param 17 (SA resets to 1 when active)

	// Yellow is the "Quick-Learner" bonus fame (combat-log: "Fama bônus do
	// Aprendizado Rápido"), earned from a maxed 100/100 node for a Learning-Points
	// cost. It appears in the gathering shape of UpdateFame as param 12 (SA does
	// not decode it). Empirically param12 ≈ 6.9×Base (node-specific) and the cost
	// (param 14) equals −UpdateLearningPoints(83) param 5 at the same timestamp.
	// We treat param 12 as the bonus alone, additive to Base (the combat log shows
	// the two as separate "+Fama" lines).
	Yellow float64 // param 12
	LPCost float64 // param 14: Learning Points spent on the yellow bonus
}

// Premium is the extra fame from the premium account bonus: a premium account
// earns 1.5× the dropped fame, so the bonus is half the base (SA: PremiumFame).
// Note: this is NOT the "yellow" Quick-Learner bonus (fame from a maxed 100/100
// node), which is a separate mechanic not carried by this event.
func (f FameUpdate) Premium() float64 {
	if f.IsPremium {
		return f.Base * 0.5
	}
	return 0
}

// Gained is the total fame credited by this event: the SA TotalGainedFame
// (base + premium + satchel) plus the Quick-Learner yellow bonus.
func (f FameUpdate) Gained() float64 {
	return (f.Base+f.Premium()+f.Satchel)*f.BonusFactor + f.Yellow
}

func DecodeUpdateFame(p protocol.Params) (FameUpdate, error) {
	var r FameUpdate
	r.PlayerObjectID, _ = p.Int64(0)

	total, ok := p.Int64(1)
	if !ok {
		return r, missing(EvUpdateFame, "TotalPlayerFame", 1)
	}
	r.TotalPlayerFame = fixPoint(total)

	if v, ok := p.Int64(2); ok {
		r.Base = fixPoint(v)
	}
	if v, ok := p.Int64(3); ok {
		r.Zone = fixPoint(v)
	}
	r.IsPremium, _ = p.Bool(5)
	if v, ok := p.Int64(10); ok {
		r.Satchel = fixPoint(v)
	}
	if v, ok := p.Int64(12); ok { // Quick-Learner yellow bonus (gathering shape)
		r.Yellow = fixPoint(v)
	}
	if v, ok := p.Int64(14); ok {
		r.LPCost = fixPoint(v)
	}

	// param 17 is an informational bonus factor. SA computes 1+param17 but resets
	// it to 1 whenever it is positive, so it never scales the gained total; a
	// negative value is not a real scenario and must not deflate fame. We keep it
	// at 1 unconditionally — the gained total is Base + Premium + Satchel.
	r.BonusFactor = 1
	return r, nil
}

// ---------------------------------------------------------------------------
// UpdateReSpecPoints — respec (Learning Points refund) points gained, a local
// player-only progression event (no object id). SA accumulates GainedReSpecPoints
// per event for the respec/hour live stat.
// SA: Network/Events/UpdateReSpecPointsEvent.cs
//   param 0: int/long array, index 1 = CurrentTotalReSpecPoints (optional)
//   param 2: GainedReSpecPoints
//   param 3: PaidSilver
// All values are FixPoint (÷10000).

type ReSpecPoints struct {
	CurrentTotal    float64 // param 0, array index 1 (lifetime total)
	HasCurrentTotal bool
	Gained          float64 // param 2: respec points gained by this event
	PaidSilver      float64 // param 3: silver paid for the respec
}

func DecodeUpdateReSpecPoints(p protocol.Params) (ReSpecPoints, error) {
	var r ReSpecPoints
	gained, ok := p.Int64(2)
	if !ok {
		return r, missing(EvUpdateReSpecPoints, "GainedReSpecPoints", 2)
	}
	r.Gained = fixPoint(gained)
	if v, ok := p.Int64(3); ok {
		r.PaidSilver = fixPoint(v)
	}
	// param 0 is an int/long array whose index 1 holds the lifetime total
	// (SA switches on Int32[]/Int64[] then reads index 1).
	if arr, ok := p.Int64s(0); ok && len(arr) > 1 {
		r.CurrentTotal = fixPoint(arr[1])
		r.HasCurrentTotal = true
	}
	return r, nil
}

// ---------------------------------------------------------------------------
// MightAndFavorReceived — Might & Favor (daily/activity progression) gained.
// Local player-only (no object id). SA accumulates Might (param 1) and Favor
// (param 4) per event for the might/hour live stat; the bonus/premium splits
// are informational.
// SA: Network/Events/MightAndFavorReceivedEvent.cs
//   1 Might, 2 BonusOfMight, 3 PremiumOfMight,
//   4 Favor, 5 BonusOfFavor, 6 PremiumOfFavor, 7 TotalFavor
// All values are FixPoint (÷10000).

type MightAndFavor struct {
	Might          float64 // param 1
	BonusOfMight   float64 // param 2
	PremiumOfMight float64 // param 3
	Favor          float64 // param 4
	BonusOfFavor   float64 // param 5
	PremiumOfFavor float64 // param 6
	TotalFavor     float64 // param 7
}

func DecodeMightAndFavorReceived(p protocol.Params) (MightAndFavor, error) {
	var r MightAndFavor
	// Might (1) is optional: the serializer omits it when the event carries only
	// favor (zero might). SA reads every field via TryGetValue.
	if v, ok := p.Int64(1); ok {
		r.Might = fixPoint(v)
	}
	if v, ok := p.Int64(2); ok {
		r.BonusOfMight = fixPoint(v)
	}
	if v, ok := p.Int64(3); ok {
		r.PremiumOfMight = fixPoint(v)
	}
	if v, ok := p.Int64(4); ok {
		r.Favor = fixPoint(v)
	}
	if v, ok := p.Int64(5); ok {
		r.BonusOfFavor = fixPoint(v)
	}
	if v, ok := p.Int64(6); ok {
		r.PremiumOfFavor = fixPoint(v)
	}
	if v, ok := p.Int64(7); ok {
		r.TotalFavor = fixPoint(v)
	}
	return r, nil
}

// ---------------------------------------------------------------------------
// Chest loot. Taking an item from a dungeon/treasure chest is NOT an
// OtherGrabbedLoot event; it is reconstructed by stitching four events by
// object id (verified against a real capture, cross-read with SA and
// ao-loot-logger since both leave the InventoryPutItem path dormant):
//
//   NewLootChest        → a chest object + its unique name
//   NewSimpleItem/Equip → each item object → item index + quantity
//   AttachItemContainer → a container's slots hold those item objects
//   InventoryPutItem    → an item object enters the local inventory (the take)
//
// SA: Network/Events/{NewLootChestEvent,AttachItemContainerEvent,
//     InventoryPutItemEvent}.cs; items: ao-loot-logger ev-new-simple-item /
//     ev-new-equipment-item. Object ids are per-zone (reset on Join).

// LootChest is a lootable chest becoming visible (SA: NewLootChestEvent.cs).
type LootChest struct {
	ObjectID   int64  // param 0 (per-zone)
	UniqueName string // param 3, e.g. "TREASURE_COFFER_SOLO"
}

func DecodeNewLootChest(p protocol.Params) (LootChest, error) {
	id, ok := p.Int64(0)
	if !ok {
		return LootChest{}, missing(EvNewLootChest, "ObjectID", 0)
	}
	name, _ := p.String(3)
	return LootChest{ObjectID: id, UniqueName: name}, nil
}

// ItemObject declares an item instance (NewSimpleItem / NewEquipmentItem).
// ao-loot-logger ev-new-simple-item / ev-new-equipment-item: 0 objectId,
// 1 itemNumId (items.txt line index), 2 quantity.
type ItemObject struct {
	ObjectID  int64
	ItemIndex int
	Quantity  int64
}

func DecodeItemObject(ev string, p protocol.Params) (ItemObject, error) {
	id, ok := p.Int64(0)
	if !ok {
		return ItemObject{}, missing(ev, "ObjectID", 0)
	}
	idx, ok := p.Int64(1)
	if !ok {
		return ItemObject{}, missing(ev, "ItemIndex", 1)
	}
	qty, ok := p.Int64(2)
	if !ok {
		qty = 1
	}
	return ItemObject{ObjectID: id, ItemIndex: int(idx), Quantity: qty}, nil
}

// ItemContainer links a container to the item objects in its slots
// (SA: AttachItemContainerEvent.cs: 0 objectId, 3 slot array of item object ids).
type ItemContainer struct {
	ObjectID int64
	GUID     string  // param 1: the container's GUID (matches InventoryMoveItem's ContainerGuid)
	Slots    []int64 // item object ids; 0 = empty slot
}

func DecodeAttachItemContainer(p protocol.Params) (ItemContainer, error) {
	id, ok := p.Int64(0)
	if !ok {
		return ItemContainer{}, missing(EvAttachItemContainer, "ObjectID", 0)
	}
	slots, ok := p.Int64s(3)
	if !ok {
		return ItemContainer{}, missing(EvAttachItemContainer, "Slots", 3)
	}
	guid, _ := p.GUID(1)
	return ItemContainer{ObjectID: id, GUID: guid, Slots: slots}, nil
}

// InventoryMove is a slot-based move the local player makes between a container
// and their inventory (SA: InventoryMoveItemRequest.cs — 0 ContainerSlot,
// 1 ContainerGuid, 3 InventorySlot, 4 UserInteractGuid). Taking from a loot
// body / chest shows up here as a move out of the container slot.
type InventoryMove struct {
	ContainerSlot int
	ContainerGUID string
	UserGUID      string
}

func DecodeInventoryMoveItem(p protocol.Params) (InventoryMove, error) {
	guid, ok := p.GUID(1)
	if !ok {
		return InventoryMove{}, missing(OpInventoryMoveItem, "ContainerGuid", 1)
	}
	slot, _ := p.Int64(0)
	user, _ := p.GUID(4)
	return InventoryMove{ContainerSlot: int(slot), ContainerGUID: guid, UserGUID: user}, nil
}

// InventoryPut is an item entering an inventory (SA: InventoryPutItemEvent.cs:
// 0 item objectId, 2 destination inventory GUID). It is a client-local event.
type InventoryPut struct {
	ItemObjectID  int64
	InventoryGUID string // param 2; matches JoinResponse.InventoryGUID for the local player
}

func DecodeInventoryPutItem(p protocol.Params) (InventoryPut, error) {
	id, ok := p.Int64(0)
	if !ok {
		return InventoryPut{}, missing(EvInventoryPutItem, "ItemObjectID", 0)
	}
	guid, _ := p.GUID(2)
	return InventoryPut{ItemObjectID: id, InventoryGUID: guid}, nil
}

// InventoryMoveGivenItems is the batch "move these items" request the client
// sends when looting a dungeon chest via the open-chest UI (take-all), which
// never emits per-item InventoryPutItem.
// SA: Network/Operations/Request/InventoryMoveGivenItemsRequest.cs —
//
//	0 containerGuid (source), 2 userInteractGuid (destination inventory),
//	4 itemObjectIds (array), 5 quantities (array).
type InventoryMoveGivenItems struct {
	ContainerGUID string
	ItemObjectIDs []int64
}

func DecodeInventoryMoveGivenItems(p protocol.Params) (InventoryMoveGivenItems, error) {
	ids, ok := p.Int64s(4)
	if !ok {
		return InventoryMoveGivenItems{}, missing(OpInventoryMoveGivenItems, "ItemObjectIds", 4)
	}
	guid, _ := p.GUID(0)
	return InventoryMoveGivenItems{ContainerGUID: guid, ItemObjectIDs: ids}, nil
}

// ---------------------------------------------------------------------------
// Killfeed — nearby deaths and kills, broadcast for entities around the player.
// SA: Network/Events/DiedEvent.cs, KilledPlayerEvent.cs

// Died is the rich, broadcast death event (SA DiedEvent.cs): who died, who
// killed them, both guilds, and whether it was the lethal (final) blow.
type Died struct {
	VictimID    int64
	Victim      string // param 2
	VictimGuild string // param 3
	KillerID    int64
	Killer      string // param 10
	KillerGuild string // param 11
	Lethal      bool   // param 17
}

func DecodeDied(p protocol.Params) (Died, error) {
	var r Died
	if _, ok := p.String(2); !ok {
		return r, missing(EvDied, "Victim", 2)
	}
	r.Victim, _ = p.String(2)
	r.VictimID, _ = p.Int64(1)
	r.VictimGuild, _ = p.String(3)
	r.KillerID, _ = p.Int64(9)
	r.Killer, _ = p.String(10)
	r.KillerGuild, _ = p.String(11)
	r.Lethal, _ = p.Bool(17)
	return r, nil
}

// KilledPlayer credits the local player with a killing blow (SA
// KilledPlayerEvent.cs): 0 killer object id, 1 killed object id, 2 killed name.
type KilledPlayer struct {
	KillerID int64
	VictimID int64
	Victim   string
}

func DecodeKilledPlayer(p protocol.Params) (KilledPlayer, error) {
	var r KilledPlayer
	r.KillerID, _ = p.Int64(0)
	id, ok := p.Int64(1)
	if !ok {
		return r, missing(EvKilledPlayer, "VictimID", 1)
	}
	r.VictimID = id
	r.Victim, _ = p.String(2)
	return r, nil
}

// ---------------------------------------------------------------------------
// HarvestFinished — a resource was gathered (fiber/ore/hide/wood/stone/fish).
// SA: Network/Events/HarvestFinishedEvent.cs
//   0 userObjectId, 3 node objectId, 4 itemId (items.txt index),
//   5 standardAmount, 6 collectorBonusAmount, 7 premiumBonusAmount.

type Harvest struct {
	UserObjectID int64
	ItemIndex    int
	Standard     int64
	Collector    int64
	Premium      int64
}

// Amount is the total gathered (base + collector + premium bonuses).
func (h Harvest) Amount() int64 { return h.Standard + h.Collector + h.Premium }

func DecodeHarvestFinished(p protocol.Params) (Harvest, error) {
	var r Harvest
	r.UserObjectID, _ = p.Int64(0)
	idx, ok := p.Int64(4)
	if !ok {
		return r, missing(EvHarvestFinished, "ItemIndex", 4)
	}
	r.ItemIndex = int(idx)
	r.Standard, _ = p.Int64(5)
	r.Collector, _ = p.Int64(6)
	r.Premium, _ = p.Int64(7)
	return r, nil
}
