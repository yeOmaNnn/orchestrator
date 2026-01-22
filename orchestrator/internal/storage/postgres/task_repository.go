package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/yeOmaNnn/orchestrator/internal/domain"
)

type TaskRepo struct {
	db *sql.DB
}

func NewTaskRepo(db *sql.DB) *TaskRepo {
	return &TaskRepo{db: db}
}

func (r *TaskRepo) Create(
	ctx context.Context,
	task *domain.Task,
) error {
	_, err := r.db.ExecContext(
		ctx,
		`INSERT INTO tasks (id, goal, status)
		 VALUES ($1, $2, $3)`,
		task.ID,
		task.Goal,
		task.Status,
	)
	return err
}

func (r *TaskRepo) GetByID(
	ctx context.Context,
	id uuid.UUID,
) (*domain.Task, error) {

	row := r.db.QueryRowContext(
		ctx,
		`SELECT id, goal, status
		 FROM tasks
		 WHERE id = $1`,
		id,
	)

	var task domain.Task
	if err := row.Scan(
		&task.ID,
		&task.Goal,
		&task.Status,
	); err != nil {
		return nil, err
	}

	return &task, nil
}

func (r *TaskRepo) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	status domain.TaskStatus,
) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE tasks
		 SET status = $2
		 WHERE id = $1`,
		id,
		status,
	)
	return err
}

func (r *TaskRepo) ListActive(
	ctx context.Context,
) ([]domain.Task, error) {

	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, goal, status
		 FROM tasks
		 WHERE status IN ($1, $2)`,
		domain.TaskPending,
		domain.TaskRunning,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []domain.Task

	for rows.Next() {
		var t domain.Task
		if err := rows.Scan(
			&t.ID,
			&t.Goal,
			&t.Status,
		); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}

	return tasks, nil
}
