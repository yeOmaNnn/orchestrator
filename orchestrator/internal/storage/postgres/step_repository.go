package postgres

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"orchestrator/internal/domain"
)

type StepRepository struct {
	db *sql.DB
}

func NewStepRepository(db *sql.DB) *StepRepository {
	return &StepRepository{db: db}
}

func (r *StepRepository) CreateMany(
	ctx context.Context,
	steps []domain.Step,
) error {

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO steps (
			id, task_id, agent, input, status, retry_count
		) VALUES ($1, $2, $3, $4, $5, $6)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, s := range steps {
		_, err := stmt.ExecContext(
			ctx,
			s.ID,
			s.TaskID,
			s.Agent,
			s.Input,
			s.Status,
			s.RetryCount,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *StepRepository) GetByTask(
	ctx context.Context,
	taskID uuid.UUID,
) ([]domain.Step, error) {

	rows, err := r.db.QueryContext(ctx, `
		SELECT
			id,
			task_id,
			agent,
			input,
			output,
			status,
			retry_count
		FROM steps
		WHERE task_id = $1
	`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var steps []domain.Step

	for rows.Next() {
		var s domain.Step
		if err := rows.Scan(
			&s.ID,
			&s.TaskID,
			&s.Agent,
			&s.Input,
			&s.Output,
			&s.Status,
			&s.RetryCount,
		); err != nil {
			return nil, err
		}
		steps = append(steps, s)
	}

	return steps, rows.Err()
}

func (r *StepRepository) Update(
	ctx context.Context,
	step *domain.Step,
) error {

	_, err := r.db.ExecContext(ctx, `
		UPDATE steps
		SET
			output = $1,
			status = $2,
			retry_count = $3
		WHERE id = $4
	`,
		step.Output,
		step.Status,
		step.RetryCount,
		step.ID,
	)

	return err
}
