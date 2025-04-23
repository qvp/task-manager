package api

import (
	"github.com/google/uuid"
	"task/task"
)

type TaskResponse struct {
	ID     uuid.UUID
	Data   string
	Result string
	Status task.Status
}

type TaskCreateRequest struct {
	Data string
}

type TaskCreateResponse struct {
	ID uuid.UUID
}
