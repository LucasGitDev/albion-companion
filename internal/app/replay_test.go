package app

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/LucasGitDev/albion-companion/internal/capture"
	"github.com/LucasGitDev/albion-companion/internal/protocol"
)

func fixtureAorec() string { return filepath.Join("..", "..", "testdata", "party-dungeon.aorec") }

func TestAorecSpan(t *testing.T) {
	first, last, count, err := capture.AorecSpan(fixtureAorec())
	if err != nil {
		t.Fatalf("span: %v", err)
	}
	if count == 0 {
		t.Fatal("expected packets in the fixture")
	}
	if last.Before(first) {
		t.Fatalf("last %v before first %v", last, first)
	}
}

func TestReplayLifecycle(t *testing.T) {
	dir := t.TempDir()
	svc := NewService(Options{Codes: protocol.DefaultCodes(), SessionsDir: dir})

	if err := svc.StartReplay(fixtureAorec(), "Test replay"); err != nil {
		t.Fatalf("start replay: %v", err)
	}
	svc.SetReplaySpeed(maxReplaySpeed) // as fast as the (bounded, paced) playback allows

	if st := svc.Status(); !st.Replay || st.ReplayLabel != "Test replay" {
		t.Fatalf("status during replay = %+v", st)
	}

	// Playback is intentionally paced (never a tight loop), so don't wait for the
	// whole file — just confirm it makes progress, then stop it.
	deadline := time.Now().Add(10 * time.Second)
	for svc.Engine().Stats().Packets == 0 {
		if time.Now().After(deadline) {
			svc.StopReplay()
			t.Fatal("replay made no progress")
		}
		time.Sleep(20 * time.Millisecond)
	}
	// Replaying never persists a session (non-destructive).
	if metas, _ := svc.HistoryList(); len(metas) != 0 {
		t.Fatalf("replay must not persist a session, got %d", len(metas))
	}

	svc.StopReplay()
	stopBy := time.Now().Add(10 * time.Second)
	for svc.Status().Replay {
		if time.Now().After(stopBy) {
			t.Fatal("replay did not stop")
		}
		time.Sleep(20 * time.Millisecond)
	}
	if metas, _ := svc.HistoryList(); len(metas) != 0 {
		t.Fatalf("replay must not persist a session after stop, got %d", len(metas))
	}
}

func TestSeekReplay(t *testing.T) {
	svc := NewService(Options{Codes: protocol.DefaultCodes(), SessionsDir: t.TempDir()})
	if err := svc.StartReplay(fixtureAorec(), "seek"); err != nil {
		t.Fatalf("start: %v", err)
	}
	// Seek to a point; playback restarts fast-forwarding there (race-checked).
	svc.SeekReplay(1.0)
	if st := svc.Status(); !st.Replay {
		t.Fatalf("still replaying after seek, got %+v", st)
	}
	// Seeking past the end clamps and stays valid.
	svc.SeekReplay(1e9)
	if metas, _ := svc.HistoryList(); len(metas) != 0 {
		t.Fatalf("seek must not persist a session, got %d", len(metas))
	}
	svc.StopReplay()
	deadline := time.Now().Add(10 * time.Second)
	for svc.Status().Replay {
		if time.Now().After(deadline) {
			t.Fatal("replay did not stop after seek")
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func TestClampSpeed(t *testing.T) {
	if clampSpeed(0) != maxReplaySpeed || clampSpeed(-5) != maxReplaySpeed {
		t.Fatal("0/negative should clamp to the bounded max (never unbounded)")
	}
	if clampSpeed(1000) != maxReplaySpeed {
		t.Fatal("above max should clamp")
	}
	if clampSpeed(8) != 8 {
		t.Fatal("in-range speed should pass through")
	}
}

func TestStartCaptureRefusedDuringReplay(t *testing.T) {
	svc := NewService(Options{Codes: protocol.DefaultCodes()})
	// Simulate an active replay.
	svc.mu.Lock()
	svc.replaying = true
	svc.replayCtl = capture.NewReplayControl(1)
	svc.mu.Unlock()

	if _, err := svc.StartCapture(); err == nil {
		t.Fatal("StartCapture should be refused during a replay")
	}
	if err := svc.StartReplay(fixtureAorec(), "x"); err == nil {
		t.Fatal("a second StartReplay should be refused")
	}

	svc.mu.Lock()
	svc.replaying = false
	svc.replayCtl = nil
	svc.mu.Unlock()
}

func TestReplayPauseResumeSpeed(t *testing.T) {
	svc := NewService(Options{Codes: protocol.DefaultCodes(), SessionsDir: t.TempDir()})
	if err := svc.StartReplay(fixtureAorec(), "p"); err != nil {
		t.Fatalf("start: %v", err)
	}
	svc.PauseReplay()
	if svc.Status().ReplayPlaying {
		t.Fatal("should report paused")
	}
	svc.SetReplaySpeed(4)
	if sp := svc.Status().ReplaySpeed; sp != 4 {
		t.Fatalf("speed = %v, want 4", sp)
	}
	svc.ResumeReplay()
	if !svc.Status().ReplayPlaying {
		t.Fatal("should report playing after resume")
	}
	svc.SetReplaySpeed(0)
	svc.StopReplay()
	// Give the goroutine a moment to unwind.
	deadline := time.Now().Add(10 * time.Second)
	for svc.Status().Replay {
		if time.Now().After(deadline) {
			t.Fatal("replay did not stop")
		}
		time.Sleep(20 * time.Millisecond)
	}
}
