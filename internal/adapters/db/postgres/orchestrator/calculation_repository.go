package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/orchestrator"
	repo "github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/repository/orchestrator"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/database"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	queryCreateCalculation = `
        INSERT INTO calculations (
            id, user_id, expression, result, status, error_message, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id, user_id, expression, result, status, error_message, created_at, updated_at`

	queryFindCalculationByID = `
        SELECT id, user_id, expression, result, status, error_message, created_at, updated_at
        FROM calculations
        WHERE id = $1`

	queryFindCalculationsByUserID = `
        SELECT id, user_id, expression, result, status, error_message, created_at, updated_at
        FROM calculations
        WHERE user_id = $1
        ORDER BY created_at DESC`

	queryUpdateCalculation = `
        UPDATE calculations
        SET user_id = $2, expression = $3, result = $4, status = $5, error_message = $6, updated_at = $7
        WHERE id = $1`

	queryUpdateCalculationStatus = `
        UPDATE calculations
        SET status = $2, result = $3, error_message = $4, updated_at = $5
        WHERE id = $1`

	queryDeleteCalculation = `DELETE FROM calculations WHERE id = $1`
)

var (
	ErrInvalidCalculationID = errors.New("invalid calculation ID")
	ErrInvalidUserID        = errors.New("invalid user ID")
	ErrInvalidCalculation   = errors.New("invalid calculation")
	ErrCalculationNotFound  = errors.New("calculation not found")
)

type PgCalculationRepository struct {
	db *database.Handler
}

var _ repo.CalculationRepository = (*PgCalculationRepository)(nil)

func NewCalculationRepository(db *database.Handler) *PgCalculationRepository {
	return &PgCalculationRepository{db: db}
}

func (r *PgCalculationRepository) Create(ctx context.Context, calculation *orchestrator.Calculation) (*orchestrator.Calculation, error) {
	const op = "PgCalculationRepository.Create"

	if calculation.ID == uuid.Nil {
		calculation.ID = uuid.New()
	}

	now := time.Now()
	if calculation.CreatedAt.IsZero() {
		calculation.CreatedAt = now
	}
	if calculation.UpdatedAt.IsZero() {
		calculation.UpdatedAt = now
	}

	conn, err := r.acquireConn(ctx, op)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	var result orchestrator.Calculation
	err = conn.QueryRow(ctx, queryCreateCalculation,
		calculation.ID,
		calculation.UserID,
		calculation.Expression,
		calculation.Result,
		calculation.Status,
		calculation.ErrorMessage,
		calculation.CreatedAt,
		calculation.UpdatedAt,
	).Scan(
		&result.ID,
		&result.UserID,
		&result.Expression,
		&result.Result,
		&result.Status,
		&result.ErrorMessage,
		&result.CreatedAt,
		&result.UpdatedAt,
	)

	if err != nil {
		return nil, r.logError(ctx, op, "create calculation", err)
	}

	logger.Info(ctx, nil, "Calculation created", zap.String("id", result.ID.String()))
	return &result, nil
}

func (r *PgCalculationRepository) FindByID(ctx context.Context, id uuid.UUID) (*orchestrator.Calculation, error) {
	const op = "PgCalculationRepository.FindByID"

	if id == uuid.Nil {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidCalculationID)
	}

	conn, err := r.acquireConn(ctx, op)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	var calculation orchestrator.Calculation
	err = conn.QueryRow(ctx, queryFindCalculationByID, id).Scan(
		&calculation.ID,
		&calculation.UserID,
		&calculation.Expression,
		&calculation.Result,
		&calculation.Status,
		&calculation.ErrorMessage,
		&calculation.CreatedAt,
		&calculation.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, r.logError(ctx, op, "find calculation", err)
	}

	return &calculation, nil
}

func (r *PgCalculationRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*orchestrator.Calculation, error) {
	const op = "PgCalculationRepository.FindByUserID"

	if userID == uuid.Nil {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidUserID)
	}

	conn, err := r.acquireConn(ctx, op)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, queryFindCalculationsByUserID, userID)
	if err != nil {
		return nil, r.logError(ctx, op, "query calculations", err)
	}
	defer rows.Close()

	calculations := make([]*orchestrator.Calculation, 0)
	for rows.Next() {
		var calc orchestrator.Calculation
		err := rows.Scan(
			&calc.ID,
			&calc.UserID,
			&calc.Expression,
			&calc.Result,
			&calc.Status,
			&calc.ErrorMessage,
			&calc.CreatedAt,
			&calc.UpdatedAt,
		)
		if err != nil {
			return nil, r.logError(ctx, op, "scan calculation row", err)
		}
		calculations = append(calculations, &calc)
	}

	if err := rows.Err(); err != nil {
		return nil, r.logError(ctx, op, "iterate rows", err)
	}

	return calculations, nil
}

func (r *PgCalculationRepository) Update(ctx context.Context, calculation *orchestrator.Calculation) error {
	const op = "PgCalculationRepository.Update"

	if calculation == nil || calculation.ID == uuid.Nil {
		return fmt.Errorf("%s: %w", op, ErrInvalidCalculation)
	}

	calculation.UpdatedAt = time.Now()

	conn, err := r.acquireConn(ctx, op)
	if err != nil {
		return err
	}
	defer conn.Release()

	// Execute query
	cmdTag, err := conn.Exec(ctx, queryUpdateCalculation,
		calculation.ID,
		calculation.UserID,
		calculation.Expression,
		calculation.Result,
		calculation.Status,
		calculation.ErrorMessage,
		calculation.UpdatedAt,
	)

	if err != nil {
		return r.logError(ctx, op, "update calculation", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, ErrCalculationNotFound)
	}

	return nil
}

func (r *PgCalculationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status orchestrator.CalculationStatus, result string, errorMsg string) error {
	const op = "PgCalculationRepository.UpdateStatus"

	if id == uuid.Nil {
		return fmt.Errorf("%s: %w", op, ErrInvalidCalculationID)
	}

	conn, err := r.acquireConn(ctx, op)
	if err != nil {
		return err
	}
	defer conn.Release()

	cmdTag, err := conn.Exec(ctx, queryUpdateCalculationStatus,
		id,
		status,
		result,
		errorMsg,
		time.Now(),
	)

	if err != nil {
		return r.logError(ctx, op, "update calculation status", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, ErrCalculationNotFound)
	}

	return nil
}

func (r *PgCalculationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const op = "PgCalculationRepository.Delete"

	if id == uuid.Nil {
		return fmt.Errorf("%s: %w", op, ErrInvalidCalculationID)
	}

	conn, err := r.acquireConn(ctx, op)
	if err != nil {
		return err
	}
	defer conn.Release()

	cmdTag, err := conn.Exec(ctx, queryDeleteCalculation, id)
	if err != nil {
		return r.logError(ctx, op, "delete calculation", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, ErrCalculationNotFound)
	}

	return nil
}

func (r *PgCalculationRepository) acquireConn(ctx context.Context, op string) (*pgxpool.Conn, error) {
	conn, err := r.db.AcquireConn(ctx)
	if err != nil {
		logger.Error(ctx, nil, "Failed to acquire connection", zap.String("op", op), zap.Error(err))
		return nil, fmt.Errorf("%s: acquire connection: %w", op, err)
	}
	return conn, nil
}

func (r *PgCalculationRepository) logError(ctx context.Context, op, action string, err error) error {
	logger.Error(ctx, nil, "Failed to "+action, zap.String("op", op), zap.Error(err))
	return fmt.Errorf("%s: %s: %w", op, action, err)
}
