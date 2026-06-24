package usecase_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"task-management/internal/models"
	"task-management/internal/repository"
	"task-management/internal/usecase"
	"task-management/internal/utils/errorutil"
)

// ── mocks ────────────────────────────────────────────────────────────────────

type mockRepo struct {
	getTasksFn    func(ctx context.Context, offset, limit int64, search string) ([]models.Task, int64, error)
	getTaskByIdFn func(ctx context.Context, id int64) (*models.Task, error)
	createTaskFn  func(ctx context.Context, task *models.Task) error
	updateTaskFn  func(ctx context.Context, task *models.Task) error
	deleteTaskFn  func(ctx context.Context, taskId int64) (*models.Task, error)
}

func (m *mockRepo) GetTasks(ctx context.Context, offset, limit int64, search string) ([]models.Task, int64, error) {
	if m.getTasksFn != nil {
		return m.getTasksFn(ctx, offset, limit, search)
	}
	panic("unexpected call: GetTasks")
}
func (m *mockRepo) GetTaskById(ctx context.Context, id int64) (*models.Task, error) {
	if m.getTaskByIdFn != nil {
		return m.getTaskByIdFn(ctx, id)
	}
	panic("unexpected call: GetTaskById")
}
func (m *mockRepo) CreateTask(ctx context.Context, task *models.Task) error {
	if m.createTaskFn != nil {
		return m.createTaskFn(ctx, task)
	}
	panic("unexpected call: CreateTask")
}
func (m *mockRepo) UpdateTask(ctx context.Context, task *models.Task) error {
	if m.updateTaskFn != nil {
		return m.updateTaskFn(ctx, task)
	}
	panic("unexpected call: UpdateTask")
}
func (m *mockRepo) DeleteTask(ctx context.Context, taskId int64) (*models.Task, error) {
	if m.deleteTaskFn != nil {
		return m.deleteTaskFn(ctx, taskId)
	}
	panic("unexpected call: DeleteTask")
}

type mockCache struct {
	getTaskFn        func(ctx context.Context, id int64) (*models.Task, error)
	setTaskFn        func(ctx context.Context, task *models.Task) error
	deleteTaskFn     func(ctx context.Context, id int64) error
	getTasksFn       func(ctx context.Context, offset, limit int64, search string) ([]models.Task, int64, error)
	setTasksFn       func(ctx context.Context, offset, limit int64, search string, tasks []models.Task, total int64) error
	invalidateTaskFn func(ctx context.Context) error
}

func (m *mockCache) GetTask(ctx context.Context, id int64) (*models.Task, error) {
	if m.getTaskFn != nil {
		return m.getTaskFn(ctx, id)
	}
	return nil, errors.New("cache miss")
}
func (m *mockCache) SetTask(ctx context.Context, task *models.Task) error {
	if m.setTaskFn != nil {
		return m.setTaskFn(ctx, task)
	}
	return nil
}
func (m *mockCache) DeleteTask(ctx context.Context, id int64) error {
	if m.deleteTaskFn != nil {
		return m.deleteTaskFn(ctx, id)
	}
	return nil
}
func (m *mockCache) GetTasks(ctx context.Context, offset, limit int64, search string) ([]models.Task, int64, error) {
	if m.getTasksFn != nil {
		return m.getTasksFn(ctx, offset, limit, search)
	}
	return nil, 0, errors.New("cache miss")
}
func (m *mockCache) SetTasks(ctx context.Context, offset, limit int64, search string, tasks []models.Task, total int64) error {
	if m.setTasksFn != nil {
		return m.setTasksFn(ctx, offset, limit, search, tasks, total)
	}
	return nil
}
func (m *mockCache) InvalidateTasks(ctx context.Context) error {
	if m.invalidateTaskFn != nil {
		return m.invalidateTaskFn(ctx)
	}
	return nil
}

// ── helpers ──────────────────────────────────────────────────────────────────

func appErrCode(err error) int {
	var e *errorutil.Error
	if errors.As(err, &e) {
		return e.Code()
	}
	return 0
}

var ctx = context.Background()

// ── GetTasks ─────────────────────────────────────────────────────────────────

func TestGetTasks_CacheHit(t *testing.T) {
	want := []models.Task{{Id: 1, Title: "cached"}}
	repo := &mockRepo{}
	cache := &mockCache{
		getTasksFn: func(_ context.Context, _, _ int64, _ string) ([]models.Task, int64, error) {
			return want, 1, nil
		},
	}

	uc := usecase.NewTaskUsecase(repo, cache)
	got, total, err := uc.GetTasks(ctx, 0, 10, "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(got) != 1 || got[0].Id != 1 {
		t.Fatalf("expected cached task, got %+v", got)
	}
	if total != 1 {
		t.Errorf("expected total 1, got %d", total)
	}
}

func TestGetTasks_CacheMiss_DBSuccess(t *testing.T) {
	want := []models.Task{{Id: 2, Title: "from db"}}
	setCalled := false

	repo := &mockRepo{
		getTasksFn: func(_ context.Context, _, _ int64, _ string) ([]models.Task, int64, error) {
			return want, 5, nil
		},
	}
	cache := &mockCache{
		setTasksFn: func(_ context.Context, _, _ int64, _ string, _ []models.Task, _ int64) error {
			setCalled = true
			return nil
		},
	}

	uc := usecase.NewTaskUsecase(repo, cache)
	got, total, err := uc.GetTasks(ctx, 0, 10, "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(got) != 1 || got[0].Id != 2 {
		t.Fatalf("unexpected result: %+v", got)
	}
	if total != 5 {
		t.Errorf("expected total 5, got %d", total)
	}
	if !setCalled {
		t.Error("expected cache.SetTasks to be called")
	}
}

func TestGetTasks_CacheMiss_DBError(t *testing.T) {
	repo := &mockRepo{
		getTasksFn: func(_ context.Context, _, _ int64, _ string) ([]models.Task, int64, error) {
			return nil, 0, errors.New("db error")
		},
	}

	uc := usecase.NewTaskUsecase(repo, &mockCache{})
	_, _, err := uc.GetTasks(ctx, 0, 10, "")

	if appErrCode(err) != http.StatusInternalServerError {
		t.Fatalf("expected 500 error, got %v", err)
	}
}

func TestGetTasks_NilCache(t *testing.T) {
	want := []models.Task{{Id: 3}}
	repo := &mockRepo{
		getTasksFn: func(_ context.Context, _, _ int64, _ string) ([]models.Task, int64, error) {
			return want, 1, nil
		},
	}

	uc := usecase.NewTaskUsecase(repo, nil)
	got, _, err := uc.GetTasks(ctx, 0, 10, "")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(got) != 1 {
		t.Fatalf("unexpected result: %+v", got)
	}
}

// ── GetTaskByID ───────────────────────────────────────────────────────────────

func TestGetTaskByID_CacheHit(t *testing.T) {
	want := &models.Task{Id: 1, Title: "cached"}
	repo := &mockRepo{}
	cache := &mockCache{
		getTaskFn: func(_ context.Context, _ int64) (*models.Task, error) {
			return want, nil
		},
	}

	uc := usecase.NewTaskUsecase(repo, cache)
	got, err := uc.GetTaskByID(ctx, 1)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.Id != 1 {
		t.Fatalf("expected task id 1, got %+v", got)
	}
}

func TestGetTaskByID_CacheMiss_DBSuccess(t *testing.T) {
	want := &models.Task{Id: 5, Title: "from db"}
	setCalled := false

	repo := &mockRepo{
		getTaskByIdFn: func(_ context.Context, _ int64) (*models.Task, error) {
			return want, nil
		},
	}
	cache := &mockCache{
		setTaskFn: func(_ context.Context, _ *models.Task) error {
			setCalled = true
			return nil
		},
	}

	uc := usecase.NewTaskUsecase(repo, cache)
	got, err := uc.GetTaskByID(ctx, 5)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.Id != 5 {
		t.Fatalf("unexpected result: %+v", got)
	}
	if !setCalled {
		t.Error("expected cache.SetTask to be called")
	}
}

func TestGetTaskByID_DBError(t *testing.T) {
	repo := &mockRepo{
		getTaskByIdFn: func(_ context.Context, _ int64) (*models.Task, error) {
			return nil, errors.New("not found")
		},
	}

	uc := usecase.NewTaskUsecase(repo, &mockCache{})
	_, err := uc.GetTaskByID(ctx, 99)

	if appErrCode(err) != http.StatusNotFound {
		t.Fatalf("expected 404 error, got %v", err)
	}
}

func TestGetTaskByID_NilCache(t *testing.T) {
	want := &models.Task{Id: 7}
	repo := &mockRepo{
		getTaskByIdFn: func(_ context.Context, _ int64) (*models.Task, error) {
			return want, nil
		},
	}

	uc := usecase.NewTaskUsecase(repo, nil)
	got, err := uc.GetTaskByID(ctx, 7)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.Id != 7 {
		t.Fatalf("unexpected result: %+v", got)
	}
}

// ── CreateTask ────────────────────────────────────────────────────────────────

func TestCreateTask_Success(t *testing.T) {
	setCalled, invalidateCalled := false, false

	repo := &mockRepo{
		createTaskFn: func(_ context.Context, task *models.Task) error {
			task.Id = 42
			return nil
		},
	}
	cache := &mockCache{
		setTaskFn: func(_ context.Context, _ *models.Task) error {
			setCalled = true
			return nil
		},
		invalidateTaskFn: func(_ context.Context) error {
			invalidateCalled = true
			return nil
		},
	}

	uc := usecase.NewTaskUsecase(repo, cache)
	task := &models.Task{Title: "new task", Priority: models.PriorityHigh}
	got, err := uc.CreateTask(ctx, task)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.Status != models.StatusTodo {
		t.Errorf("expected status todo, got %s", got.Status)
	}
	if got.CreatedAt == 0 || got.UpdatedAt == 0 {
		t.Error("expected timestamps to be set")
	}
	if got.Id != 42 {
		t.Errorf("expected id 42, got %d", got.Id)
	}
	if !setCalled {
		t.Error("expected cache.SetTask to be called")
	}
	if !invalidateCalled {
		t.Error("expected cache.InvalidateTasks to be called")
	}
}

func TestCreateTask_DBError(t *testing.T) {
	repo := &mockRepo{
		createTaskFn: func(_ context.Context, _ *models.Task) error {
			return errors.New("insert failed")
		},
	}

	uc := usecase.NewTaskUsecase(repo, nil)
	_, err := uc.CreateTask(ctx, &models.Task{Title: "x"})

	if appErrCode(err) != http.StatusInternalServerError {
		t.Fatalf("expected 500 error, got %v", err)
	}
}

// ── UpdateTask ────────────────────────────────────────────────────────────────

func TestUpdateTask_Success(t *testing.T) {
	existing := &models.Task{Id: 1, Title: "old", Priority: models.PriorityLow, Status: models.StatusTodo}

	repo := &mockRepo{
		getTaskByIdFn: func(_ context.Context, _ int64) (*models.Task, error) {
			return existing, nil
		},
		updateTaskFn: func(_ context.Context, task *models.Task) error {
			return nil
		},
	}

	uc := usecase.NewTaskUsecase(repo, nil)
	incoming := &models.Task{Id: 1, Title: "new title", Description: "desc", Priority: models.PriorityHigh}
	got, err := uc.UpdateTask(ctx, incoming)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.Title != "new title" {
		t.Errorf("expected title 'new title', got %q", got.Title)
	}
	if got.Priority != models.PriorityHigh {
		t.Errorf("expected priority high, got %s", got.Priority)
	}
	if got.Status != models.StatusTodo {
		t.Errorf("expected status to be preserved, got %s", got.Status)
	}
}

func TestUpdateTask_NotFound(t *testing.T) {
	repo := &mockRepo{
		getTaskByIdFn: func(_ context.Context, _ int64) (*models.Task, error) {
			return nil, errors.New("not found")
		},
	}

	uc := usecase.NewTaskUsecase(repo, nil)
	_, err := uc.UpdateTask(ctx, &models.Task{Id: 99})

	if appErrCode(err) != http.StatusNotFound {
		t.Fatalf("expected 404 error, got %v", err)
	}
}

func TestUpdateTask_DBError(t *testing.T) {
	repo := &mockRepo{
		getTaskByIdFn: func(_ context.Context, _ int64) (*models.Task, error) {
			return &models.Task{Id: 1}, nil
		},
		updateTaskFn: func(_ context.Context, _ *models.Task) error {
			return errors.New("update failed")
		},
	}

	uc := usecase.NewTaskUsecase(repo, nil)
	_, err := uc.UpdateTask(ctx, &models.Task{Id: 1, Title: "x"})

	if appErrCode(err) != http.StatusInternalServerError {
		t.Fatalf("expected 500 error, got %v", err)
	}
}

// ── DeleteTask ────────────────────────────────────────────────────────────────

func TestDeleteTask_Success(t *testing.T) {
	deleted := &models.Task{Id: 10, Title: "to delete"}
	deleteCacheCalled, invalidateCalled := false, false

	repo := &mockRepo{
		deleteTaskFn: func(_ context.Context, id int64) (*models.Task, error) {
			return deleted, nil
		},
	}
	cache := &mockCache{
		deleteTaskFn: func(_ context.Context, _ int64) error {
			deleteCacheCalled = true
			return nil
		},
		invalidateTaskFn: func(_ context.Context) error {
			invalidateCalled = true
			return nil
		},
	}

	uc := usecase.NewTaskUsecase(repo, cache)
	got, err := uc.DeleteTask(ctx, &models.Task{Id: 10})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.Id != 10 {
		t.Errorf("expected deleted task id 10, got %d", got.Id)
	}
	if !deleteCacheCalled {
		t.Error("expected cache.DeleteTask to be called")
	}
	if !invalidateCalled {
		t.Error("expected cache.InvalidateTasks to be called")
	}
}

func TestDeleteTask_NotFound(t *testing.T) {
	repo := &mockRepo{
		deleteTaskFn: func(_ context.Context, _ int64) (*models.Task, error) {
			return nil, repository.ErrTaskNotFound
		},
	}

	uc := usecase.NewTaskUsecase(repo, nil)
	_, err := uc.DeleteTask(ctx, &models.Task{Id: 99})

	if appErrCode(err) != http.StatusNotFound {
		t.Fatalf("expected 404 error, got %v", err)
	}
}

func TestDeleteTask_InternalError(t *testing.T) {
	repo := &mockRepo{
		deleteTaskFn: func(_ context.Context, _ int64) (*models.Task, error) {
			return nil, errors.New("connection lost")
		},
	}

	uc := usecase.NewTaskUsecase(repo, nil)
	_, err := uc.DeleteTask(ctx, &models.Task{Id: 1})

	if appErrCode(err) != http.StatusInternalServerError {
		t.Fatalf("expected 500 error, got %v", err)
	}
}

// ── UpdateTaskStatus ──────────────────────────────────────────────────────────

func TestUpdateTaskStatus_Success(t *testing.T) {
	task := &models.Task{Id: 1, Status: models.StatusTodo}

	repo := &mockRepo{
		getTaskByIdFn: func(_ context.Context, _ int64) (*models.Task, error) {
			return task, nil
		},
		updateTaskFn: func(_ context.Context, _ *models.Task) error {
			return nil
		},
	}

	uc := usecase.NewTaskUsecase(repo, nil)
	got, err := uc.UpdateTaskStatus(ctx, 1, models.StatusInProgress)

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got.Status != models.StatusInProgress {
		t.Errorf("expected status in_progress, got %s", got.Status)
	}
}

func TestUpdateTaskStatus_SameStatus_IsNoop(t *testing.T) {
	task := &models.Task{Id: 1, Status: models.StatusTodo}

	repo := &mockRepo{
		getTaskByIdFn: func(_ context.Context, _ int64) (*models.Task, error) {
			return task, nil
		},
		updateTaskFn: func(_ context.Context, _ *models.Task) error {
			return nil
		},
	}

	uc := usecase.NewTaskUsecase(repo, nil)
	got, err := uc.UpdateTaskStatus(ctx, 1, models.StatusTodo)

	if err != nil {
		t.Fatalf("expected no error for same-status update, got %v", err)
	}
	if got.Status != models.StatusTodo {
		t.Errorf("expected status todo, got %s", got.Status)
	}
}

func TestUpdateTaskStatus_InvalidTransition(t *testing.T) {
	cases := []struct {
		from models.Status
		to   models.Status
	}{
		{models.StatusTodo, models.StatusDone},
		{models.StatusDone, models.StatusTodo},
		{models.StatusDone, models.StatusInProgress},
	}

	for _, tc := range cases {
		task := &models.Task{Id: 1, Status: tc.from}
		repo := &mockRepo{
			getTaskByIdFn: func(_ context.Context, _ int64) (*models.Task, error) {
				return task, nil
			},
		}

		uc := usecase.NewTaskUsecase(repo, nil)
		_, err := uc.UpdateTaskStatus(ctx, 1, tc.to)

		if appErrCode(err) != http.StatusUnprocessableEntity {
			t.Errorf("transition %s→%s: expected 422, got %v", tc.from, tc.to, err)
		}
	}
}

func TestUpdateTaskStatus_NotFound(t *testing.T) {
	repo := &mockRepo{
		getTaskByIdFn: func(_ context.Context, _ int64) (*models.Task, error) {
			return nil, errors.New("not found")
		},
	}

	uc := usecase.NewTaskUsecase(repo, nil)
	_, err := uc.UpdateTaskStatus(ctx, 99, models.StatusDone)

	if appErrCode(err) != http.StatusNotFound {
		t.Fatalf("expected 404 error, got %v", err)
	}
}

func TestUpdateTaskStatus_DBError(t *testing.T) {
	repo := &mockRepo{
		getTaskByIdFn: func(_ context.Context, _ int64) (*models.Task, error) {
			return &models.Task{Id: 1, Status: models.StatusTodo}, nil
		},
		updateTaskFn: func(_ context.Context, _ *models.Task) error {
			return errors.New("update failed")
		},
	}

	uc := usecase.NewTaskUsecase(repo, nil)
	_, err := uc.UpdateTaskStatus(ctx, 1, models.StatusInProgress)

	if appErrCode(err) != http.StatusInternalServerError {
		t.Fatalf("expected 500 error, got %v", err)
	}
}
