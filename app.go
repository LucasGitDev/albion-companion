package main

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/LucasGitDev/albion-companion/internal/app"
	"github.com/LucasGitDev/albion-companion/internal/capture"
	"github.com/LucasGitDev/albion-companion/internal/combat"
	"github.com/LucasGitDev/albion-companion/internal/fame"
	"github.com/LucasGitDev/albion-companion/internal/itemdb"
	"github.com/LucasGitDev/albion-companion/internal/loot"
	"github.com/LucasGitDev/albion-companion/internal/protocol"
)

// App is the thin Wails binding layer. Every exported method delegates to
// app.Service so the engine stays UI-agnostic; the struct holds only the Wails
// context needed to emit events and run the flush loop.
type App struct {
	ctx   context.Context
	svc   *app.Service
	stop  context.CancelFunc
	codes *protocol.Codes

	mu    sync.Mutex
	items *itemdb.DB
}

func NewApp() *App { return &App{} }

// startup builds the Service once the Wails runtime context exists, so the
// emitter can forward engine notifications to the frontend.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	a.codes = protocol.DefaultCodes()
	opts := app.Options{
		Codes:         a.codes,
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
			a.mu.Lock()
			a.items = db
			a.mu.Unlock()
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
func (a *App) ResetFame()                               { a.svc.ResetFame() }
func (a *App) FameView() fame.Snapshot                  { return a.svc.FameView() }
func (a *App) TimelineView(recent int) app.TimelineView { return a.svc.TimelineView(recent) }
func (a *App) ResetTimeline()                           { a.svc.ResetTimeline() }

// ExportLootCSV opens a native save dialog and writes the current (filtered)
// loot in ao-loot-logger format. Returns the saved path, or "" if the user
// cancelled the dialog.
func (a *App) ExportLootCSV() (string, error) {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export loot",
		DefaultFilename: "albion-loot-" + time.Now().Format("2006-01-02-150405") + ".csv",
		Filters:         []runtime.FileFilter{{DisplayName: "CSV (*.csv)", Pattern: "*.csv"}},
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil // cancelled
	}
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if err := a.svc.ExportLootCSV(f); err != nil {
		return "", err
	}
	return path, nil
}

// Interfaces lists the physical capture interfaces for the Settings picker.
func (a *App) Interfaces() ([]string, error) { return capture.Interfaces() }

// CodesInfo describes the loaded event/operation code table.
type CodesInfo struct {
	Source     string `json:"source"`
	Events     int    `json:"events"`
	Operations int    `json:"operations"`
}

func (a *App) CodesInfo() CodesInfo {
	return CodesInfo{Source: a.codes.Source, Events: len(a.codes.Events), Operations: len(a.codes.Operations)}
}

// ItemDBInfo reports how many items are loaded.
type ItemDBInfo struct {
	Count int `json:"count"`
}

func (a *App) ItemDBInfo() ItemDBInfo {
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.items == nil {
		return ItemDBInfo{}
	}
	return ItemDBInfo{Count: a.items.Len()}
}

// RefreshItemDB forces a fresh item index download (maxAge 0) and rebinds it.
func (a *App) RefreshItemDB() (ItemDBInfo, error) {
	db, err := itemdb.LoadCached(a.ctx, itemsCacheDir(), 0)
	if err != nil {
		return a.ItemDBInfo(), err
	}
	a.svc.Engine().Loot().SetItems(db)
	a.mu.Lock()
	a.items = db
	a.mu.Unlock()
	runtime.EventsEmit(a.ctx, "loot:changed", nil)
	return ItemDBInfo{Count: db.Len()}, nil
}
