package handlers

import (
	"net/http"
	"strconv"
	"task-management/internal/delivery/dto"
	"task-management/internal/models"
	"task-management/internal/ports"
	"task-management/internal/utils/errorutil"

	"github.com/gin-gonic/gin"
)

type TaskHandler struct {
	usecase ports.UseCase
}

func NewTaskHandler(usecase ports.UseCase) *TaskHandler {
	return &TaskHandler{usecase: usecase}
}

func respondError(c *gin.Context, err error) {
	if e, ok := err.(*errorutil.Error); ok {
		c.JSON(e.Code(), dto.ErrorResponse{Error: e.Error()})
		return
	}
	c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal server error"})
}

func parseID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid id"})
		return 0, false
	}
	return id, true
}

// GetTasks godoc
// @Summary      List tasks
// @Description  Returns a paginated list of tasks. Optionally filter by title or description.
// @Tags         tasks
// @Produce      json
// @Param        offset  query     int     false  "Pagination offset"  default(0)
// @Param        limit   query     int     false  "Page size"          default(20)
// @Param        search  query     string  false  "Search in title and description"
// @Success      200     {object}  dto.TaskListResponse
// @Failure      500     {object}  dto.ErrorResponse
// @Router       /tasks [get]
func (h *TaskHandler) GetTasks(c *gin.Context) {
	offset, _ := strconv.ParseInt(c.DefaultQuery("offset", "0"), 10, 64)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 64)
	search := c.Query("search")

	tasks, total, err := h.usecase.GetTasks(c.Request.Context(), offset, limit, search)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, dto.TaskListResponse{Tasks: tasks, Total: total, Offset: offset, Limit: limit})
}

// GetTaskById godoc
// @Summary      Get task by ID
// @Description  Returns a single task by its numeric ID.
// @Tags         tasks
// @Produce      json
// @Param        id   path      int  true  "Task ID"
// @Success      200  {object}  models.Task
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Router       /tasks/{id} [get]
func (h *TaskHandler) GetTaskById(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	task, err := h.usecase.GetTaskByID(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, task)
}

// CreateTask godoc
// @Summary      Create task
// @Description  Creates a new task with the given title, description, and priority.
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        task  body      dto.CreateTaskRequest  true  "Task payload"
// @Success      201   {object}  models.Task
// @Failure      422   {object}  dto.ErrorResponse
// @Failure      500   {object}  dto.ErrorResponse
// @Router       /tasks [post]
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var req dto.CreateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse{Error: err.Error()})
		return
	}

	task := &models.Task{
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
	}

	created, err := h.usecase.CreateTask(c.Request.Context(), task)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, created)
}

// UpdateTask godoc
// @Summary      Update task
// @Description  Updates the title, description, and priority of an existing task.
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id    path      int                    true  "Task ID"
// @Param        task  body      dto.UpdateTaskRequest  true  "Task payload"
// @Success      200   {object}  models.Task
// @Failure      400   {object}  dto.ErrorResponse
// @Failure      404   {object}  dto.ErrorResponse
// @Failure      422   {object}  dto.ErrorResponse
// @Failure      500   {object}  dto.ErrorResponse
// @Router       /tasks/{id} [put]
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req dto.UpdateTaskRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse{Error: err.Error()})
		return
	}

	task := &models.Task{
		Id:          id,
		Title:       req.Title,
		Description: req.Description,
		Priority:    req.Priority,
	}

	updated, err := h.usecase.UpdateTask(c.Request.Context(), task)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, updated)
}

// DeleteTask godoc
// @Summary      Delete task
// @Description  Permanently deletes a task and returns the deleted record.
// @Tags         tasks
// @Produce      json
// @Param        id   path      int  true  "Task ID"
// @Success      200  {object}  models.Task
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /tasks/{id} [delete]
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	deleted, err := h.usecase.DeleteTask(c.Request.Context(), &models.Task{Id: id})
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, deleted)
}

// UpdateTaskStatus godoc
// @Summary      Update task status
// @Description  Changes the lifecycle status of a task (todo → in_progress → done).
// @Tags         tasks
// @Accept       json
// @Produce      json
// @Param        id      path      int                      true  "Task ID"
// @Param        status  body      dto.UpdateStatusRequest  true  "Status payload"
// @Success      200     {object}  models.Task
// @Failure      400     {object}  dto.ErrorResponse
// @Failure      404     {object}  dto.ErrorResponse
// @Failure      422     {object}  dto.ErrorResponse
// @Failure      500     {object}  dto.ErrorResponse
// @Router       /tasks/{id}/status [patch]
func (h *TaskHandler) UpdateTaskStatus(c *gin.Context) {
	id, ok := parseID(c)
	if !ok {
		return
	}

	var req dto.UpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusUnprocessableEntity, dto.ErrorResponse{Error: err.Error()})
		return
	}

	updated, err := h.usecase.UpdateTaskStatus(c.Request.Context(), id, req.Status)
	if err != nil {
		respondError(c, err)
		return
	}

	c.JSON(http.StatusOK, updated)
}
