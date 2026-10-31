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
)

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
	ObjectID int64
	GUID     string
	Name     string
	Location string // map index / cluster id
	Guild    string
	Alliance string
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

// SilverAmount converts the silver quantity from Albion's FixPoint (x10000).
// TODO(verify): confirm the scale with a real capture.
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
	if r.Change, ok = p.Float64(2); !ok {
		return r, missing(EvHealthUpdate, "Change", 2)
	}
	r.NewHealth, r.HasNewHealth = p.Float64(3)
	if v, ok := p.Int64(4); ok {
		r.EffectType = EffectType(v)
	}
	r.EffectOrigin = 9
	if v, ok := p.Int64(5); ok {
		r.EffectOrigin = EffectOrigin(v)
	}
	if r.CauserID, ok = p.Int64(6); !ok {
		return r, missing(EvHealthUpdate, "CauserID", 6)
	}
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
	guids, ok := p.GUIDs(8)
	if !ok {
		return PartyJoined{}, missing(EvPartyJoined, "GUIDs", 8)
	}
	names, ok := p.Strings(9)
	if !ok || len(names) != len(guids) {
		return PartyJoined{}, missing(EvPartyJoined, "Names", 9)
	}
	r := PartyJoined{Members: make([]PartyMember, len(names))}
	for i := range names {
		r.Members[i] = PartyMember{GUID: guids[i], Name: names[i]}
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
