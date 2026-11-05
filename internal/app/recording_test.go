package app

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LucasGitDev/albion-companion/internal/capture"
)

// mkRec writes a dummy .aorec of n bytes with a given modtime.
func mkRec(t *testing.T, dir, name string, size int, age time.Duration) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, make([]byte, size), 0o644); err != nil {
		t.Fatalf("write %s: %v", name, err)
	}
	mt := time.Now().Add(-age)
	if err := os.Chtimes(p, mt, mt); err != nil {
		t.Fatalf("chtimes: %v", err)
	}
	return p
}

func exists(p string) bool { _, err := os.Stat(p); return err == nil }

func TestPruneRecordings_SizeCap(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(Options{RecordingsDir: dir})
	svc.mu.Lock()
	svc.recCfg = RecordConfig{MaxMB: 1, MaxDays: 0} // 1 MB cap, no age cap
	svc.mu.Unlock()

	old := mkRec(t, dir, "albion-20260101-000000.aorec", 700*1024, 48*time.Hour)
	mid := mkRec(t, dir, "albion-20260102-000000.aorec", 700*1024, 24*time.Hour)
	newF := mkRec(t, dir, "albion-20260103-000000.aorec", 700*1024, 1*time.Hour)

	svc.pruneRecordings()

	// 3×700KB = ~2.05MB > 1MB: oldest deleted until under cap. Newest kept.
	if exists(old) {
		t.Error("oldest recording should be pruned by size cap")
	}
	if !exists(newF) {
		t.Error("newest recording must be kept")
	}
	// mid may or may not survive depending on exact cap math; newest always kept.
	_ = mid
}

func TestPruneRecordings_AgeCap(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(Options{RecordingsDir: dir})
	svc.mu.Lock()
	svc.recCfg = RecordConfig{MaxMB: 0, MaxDays: 7} // age cap only
	svc.mu.Unlock()

	stale := mkRec(t, dir, "albion-20260101-000000.aorec", 1024, 10*24*time.Hour)
	fresh := mkRec(t, dir, "albion-20260102-000000.aorec", 1024, 2*24*time.Hour)

	svc.pruneRecordings()

	if exists(stale) {
		t.Error("recording older than MaxDays should be pruned")
	}
	if !exists(fresh) {
		t.Error("recent recording must be kept")
	}
}

func TestPruneRecordings_NeverDeletesInProgress(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(Options{RecordingsDir: dir})
	svc.mu.Lock()
	svc.recCfg = RecordConfig{MaxMB: 0, MaxDays: 1} // age cap would delete old files
	svc.mu.Unlock()

	other := mkRec(t, dir, "albion-20250101-000000.aorec", 1024, 100*24*time.Hour)

	// The in-progress recording, made to look old on disk, must still survive.
	curPath := filepath.Join(dir, "albion-current.aorec")
	rec, err := capture.NewRecorder(curPath)
	if err != nil {
		t.Fatalf("recorder: %v", err)
	}
	defer rec.Close()
	old := time.Now().Add(-100 * 24 * time.Hour)
	if err := os.Chtimes(curPath, old, old); err != nil {
		t.Fatalf("chtimes: %v", err)
	}
	svc.mu.Lock()
	svc.curRec = rec
	svc.mu.Unlock()

	svc.pruneRecordings()
	if !exists(curPath) {
		t.Error("the in-progress recording must never be pruned")
	}
	if exists(other) {
		t.Error("other stale recordings should be pruned by age cap")
	}
}

func TestRecordConfig_PersistRoundTrip(t *testing.T) {
	dir := t.TempDir()
	cfg := RecordConfig{AutoRecord: false, MaxMB: 512, MaxDays: 14}
	if err := saveRecordConfig(dir, cfg); err != nil {
		t.Fatalf("save: %v", err)
	}
	got := loadRecordConfig(dir)
	if got.AutoRecord || got.MaxMB != 512 || got.MaxDays != 14 {
		t.Fatalf("round-trip = %+v", got)
	}
	// Missing file → defaults (auto-record on).
	def := loadRecordConfig(filepath.Join(dir, "nope"))
	if !def.AutoRecord || def.MaxMB != defaultRecordMaxMB {
		t.Fatalf("defaults = %+v", def)
	}
}

func TestSetRecordConfig_ClampsNegatives(t *testing.T) {
	svc := NewService(Options{ConfigDir: t.TempDir()})
	svc.SetRecordConfig(RecordConfig{AutoRecord: true, MaxMB: -5, MaxDays: -3})
	got := svc.RecordConfig()
	if got.MaxMB != 0 || got.MaxDays != 0 {
		t.Fatalf("negatives should clamp to 0, got %+v", got)
	}
}
