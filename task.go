package main

import (
	"time"

	"github.com/google/uuid"
)

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
