package core

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/ryanberckmans/est/core/worktimes"
)

// const ansiBoldGreen = "\033[92m"
// const ansiBoldBlue = "\033[94m"
// const ansiBoldCyan = "\033[96m"
// const ansiBoldWhite = "\033[97m"

const ansiReset = "\033[0m"
const ansiBold = "\033[1m"
const ansiBoldRed = "\033[92m"
const ansiBoldYellow = "\033[93m"
const ansiBoldMagenta = "\033[95m"
const ansiDangerOrange = "\033[38;5;202m" // color 202 of 256

type tasks []*Task

// TODO make use of event
type event struct {
	When time.Time
	Type string
	Msg  string
}

// Task is a wrapper around task, preventing illegal state and state
// transitions. Some task state can be updated only in the context
// of a collection of other tasks. The root cause here is that task
// fields must be exported to be automatically serializeable, so this
// wrapper tries to give us both a nice API and easy serialization.
type Task struct {
	task task
}

// NewTask returns a new Task.
func NewTask() *Task {
	return &Task{task: newTask()}
}

// ID returns this task's ID.
func (t *Task) ID() uuid.UUID {
	return t.task.ID
}

// Name returns this task's name.
func (t *Task) Name() string {
	return t.task.Name
}

const taskNameMaxLen = 120

// SetName sets this task's name.
func (t *Task) SetName(n string) error {
	n2 := strings.TrimSpace(n)
	if n2 == "" {
		return errors.New("task name cannot be empty")
	}
	if len(n2) > taskNameMaxLen {
		return fmt.Errorf("task name can be at most %d characters", taskNameMaxLen)
	}
	t.task.Name = n2
	return nil
}

// IsEstimated returns true iff this task has a non-zero estimated duration.
func (t *Task) IsEstimated() bool {
	return t.task.Estimated != 0
}

// IsNeverStarted returns true iff this task was never started.
func (t *Task) IsNeverStarted() bool {
	return t.task.ActualUpdatedAt.IsZero()
}

// IsStarted returns true iff this task is currently started.
func (t *Task) IsStarted() bool {
	return !t.IsNeverStarted() && !t.task.IsDone && !t.task.IsPaused
}

// IsPaused returns true iff this task is currently paused.
func (t *Task) IsPaused() bool {
	return !t.IsNeverStarted() && !t.task.IsDone && t.task.IsPaused
}

// IsDone returns true iff this task is currently done.
func (t *Task) IsDone() bool {
	return !t.IsNeverStarted() && t.task.IsDone
}

// IsDeleted returns true iff this task is currently deleted.
func (t *Task) IsDeleted() bool {
	return t.task.IsDeleted
}

// Delete this task.
func (t *Task) Delete() error {
	if t.IsStarted() {
		return errors.New("cannot delete task which is started")
	}
	if t.IsDeleted() {
		return errors.New("task is already deleted")
	}
	t.task.IsDeleted = true
	t.task.DeletedAt = time.Now()
	return nil
}

// Undelete this task.
func (t *Task) Undelete() error {
	if !t.IsDeleted() {
		return errors.New("cannot undelete task which isn't deleted")
	}
	if t.IsStarted() {
		// We don't allow deleting started tasks, and so expect deleted tasks to be unstarted.
		panic("expected task to be unstarted")
	}
	t.task.IsDeleted = false
	return nil
}

// Estimated returns the estimated duration for this task.
func (t *Task) Estimated() time.Duration {
	return t.task.Estimated
}

// SetEstimated sets this task's estimated duration.
func (t *Task) SetEstimated(d time.Duration) error {
	if !t.IsNeverStarted() {
		return errors.New("cannot re-estimate a task which has been started")
	}
	t.task.Estimated = d
	t.task.EstimatedAt = time.Now()
	return nil
}

// Actual returns the actual duration elapsed for this task.
func (t *Task) Actual() time.Duration {
	return t.task.Actual
}

// AddActual logs actual time spent against this task. Most
// tasks should use auto time tracking. AddActual() provides
// an escape hatch for auto time tracking edge cases.
func (t *Task) AddActual(d time.Duration, now time.Time) error {
	if t.IsNeverStarted() {
		return errors.New("cannot add actual time to a task which has never been started")
	}
	t.task.Actual += d
	t.task.ActualUpdatedAt = now
	return nil
}

// EstimateAccuracyRatio returns a ratio of estimate / actual hours for a done task.
// I.e. 1.0 is perfect estimate, 2.0 means task was twice as fast, 0.5 task twice as long.
func (t *Task) EstimateAccuracyRatio() AccuracyRatio {
	return AccuracyRatio{
		time:     t.DoneAt(),
		duration: t.Estimated(),
		ratio:    t.estimateAccuracyRatio(),
	}
}

func (t *Task) estimateAccuracyRatio() float64 {
	if t.Actual() == 0 {
		// Downstream expects non-zero ratios: panic so that over time all
		// clients of EstimateAccuracyRatio() will ensure non-zero Actual.
		panic("expected non-zero task.Actual")
	}
	// Canonically we want an accuracy ratio for only done tasks, but we'll allow computing it on any task, because it's simple and may be useful.
	return t.Estimated().Hours() / t.Actual().Hours()
}

// CreatedAt returns the time at which this task was created.
func (t *Task) CreatedAt() time.Time {
	return t.task.CreatedAt
}

// EstimatedAt returns the most recent time at which this task was estimated.
func (t *Task) EstimatedAt() time.Time {
	return t.task.EstimatedAt
}

// StartedAt returns the most recent time at which this task was started.
func (t *Task) StartedAt() time.Time {
	return t.task.StartedAt
}

// PausedAt returns the most recent time at which this task was paused.
func (t *Task) PausedAt() time.Time {
	return t.task.PausedAt
}

// DoneAt returns the most recent time at which this task was marked done.
func (t *Task) DoneAt() time.Time {
	return t.task.DoneAt
}

// DeletedAt returns the most recent time at which this task was deleted.
func (t *Task) DeletedAt() time.Time {
	return t.task.DeletedAt
}

// Return this task's status and date of that status.
func (t *Task) status() (taskStatus, time.Time) {
	switch {
	// The order of these cases is significant. Deleted is orthogonal to Done | Started; both are orthogonal to Estimated | Unestimated.
	case t.task.IsDeleted:
		return taskStatusDeleted, t.DeletedAt()
	case t.IsDone():
		return taskStatusDone, t.DoneAt()
	case t.IsPaused():
		return taskStatusPaused, t.PausedAt()
	case t.IsStarted():
		return taskStatusStarted, t.StartedAt()
	case t.IsEstimated():
		return taskStatusEstimated, t.EstimatedAt()
	}
	return taskStatusUnestimated, t.CreatedAt()
}

// taskStatus is transient: task doesn't actually have a field
// taskStatus, because taskStatus is a projection of orthogonal
// state into one dimension. It helps explain task state to humans.
type taskStatus int

const (
	taskStatusDeleted = iota
	taskStatusDone
	taskStatusEstimated
	taskStatusPaused
	taskStatusStarted
	taskStatusUnestimated
)

// RenderYesterdayTasks returns a user-suitable summary of task activity on
// first business day prior to now.
func RenderYesterdayTasks(wt worktimes.WorkTimes, ts tasks, now time.Time) string {
	for {
		// Find previous business day by searching for first day in past
		// with some worktimes. Will never terminate if wt has no worktimes.
		now = now.Add(-time.Hour * 24)
		if len(wt.GetWorkTimesOnDay(now)) > 0 {
			break
		}
	}
	start := worktimes.StartOfDay(now)
	end := worktimes.EndOfDay(now)
	var ts2 tasks
	for _, t := range ts {
		_, taskTime := t.status()
		if t.IsStarted() || taskTime.After(start) && taskTime.Before(end) {
			ts2 = append(ts2, t)
		}
	}
	rs := make([]string, len(ts2)+2) // +1 causes the last element to be empty string, which causes the Join to add an extra newline
	rs[0] = "Activity on " + now.Format("Monday, January 2, 2006")
	for i := range ts2 {
		rs[i+1] = RenderTaskOneLineSummary(ts2[i], i == 0)
	}
	return strings.Join(rs, "\n")
}

// RenderTaskOneLineSummary returns a string rendering of passed task
// suitable to be included in a one-task-per-line output to user.
func RenderTaskOneLineSummary(t *Task, includeHeaders bool) string {
	var status string
	statusCode, statusTime := t.status()
	_, month, day := statusTime.Date()
	switch statusCode {
	case taskStatusDeleted:
		// right padding is to align table because "deleted" is a shorter word
		status = fmt.Sprintf("%sdeleted%s on %d/%d   ", ansiBold+ansiBoldRed, ansiReset, month, day)
	case taskStatusDone:
		if t.Actual() < time.Minute {
			status = fmt.Sprintf("done in %.1fs on %d/%d", t.Actual().Seconds(), month, day)
		} else if t.Actual() < time.Hour {
			status = fmt.Sprintf("done in %.1fm on %d/%d", t.Actual().Minutes(), month, day)
		} else {
			status = fmt.Sprintf("done in %.1fh on %d/%d", t.Actual().Hours(), month, day)
		}
	case taskStatusStarted:
		// right padding is to align table because "started" is a shorter word
		status = fmt.Sprintf("%sstarted%s on %d/%d   ", ansiBold+ansiBoldYellow, ansiReset, month, day)
	case taskStatusPaused:
		// right padding is to align table because "paused" is a shorter word
		status = fmt.Sprintf("%spaused%s on %d/%d   ", ansiBold+ansiBoldMagenta, ansiReset, month, day)
	case taskStatusEstimated:
		status = fmt.Sprintf("estimated on %d/%d", month, day)
	default:
		status = fmt.Sprintf("%sunestimated%s on %d/%d", ansiBold+ansiDangerOrange, ansiReset, month, day)
	}
	var headers string
	if includeHeaders {
		headers = "STATUS\t\t\t\tESTIMATE\tID\tNAME\n"
	}
	return fmt.Sprintf("%s%s\t\t%.1fh\t\t%s\t%s",
		headers,
		status,
		t.Estimated().Hours(),
		t.task.ID.String()[0:5], // TODO dynamic length of ID prefix based on uniqueness of all task IDs. (Inject IDPrefixLen)
		t.Name(),                // name has arbitrary length and so is last
	)
}

// task is the unit of estimation for est. Users estimate and do
// tasks, and then est predicts future tasks' delivery schedule.
// A task is the same thing as a story, feature, bug, etc.
type task struct {
	ID              uuid.UUID
	Name            string
	Events          []event       // event log to show history to humans  TODO generate event log
	Estimated       time.Duration // estimated duration for this task (as estimated by a human)
	Actual          time.Duration // actual duration spent on this task
	ActualUpdatedAt time.Time     // ActualUpdatedAt is last time this task had time logged. This task was never started iff ActualUpdatedAt is zero.
	IsPaused        bool          // if ActualUpdatedAt is zero or IsDone is true, IsPaused is undefined. Otherwise, this task is paused iff IsPaused.
	IsDone          bool          // if ActualUpdatedAt is zero, IsDone is undefined. Otherwise, this task is done if IsDone else this task is started.
	IsDeleted       bool          // this task is deleted iff IsDeleted; orthogonal to other task state.

	// These times aren't needed for tasks to work properly; they exist to
	// show to humans.
	// Each time is the most recent time at which the thing occurred.
	CreatedAt   time.Time
	EstimatedAt time.Time
	StartedAt   time.Time
	PausedAt    time.Time
	DoneAt      time.Time
	DeletedAt   time.Time
}

func newTask() task {
	return task{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
	}
}

func toExportedTasks(ts []task) tasks {
	ts2 := make(tasks, len(ts))
	for i := range ts {
		// WARNING there is still shared memory allocation between ts2[i] and ts[i], e.g. []event
		ts2[i] = &Task{task: ts[i]}
	}
	return ts2
}

func toUnexportedTasks(ts tasks) []task {
	ts2 := make([]task, len(ts))
	for i := range ts {
		// WARNING there is still shared memory allocation between ts[i] and ts[i], e.g. []event
		ts2[i] = ts[i].task
	}
	return ts2
}

// Start the ith task of tasks. When a task transitions to or from started,
// auto time tracking is updated for all started tasks. Auto time tracking is
// relative to a set of tasks, so that multiple tasks in progress share the
// passage of time.
func (ts tasks) Start(wt worktimes.WorkTimes, i int, now time.Time) error {
	t := ts[i]
	if t.IsDeleted() {
		return errors.New("cannot start deleted task")
	}
	if !t.IsEstimated() {
		return errors.New("cannot start unestimated task")
	}
	if t.IsStarted() {
		return errors.New("cannot start task which is already started")
	}

	// Auto track time against current tasks in progress. This must be done prior to
	// starting i'th task, because shared passage of time for current started tasks
	// must exclude this newly started task (as it wasn't auto ticking until now).
	autoAddActual(wt, ts.IsStarted().IsNotDeleted(), now) // IsNotDeleted is sanity because we expect started tasks to never be deleted
	t.task.ActualUpdatedAt = now
	t.task.StartedAt = now
	t.task.IsPaused = false
	t.task.IsDone = false

	return nil
}

// Pause the ith task of tasks. See note on Start().
func (ts tasks) Pause(wt worktimes.WorkTimes, i int, now time.Time) error {
	t := ts[i]
	if !t.IsStarted() {
		return errors.New("cannot pause a task which isn't started")
	}
	if t.IsDeleted() {
		// We don't allow starting deleted tasks or deleting a started task, and so never expect a started task to be deleted.
		panic("expected started to be not deleted")
	}

	// Auto track time against current tasks in progress. This must
	// be done prior to pausing i'th task, because shared
	// passage of time for current started tasks must include this
	// previously started task (so all started tasks tick together).
	autoAddActual(wt, ts.IsStarted().IsNotDeleted(), now) // IsNotDeleted is sanity because we expect started tasks to never be deleted
	// We don't set t.ActualUpdatedAt because it's been set inside autoAddActual() XOR ActualUpdatedAt is in the future and shouldn't be overwritten.
	t.task.PausedAt = now
	t.task.IsPaused = true

	return nil
}

// Mark the ith task of tasks as done. See note on Start().
func (ts tasks) Done(wt worktimes.WorkTimes, i int, now time.Time) error {
	t := ts[i]
	if !t.IsStarted() && !t.IsPaused() {
		return errors.New("cannot mark done a task which isn't started or paused")
	}
	if t.IsDeleted() {
		// We don't allow starting deleted tasks or deleting a started task, and so never expect a started task to be deleted.
		panic("expected started to be not deleted")
	}

	// Auto track time against current tasks in progress. This must
	// be done prior to marking done the i'th task, because shared
	// passage of time for current started tasks must include this
	// previously started task (so all started tasks tick together).
	autoAddActual(wt, ts.IsStarted().IsNotDeleted(), now) // IsNotDeleted is sanity because we expect started tasks to never be deleted
	// We don't set t.ActualUpdatedAt because it's been set inside autoAddActual() XOR ActualUpdatedAt is in the future and shouldn't be overwritten.
	t.task.DoneAt = now
	t.task.IsPaused = false
	t.task.IsDone = true

	return nil
}

// TODO unit tests
func autoAddActual(wt worktimes.WorkTimes, ts tasks, end time.Time) {
	if len(ts) < 1 {
		return
	}
	for i := range ts {
		if !ts[i].IsStarted() {
			panic("sanity: expected ts to be all started")
		}
	}

	ts = ts.sortByActualUpdatedAtAscending()

	for {
		lowest := ts[0].task.ActualUpdatedAt
		if !lowest.Before(end) {
			// All tasks' ActualUpdatedAt is >= end, i.e. job is done
			return
		}
		var ts2 tasks
		for i := 0; i < len(ts) && lowest.Equal(ts[i].task.ActualUpdatedAt); i++ {
			ts2 = append(ts2, ts[i])
		}

		if len(ts2) < 1 {
			panic("sanity: expected ts2 to be non-empty")
		}

		// ts2 is now the tasks which have lowest lastUpdatedAt and lastUpdatedAt
		// < end. We will auto track shared passage of actual time for these tasks.
		// Each task will get time at a rate of 1/len(ts2) vs. real time. To properly
		// share the passage of time, we tick a cohort of tasks with same start time
		// to the same end time. The start time here is `lowest`. The end time here
		// is the lowest time in ts which is after `lowest`, or `end` if none exists.

		var nextEnd time.Time
		for i := range ts {
			if ts[i].task.ActualUpdatedAt.After(lowest) {
				nextEnd = ts[i].task.ActualUpdatedAt
				break
			}
		}
		if nextEnd.IsZero() {
			// i.e. ts2 is all the tasks in ts.
			if len(ts2) != len(ts) {
				panic(fmt.Sprintf("sanity: expected len(ts2)==len(ts) because nextEnd == end, len(ts2)==%d, len(ts)==%d", len(ts2), len(ts)))
			}
			nextEnd = end
		}

		// We'll now tick ts2 in shared passage of time. ts2's start time is
		// the same and lowest of ts. nextEnd is the next lowest time after ts2.

		autoActual := wt.DurationBetween(lowest, nextEnd) // auto time tracking includes workin hours only, otherwise weekends, sleep, etc., would count as time on task.
		autoActualShared := autoActual / time.Duration(len(ts2))
		// fmt.Printf("count=%d lowest=%v nextEnd=%v autoActual=%v autoActualShared=%v end=%v\n", len(ts2), lowest, nextEnd, autoActual, autoActualShared, end)
		for i := range ts2 {
			err := ts2[i].AddActual(autoActualShared, nextEnd)
			if err != nil {
				panic(err)
			}
		}
	}
}

// FindByIDPrefix returns the index of the first task to match passed Task.ID prefix.
// Returns -1 if no task found.
func (ts tasks) FindByIDPrefix(prefix string) int {
	if prefix == "" {
		return -1
	}
	return searchTasks(ts, func(t *Task) bool {
		return strings.HasPrefix(t.ID().String(), prefix)
	})
}

func (ts tasks) IsNotDeleted() tasks {
	return filterTasks(ts, func(t *Task) bool {
		return !t.IsDeleted()
	})
}

func (ts tasks) IsDone() tasks {
	return filterTasks(ts, func(t *Task) bool {
		return t.IsDone()
	})
}

func (ts tasks) IsNotDone() tasks {
	return filterTasks(ts, func(t *Task) bool {
		return !t.IsDone()
	})
}

func (ts tasks) IsEstimated() tasks {
	return filterTasks(ts, func(t *Task) bool {
		return t.IsEstimated()
	})
}

func (ts tasks) IsStarted() tasks {
	return filterTasks(ts, func(t *Task) bool {
		return t.IsStarted()
	})
}

func (ts tasks) IsNotStarted() tasks {
	return filterTasks(ts, func(t *Task) bool {
		return !t.IsStarted()
	})
}

func (ts tasks) IsNonZeroActual() tasks {
	return filterTasks(ts, func(t *Task) bool {
		return t.Actual() != 0
	})
}

func (ts tasks) SortByStartedAtDescending() tasks {
	sort.Sort(sortByStartedAtDescending(ts))
	return ts
}

func (ts tasks) SortByStatusDescending() tasks {
	sort.Sort(sortByStatusDescending(ts))
	return ts
}

func (ts tasks) sortByActualUpdatedAtAscending() tasks {
	sort.Sort(sortByActualUpdatedAtAscending(ts))
	return ts
}

func searchTasks(ts tasks, fn func(t *Task) bool) int {
	for i := range ts {
		if fn(ts[i]) {
			return i
		}
	}
	return -1
}

func filterTasks(ts tasks, fn func(t *Task) bool) tasks {
	if ts == nil {
		return nil
	}
	var res tasks
	for i := range ts {
		if fn(ts[i]) {
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
	return ts[i].task.StartedAt.After(ts[j].task.StartedAt)
}
func (ts sortByStartedAtDescending) Swap(i, j int) {
	tmp := ts[j]
	ts[j] = ts[i]
	ts[i] = tmp
}

type sortByStatusDescending tasks

func (ts sortByStatusDescending) Len() int {
	return len(ts)
}
func (ts sortByStatusDescending) Less(i, j int) bool {
	iStatus, iStatusDate := ts[i].status()
	jStatus, jStatusDate := ts[j].status()
	if iStatus != jStatus {
		return iStatus > jStatus
	}
	return iStatusDate.After(jStatusDate)
}
func (ts sortByStatusDescending) Swap(i, j int) {
	tmp := ts[j]
	ts[j] = ts[i]
	ts[i] = tmp
}

type sortByActualUpdatedAtAscending tasks

func (ts sortByActualUpdatedAtAscending) Len() int {
	return len(ts)
}
func (ts sortByActualUpdatedAtAscending) Less(i, j int) bool {
	return ts[i].task.ActualUpdatedAt.Before(ts[j].task.ActualUpdatedAt)
}
func (ts sortByActualUpdatedAtAscending) Swap(i, j int) {
	tmp := ts[j]
	ts[j] = ts[i]
	ts[i] = tmp
}
