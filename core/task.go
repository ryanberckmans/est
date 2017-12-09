package core

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

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

const taskNameMaxLen = 80

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
	return !t.IsNeverStarted() && !t.task.IsDone
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
		return errors.New("cannot re-estimate a task which has been started; the idea is to get better at up-front estimation")
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
func (t *Task) AddActual(d time.Duration) error {
	if t.IsNeverStarted() {
		return errors.New("cannot add actual time to a task which has never been started")
	}
	t.task.Actual += d
	t.task.ActualUpdatedAt = time.Now() // TODO it's unclear to me if now should be injected. I.e. for business hours tracking, we typically don't want to consider "nows" outside of business hours. But I don't think that extends to ActualUpdatedAt; I think business hours are completely ignored outside of auto time tracking and this should always just be time.Now().
	return nil
}

// EstimateAccuracyRatio returns a ratio of estimate / actual hours for a done task.
// I.e. 1.0 is perfect estimate, 2.0 means task was twice as fast, 0.5 task twice as long.
func (t *Task) EstimateAccuracyRatio() float64 {
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

// DoneAt returns the most recent time at which this task was marked done.
func (t *Task) DoneAt() time.Time {
	return t.task.DoneAt
}

// DeletedAt returns the most recent time at which this task was deleted.
func (t *Task) DeletedAt() time.Time {
	return t.task.DeletedAt
}

func (t *Task) status() taskStatus {
	switch {
	// The order of these cases is significant. Deleted is orthogonal to Done | Started; both are orthogonal to Estimated | Unestimated.
	case t.task.IsDeleted:
		return taskStatusDeleted
	case t.IsDone():
		return taskStatusDone
	case t.IsStarted():
		return taskStatusStarted
	case t.IsEstimated():
		return taskStatusEstimated
	}
	return taskStatusUnestimated
}

// task doesn't actually have a field taskStatus, because taskStatus is a projection
// of orthogonal state into one dimension. It helps explain task state to humans.
type taskStatus int

const (
	taskStatusDeleted = iota
	taskStatusDone
	taskStatusEstimated
	taskStatusStarted
	taskStatusUnestimated
)

// RenderTaskOneLineSummary returns a string rendering of passed task
// suitable to be included in a one-task-per-line output to user.
func RenderTaskOneLineSummary(t *Task) string {
	idPrefix := t.task.ID.String()[0:5]      // TODO dynamic length of ID prefix based on uniqueness of all task IDs. (Inject IDPrefixLen)
	_, month, day := t.task.CreatedAt.Date() // TODO this should be "last updated at"; maybe we have actually an UpdatedAt or dynamically select from latest of the dates
	estimate := t.Estimated().Hours()        // TODO this should be nice format "12h, 2d"; maybe represent estimate as a Duration
	// TODO replace with something nice, also use ShortName/Summary
	nameFixedWidth := 12
	lenRemaining := nameFixedWidth - len(t.task.Name)
	var name string
	if lenRemaining > 0 {
		name = t.task.Name
		for ; lenRemaining > 0; lenRemaining-- {
			name += " "
		}
	} else {
		name = t.task.Name[:nameFixedWidth]
	}

	status := "unestimated" // TODO danger orange in colors package :)
	switch t.status() {
	case taskStatusDeleted:
		status = "deleted"
	case taskStatusDone:
		status = fmt.Sprintf("done in %.1fh", t.Actual().Hours())
	case taskStatusStarted:
		status = "started"
	case taskStatusEstimated:
		status = "estimated"
	}

	return fmt.Sprintf("%s\t%d/%d\t%.1fh\t%s\t%s", idPrefix, month, day, estimate, name, status)
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
	ActualUpdatedAt time.Time     // ActualUpdatedAt is last time this task had time logged. This task was never stared iff ActualUpdatedAt is zero.
	IsDone          bool          // if ActualUpdatedAt is zero, IsDone is undefined. Otherwise, this task is done if IsDone else this task is started.
	IsDeleted       bool          // this task is deleted iff IsDeleted; orthogonal to other task state.

	// These times aren't needed for tasks to work properly; they exist to
	// show to humans.
	// Each time is the most recent time at which the thing occurred.
	CreatedAt   time.Time
	EstimatedAt time.Time
	StartedAt   time.Time
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
// TODO what is signature of start? We must consider at least an injected time.Now() and also business hours to consider for auto time tracking.
func (ts tasks) Start(i int) error {
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

	// TODO impl

	// Previous code for elapsed might be useful:
	// elapsed := math.Max(now.Sub(t.StartedAt).Hours(), 0) // disallow negative elapsed, which is philosophically interesting but produces invalid accuracy ratios.

	now := time.Now()
	t.task.ActualUpdatedAt = now
	t.task.StartedAt = now
	t.task.IsDone = false

	return nil
}

// Mark the ith task of tasks as done. See note on Start().
func (ts tasks) Done(i int) error {
	t := ts[i]
	if !t.IsStarted() {
		return errors.New("cannot mark done a task which isn't started")
	}
	if t.IsDeleted() {
		// We don't allow starting deleted tasks or deleting a started task, and so never expect a started task to be deleted.
		panic("expected started to be not deleted")
	}

	// TODO impl

	now := time.Now()
	// t.task.ActualUpdatedAt = now TODO this should be set to "now" but only after tracking time
	t.task.DoneAt = now
	t.task.IsDone = true

	return nil
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
	// TODO StartedAt
	return false
	// return ts[i].StartedAt.After(ts[j].StartedAt)
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
	return ts[i].status() > ts[j].status()
}
func (ts sortByStatusDescending) Swap(i, j int) {
	tmp := ts[j]
	ts[j] = ts[i]
	ts[i] = tmp
}
