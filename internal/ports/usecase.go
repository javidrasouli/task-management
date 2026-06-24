package ports

import (
	"context"
	"task-management/internal/models"
)

type UseCase interface {
	GetTasks(ctx context.Context, offset, limit int64, search string) ([]models.Task, int64, error)
	CreateTask(ctx context.Context, task *models.Task) (*models.Task, error)
	UpdateTask(ctx context.Context, task *models.Task) (*models.Task, error)
	DeleteTask(ctx context.Context, task *models.Task) (*models.Task, error)
	GetTaskByID(ctx context.Context, id int64) (*models.Task, error)
	UpdateTaskStatus(ctx context.Context, taskId int64, status models.Status) (*models.Task, error)
}
