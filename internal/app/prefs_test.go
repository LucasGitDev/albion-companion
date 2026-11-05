package app

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPreferences_PersistRoundTrip(t *testing.T) {
	dir := t.TempDir()
	p := Preferences{Recording: RecordConfig{AutoRecord: false, MaxMB: 512, MaxDays: 14}, RecordingsDir: "/tmp/recs"}
	if err := savePreferences(dir, p); err != nil {
		t.Fatalf("save: %v", err)
	}
	got := loadPreferences(dir)
	if got.Recording.AutoRecord || got.Recording.MaxMB != 512 || got.RecordingsDir != "/tmp/recs" {
		t.Fatalf("round-trip = %+v", got)
	}
}

func TestPreferences_MigratesLegacyRecordingJSON(t *testing.T) {
	dir := t.TempDir()
	// Only a legacy recording.json exists (no preferences.json yet).
	if err := saveRecordConfig(dir, RecordConfig{AutoRecord: false, MaxMB: 256, MaxDays: 7}); err != nil {
		t.Fatalf("seed legacy: %v", err)
	}
	p := loadPreferences(dir)
	if p.Recording.AutoRecord || p.Recording.MaxMB != 256 || p.Recording.MaxDays != 7 {
		t.Fatalf("legacy not migrated: %+v", p.Recording)
	}
	// preferences.json is now authoritative.
	if _, err := os.Stat(preferencesFile(dir)); err != nil {
		t.Fatalf("preferences.json should exist after migration: %v", err)
	}
}

func TestPreferences_DefaultsWhenMissing(t *testing.T) {
	p := loadPreferences(filepath.Join(t.TempDir(), "nope"))
	if !p.Recording.AutoRecord || p.RecordingsDir != "" {
		t.Fatalf("defaults = %+v", p)
	}
}

func TestEffectiveRecordingsDir(t *testing.T) {
	// Default from Options when no override.
	svc := NewService(Options{RecordingsDir: "/default/recs"})
	svc.mu.Lock()
	def := svc.effectiveRecordingsDirLocked()
	svc.mu.Unlock()
	if def != "/default/recs" {
		t.Fatalf("default dir = %q", def)
	}
	// User override wins.
	svc.SetRecordingsDir("/custom/recs")
	svc.mu.Lock()
	got := svc.effectiveRecordingsDirLocked()
	svc.mu.Unlock()
	if got != "/custom/recs" {
		t.Fatalf("override dir = %q", got)
	}
	if svc.RecordingsInfo().Dir != "/custom/recs" {
		t.Fatalf("RecordingsInfo.Dir = %q", svc.RecordingsInfo().Dir)
	}
	// Clearing reverts to the default.
	svc.SetRecordingsDir("")
	if svc.RecordingsInfo().Dir != "/default/recs" {
		t.Fatalf("cleared dir = %q", svc.RecordingsInfo().Dir)
	}
}
