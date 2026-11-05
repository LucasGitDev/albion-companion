package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/LucasGitDev/albion-companion/internal/capture"
)

// Auto-recording keeps a per-capture .aorec on disk so any session can be
// reprocessed later (see Service.ReprocessAorec). It is on by default, with a
// disk-retention policy so recordings don't grow without bound. An explicit
// ALBION_COMPANION_RECORD override (wired in app.go) disables auto-recording.

const (
	defaultRecordMaxMB   = 2048 // ~2 GB of recordings kept
	defaultRecordMaxDays = 30
)

// RecordConfig is the user's auto-record preference, persisted to disk.
type RecordConfig struct {
	AutoRecord bool `json:"autoRecord"`
	MaxMB      int  `json:"maxMB"`   // prune oldest once total exceeds this (0 = no size cap)
	MaxDays    int  `json:"maxDays"` // prune recordings older than this (0 = no age cap)
}

func defaultRecordConfig() RecordConfig {
	return RecordConfig{AutoRecord: true, MaxMB: defaultRecordMaxMB, MaxDays: defaultRecordMaxDays}
}

// RecordingsInfo reports the current recording settings and disk usage for the
// Settings UI.
type RecordingsInfo struct {
	Dir      string       `json:"dir"`
	Enabled  bool         `json:"enabled"`  // auto-record on AND a dir is available
	Override bool         `json:"override"` // ALBION_COMPANION_RECORD is set (auto-record disabled)
	Count    int          `json:"count"`
	Bytes    int64        `json:"bytes"`
	Config   RecordConfig `json:"config"`
}

// RecordConfig returns the current auto-record settings.
func (s *Service) RecordConfig() RecordConfig {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.recCfg
}

// SetRecordConfig updates and persists the auto-record settings. It does not
// touch an in-progress recording; the change takes effect on the next capture.
func (s *Service) SetRecordConfig(cfg RecordConfig) {
	if cfg.MaxMB < 0 {
		cfg.MaxMB = 0
	}
	if cfg.MaxDays < 0 {
		cfg.MaxDays = 0
	}
	s.mu.Lock()
	s.recCfg = cfg
	dir := s.opts.ConfigDir
	s.mu.Unlock()
	if dir != "" {
		_ = saveRecordConfig(dir, cfg)
	}
	go s.pruneRecordings() // apply new caps immediately, off the UI thread
	s.opts.Emit("recording:changed", nil)
}

// RecordingsInfo lists the recordings directory usage.
func (s *Service) RecordingsInfo() RecordingsInfo {
	s.mu.Lock()
	cfg := s.recCfg
	dir := s.opts.RecordingsDir
	override := s.opts.Recorder != nil
	s.mu.Unlock()
	info := RecordingsInfo{Dir: dir, Override: override, Config: cfg, Enabled: cfg.AutoRecord && dir != "" && !override}
	for _, f := range listRecordings(dir) {
		info.Count++
		info.Bytes += f.size
	}
	return info
}

// startRecording opens a new per-capture recording and attaches it to the
// engine. No-op when auto-record is off, there's no dir, or an external
// recorder (env override) is in use. Caller holds no lock.
func (s *Service) startRecording(start time.Time) {
	s.mu.Lock()
	cfg := s.recCfg
	dir := s.opts.RecordingsDir
	override := s.opts.Recorder != nil
	s.mu.Unlock()
	if !cfg.AutoRecord || dir == "" || override {
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return
	}
	path := filepath.Join(dir, "albion-"+start.Format("20060102-150405")+".aorec")
	r, err := capture.NewRecorder(path)
	if err != nil {
		return
	}
	if prev := s.eng.SetRecorder(r); prev != nil {
		_ = prev.Close()
	}
	s.mu.Lock()
	s.curRec = r
	s.mu.Unlock()
}

// stopRecording closes the current recording (if any) and prunes old ones.
func (s *Service) stopRecording() {
	s.mu.Lock()
	r := s.curRec
	s.curRec = nil
	override := s.opts.Recorder != nil
	s.mu.Unlock()
	if override {
		return // the external recorder is owned by app.go
	}
	s.eng.SetRecorder(nil)
	if r != nil {
		_ = r.Close()
	}
	s.pruneRecordings()
}

type recFile struct {
	path string
	size int64
	mod  time.Time
}

// listRecordings returns albion-*.aorec files in dir, oldest first.
func listRecordings(dir string) []recFile {
	if dir == "" {
		return nil
	}
	ents, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	out := make([]recFile, 0, len(ents))
	for _, e := range ents {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".aorec") {
			continue
		}
		fi, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, recFile{path: filepath.Join(dir, e.Name()), size: fi.Size(), mod: fi.ModTime()})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].mod.Before(out[j].mod) })
	return out
}

// pruneRecordings enforces the age and size caps, deleting oldest first. The
// in-progress recording (curRec) is never deleted.
func (s *Service) pruneRecordings() {
	s.mu.Lock()
	cfg := s.recCfg
	dir := s.opts.RecordingsDir
	var keepPath string
	if s.curRec != nil {
		keepPath = s.curRec.Path()
	}
	s.mu.Unlock()
	if dir == "" {
		return
	}

	files := listRecordings(dir)
	// Age cap: drop anything older than MaxDays.
	if cfg.MaxDays > 0 {
		cutoff := time.Now().AddDate(0, 0, -cfg.MaxDays)
		kept := files[:0]
		for _, f := range files {
			if f.path != keepPath && f.mod.Before(cutoff) {
				_ = os.Remove(f.path)
				continue
			}
			kept = append(kept, f)
		}
		files = kept
	}
	// Size cap: drop oldest until total <= MaxMB.
	if cfg.MaxMB > 0 {
		limit := int64(cfg.MaxMB) * 1024 * 1024
		var total int64
		for _, f := range files {
			total += f.size
		}
		for _, f := range files {
			if total <= limit {
				break
			}
			if f.path == keepPath {
				continue
			}
			if os.Remove(f.path) == nil {
				total -= f.size
			}
		}
	}
}

// --- persistence ------------------------------------------------------------

func recordConfigFile(dir string) string { return filepath.Join(dir, "recording.json") }

// loadRecordConfig reads the settings; a missing file yields the defaults.
func loadRecordConfig(dir string) RecordConfig {
	cfg := defaultRecordConfig()
	if dir == "" {
		return cfg
	}
	b, err := os.ReadFile(recordConfigFile(dir))
	if err != nil {
		return cfg
	}
	_ = json.Unmarshal(b, &cfg)
	return cfg
}

func saveRecordConfig(dir string, cfg RecordConfig) error {
	if dir == "" {
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	b, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	tmp := recordConfigFile(dir) + ".tmp"
	if err := os.WriteFile(tmp, b, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, recordConfigFile(dir))
}
