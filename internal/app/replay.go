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
	ctx, cancel := context.WithCancel(context.Background())
	ctl := capture.NewReplayControl(defaultReplaySpeed)
	ctl.OnProgress = func(at time.Time) { s.replayAtNano.Store(at.UnixNano()) }
	s.replaying = true
	s.replayCtl = ctl
	s.replayCancel = cancel
	s.replayLabel = label
	s.replayFirstNano = first.UnixNano()
	s.replayLastNano = last.UnixNano()
	s.replayAtNano.Store(first.UnixNano())
	s.status = CaptureStatus{} // clear any stale live status
	s.mu.Unlock()

	// Fresh state so the replay builds from zero, like a live run.
	s.resetAll()
	s.opts.Emit("capture:status", s.Status())
	s.opts.Emit("replay:changed", nil)

	go func() {
		src := capture.ReplaySource{Path: path, Control: ctl}
		_ = s.eng.Run(ctx, src)
		s.mu.Lock()
		s.replaying = false
		s.replayCtl = nil
		s.replayCancel = nil
		s.mu.Unlock()
		s.opts.Emit("capture:status", s.Status())
		s.opts.Emit("replay:changed", nil)
	}()

	// Progress heartbeat so the transport's elapsed time advances smoothly even
	// between the engine's throttled change events.
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
	return nil
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

func (s *Service) SetReplaySpeed(speed float64) {
	if speed < 0 {
		speed = 0
	}
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
