package repository

import (
	"context"
	"errors"
	"strings"
	"task-management/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrTaskNotFound = errors.New("task not found")

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

const taskColumns = `id, title, description, priority, status, created_at, updated_at`

func escapeLike(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `%`, `\%`)
	s = strings.ReplaceAll(s, `_`, `\_`)
	return s
}

func scanTask(row pgx.Row) (*models.Task, error) {
	var t models.Task
	err := row.Scan(&t.Id, &t.Title, &t.Description, &t.Priority, &t.Status, &t.CreatedAt, &t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrTaskNotFound
	}
	return &t, err
}

func (r *PostgresRepository) GetTasks(ctx context.Context, offset, limit int64, search string) ([]models.Task, int64, error) {
	var total int64

	if search != "" {
		if err := r.db.QueryRow(ctx, `
			SELECT COUNT(*) FROM tasks WHERE title ILIKE $1 ESCAPE '\' OR description ILIKE $1 ESCAPE '\'
		`, "%"+escapeLike(search)+"%").Scan(&total); err != nil {
			return nil, 0, err
		}
	} else {
		if err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM tasks`).Scan(&total); err != nil {
			return nil, 0, err
		}
	}

	var (
		rows pgx.Rows
		err  error
	)

	if search != "" {
		rows, err = r.db.Query(ctx, `
			SELECT `+taskColumns+`
			FROM tasks
			WHERE title ILIKE $1 ESCAPE '\' OR description ILIKE $1 ESCAPE '\'
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
		`, "%"+escapeLike(search)+"%", limit, offset)
	} else {
		rows, err = r.db.Query(ctx, `
			SELECT `+taskColumns+`
			FROM tasks
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2
		`, limit, offset)
	}

	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		if err := rows.Scan(&t.Id, &t.Title, &t.Description, &t.Priority, &t.Status, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, 0, err
		}
		tasks = append(tasks, t)
	}
	return tasks, total, rows.Err()
}

func (r *PostgresRepository) GetTaskById(ctx context.Context, id int64) (*models.Task, error) {
	row := r.db.QueryRow(ctx, `
		SELECT `+taskColumns+` FROM tasks WHERE id = $1
	`, id)
	return scanTask(row)
}

func (r *PostgresRepository) CreateTask(ctx context.Context, task *models.Task) error {
	return r.db.QueryRow(ctx, `
		INSERT INTO tasks (title, description, priority, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, task.Title, task.Description, task.Priority, task.Status, task.CreatedAt, task.UpdatedAt).Scan(&task.Id)
}

func (r *PostgresRepository) UpdateTask(ctx context.Context, task *models.Task) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	var id int64
	err = tx.QueryRow(ctx, `SELECT id FROM tasks WHERE id = $1 FOR UPDATE`, task.Id).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrTaskNotFound
	}
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		UPDATE tasks
		SET title = $1, description = $2, priority = $3, status = $4, updated_at = $5
		WHERE id = $6
	`, task.Title, task.Description, task.Priority, task.Status, task.UpdatedAt, task.Id)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *PostgresRepository) DeleteTask(ctx context.Context, taskId int64) (*models.Task, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx) //nolint:errcheck

	task, err := scanTask(tx.QueryRow(ctx, `
		SELECT `+taskColumns+` FROM tasks WHERE id = $1 FOR UPDATE
	`, taskId))
	if err != nil {
		return nil, err
	}

	if _, err = tx.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, taskId); err != nil {
		return nil, err
	}

	return task, tx.Commit(ctx)
}
