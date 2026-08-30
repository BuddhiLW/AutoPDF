package deck

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAQuietFileNeverRebuilds(t *testing.T) {
	state := watchState{fingerprint: "a"}
	now := time.Now()
	for range 10 {
		var due bool
		now = now.Add(time.Second)
		state, due = state.step("a", now, DefaultDebounce)
		assert.False(t, due)
	}
	assert.Zero(t, state.revision)
}

func TestAChangeRebuildsOnceItSettles(t *testing.T) {
	start := time.Now()
	state := watchState{fingerprint: "a"}

	state, due := state.step("b", start, 250*time.Millisecond)
	assert.False(t, due, "a change starts the debounce, it does not build immediately")

	state, due = state.step("b", start.Add(100*time.Millisecond), 250*time.Millisecond)
	assert.False(t, due, "still inside the quiet period")

	state, due = state.step("b", start.Add(300*time.Millisecond), 250*time.Millisecond)
	assert.True(t, due)
	assert.Equal(t, uint64(1), state.revision)

	state, due = state.step("b", start.Add(900*time.Millisecond), 250*time.Millisecond)
	assert.False(t, due, "one change is one build")
	assert.Equal(t, uint64(1), state.revision)
}

func TestWritesInChunksProduceOneBuild(t *testing.T) {
	// An editor that saves in several writes would otherwise start a build per
	// write, and every one but the last would be superseded work.
	start := time.Now()
	state := watchState{fingerprint: "a"}
	builds := 0
	for index, fingerprint := range []string{"b", "c", "d", "d", "d"} {
		var due bool
		state, due = state.step(fingerprint, start.Add(time.Duration(index)*50*time.Millisecond), 250*time.Millisecond)
		if due {
			builds++
		}
	}
	assert.Zero(t, builds, "no build while the file is still moving")

	state, due := state.step("d", start.Add(time.Second), 250*time.Millisecond)
	assert.True(t, due)
	assert.Equal(t, uint64(1), state.revision, "the whole burst is one revision")
}

func TestRevisionsIncreaseMonotonically(t *testing.T) {
	now := time.Now()
	state := watchState{fingerprint: "a"}
	for index, fingerprint := range []string{"b", "c", "d"} {
		var due bool
		now = now.Add(time.Second)
		state, due = state.step(fingerprint, now, time.Millisecond)
		require.False(t, due)

		now = now.Add(time.Second)
		state, due = state.step(fingerprint, now, time.Millisecond)
		require.True(t, due)
		assert.Equal(t, uint64(index+1), state.revision)
	}
}

func TestAnUnreadableFileIsWaitedThroughNotTreatedAsAChange(t *testing.T) {
	assert.Empty(t, fingerprintFile(filepath.Join(t.TempDir(), "absent.json")))

	// An atomic save makes the file briefly absent, then present again with
	// its original content. That round trip must not look like two changes.
	state := watchState{fingerprint: "a"}
	now := time.Now()
	state, _ = state.step("", now, time.Hour)
	state, due := state.step("a", now.Add(time.Millisecond), time.Hour)
	assert.False(t, due)
	assert.Equal(t, "a", state.fingerprint)
}

func TestFingerprintTracksContentNotTimestamps(t *testing.T) {
	path := filepath.Join(t.TempDir(), "spec.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"a":1}`), 0o644))
	first := fingerprintFile(path)

	require.NoError(t, os.WriteFile(path, []byte(`{"a":1}`), 0o644))
	assert.Equal(t, first, fingerprintFile(path), "a touch with identical content is not a change")

	require.NoError(t, os.WriteFile(path, []byte(`{"a":2}`), 0o644))
	assert.NotEqual(t, first, fingerprintFile(path))
}

func TestRunWatchBuildsOnceAtStartupThenOnChange(t *testing.T) {
	path := filepath.Join(t.TempDir(), "spec.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"a":1}`), 0o644))

	var mutex sync.Mutex
	var revisions []uint64
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = runWatch(ctx, watchOptions{
			WatchPath:    path,
			PollInterval: 5 * time.Millisecond,
			Debounce:     10 * time.Millisecond,
		}, func(context.Context, uint64) error {
			mutex.Lock()
			defer mutex.Unlock()
			revisions = append(revisions, uint64(len(revisions)+1))
			return nil
		})
	}()

	require.Eventually(t, func() bool {
		mutex.Lock()
		defer mutex.Unlock()
		return len(revisions) == 1
	}, time.Second, 5*time.Millisecond, "a watch must build once at startup")

	require.NoError(t, os.WriteFile(path, []byte(`{"a":2}`), 0o644))
	require.Eventually(t, func() bool {
		mutex.Lock()
		defer mutex.Unlock()
		return len(revisions) == 2
	}, 2*time.Second, 5*time.Millisecond, "a change must produce a build")

	cancel()
	<-done
}

func TestABuildFailureDoesNotEndTheWatch(t *testing.T) {
	path := filepath.Join(t.TempDir(), "spec.json")
	require.NoError(t, os.WriteFile(path, []byte(`{"a":1}`), 0o644))

	var mutex sync.Mutex
	attempts := 0
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		_ = runWatch(ctx, watchOptions{
			WatchPath:    path,
			PollInterval: 5 * time.Millisecond,
			Debounce:     10 * time.Millisecond,
		}, func(context.Context, uint64) error {
			mutex.Lock()
			defer mutex.Unlock()
			attempts++
			return assert.AnError
		})
	}()

	require.Eventually(t, func() bool {
		mutex.Lock()
		defer mutex.Unlock()
		return attempts >= 1
	}, time.Second, 5*time.Millisecond)

	require.NoError(t, os.WriteFile(path, []byte(`{"a":2}`), 0o644))
	require.Eventually(t, func() bool {
		mutex.Lock()
		defer mutex.Unlock()
		return attempts >= 2
	}, 2*time.Second, 5*time.Millisecond, "the next save is usually the fix, so the watch must survive a failure")

	cancel()
	<-done
}

func TestWatchOptionsDefaultToTheSpec(t *testing.T) {
	parsed, err := parseRequest([]string{"talk.json"})
	require.NoError(t, err)
	assert.Equal(t, "talk.json", parsed.watch.WatchPath)

	parsed, err = parseRequest([]string{"talk.json", "watch=talk.org", "command=plato spec talk.org"})
	require.NoError(t, err)
	assert.Equal(t, "talk.org", parsed.watch.WatchPath)
	assert.Equal(t, "plato spec talk.org", parsed.watch.Command)
}

func TestDurationsAreValidated(t *testing.T) {
	_, err := parseRequest([]string{"talk.json", "poll=soon"})
	require.Error(t, err)
	_, err = parseRequest([]string{"talk.json", "debounce=-1s"})
	require.Error(t, err)

	parsed, err := parseRequest([]string{"talk.json", "poll=1s", "debounce=100ms"})
	require.NoError(t, err)
	assert.Equal(t, time.Second, parsed.watch.PollInterval)
	assert.Equal(t, 100*time.Millisecond, parsed.watch.Debounce)
}
