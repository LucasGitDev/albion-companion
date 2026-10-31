package main

import (
	"context"
	"os"
	"path/filepath"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/LucasGitDev/albion-companion/internal/app"
	"github.com/LucasGitDev/albion-companion/internal/capture"
	"github.com/LucasGitDev/albion-companion/internal/combat"
	"github.com/LucasGitDev/albion-companion/internal/itemdb"
	"github.com/LucasGitDev/albion-companion/internal/loot"
	"github.com/LucasGitDev/albion-companion/internal/protocol"
)

// App is the thin Wails binding layer. Every exported method delegates to
// app.Service so the engine stays UI-agnostic; the struct holds only the Wails
// context needed to emit events and run the flush loop.
type App struct {
	ctx  context.Context
	svc  *app.Service
	stop context.CancelFunc
}

func NewApp() *App { return &App{} }

// startup builds the Service once the Wails runtime context exists, so the
// emitter can forward engine notifications to the frontend.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	opts := app.Options{
		Codes:         protocol.DefaultCodes(),
		Throttle:      250 * time.Millisecond,
		PartyOnlyLoot: true,
		Combat:        combat.Options{PartyOnly: true},
		Emit:          func(event string, payload any) { runtime.EventsEmit(a.ctx, event, payload) },
	}
	// Dev without the game: ALBION_COMPANION_REPLAY=<file> replays a recording
	// at real time instead of capturing live (see UI spec).
	if replay := os.Getenv("ALBION_COMPANION_REPLAY"); replay != "" {
		opts.Sources = func() ([]capture.Source, error) {
			return []capture.Source{capture.ReplaySource{Path: replay, Speed: 1}}, nil
		}
	}
	a.svc = app.NewService(opts)

	// Flush throttled "*:changed" events on a fixed cadence.
	flushCtx, cancel := context.WithCancel(ctx)
	a.stop = cancel
	go func() {
		t := time.NewTicker(opts.Throttle)
		defer t.Stop()
		for {
			select {
			case <-flushCtx.Done():
				return
			case <-t.C:
				a.svc.Flush()
			}
		}
	}()

	// Resolve item names/icons in the background; the engine runs without them.
	go func() {
		db, err := itemdb.LoadCached(ctx, itemsCacheDir(), 24*time.Hour)
		if err == nil {
			a.svc.Engine().Loot().SetItems(db)
			runtime.EventsEmit(a.ctx, "loot:changed", nil)
		}
	}()

	// In replay mode, start immediately so the debug view fills without a click.
	if os.Getenv("ALBION_COMPANION_REPLAY") != "" {
		_, _ = a.svc.StartCapture()
	}
}

func (a *App) shutdown(context.Context) {
	if a.stop != nil {
		a.stop()
	}
	if a.svc != nil {
		a.svc.StopCapture()
	}
}

func itemsCacheDir() string {
	if d, err := os.UserCacheDir(); err == nil {
		return filepath.Join(d, "albion-companion")
	}
	return filepath.Join(os.TempDir(), "albion-companion")
}

// --- Bindings exposed to the frontend (JSON-friendly signatures only) ---

func (a *App) Status() app.CaptureStatus                { return a.svc.Status() }
func (a *App) StartCapture() (app.CaptureStatus, error) { return a.svc.StartCapture() }
func (a *App) StopCapture() app.CaptureStatus           { return a.svc.StopCapture() }
func (a *App) LootView(recent int) app.LootView         { return a.svc.LootView(recent) }
func (a *App) CombatView() app.CombatView               { return a.svc.CombatView() }
func (a *App) PartyView() app.PartyView                 { return a.svc.PartyView() }
func (a *App) SetLootFilter(f loot.Filter)              { a.svc.SetLootFilter(f) }
func (a *App) SetCombatOptions(o combat.Options)        { a.svc.SetCombatOptions(o) }
func (a *App) ResetLoot()                               { a.svc.ResetLoot() }
func (a *App) ResetCombat()                             { a.svc.ResetCombat() }
