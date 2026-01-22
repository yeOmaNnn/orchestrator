package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/yeOmaNnn/orchestrator/internal/domain"
)

type StepRepo struct {
	db *sql.DB
}

func NewStepRepo(db *sql.DB) *StepRepo {
	return &StepRepo{db: db}
}

func (r *StepRepo) CreateMany(
	ctx context.Context,
	steps []domain.Step,
) error {

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	for _, s := range steps {
		_, err := tx.ExecContext(
			ctx,
			`INSERT INTO steps
			 (id, task_id, agent, input, status, depends_on, created_at, updated_at)
			 VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())`,
			s.ID,
			s.TaskID,
			s.Agent,
			s.Input,
			s.Status,
			s.DependsOn,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *StepRepo) GetByTask(
	ctx context.Context,
	taskID uuid.UUID,
) ([]domain.Step, error) {

	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, task_id, agent, input, output,
		        status, retry_count, depends_on, 
				locked_at, locked_by, created_at, update_at
		 FROM steps
		 WHERE task_id = $1`,
		taskID,
	)
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
			&s.DependsOn,
			&s.LockedAt, 
			&s.LockedBy, 
			&s.CreatedAt, 
			&s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		steps = append(steps, s)
	}

	return steps, nil
}

func (r *StepRepo) Update(
	ctx context.Context,
	step *domain.Step,
) error {
	_, err := r.db.ExecContext(
		ctx,
		`UPDATE steps
		 SET status = $2,
		     output = $3,
		     retry_count = $4, 
			 locked_at = $5, 
			 locked_by = $6, 
			 updated_at = NOW()
		 WHERE id = $1`,
		step.ID,
		step.Status,
		step.Output,
		step.RetryCount,
		step.LockedAt, 
		step.LockedBy,
	)
	return err
}

func (r *StepRepo) AcquireReadySteps(
	ctx context.Context,
	taskID uuid.UUID,
	limit int,
	workerID string,
) ([]domain.Step, error) {

	rows, err := r.db.QueryContext(ctx, `
		UPDATE steps
		SET
			status = 'IN_PROGRESS',
			locked_at = NOW(),
			locked_by = $3,
			updated_at = NOW()
		WHERE id IN (
			SELECT s.id
			FROM steps s
			WHERE s.task_id = $1
			  AND s.status = 'WAITING'
			  AND (s.next_run_at IS NULL OR s.next_run_at <= NOW())
			  AND NOT EXISTS (
				SELECT 1
				FROM step_dependencies d
				JOIN steps dep ON dep.id = d.depends_on
				WHERE d.step_id = s.id
				  AND dep.status != 'DONE'
			  )
			ORDER BY s.created_at
			LIMIT $2
			FOR UPDATE SKIP LOCKED
		)
		RETURNING id, task_id, agent, input, output, status, retry_count
	`, taskID, limit, workerID)

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

	return steps, nil
}

func (r *StepRepo) CancelByTask(
	ctx context.Context,
	taskID uuid.UUID,
) error {

	_, err := r.db.ExecContext(
		ctx,
		`UPDATE steps
		 SET status = $2,
		     updated_at = NOW()
		 WHERE task_id = $1
		   AND status NOT IN ($3, $4)`,
		taskID,
		domain.StepCancelled,
		domain.StepDone,
		domain.StepError,
	)

	return err
}

func (r *StepRepo) ReleaseStaleLocks(
	ctx context.Context,
	ttl time.Duration,
) error {

	_, err := r.db.ExecContext(
		ctx,
		`UPDATE steps
		 SET
			status = 'WAITING',
			locked_at = NULL,
			locked_by = NULL,
			updated_at = NOW()
		 WHERE status = 'IN_PROGRESS'
		   AND locked_at < NOW() - $1::interval`,
		fmt.Sprintf("%f seconds", ttl.Seconds()),
	)

	return err
}