package main

import (
	"time"

	"github.com/google/uuid"
)

type tasks []task

type task struct {
	ID             uuid.UUID
	CreatedAt      time.Time
	EstimatedAt    time.Time
	EstimatedHours float64
	Timeline       []time.Time // one Time per start, stop, start, stop, ... see isDone()
	IsDeleted      bool
}

func newTask() *task {
	return &task{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
	}
}

func (t *task) isDone() bool {
	return len(t.Timeline) > 0 && len(t.Timeline)%2 == 0
}

func (ts tasks) notDeleted() tasks {
	return filterTasks(ts, func(t *task) bool {
		return !t.IsDeleted
	})
}

func (ts tasks) notDone() tasks {
	return filterTasks(ts, func(t *task) bool {
		return !t.isDone()
	})
}

func (ts tasks) notStarted() tasks {
	return filterTasks(ts, func(t *task) bool {
		return len(t.Timeline) == 0
	})
}

func (ts tasks) estimated() tasks {
	return filterTasks(ts, func(t *task) bool {
		return t.EstimatedHours > 0
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
