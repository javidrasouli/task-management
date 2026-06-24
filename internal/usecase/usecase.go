package usecase

import (
	"context"
	"errors"
	"fmt"
	"task-management/internal/models"
	"task-management/internal/ports"
	"task-management/internal/repository"
	"task-management/internal/utils/errorutil"
	"task-management/internal/utils/logutil"
	"time"

	"go.uber.org/zap"
)

var allowedTransitions = map[models.Status]map[models.Status]bool{
	models.StatusTodo:       {models.StatusInProgress: true},
	models.StatusInProgress: {models.StatusDone: true, models.StatusTodo: true},
	models.StatusDone:       {},
}

type TaskUsecase struct {
	db    ports.Repository
	cache ports.Cache
}

func NewTaskUsecase(db ports.Repository, cache ports.Cache) *TaskUsecase {
	return &TaskUsecase{
		db:    db,
		cache: cache,
	}
}

func (u *TaskUsecase) GetTasks(ctx context.Context, offset, limit int64, search string) ([]models.Task, int64, error) {
	log := logutil.FromContext(ctx).With(zap.String("op", "GetTasks"), zap.Int64("offset", offset), zap.Int64("limit", limit))

	if u.cache != nil {
		if tasks, total, err := u.cache.GetTasks(ctx, offset, limit, search); err == nil {
			return tasks, total, nil
		}
	}

	tasks, total, err := u.db.GetTasks(ctx, offset, limit, search)
	if err != nil {
		log.Error("db fetch failed", zap.Error(err))
		return nil, 0, errorutil.NewError(errorutil.ErrInternal, err)
	}

	if u.cache != nil {
		_ = u.cache.SetTasks(ctx, offset, limit, search, tasks, total)
	}

	return tasks, total, nil
}

func (u *TaskUsecase) GetTaskByID(ctx context.Context, id int64) (*models.Task, error) {
	log := logutil.FromContext(ctx).With(zap.String("op", "GetTaskByID"), zap.Int64("task_id", id))

	if u.cache != nil {
		if task, err := u.cache.GetTask(ctx, id); err == nil {
			return task, nil
		}
	}

	task, err := u.db.GetTaskById(ctx, id)
	if err != nil {
		log.Error("task not found", zap.Error(err))
		return nil, errorutil.NewError(errorutil.ErrNotFound, err)
	}

	if u.cache != nil {
		_ = u.cache.SetTask(ctx, task)
	}

	return task, nil
}

func (u *TaskUsecase) CreateTask(ctx context.Context, task *models.Task) (*models.Task, error) {
	log := logutil.FromContext(ctx).With(zap.String("op", "CreateTask"))

	task.Status = models.StatusTodo
	task.CreatedAt = time.Now().Unix()
	task.UpdatedAt = time.Now().Unix()

	if err := u.db.CreateTask(ctx, task); err != nil {
		log.Error("create task failed", zap.Error(err))
		return nil, errorutil.NewError(errorutil.ErrInternal, err)
	}

	if u.cache != nil {
		_ = u.cache.SetTask(ctx, task)
		_ = u.cache.InvalidateTasks(ctx)
	}

	return task, nil
}

func (u *TaskUsecase) UpdateTask(ctx context.Context, task *models.Task) (*models.Task, error) {
	log := logutil.FromContext(ctx).With(zap.String("op", "UpdateTask"), zap.Int64("task_id", task.Id))

	existing, err := u.db.GetTaskById(ctx, task.Id)
	if err != nil {
		log.Error("task not found", zap.Error(err))
		return nil, errorutil.NewError(errorutil.ErrNotFound, err)
	}

	existing.Title = task.Title
	existing.Description = task.Description
	existing.Priority = task.Priority
	existing.UpdatedAt = time.Now().Unix()

	if err := u.db.UpdateTask(ctx, existing); err != nil {
		log.Error("update task failed", zap.Error(err))
		return nil, errorutil.NewError(errorutil.ErrInternal, err)
	}

	if u.cache != nil {
		_ = u.cache.SetTask(ctx, existing)
		_ = u.cache.InvalidateTasks(ctx)
	}

	return existing, nil
}

func (u *TaskUsecase) DeleteTask(ctx context.Context, task *models.Task) (*models.Task, error) {
	log := logutil.FromContext(ctx).With(zap.String("op", "DeleteTask"), zap.Int64("task_id", task.Id))

	deleted, err := u.db.DeleteTask(ctx, task.Id)
	if err != nil {
		log.Error("delete task failed", zap.Error(err))
		if errors.Is(err, repository.ErrTaskNotFound) {
			return nil, errorutil.NewError(errorutil.ErrNotFound, err)
		}
		return nil, errorutil.NewError(errorutil.ErrInternal, err)
	}

	if u.cache != nil {
		_ = u.cache.DeleteTask(ctx, task.Id)
		_ = u.cache.InvalidateTasks(ctx)
	}

	return deleted, nil
}

func (u *TaskUsecase) UpdateTaskStatus(ctx context.Context, taskId int64, status models.Status) (*models.Task, error) {
	log := logutil.FromContext(ctx).With(zap.String("op", "UpdateTaskStatus"), zap.Int64("task_id", taskId), zap.String("status", string(status)))

	task, err := u.db.GetTaskById(ctx, taskId)
	if err != nil {
		log.Error("task not found", zap.Error(err))
		return nil, errorutil.NewError(errorutil.ErrNotFound, err)
	}

	if task.Status != status {
		if !allowedTransitions[task.Status][status] {
			return nil, errorutil.NewError(errorutil.ErrInvalidTransition,
				fmt.Errorf("cannot transition from %s to %s", task.Status, status))
		}
	}

	task.Status = status
	task.UpdatedAt = time.Now().Unix()

	if err := u.db.UpdateTask(ctx, task); err != nil {
		log.Error("update status failed", zap.Error(err))
		return nil, errorutil.NewError(errorutil.ErrInternal, err)
	}

	if u.cache != nil {
		_ = u.cache.SetTask(ctx, task)
		_ = u.cache.InvalidateTasks(ctx)
	}

	return task, nil
}
