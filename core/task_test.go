package core

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/ryanberckmans/est/core/worktimes"
)

func getStartedTask() *Task {
	t := NewTask()
	t.task.ActualUpdatedAt = time.Now()
	t.task.Estimated = time.Minute * 7
	if !t.IsStarted() {
		panic("expected started task")
	}
	return t
}

func getPausedTask() *Task {
	t := getStartedTask()
	t.task.IsPaused = true
	if !t.IsPaused() {
		panic("expected paused task")
	}
	return t
}

func TestPause(t *testing.T) {
	t.Run("pause a started task", func(t *testing.T) {
		ts := tasks{getStartedTask()}
		assert.True(t, ts[0].IsStarted(), "sanity")
		now := time.Now().Add(time.Minute)
		wt := worktimes.GetAnonymousWorkTimes()
		assert.NoError(t, ts.Pause(wt, 0, now), "can pause started task")
		assert.Equal(t, ts[0].task.PausedAt, now)
		assert.Equal(t, ts[0].task.ActualUpdatedAt, now)
		assert.True(t, ts[0].IsPaused())
		assert.False(t, ts[0].IsStarted(), "sanity")
	})
	t.Run("disallowed on paused task", func(t *testing.T) {
		ts := tasks{getPausedTask()}
		assert.True(t, ts[0].IsPaused(), "sanity")
		assert.Error(t, ts[0].SetEstimated(time.Minute), "cannot re-estimate paused task")
		now := time.Now().Add(time.Minute)
		wt := worktimes.GetAnonymousWorkTimes()
		assert.Error(t, ts.Pause(wt, 0, now), "cannot pause paused task")
		assert.True(t, ts[0].IsPaused(), "still paused")
	})
	t.Run("start a paused task", func(t *testing.T) {
		ts := tasks{getPausedTask()}
		assert.True(t, ts[0].IsPaused(), "sanity")
		now := time.Now().Add(time.Minute)
		wt := worktimes.GetAnonymousWorkTimes()
		assert.NoError(t, ts.Start(wt, 0, now), "can start paused task")
		assert.Equal(t, ts[0].task.StartedAt, now)
		assert.Equal(t, ts[0].task.ActualUpdatedAt, now)
		assert.True(t, ts[0].IsStarted())
		assert.False(t, ts[0].IsPaused(), "sanity")
	})
	t.Run("mark done a paused task", func(t *testing.T) {
		ts := tasks{getPausedTask()}
		assert.True(t, ts[0].IsPaused(), "sanity")
		oldActualUpdatedAt := ts[0].task.ActualUpdatedAt
		now := time.Now().Add(time.Minute)
		wt := worktimes.GetAnonymousWorkTimes()
		assert.NoError(t, ts.Done(wt, 0, now), "can mark done paused task")
		assert.Equal(t, ts[0].task.DoneAt, now)
		assert.Equal(t, ts[0].task.ActualUpdatedAt, oldActualUpdatedAt, "actual updated at not updated because this task was paused and had no time tracked")
		assert.True(t, ts[0].IsDone())
		assert.False(t, ts[0].IsPaused(), "sanity")
	})
	t.Run("delete a paused task", func(t *testing.T) {
		ts := tasks{getPausedTask()}
		assert.True(t, ts[0].IsPaused(), "sanity")
		assert.NoError(t, ts[0].Delete(), "can delete paused task")
		assert.True(t, ts[0].IsDeleted())
		// ts[0].IsPaused is still true because IsDeleted is orthogonal state
	})
}
