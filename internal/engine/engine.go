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
	"github.com/LucasGitDev/albion-companion/internal/combat"
	"github.com/LucasGitDev/albion-companion/internal/events"
	"github.com/LucasGitDev/albion-companion/internal/game"
	"github.com/LucasGitDev/albion-companion/internal/loot"
	"github.com/LucasGitDev/albion-companion/internal/protocol"
)

// Topic tells listeners what changed, so the UI can refresh only that view.
type Topic string

const (
	TopicLoot   Topic = "loot"
	TopicCombat Topic = "combat"
	TopicParty  Topic = "party"
	TopicZone   Topic = "zone"
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
	}
	e.dec = protocol.NewDecoder(cfg.Codes, e.handle)
	return e
}

func (e *Engine) State() *game.State    { return e.state }
func (e *Engine) Loot() *loot.Tracker   { return e.loot }
func (e *Engine) Combat() *combat.Meter { return e.combat }

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
				e.changed(TopicZone)
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
		}
	case events.EvPartyPlayerJoined:
		if p, err := events.DecodePartyPlayerJoined(m.Params); err != nil {
			e.fail(m.Name, err)
		} else {
			e.state.ApplyPartyPlayerJoined(p)
			e.changed(TopicParty)
		}
	case events.EvPartyPlayerLeft:
		if p, err := events.DecodePartyPlayerLeft(m.Params); err != nil {
			e.fail(m.Name, err)
		} else {
			e.state.ApplyPartyPlayerLeft(p)
			e.changed(TopicParty)
		}
	case events.EvPartyDisbanded:
		e.state.ApplyPartyDisbanded()
		e.changed(TopicParty)
	}
}
