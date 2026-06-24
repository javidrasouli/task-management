package dto

import "task-management/internal/models"

type CreateTaskRequest struct {
	Title       string          `json:"title" binding:"required"`
	Description string          `json:"description"`
	Priority    models.Priority `json:"priority" binding:"required,oneof=low medium high"`
}

type UpdateTaskRequest struct {
	Title       string          `json:"title" binding:"required"`
	Description string          `json:"description"`
	Priority    models.Priority `json:"priority" binding:"required,oneof=low medium high"`
}

type UpdateStatusRequest struct {
	Status models.Status `json:"status" binding:"required,oneof=todo in_progress done"`
}

type TaskListResponse struct {
	Tasks  []models.Task `json:"tasks"`
	Total  int64         `json:"total"`
	Offset int64         `json:"offset"`
	Limit  int64         `json:"limit"`
}
