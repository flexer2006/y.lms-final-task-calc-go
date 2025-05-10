package orchestrator

import (
	"context"
	"errors"
	"fmt"

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
	queryCreateOperation = `
        INSERT INTO operations (
            id, calculation_id, operation_type, operand1, operand2, result, status, error_message, processing_time_ms, agent_id
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
        ) RETURNING id, calculation_id, operation_type, operand1, operand2, result, status, error_message, processing_time_ms, agent_id`

	queryFindOperationByID = `
        SELECT id, calculation_id, operation_type, operand1, operand2, result, status, error_message, processing_time_ms, agent_id
        FROM operations
        WHERE id = $1`

	queryFindOperationsByCalculationID = `
        SELECT id, calculation_id, operation_type, operand1, operand2, result, status, error_message, processing_time_ms, agent_id
        FROM operations
        WHERE calculation_id = $1
        ORDER BY id`

	queryGetPendingOperations = `
        SELECT id, calculation_id, operation_type, operand1, operand2, result, status, error_message, processing_time_ms, agent_id
        FROM operations
        WHERE status = $1
        ORDER BY id
        LIMIT $2`

	queryUpdateOperation = `
        UPDATE operations
        SET calculation_id = $2, operation_type = $3, operand1 = $4, operand2 = $5, 
            result = $6, status = $7, error_message = $8, processing_time_ms = $9, agent_id = $10
        WHERE id = $1`

	queryUpdateOperationStatus = `
        UPDATE operations
        SET status = $2, result = $3, error_message = $4
        WHERE id = $1`

	queryAssignAgent = `
        UPDATE operations
        SET agent_id = $2, status = $3
        WHERE id = $1 AND status = $4`

	batchInsertOperation = `
        INSERT INTO operations (
            id, calculation_id, operation_type, operand1, operand2, result, status, error_message, processing_time_ms, agent_id
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
        )`
)

var (
	ErrOperationNil               = errors.New("operation cannot be nil")
	ErrOperationHasNoCalcID       = errors.New("operation has no calculation ID")
	ErrInvalidOperationID         = errors.New("invalid operation ID")
	ErrInvalidOperation           = errors.New("invalid operation")
	ErrOperationNotFound          = errors.New("operation not found")
	ErrInvalidOperationOrAgentID  = errors.New("invalid operation ID or agent ID")
	ErrOperationNotInPendingState = errors.New("operation not found or not in pending state")
	ErrInvalidCalculationID2      = errors.New("invalid calculation ID")
)

type PgOperationRepository struct {
	db *database.Handler
}

var _ repo.OperationRepository = (*PgOperationRepository)(nil)

func NewOperationRepository(db *database.Handler) *PgOperationRepository {
	return &PgOperationRepository{db: db}
}

func (r *PgOperationRepository) Create(ctx context.Context, operation *orchestrator.Operation) (*orchestrator.Operation, error) {
	const op = "PgOperationRepository.Create"

	if operation == nil {
		return nil, fmt.Errorf("%s: %w", op, ErrOperationNil)
	}

	if operation.ID == uuid.Nil {
		operation.ID = uuid.New()
	}

	conn, err := r.acquireConn(ctx, op)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	var result orchestrator.Operation
	err = conn.QueryRow(ctx, queryCreateOperation,
		operation.ID,
		operation.CalculationID,
		operation.OperationType,
		operation.Operand1,
		operation.Operand2,
		operation.Result,
		operation.Status,
		operation.ErrorMessage,
		operation.ProcessingTime,
		operation.AgentID,
	).Scan(
		&result.ID,
		&result.CalculationID,
		&result.OperationType,
		&result.Operand1,
		&result.Operand2,
		&result.Result,
		&result.Status,
		&result.ErrorMessage,
		&result.ProcessingTime,
		&result.AgentID,
	)

	if err != nil {
		return nil, r.logError(ctx, op, "create operation", err)
	}

	logger.Info(ctx, nil, "Operation created", zap.String("id", result.ID.String()))
	return &result, nil
}

func (r *PgOperationRepository) CreateBatch(ctx context.Context, operations []*orchestrator.Operation) error {
	const op = "PgOperationRepository.CreateBatch"

	if len(operations) == 0 {
		return nil
	}

	conn, err := r.acquireConn(ctx, op)
	if err != nil {
		return err
	}
	// Release connection only after all operations are complete
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return r.logError(ctx, op, "begin transaction", err)
	}

	// Track if we need to rollback
	var committed bool
	// Add the missing rollback code in the defer function
	defer func() {
		if !committed {
			// Only try to rollback if not committed
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				logger.Error(ctx, nil, "Failed to rollback transaction",
					zap.String("op", op),
					zap.Error(rbErr))
			}
		}
	}()

	// Create batch requests
	batch := &pgx.Batch{}
	for _, operation := range operations {
		if operation.ID == uuid.Nil {
			operation.ID = uuid.New()
		}

		// Validate required field
		if operation.CalculationID == uuid.Nil {
			return fmt.Errorf("%s: %w", op, ErrOperationHasNoCalcID)
		}

		batch.Queue(batchInsertOperation,
			operation.ID,
			operation.CalculationID,
			operation.OperationType,
			operation.Operand1,
			operation.Operand2,
			operation.Result,
			operation.Status,
			operation.ErrorMessage,
			operation.ProcessingTime,
			operation.AgentID,
		)
	}

	// Send batch and get results
	batchResults := tx.SendBatch(ctx, batch)

	// Important: always close batch results before other operations
	err = func() error {
		defer func() {
			if closeErr := batchResults.Close(); closeErr != nil {
				logger.Error(ctx, nil, "Failed to close batch results",
					zap.String("op", op), zap.Error(closeErr))
			}
		}()

		// Process all results
		for i := 0; i < batch.Len(); i++ {
			_, err := batchResults.Exec()
			if err != nil {
				return r.logError(ctx, op, fmt.Sprintf("execute batch query at index %d", i), err)
			}
		}
		return nil
	}()

	if err != nil {
		return err
	}

	// Commit transaction after successfully processing all results
	if err = tx.Commit(ctx); err != nil {
		return r.logError(ctx, op, "commit transaction", err)
	}

	// Mark as committed to avoid rollback attempt in defer
	committed = true

	logger.Info(ctx, nil, "Created operations batch", zap.Int("count", len(operations)))
	return nil
}

func (r *PgOperationRepository) FindByID(ctx context.Context, id uuid.UUID) (*orchestrator.Operation, error) {
	const op = "PgOperationRepository.FindByID"

	if id == uuid.Nil {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidOperationID)
	}

	conn, err := r.acquireConn(ctx, op)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	var operation orchestrator.Operation
	err = conn.QueryRow(ctx, queryFindOperationByID, id).Scan(
		&operation.ID,
		&operation.CalculationID,
		&operation.OperationType,
		&operation.Operand1,
		&operation.Operand2,
		&operation.Result,
		&operation.Status,
		&operation.ErrorMessage,
		&operation.ProcessingTime,
		&operation.AgentID,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, r.logError(ctx, op, "find operation", err)
	}

	return &operation, nil
}

func (r *PgOperationRepository) FindByCalculationID(ctx context.Context, calculationID uuid.UUID) ([]*orchestrator.Operation, error) {
	const op = "PgOperationRepository.FindByCalculationID"

	if calculationID == uuid.Nil {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidCalculationID2)
	}

	conn, err := r.acquireConn(ctx, op)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, queryFindOperationsByCalculationID, calculationID)
	if err != nil {
		return nil, r.logError(ctx, op, "query operations", err)
	}
	defer rows.Close()

	operations := make([]*orchestrator.Operation, 0)
	for rows.Next() {
		var operation orchestrator.Operation
		err := rows.Scan(
			&operation.ID,
			&operation.CalculationID,
			&operation.OperationType,
			&operation.Operand1,
			&operation.Operand2,
			&operation.Result,
			&operation.Status,
			&operation.ErrorMessage,
			&operation.ProcessingTime,
			&operation.AgentID,
		)
		if err != nil {
			return nil, r.logError(ctx, op, "scan row", err)
		}
		operations = append(operations, &operation)
	}

	if err := rows.Err(); err != nil {
		return nil, r.logError(ctx, op, "iterate rows", err)
	}

	return operations, nil
}

func (r *PgOperationRepository) GetPendingOperations(ctx context.Context, limit int) ([]*orchestrator.Operation, error) {
	const op = "PgOperationRepository.GetPendingOperations"

	if limit <= 0 {
		limit = 10
	}

	conn, err := r.acquireConn(ctx, op)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	rows, err := conn.Query(ctx, queryGetPendingOperations, orchestrator.OperationStatusPending, limit)
	if err != nil {
		return nil, r.logError(ctx, op, "query pending operations", err)
	}
	defer rows.Close()

	operations := make([]*orchestrator.Operation, 0, limit)

	for rows.Next() {
		var operation orchestrator.Operation
		err := rows.Scan(
			&operation.ID,
			&operation.CalculationID,
			&operation.OperationType,
			&operation.Operand1,
			&operation.Operand2,
			&operation.Result,
			&operation.Status,
			&operation.ErrorMessage,
			&operation.ProcessingTime,
			&operation.AgentID,
		)
		if err != nil {
			return nil, r.logError(ctx, op, "scan row", err)
		}
		operations = append(operations, &operation)
	}

	if err := rows.Err(); err != nil {
		return nil, r.logError(ctx, op, "iterate rows", err)
	}

	return operations, nil
}

func (r *PgOperationRepository) Update(ctx context.Context, operation *orchestrator.Operation) error {
	const op = "PgOperationRepository.Update"

	if operation == nil || operation.ID == uuid.Nil {
		return fmt.Errorf("%s: %w", op, ErrInvalidOperation)
	}

	conn, err := r.acquireConn(ctx, op)
	if err != nil {
		return err
	}
	defer conn.Release()

	cmdTag, err := conn.Exec(ctx, queryUpdateOperation,
		operation.ID,
		operation.CalculationID,
		operation.OperationType,
		operation.Operand1,
		operation.Operand2,
		operation.Result,
		operation.Status,
		operation.ErrorMessage,
		operation.ProcessingTime,
		operation.AgentID,
	)

	if err != nil {
		return r.logError(ctx, op, "update operation", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, ErrOperationNotFound)
	}

	return nil
}

func (r *PgOperationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status orchestrator.OperationStatus, result string, errorMsg string) error {
	const op = "PgOperationRepository.UpdateStatus"

	if id == uuid.Nil {
		return fmt.Errorf("%s: %w", op, ErrInvalidOperationID)
	}

	conn, err := r.acquireConn(ctx, op)
	if err != nil {
		return err
	}
	defer conn.Release()

	cmdTag, err := conn.Exec(ctx, queryUpdateOperationStatus,
		id,
		status,
		result,
		errorMsg,
	)

	if err != nil {
		return r.logError(ctx, op, "update operation status", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, ErrOperationNotFound)
	}

	return nil
}

func (r *PgOperationRepository) AssignAgent(ctx context.Context, operationID uuid.UUID, agentID string) error {
	const op = "PgOperationRepository.AssignAgent"

	if operationID == uuid.Nil || agentID == "" {
		return fmt.Errorf("%s: %w", op, ErrInvalidOperationOrAgentID)
	}

	conn, err := r.acquireConn(ctx, op)
	if err != nil {
		return err
	}
	defer conn.Release()

	cmdTag, err := conn.Exec(ctx, queryAssignAgent,
		operationID,
		agentID,
		orchestrator.OperationStatusInProgress,
		orchestrator.OperationStatusPending,
	)

	if err != nil {
		return r.logError(ctx, op, "assign agent", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, ErrOperationNotInPendingState)
	}

	return nil
}

func (r *PgOperationRepository) acquireConn(ctx context.Context, op string) (*pgxpool.Conn, error) {
	conn, err := r.db.AcquireConn(ctx)
	if err != nil {
		logger.Error(ctx, nil, "Failed to acquire connection", zap.String("op", op), zap.Error(err))
		return nil, fmt.Errorf("%s: acquire connection: %w", op, err)
	}
	return conn, nil
}

func (r *PgOperationRepository) logError(ctx context.Context, op, action string, err error) error {
	logger.Error(ctx, nil, "Failed to "+action, zap.String("op", op), zap.Error(err))
	return fmt.Errorf("%s: %s: %w", op, action, err)
}
