// Package engine wires capture → protocol → events → game/loot/combat.
//
// It is UI-agnostic: the Wails app and the CLI both drive the same Engine.
// All decoding runs on a single goroutine (the Photon parser keeps fragment
// state), while loot/combat/state are safe to read concurrently.
package engine

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/LucasGitDev/albion-companion/internal/capture"
	"github.com/LucasGitDev/albion-companion/internal/chest"
	"github.com/LucasGitDev/albion-companion/internal/combat"
	"github.com/LucasGitDev/albion-companion/internal/events"
	"github.com/LucasGitDev/albion-companion/internal/fame"
	"github.com/LucasGitDev/albion-companion/internal/game"
	"github.com/LucasGitDev/albion-companion/internal/loot"
	"github.com/LucasGitDev/albion-companion/internal/protocol"
	"github.com/LucasGitDev/albion-companion/internal/timeline"
)

// Topic tells listeners what changed, so the UI can refresh only that view.
type Topic string

const (
	TopicLoot     Topic = "loot"
	TopicCombat   Topic = "combat"
	TopicParty    Topic = "party"
	TopicZone     Topic = "zone"
	TopicFame     Topic = "fame"
	TopicTimeline Topic = "timeline"
)

type Config struct {
	Codes         *protocol.Codes   // nil = embedded defaults
	Items         loot.ItemResolver // may be nil and set later via Loot().SetItems
	Combat        combat.Options
	Recorder      *capture.Recorder            // optional: record every packet to .aorec
	OnMessage     func(protocol.Message)       // optional: raw message tap (debug/CLI dump)
	OnLoot        func(loot.Entry)             // optional
	OnChange      func(Topic)                  // optional: called on the engine goroutine; keep it cheap
	OnDecodeError func(name string, err error) // optional
	TimelineCap   int                          // max timeline events retained (0 = default)
}

// Stats are diagnostics shown in the UI/CLI.
type Stats struct {
	Packets      uint64 `json:"packets"`
	Messages     uint64 `json:"messages"`
	Unknown      uint64 `json:"unknown"` // codes not present in codes.json
	DecodeErrors uint64 `json:"decodeErrors"`
	Encrypted    uint64 `json:"encrypted"`
}

type Engine struct {
	cfg    Config
	dec    *protocol.Decoder
	state  *game.State
	loot   *loot.Tracker
	combat *combat.Meter
	fame   *fame.Tracker
	chest  *chest.Tracker
	tl     *timeline.Ring

	packets, messages, unknown, decodeErrors atomic.Uint64

	mu      sync.Mutex
	running bool
}

func New(cfg Config) *Engine {
	if cfg.Codes == nil {
		cfg.Codes = protocol.DefaultCodes()
	}
	e := &Engine{
		cfg:    cfg,
		state:  game.NewState(),
		loot:   loot.NewTracker(cfg.Items),
		combat: combat.NewMeter(cfg.Combat),
		fame:   fame.New(),
		chest:  chest.New(),
		tl:     timeline.NewRing(cfg.TimelineCap),
	}
	e.dec = protocol.NewDecoder(cfg.Codes, e.handle)
	return e
}

func (e *Engine) State() *game.State    { return e.state }
func (e *Engine) Loot() *loot.Tracker   { return e.loot }
func (e *Engine) Combat() *combat.Meter { return e.combat }
func (e *Engine) Fame() *fame.Tracker   { return e.fame }
func (e *Engine) Chest() *chest.Tracker { return e.chest }

// Timeline is the merged chronological feed of session events.
func (e *Engine) Timeline() *timeline.Ring { return e.tl }

func (e *Engine) Stats() Stats {
	return Stats{
		Packets:      e.packets.Load(),
		Messages:     e.messages.Load(),
		Unknown:      e.unknown.Load(),
		DecodeErrors: e.decodeErrors.Load(),
		Encrypted:    uint64(e.dec.Encrypted),
	}
}

// Run blocks until ctx is cancelled or all sources are exhausted.
func (e *Engine) Run(ctx context.Context, sources ...capture.Source) error {
	e.mu.Lock()
	if e.running {
		e.mu.Unlock()
		return errors.New("engine: already running")
	}
	e.running = true
	e.mu.Unlock()
	defer func() {
		e.mu.Lock()
		e.running = false
		e.mu.Unlock()
	}()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	in := make(chan capture.Packet, 4096)
	errCh := make(chan error, 1)
	go func() {
		errCh <- capture.RunAll(ctx, in, sources...)
		close(in)
	}()

	for p := range in {
		e.Feed(p)
	}
	return <-errCh
}

// Feed processes one packet. Exposed for tests; must not be called concurrently with Run.
func (e *Engine) Feed(p capture.Packet) {
	e.packets.Add(1)
	if e.cfg.Recorder != nil {
		_ = e.cfg.Recorder.Write(p)
	}
	e.dec.Feed(p.Payload, p.At)
}

func (e *Engine) changed(t Topic) {
	if e.cfg.OnChange != nil {
		e.cfg.OnChange(t)
	}
}

// pushTL appends one entry to the merged timeline and signals the UI. It runs
// on the decode goroutine, right after the matching tracker/state update, so
// the timeline stays consistent with the per-view data.
func (e *Engine) pushTL(ev timeline.Event) {
	e.tl.Add(ev)
	e.changed(TopicTimeline)
}

func (e *Engine) fail(name string, err error) {
	e.decodeErrors.Add(1)
	if e.cfg.OnDecodeError != nil {
		e.cfg.OnDecodeError(name, err)
	}
}

func (e *Engine) handle(m protocol.Message) {
	e.messages.Add(1)
	if m.Name == "" {
		e.unknown.Add(1)
	}
	if e.cfg.OnMessage != nil {
		e.cfg.OnMessage(m)
	}

	switch m.Kind {
	case protocol.KindResponse:
		if m.Name == events.OpJoin {
			if j, err := events.DecodeJoinResponse(m.Params); err != nil {
				e.fail(m.Name, err)
			} else {
				e.state.ApplyJoin(j)
				e.chest.Reset() // object ids are per-zone
				e.changed(TopicZone)
				e.pushTL(timeline.Event{At: m.At, Kind: timeline.KindZone, Zone: e.state.Location(), Player: j.Name})
			}
		}
	case protocol.KindEvent:
		e.handleEvent(m)
	}
}

func (e *Engine) handleEvent(m protocol.Message) {
	switch m.Name {
	case events.EvNewCharacter:
		if c, err := events.DecodeNewCharacter(m.Params); err != nil {
			e.fail(m.Name, err)
		} else {
			e.state.ApplyNewCharacter(c)
		}
	case events.EvLeave:
		if l, err := events.DecodeLeave(m.Params); err == nil {
			e.state.ApplyLeave(l)
		}
	case events.EvOtherGrabbedLoot:
		g, err := events.DecodeGrabbedLoot(m.Params)
		if err != nil {
			e.fail(m.Name, err)
			return
		}
		entry := e.loot.Record(g, m.At, e.state)
		if e.cfg.OnLoot != nil {
			e.cfg.OnLoot(entry)
		}
		e.changed(TopicLoot)
		tev := timeline.Event{At: entry.At, Player: entry.Looter, Source: entry.LootedFrom}
		if entry.IsSilver {
			tev.Kind, tev.Silver = timeline.KindSilver, entry.Silver
		} else {
			tev.Kind, tev.Item, tev.ItemID, tev.Quantity = timeline.KindLoot, entry.ItemName, entry.ItemID, entry.Quantity
		}
		e.pushTL(tev)
	case events.EvHealthUpdate:
		u, err := events.DecodeHealthUpdate(m.Params)
		if err != nil {
			e.fail(m.Name, err)
			return
		}
		e.combat.Apply(u, m.At, e.state)
		e.changed(TopicCombat)
	case events.EvHealthUpdates:
		us, err := events.DecodeHealthUpdates(m.Params)
		if err != nil {
			e.fail(m.Name, err)
			return
		}
		for _, u := range us {
			e.combat.Apply(u, m.At, e.state)
		}
		e.changed(TopicCombat)
	case events.EvPartyJoined:
		if p, err := events.DecodePartyJoined(m.Params); err != nil {
			e.fail(m.Name, err)
		} else {
			e.state.ApplyPartyJoined(p)
			e.changed(TopicParty)
			for _, mem := range p.Members {
				e.pushTL(timeline.Event{At: m.At, Kind: timeline.KindPartyJoined, Player: mem.Name})
			}
		}
	case events.EvPartyPlayerJoined:
		if p, err := events.DecodePartyPlayerJoined(m.Params); err != nil {
			e.fail(m.Name, err)
		} else {
			e.state.ApplyPartyPlayerJoined(p)
			e.changed(TopicParty)
			e.pushTL(timeline.Event{At: m.At, Kind: timeline.KindPartyJoined, Player: p.Name})
		}
	case events.EvPartyPlayerLeft:
		if p, err := events.DecodePartyPlayerLeft(m.Params); err != nil {
			e.fail(m.Name, err)
		} else {
			// Resolve the name before the member is removed from state.
			name := p.GUID
			for _, mem := range e.state.PartyMembers() {
				if mem.GUID == p.GUID {
					name = mem.Name
					break
				}
			}
			e.state.ApplyPartyPlayerLeft(p)
			e.changed(TopicParty)
			e.pushTL(timeline.Event{At: m.At, Kind: timeline.KindPartyLeft, Player: name})
		}
	case events.EvPartyDisbanded:
		e.state.ApplyPartyDisbanded()
		e.changed(TopicParty)
	case events.EvUpdateFame:
		f, err := events.DecodeUpdateFame(m.Params)
		if err != nil {
			e.fail(m.Name, err)
			return
		}
		// Fame is private: count only events whose subject is the local player,
		// and only once the local player is known (after Join). This avoids
		// attributing an event with an absent/zero subject id to the session.
		if local := e.state.LocalPlayer(); local.ObjectID != 0 && f.PlayerObjectID == local.ObjectID {
			e.fame.Add(f, m.At)
			e.changed(TopicFame)
			e.pushTL(timeline.Event{At: m.At, Kind: timeline.KindFame, Player: local.Name, Fame: f.Gained()})
		}

	// --- Chest loot: stitched across four events (see internal/chest) ---
	case events.EvNewLootChest:
		if c, err := events.DecodeNewLootChest(m.Params); err == nil {
			e.chest.NewChest(c)
		}
	case events.EvNewSimpleItem, events.EvNewEquipmentItem:
		if it, err := events.DecodeItemObject(m.Name, m.Params); err == nil {
			e.chest.NewItem(it)
		}
	case events.EvAttachItemContainer:
		if c, err := events.DecodeAttachItemContainer(m.Params); err == nil {
			e.chest.Attach(c)
		}
	case events.EvInventoryPutItem:
		put, err := events.DecodeInventoryPutItem(m.Params)
		if err != nil {
			return
		}
		take, ok := e.chest.Take(put.ItemObjectID)
		if !ok {
			return
		}
		g := events.GrabbedLoot{
			Looter:     e.state.LocalPlayer().Name,
			LootedFrom: take.ChestName,
			ItemIndex:  take.ItemIndex,
			Quantity:   take.Quantity,
		}
		entry := e.loot.Record(g, m.At, e.state)
		if e.cfg.OnLoot != nil {
			e.cfg.OnLoot(entry)
		}
		e.changed(TopicLoot)
		e.pushTL(timeline.Event{
			At: entry.At, Kind: timeline.KindLoot, Player: entry.Looter,
			Source: entry.LootedFrom, Item: entry.ItemName, ItemID: entry.ItemID, Quantity: entry.Quantity,
		})

	// --- Killfeed: nearby deaths (broadcast). KilledPlayer (local-kill credit)
	// is decoded too but the feed is driven by the richer Died event. ---
	case events.EvDied:
		d, err := events.DecodeDied(m.Params)
		if err != nil {
			e.fail(m.Name, err)
			return
		}
		e.pushTL(timeline.Event{
			At: m.At, Kind: timeline.KindKill,
			Player: d.Victim, Guild: d.VictimGuild,
			Killer: d.Killer, KillerGuild: d.KillerGuild, Lethal: d.Lethal,
		})
	}
}
