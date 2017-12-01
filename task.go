package main

import (
	"math"
	"time"

	"github.com/google/uuid"
)

type tasks []task

type task struct {
	// TODO do we want to allow tasks to be created without estimates? is this a backlog tool, then, too? I can see this being really useful, you have 4 tasks, want to record them all, and not yet sure what estimate will be for last one, because it depends on formers.
	/*
		Ideas

		$ est add  // -e, --estimate parameter is optional
		$ est est // estimate a task, if it already has an estimate it will fail unless you give --force

		instead of EstimatedAt, could have EstimateCount to keep track of flaky estimates
	*/
	ID        uuid.UUID
	Name      string
	CreatedAt time.Time
	StartedAt time.Time // If len(Hours) is 0, StartedAt is undefined, because an unestimated task cannot be started. (I.e. StartedAt is not orthogonal to Hours). Otherwise, this task is unstarted iff StartedAt.IsZero(), else this task is in progress as of StartedAt.
	// TODO DoneAt ?? not used for math but for backlog purposes. Maybe a simple event log. But, if DoneAt is buried in event log, will make it more difficult to sort by doneAt.
	Hours     []float64 // Hours[0] is the estimate for this task. [1] is time elapsed between initial start/stop. [N] is subsequent starts/stops. An alternative to `Hours []float64` is `Durations []time.Duration`, however our monte carlo algorithms use float64 and hours is easier to read in raw estfile.
	IsDeleted bool
}

func newTask() *task {
	return &task{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
	}
}

// TODO setting task name should trim whitespace and have a maximum name length. same for shortname.

// start this task, panicking if this would create illegal state.
func (t *task) start(now time.Time) {
	if !t.isEstimated() {
		panic("cannot start unestimated task") // an error would be better if this were a library
	}
	if t.isStarted() {
		panic("cannot start task which is already started")
	}
	t.StartedAt = now
}

// stop this task, panicking if this would create illegal state.
func (t *task) stop(now time.Time) {
	if !t.isStarted() {
		panic("cannot stop task which is unstarted")
	}
	elapsed := math.Max(now.Sub(t.StartedAt).Hours(), 0) // disallow negative elapsed, which is philosophically interesting but produces invalid accuracy ratios.
	t.Hours = append(t.Hours, elapsed)
	t.StartedAt = time.Time{}
}

func (t *task) isEstimated() bool {
	return len(t.Hours) > 0
}

func (t *task) isStarted() bool {
	return t.isEstimated() && !t.StartedAt.IsZero()
}

func (t *task) isDone() bool {
	return len(t.Hours) > 1 && t.StartedAt.IsZero()
}

func (t *task) estimatedHours() float64 {
	if len(t.Hours) < 1 {
		return 0
	}
	return t.Hours[0]
}

// actualHours is the sum of elapsed time spent on this task for start-stop intervals.
func (t *task) actualHours() float64 {
	if len(t.Hours) < 2 {
		return 0
	}
	var hours float64
	for i := 1; i < len(t.Hours); i++ {
		hours += t.Hours[i]
	}
	return hours
}

// estimateAccuracyRatio returns a ratio of estimate / actual hours for a done task.
// I.e. 1.0 is perfect estimate, 2.0 means task was twice as fast, 0.5 task twice as long.
func (t *task) estimateAccuracyRatio() float64 {
	if len(t.Hours) < 2 {
		// we need an estimate and elapsed time to calculate accuracy ratio
		return 0
	}
	// It's possible this task isStarted(), but we'll allow computing accuracy ratio on a previously-done task which was restarted, because it's simple and may be useful
	return t.estimatedHours() / t.actualHours()
}

func (ts tasks) notDeleted() tasks {
	return filterTasks(ts, func(t *task) bool {
		return !t.IsDeleted
	})
}

func (ts tasks) done() tasks {
	return filterTasks(ts, func(t *task) bool {
		return t.isDone()
	})
}

func (ts tasks) notDone() tasks {
	return filterTasks(ts, func(t *task) bool {
		return !t.isDone()
	})
}

func (ts tasks) estimated() tasks {
	return filterTasks(ts, func(t *task) bool {
		return t.isEstimated()
	})
}

func filterTasks(ts []task, fn func(t *task) bool) []task {
	if ts == nil {
		return nil
	}
	var res []task
	for i := range ts {
		if fn(&ts[i]) {
			res = append(res, ts[i])
		}
	}
	return res
}
