package core

import (
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
)

type tasks []Task

type event struct {
	When time.Time
	Type string
	Msg  string
}

/*
	TODO doc orthogonal state spaces
		Deleted
		Unestimated | Estimated
		NeverStarted | Started | Done
			but, NeverStarted == ActualUpdatedAt.IsZero()
*/

// Task is a wrapper around task. Illegal state is highly representable in a task,
// and some task state can be updated only in the context of a collection of other
// tasks. This wrapper is an experiment in information hiding, so we can guarantee
// only legal task state. The root cause here is that task fields must be exported
// to be automatically serializeable, so this wrapper tries to give us both
// a nice API and easy serialization.
type Task struct {
	task task
}

func NewTask() *Task {
	return &Task{task: newTask()}
}

// LogActual logs actual time against this task. Most tasks should use auto time
// tracking. LogActual() provides an escape hatch for auto time tracking edge cases.
func (t *Task) LogActual(d time.Duration) {
	t.task.Actual += d
	t.task.ActualUpdatedAt = time.Now() // TODO it's unclear to me if now should be injected. I.e. for business hours tracking, we typically don't want to consider "nows" outside of business hours. But I don't think that extends to ActualUpdatedAt; I think business hours are completely ignored outside of auto time tracking and this should always just be time.Now().
}

func (t *Task) ID() uuid.UUID {
	return t.task.ID
}

func (t *Task) Name() string {
	return t.task.Name
}

func (t *Task) SetName(n string) {
	// TODO setting task name should trim whitespace and have a maximum name length. same for shortname.
	t.task.Name = n
}

func (t *Task) CreatedAt() time.Time {
	return t.task.CreatedAt
}

func (t *Task) IsEstimated() bool {
	return t.task.Estimated != 0
}

func (t *Task) IsNeverStarted() bool {
	return t.task.ActualUpdatedAt.IsZero()
}

func (t *Task) IsStarted() bool {
	return !t.IsNeverStarted() && !t.task.IsDone
}

func (t *Task) IsDone() bool {
	return !t.IsNeverStarted() && t.task.IsDone
}

func (t *Task) IsDeleted() bool {
	return t.task.IsDeleted
}

func (t *Task) Delete() error {
	if t.IsStarted() {
		return errors.New("cannot delete a started task")
	}
	if t.IsDeleted() {
		return errors.New("task is already deleted")
	}
	t.task.IsDeleted = true
	t.task.DeletedAt = time.Now()
	return nil
}

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

func (t *Task) EstimatedHours() float64 {
	return t.task.Estimated.Hours()
}

// ActualHours is the sum of elapsed time spent on this task for start-stop intervals.
func (t *Task) ActualHours() float64 {
	return t.task.Actual.Hours()
}

// EstimateAccuracyRatio returns a ratio of estimate / actual hours for a done task.
// I.e. 1.0 is perfect estimate, 2.0 means task was twice as fast, 0.5 task twice as long.
func (t *Task) EstimateAccuracyRatio() float64 {
	if t.ActualHours() == 0 {
		return 0
	}
	// Canonically we want an accuracy ratio for only done tasks, but we'll allow computing it on any task, because it's simple and may be useful.
	return t.EstimatedHours() / t.ActualHours()
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

// task is the unit of estimation for est.
// Users estimate and do tasks, and then est predicts future tasks' delivery schedule.
// A task is the same thing as a story, feature, bug, etc.
type task struct {
	ID              uuid.UUID
	Name            string
	Events          []event       // event log to show history to humans
	Estimated       time.Duration // estimated hours for this task (as estimated by a human)
	Actual          time.Duration // actual hours logged for this task
	ActualUpdatedAt time.Time     // ActualUpdatedAt is last time this task had hours logged. This task was never stared iff ActualUpdatedAt is zero.
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

// RenderTaskOneLineSummary returns a string rendering of passed task
// suitable to be included in a one-task-per-line output to user.
func RenderTaskOneLineSummary(t *Task) string {
	idPrefix := t.task.ID.String()[0:5]      // TODO dynamic length of ID prefix based on uniqueness of all task IDs. (Inject IDPrefixLen)
	_, month, day := t.task.CreatedAt.Date() // TODO this should be "last updated at"; maybe we have actually an UpdatedAt or dynamically select from latest of the dates
	estimate := t.EstimatedHours()           // TODO this should be nice format "12h, 2d"; maybe represent estimate as a Duration
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
		status = fmt.Sprintf("done in %.1fh", t.ActualHours())
	case taskStatusStarted:
		status = "started"
	case taskStatusEstimated:
		status = "estimated"
	}

	return fmt.Sprintf("%s\t%d/%d\t%.1fh\t%s\t%s", idPrefix, month, day, estimate, name, status)
}

// Start the ith task of tasks. When a task transitions to or from started,
// auto time tracking is updated for all started tasks. Auto time tracking is
// relative to a set of tasks, so that multiple tasks in progress share the
// passage of time.
// TODO what is signature of start? We must consider at least an injected time.Now() and also business hours to consider for auto time tracking.
func (ts tasks) Start(i int) error {
	t := &ts[i]
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

	// TODO set StartedAt

	return nil
}

// Mark the ith task of tasks as done. See note on Start().
func (ts tasks) Done(i int) error {
	t := &ts[i]
	if !t.IsStarted() {
		return errors.New("cannot mark done a task which isn't started")
	}
	if t.IsDeleted() {
		// We don't allow starting deleted tasks or deleting a started task, and so never expect a started task to be deleted.
		panic("expected started to be not deleted")
	}

	// TODO impl

	// TODO set DoneAt

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

func (ts tasks) SortByStartedAtDescending() tasks {
	sort.Sort(sortByStartedAtDescending(ts))
	return ts
}

func (ts tasks) SortByStatusDescending() tasks {
	sort.Sort(sortByStatusDescending(ts))
	return ts
}

func searchTasks(ts []Task, fn func(t *Task) bool) int {
	for i := range ts {
		if fn(&ts[i]) {
			return i
		}
	}
	return -1
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
