package ports

import (
	"context"
	"task-management/internal/models"
)

type Repository interface {
	GetTasks(ctx context.Context, offset, limit int64, search string) ([]models.Task, int64, error)
	GetTaskById(ctx context.Context, id int64) (*models.Task, error)
	CreateTask(ctx context.Context, task *models.Task) error
	UpdateTask(ctx context.Context, task *models.Task) error
	DeleteTask(ctx context.Context, taskId int64) (*models.Task, error)
}
