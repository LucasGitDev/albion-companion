package app

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Preferences is the app-level user config (preferences.json in the config dir).
// Feature DATA (the Loot Guard roster, the known-players registry) stays in its
// own files; this holds settings the user tweaks.
type Preferences struct {
	Recording RecordConfig `json:"recording"`
	// RecordingsDir overrides where recordings are written; empty = app default.
	RecordingsDir string `json:"recordingsDir,omitempty"`
}

func defaultPreferences() Preferences {
	return Preferences{Recording: defaultRecordConfig()}
}

func preferencesFile(dir string) string { return filepath.Join(dir, "preferences.json") }

// loadPreferences reads preferences.json. On first run it migrates a legacy
// recording.json into the new file (then preferences.json is authoritative).
func loadPreferences(dir string) Preferences {
	p := defaultPreferences()
	if dir == "" {
		return p
	}
	if b, err := os.ReadFile(preferencesFile(dir)); err == nil {
		_ = json.Unmarshal(b, &p)
		return p
	}
	// Migrate legacy recording.json if present.
	if b, err := os.ReadFile(recordConfigFile(dir)); err == nil {
		var rc RecordConfig
		if json.Unmarshal(b, &rc) == nil {
			p.Recording = rc
			_ = savePreferences(dir, p)
		}
	}
	return p
}

func savePreferences(dir string, p Preferences) error {
	if dir == "" {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return err
	}
	tmp := preferencesFile(dir) + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, preferencesFile(dir))
}
