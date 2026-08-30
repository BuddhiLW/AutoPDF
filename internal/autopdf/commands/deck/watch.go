// Copyright 2025 AutoPDF BuddhiLW
// SPDX-License-Identifier: Apache-2.0

package deck

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// DefaultPollInterval is how often a watched file is checked.
const DefaultPollInterval = 400 * time.Millisecond

// DefaultDebounce is how long a file must be quiet before a rebuild starts.
const DefaultDebounce = 250 * time.Millisecond

// watchState is the inert part of the loop: what was last seen, and when.
type watchState struct {
	fingerprint string
	pendingAt   time.Time
	pending     bool
	revision    uint64
}

// step decides what the loop does with one observation. It performs no I/O, so
// the debounce and supersession rules are testable without a clock or a disk.
//
// A rebuild is due only when the file has been quiet for `debounce`: an editor
// that writes in several chunks would otherwise start a build per chunk, and
// every one but the last would be superseded work.
func (state watchState) step(fingerprint string, now time.Time, debounce time.Duration) (watchState, bool) {
	if fingerprint != state.fingerprint {
		state.fingerprint = fingerprint
		state.pendingAt = now
		state.pending = true
		return state, false
	}
	if state.pending && !now.Before(state.pendingAt.Add(debounce)) {
		state.pending = false
		state.revision++
		return state, true
	}
	return state, false
}

// fingerprintFile returns a content hash, or "" when the file cannot be read.
// An unreadable file is a state to wait through, not an error: an editor
// replacing a file atomically makes it briefly absent.
func fingerprintFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	digest := sha256.Sum256(data)
	return hex.EncodeToString(digest[:])
}

// watchOptions configures one watch loop.
type watchOptions struct {
	// WatchPath is the file whose changes drive rebuilds.
	WatchPath string
	// Command, when set, runs before each rebuild to regenerate the spec.
	Command string
	// PollInterval and Debounce default to the constants above.
	PollInterval time.Duration
	Debounce     time.Duration
}

// rebuild is what a watch loop performs once the watched file settles.
type rebuild func(ctx context.Context, revision uint64) error

// runWatch polls WatchPath and calls onChange once per settled change, plus
// once at startup. It returns when ctx is cancelled.
func runWatch(ctx context.Context, options watchOptions, onChange rebuild) error {
	interval := options.PollInterval
	if interval <= 0 {
		interval = DefaultPollInterval
	}
	debounce := options.Debounce
	if debounce <= 0 {
		debounce = DefaultDebounce
	}

	state := watchState{fingerprint: fingerprintFile(options.WatchPath)}
	if err := runOnce(ctx, options, 1, onChange); err != nil {
		fmt.Fprintf(os.Stderr, "build failed: %v\n", err)
	}
	state.revision = 1

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case now := <-ticker.C:
			var due bool
			state, due = state.step(fingerprintFile(options.WatchPath), now, debounce)
			if !due {
				continue
			}
			if err := runOnce(ctx, options, state.revision, onChange); err != nil {
				// A failed build must not end the watch: the next save is
				// usually the fix.
				fmt.Fprintf(os.Stderr, "build failed: %v\n", err)
			}
		}
	}
}

func runOnce(ctx context.Context, options watchOptions, revision uint64, onChange rebuild) error {
	if strings.TrimSpace(options.Command) != "" {
		command := exec.CommandContext(ctx, "sh", "-c", options.Command)
		command.Stderr = os.Stderr
		if err := command.Run(); err != nil {
			return fmt.Errorf("regenerate command failed: %w", err)
		}
	}
	return onChange(ctx, revision)
}
