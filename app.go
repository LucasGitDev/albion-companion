package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"

	"github.com/LucasGitDev/albion-companion/internal/app"
	"github.com/LucasGitDev/albion-companion/internal/capture"
	"github.com/LucasGitDev/albion-companion/internal/chestlog"
	"github.com/LucasGitDev/albion-companion/internal/combat"
	"github.com/LucasGitDev/albion-companion/internal/fame"
	"github.com/LucasGitDev/albion-companion/internal/itemdb"
	"github.com/LucasGitDev/albion-companion/internal/loot"
	"github.com/LucasGitDev/albion-companion/internal/prices"
	"github.com/LucasGitDev/albion-companion/internal/protocol"
	"github.com/LucasGitDev/albion-companion/internal/route"
	"github.com/LucasGitDev/albion-companion/internal/store"
)

// App is the thin Wails binding layer. Every exported method delegates to
// app.Service so the engine stays UI-agnostic; the struct holds only the Wails
// context needed to emit events and run the flush loop.
type App struct {
	ctx      context.Context
	svc      *app.Service
	stop     context.CancelFunc
	codes    *protocol.Codes
	recorder *capture.Recorder

	mu    sync.Mutex
	items *itemdb.DB
}

func NewApp() *App { return &App{} }

// startup builds the Service once the Wails runtime context exists, so the
// emitter can forward engine notifications to the frontend.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	// A user override file (hotfix codes after a patch without a new release)
	// takes precedence over the embedded codes.json when present and valid.
	a.codes = protocol.DefaultCodes()
	if path := codesOverridePath(); path != "" {
		if c, err := protocol.LoadCodes(path); err == nil {
			a.codes = c
			runtime.LogInfof(ctx, "codes: using override %s", path)
		} else if !os.IsNotExist(err) {
			runtime.LogWarningf(ctx, "codes: ignoring invalid override %s: %v", path, err)
		}
	}
	opts := app.Options{
		Codes:          a.codes,
		Throttle:       250 * time.Millisecond,
		PartyOnlyLoot:  true,
		Combat:         combat.Options{PartyOnly: true},
		SessionsDir:    sessionsDir(),
		PricesCacheDir: itemsCacheDir(),
		ConfigDir:      configDir(),
		Emit:           func(event string, payload any) { runtime.EventsEmit(a.ctx, event, payload) },
	}
	// ALBION_COMPANION_RECORD=<path|dir> records every captured packet to a
	// timestamped .aorec so a live test session can be replayed and correlated
	// with wall-clock notes. A directory (or a value not ending in .aorec) gets a
	// generated albion-YYYYMMDD-HHMMSS.aorec inside it.
	if rec := os.Getenv("ALBION_COMPANION_RECORD"); rec != "" {
		path := recordPath(rec)
		if r, err := capture.NewRecorder(path); err != nil {
			runtime.LogWarningf(ctx, "record: cannot write %s: %v", path, err)
		} else {
			a.recorder = r
			opts.Recorder = r
			runtime.LogInfof(ctx, "record: writing capture to %s", path)
		}
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
			a.svc.Engine().Gather().SetItems(db)
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
	if a.recorder != nil {
		a.recorder.Close() // flush the .aorec
	}
}

// recordPath resolves ALBION_COMPANION_RECORD: an explicit *.aorec path is used
// as-is; anything else is treated as a directory to drop a timestamped file in.
func recordPath(v string) string {
	if strings.HasSuffix(v, ".aorec") {
		return v
	}
	return filepath.Join(v, "albion-"+time.Now().Format("20060102-150405")+".aorec")
}

// codesOverridePath is the optional user file that overrides the embedded codes
// (<config dir>/albion-companion/codes.json). Returns "" if no config dir.
func codesOverridePath() string {
	d, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(d, "albion-companion", "codes.json")
}

// sessionsDir is where finished sessions are persisted for the History tab
// (<config dir>/albion-companion/sessions). Empty config dir disables history.
func sessionsDir() string {
	d, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(d, "albion-companion", "sessions")
}

// configDir is where user config (the Loot Guard roster) is persisted
// (<config dir>/albion-companion). Empty when the OS config dir is unavailable,
// which disables persistence (roster stays in-memory for the session).
func configDir() string {
	d, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(d, "albion-companion")
}

func itemsCacheDir() string {
	if d, err := os.UserCacheDir(); err == nil {
		return filepath.Join(d, "albion-companion")
	}
	return filepath.Join(os.TempDir(), "albion-companion")
}

// --- Bindings exposed to the frontend (JSON-friendly signatures only) ---

func (a *App) Status() app.CaptureStatus                 { return a.svc.Status() }
func (a *App) StartCapture() (app.CaptureStatus, error)  { return a.svc.StartCapture() }
func (a *App) StopCapture() app.CaptureStatus            { return a.svc.StopCapture() }
func (a *App) LootView(recent int) app.LootView          { return a.svc.LootView(recent) }
func (a *App) CombatView() app.CombatView                { return a.svc.CombatView() }
func (a *App) PartyView() app.PartyView                  { return a.svc.PartyView() }
func (a *App) SetLootFilter(f loot.Filter)               { a.svc.SetLootFilter(f) }
func (a *App) SetCombatOptions(o combat.Options)         { a.svc.SetCombatOptions(o) }
func (a *App) ResetLoot()                                { a.svc.ResetLoot() }
func (a *App) ResetCombat()                              { a.svc.ResetCombat() }
func (a *App) ResetFame()                                { a.svc.ResetFame() }
func (a *App) FameView() fame.Snapshot                   { return a.svc.FameView() }
func (a *App) RatesView() app.RatesView                  { return a.svc.RatesView() }
func (a *App) ResetRates()                               { a.svc.ResetRates() }
func (a *App) GatherView() app.GatherView                { return a.svc.GatherView() }
func (a *App) ResetGather()                              { a.svc.ResetGather() }
func (a *App) LootGuardView(recent int) app.GuardView    { return a.svc.LootGuardView(recent) }
func (a *App) LootGuardConfig() app.GuardConfig          { return a.svc.LootGuardConfig() }
func (a *App) SetLootGuardConfig(cfg app.GuardConfig)    { a.svc.SetLootGuardConfig(cfg) }
func (a *App) KnownPlayers() []string                    { return a.svc.KnownPlayers() }
func (a *App) SessionInfo() app.SessionInfo              { return a.svc.SessionInfo() }
func (a *App) SetSessionInfo(label, typ string)          { a.svc.SetSessionInfo(label, typ) }
func (a *App) StartSession(label, typ string)            { a.svc.StartSession(label, typ) }
func (a *App) ChestLogView(recent int) chestlog.Snapshot { return a.svc.ChestLogView(recent) }
func (a *App) ResetChestLog()                            { a.svc.ResetChestLog() }
func (a *App) RouteView(recent int) route.Snapshot       { return a.svc.RouteView(recent) }
func (a *App) ResetRoute()                               { a.svc.ResetRoute() }
func (a *App) HistoryList() ([]store.Meta, error)        { return a.svc.HistoryList() }
func (a *App) HistoryLoad(id string) (app.SessionDetail, error) {
	return a.svc.HistoryLoad(id)
}
func (a *App) DeleteSession(id string) error            { return a.svc.DeleteSession(id) }
func (a *App) TimelineView(recent int) app.TimelineView { return a.svc.TimelineView(recent) }
func (a *App) ResetTimeline()                           { a.svc.ResetTimeline() }
func (a *App) DungeonView() app.DungeonView             { return a.svc.DungeonView() }
func (a *App) RenameDungeon(id int, name string)        { a.svc.RenameDungeon(id, name) }
func (a *App) SetDungeonTier(id int, tier string)       { a.svc.SetDungeonTier(id, tier) }
func (a *App) ResetDungeons()                           { a.svc.ResetDungeons() }

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

// ExportLootGuardCSV opens a native save dialog and writes the flagged
// (non-collector) loot feed as CSV for evidence. Returns the saved path, or ""
// if the user cancelled.
func (a *App) ExportLootGuardCSV() (string, error) {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export loot guard evidence",
		DefaultFilename: "albion-loot-guard-" + time.Now().Format("2006-01-02-150405") + ".csv",
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
	if err := a.svc.ExportLootGuardCSV(f); err != nil {
		return "", err
	}
	return path, nil
}

// ReprocessAorec opens a native file picker for an .aorec recording and replays
// it into a new History session. Returns the chosen file's base name (or "" if
// cancelled). Label/type are optional; an empty label defaults to the file name.
func (a *App) ReprocessAorec(label, typ string) (string, error) {
	path, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title:   "Reprocess a recording",
		Filters: []runtime.FileFilter{{DisplayName: "Albion recording (*.aorec)", Pattern: "*.aorec"}},
	})
	if err != nil {
		return "", err
	}
	if path == "" {
		return "", nil // cancelled
	}
	if err := a.svc.ReprocessAorec(path, label, typ); err != nil {
		return "", err
	}
	return filepath.Base(path), nil
}

// ExportSessionCSV opens a native save dialog and writes a saved session's loot
// as CSV. Returns the saved path, or "" if the user cancelled.
func (a *App) ExportSessionCSV(id string) (string, error) {
	path, err := runtime.SaveFileDialog(a.ctx, runtime.SaveDialogOptions{
		Title:           "Export session loot",
		DefaultFilename: "albion-session-" + id + ".csv",
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
	if err := a.svc.WriteSessionCSV(f, id); err != nil {
		return "", err
	}
	return path, nil
}

// Interfaces lists the physical capture interfaces for the Settings picker.
func (a *App) Interfaces() ([]string, error) { return capture.Interfaces() }

// PriceRegions lists the selectable market regions for the Settings dropdown.
func (a *App) PriceRegions() []prices.Region { return a.svc.PriceRegions() }

// SetPriceRegion overrides the loot-value market region (empty = auto-detect).
func (a *App) SetPriceRegion(key string) { a.svc.SetPriceRegion(key) }

// RefreshPrices force-refreshes item prices for the current session.
func (a *App) RefreshPrices() error { return a.svc.RefreshPrices() }

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
	a.svc.Engine().Gather().SetItems(db)
	a.mu.Lock()
	a.items = db
	a.mu.Unlock()
	runtime.EventsEmit(a.ctx, "loot:changed", nil)
	return ItemDBInfo{Count: db.Len()}, nil
}
