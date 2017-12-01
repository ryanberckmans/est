package core

import (
	"math"
	"sort"
	"time"

	"github.com/google/uuid"
)

type tasks []Task

// Task is the unit of estimation for est.
// Users estimate and do tasks, and then est predicts future tasks' delivery schedule.
// A task is the same thing as a story, feature, bug, etc.
type Task struct {
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

func NewTask() *Task {
	return &Task{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
	}
}

// TODO setting task name should trim whitespace and have a maximum name length. same for shortname.

// Start this task, panicking if this would create illegal state.
func (t *Task) Start(now time.Time) {
	if !t.IsEstimated() {
		panic("cannot start unestimated task") // an error would be better if this were a library
	}
	if t.IsStarted() {
		panic("cannot start task which is already started")
	}
	t.StartedAt = now
}

// Stop this task, panicking if this would create illegal state.
func (t *Task) Stop(now time.Time) {
	if !t.IsStarted() {
		panic("cannot stop task which is unstarted")
	}
	elapsed := math.Max(now.Sub(t.StartedAt).Hours(), 0) // disallow negative elapsed, which is philosophically interesting but produces invalid accuracy ratios.
	t.Hours = append(t.Hours, elapsed)
	t.StartedAt = time.Time{}
}

func (t *Task) IsEstimated() bool {
	return len(t.Hours) > 0
}

func (t *Task) IsStarted() bool {
	return t.IsEstimated() && !t.StartedAt.IsZero()
}

func (t *Task) IsDone() bool {
	return len(t.Hours) > 1 && t.StartedAt.IsZero()
}

func (t *Task) EstimatedHours() float64 {
	if len(t.Hours) < 1 {
		return 0
	}
	return t.Hours[0]
}

// ActualHours is the sum of elapsed time spent on this task for start-stop intervals.
func (t *Task) ActualHours() float64 {
	if len(t.Hours) < 2 {
		return 0
	}
	var hours float64
	for i := 1; i < len(t.Hours); i++ {
		hours += t.Hours[i]
	}
	return hours
}

// EstimateAccuracyRatio returns a ratio of estimate / actual hours for a done task.
// I.e. 1.0 is perfect estimate, 2.0 means task was twice as fast, 0.5 task twice as long.
func (t *Task) EstimateAccuracyRatio() float64 {
	if len(t.Hours) < 2 {
		// we need an estimate and elapsed time to calculate accuracy ratio
		return 0
	}
	// It's possible this task isStarted(), but we'll allow computing accuracy ratio on a previously-done task which was restarted, because it's simple and may be useful
	return t.EstimatedHours() / t.ActualHours()
}

func (ts tasks) NotDeleted() tasks {
	return filterTasks(ts, func(t *Task) bool {
		return !t.IsDeleted
	})
}

func (ts tasks) Done() tasks {
	return filterTasks(ts, func(t *Task) bool {
		return t.IsDone()
	})
}

func (ts tasks) NotDone() tasks {
	return filterTasks(ts, func(t *Task) bool {
		return !t.IsDone()
	})
}

func (ts tasks) Estimated() tasks {
	return filterTasks(ts, func(t *Task) bool {
		return t.IsEstimated()
	})
}

func (ts tasks) Started() tasks {
	return filterTasks(ts, func(t *Task) bool {
		return t.IsStarted()
	})
}

func (ts tasks) NotStarted() tasks {
	return filterTasks(ts, func(t *Task) bool {
		return !t.IsStarted()
	})
}

func (ts tasks) SortByStartedAtDescending() tasks {
	sort.Sort(sortByStartedAtDescending(ts))
	return ts
}

func filterTasks(ts []Task, fn func(t *Task) bool) []Task {
	if ts == nil {
		return nil
	}
	var res []Task
	for i := range ts {
		if fn(&ts[i]) {
			res = append(res, ts[i])
		}
	}
	return res
}

type sortByStartedAtDescending tasks

func (ts sortByStartedAtDescending) Len() int {
	return len(ts)
}
func (ts sortByStartedAtDescending) Less(i, j int) bool {
	return ts[i].StartedAt.After(ts[j].StartedAt)
}
func (ts sortByStartedAtDescending) Swap(i, j int) {
	tmp := ts[j]
	ts[j] = ts[i]
	ts[i] = tmp
}
