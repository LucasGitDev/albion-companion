package app

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LucasGitDev/albion-companion/internal/capture"
)

// History time machine: play a saved recording back through the live engine so
// every view re-hydrates as the run happened. Live capture is blocked while a
// replay runs. Replaying never saves a session (non-destructive).

const defaultReplaySpeed = 8.0

// clampSpeed keeps playback in a safe, always-paced range: never 0 (unbounded
// tight loop) and never above maxReplaySpeed.
func clampSpeed(speed float64) float64 {
	if speed <= 0 {
		return maxReplaySpeed
	}
	if speed > maxReplaySpeed {
		return maxReplaySpeed
	}
	return speed
}

// StartReplay resets the live state and begins playing the .aorec at `path`.
// `label` is shown in the replay banner. It is refused while capturing or while
// another replay is active.
func (s *Service) StartReplay(path, label string) error {
	path = strings.TrimSpace(path)
	if path == "" {
		return fmt.Errorf("replay: no file given")
	}
	first, last, count, err := capture.AorecSpan(path)
	if err != nil {
		return fmt.Errorf("replay: %w", err)
	}
	if count == 0 {
		return fmt.Errorf("replay: recording is empty")
	}

	s.mu.Lock()
	if s.cancel != nil {
		s.mu.Unlock()
		return fmt.Errorf("replay: stop live capture first")
	}
	if s.replaying {
		s.mu.Unlock()
		return fmt.Errorf("replay: a replay is already running")
	}
	if label == "" {
		label = "recording"
	}
	s.replayLabel = label
	s.replayPath = path
	s.replayFirstNano = first.UnixNano()
	s.replayLastNano = last.UnixNano()
	s.launchLocked(defaultReplaySpeed, time.Time{})
	s.mu.Unlock()
	return nil
}

// launchLocked starts (or restarts, for a seek) the replay playback goroutine.
// Caller holds s.mu and has set replayLabel/replayPath/first/last. skipUntil>0
// fast-forwards to that time with UI emits suppressed (a seek).
func (s *Service) launchLocked(speed float64, skipUntil time.Time) {
	ctx, cancel := context.WithCancel(context.Background())
	ctl := capture.NewReplayControl(clampSpeed(speed))
	ctl.OnProgress = func(at time.Time) { s.replayAtNano.Store(at.UnixNano()) }

	first := time.Unix(0, s.replayFirstNano)
	startAt := first
	if !skipUntil.IsZero() {
		startAt = skipUntil
	}
	done := make(chan struct{})
	s.replaying = true
	s.replayCtl = ctl
	s.replayCancel = cancel
	s.replayDone = done
	s.replayAtNano.Store(startAt.UnixNano())
	s.replaySilent.Store(!skipUntil.IsZero()) // suppress emits while fast-forwarding
	s.status = CaptureStatus{}
	path := s.replayPath

	// Rebuild engine state from zero (resetAll locks s.mu internally, so do it
	// after releasing — schedule via goroutine start below).
	go func() {
		defer close(done)
		s.resetAll()
		src := capture.ReplaySource{
			Path:      path,
			Control:   ctl,
			SkipUntil: skipUntil,
			OnReachLive: func() {
				s.replaySilent.Store(false)
				s.opts.Emit("capture:status", s.Status())
				s.opts.Emit("replay:changed", nil)
			},
		}
		s.opts.Emit("capture:status", s.Status())
		s.opts.Emit("replay:changed", nil)
		_ = s.eng.Run(ctx, src)
		s.mu.Lock()
		if s.replayCtl == ctl { // only the current generation clears state
			s.replaying = false
			s.replayCtl = nil
			s.replayCancel = nil
			s.replaySilent.Store(false)
		}
		s.mu.Unlock()
		s.opts.Emit("capture:status", s.Status())
		s.opts.Emit("replay:changed", nil)
	}()

	go func() {
		t := time.NewTicker(500 * time.Millisecond)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				s.opts.Emit("replay:progress", nil)
			}
		}
	}()
}

// SeekReplay jumps the current replay to elapsedSec from the start by rebuilding
// engine state up to that point (fast-forward) and resuming paced playback.
func (s *Service) SeekReplay(elapsedSec float64) {
	if elapsedSec < 0 {
		elapsedSec = 0
	}
	s.mu.Lock()
	if !s.replaying {
		s.mu.Unlock()
		return
	}
	total := time.Duration(s.replayLastNano-s.replayFirstNano) * time.Nanosecond
	target := time.Duration(elapsedSec * float64(time.Second))
	// Never seek all the way to EOF — that would end the replay and drop the
	// transport. Keep playback alive just before the end instead.
	if max := total - time.Second; target > max {
		target = max
	}
	if target < 0 {
		target = 0
	}
	speed := defaultReplaySpeed
	if s.replayCtl != nil {
		speed = s.replayCtl.Speed()
	}
	oldCtl, oldCancel, oldDone := s.replayCtl, s.replayCancel, s.replayDone
	skipUntil := time.Unix(0, s.replayFirstNano).Add(target)
	s.mu.Unlock()

	// Fully stop the current generation (engine.Run refuses concurrent runs) so
	// the fresh fast-forward starts on a clean engine.
	if oldCtl != nil {
		oldCtl.Stop()
	}
	if oldCancel != nil {
		oldCancel()
	}
	if oldDone != nil {
		<-oldDone
	}

	s.mu.Lock()
	s.launchLocked(speed, skipUntil)
	s.mu.Unlock()
}

// ReplaySession replays a saved session by its id, using the recording path
// stored in its meta. Errors if the session has no recording on disk.
func (s *Service) ReplaySession(id string) error {
	if s.store == nil {
		return fmt.Errorf("replay: history is disabled")
	}
	m, _, err := s.store.Load(id)
	if err != nil {
		return fmt.Errorf("replay: %w", err)
	}
	if m.RecordingPath == "" {
		return fmt.Errorf("replay: this session has no recording to replay")
	}
	label := m.Label
	if label == "" {
		label = m.Zone
	}
	return s.StartReplay(m.RecordingPath, label)
}

// PauseReplay / ResumeReplay / SetReplaySpeed drive the transport.
func (s *Service) PauseReplay() {
	s.mu.Lock()
	ctl := s.replayCtl
	s.mu.Unlock()
	if ctl != nil {
		ctl.Pause()
		s.opts.Emit("replay:changed", nil)
	}
}

func (s *Service) ResumeReplay() {
	s.mu.Lock()
	ctl := s.replayCtl
	s.mu.Unlock()
	if ctl != nil {
		ctl.Resume()
		s.opts.Emit("replay:changed", nil)
	}
}

// maxReplaySpeed bounds playback so a very high speed still paces off the
// recorded timing (never an unbounded tight loop).
const maxReplaySpeed = 64.0

func (s *Service) SetReplaySpeed(speed float64) {
	speed = clampSpeed(speed)
	s.mu.Lock()
	ctl := s.replayCtl
	s.mu.Unlock()
	if ctl != nil {
		ctl.SetSpeed(speed)
		s.opts.Emit("replay:changed", nil)
	}
}

// StopReplay cancels the current replay (if any) and wakes a paused one.
func (s *Service) StopReplay() {
	s.mu.Lock()
	cancel := s.replayCancel
	ctl := s.replayCtl
	s.mu.Unlock()
	if ctl != nil {
		ctl.Stop() // wake if paused
	}
	if cancel != nil {
		cancel()
	}
}
