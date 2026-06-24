package ports

import (
	"context"
	"task-management/internal/models"
)

type Cache interface {
	GetTask(ctx context.Context, id int64) (*models.Task, error)
	SetTask(ctx context.Context, task *models.Task) error
	DeleteTask(ctx context.Context, id int64) error
	GetTasks(ctx context.Context, offset, limit int64, search string) ([]models.Task, int64, error)
	SetTasks(ctx context.Context, offset, limit int64, search string, tasks []models.Task, total int64) error
	InvalidateTasks(ctx context.Context) error
}
