// Package orchestrator содержит интерфейс для работы с хранилищем операций.
package orchestrator

import (
	"context"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/orchestrator"
	"github.com/google/uuid"
)

// OperationRepository определяет интерфейс для работы с хранилищем операций.
type OperationRepository interface {
	// Create создаёт новую операцию.
	Create(ctx context.Context, operation *orchestrator.Operation) (*orchestrator.Operation, error)

	// CreateBatch создаёт несколько операций.
	CreateBatch(ctx context.Context, operations []*orchestrator.Operation) error

	// FindByID находит операцию по ID.
	FindByID(ctx context.Context, id uuid.UUID) (*orchestrator.Operation, error)

	// FindByCalculationID находит операции по ID вычисления.
	FindByCalculationID(ctx context.Context, calculationID uuid.UUID) ([]*orchestrator.Operation, error)

	// GetPendingOperations получает список ожидающих выполнения операций.
	GetPendingOperations(ctx context.Context, limit int) ([]*orchestrator.Operation, error)

	// Update обновляет операцию.
	Update(ctx context.Context, operation *orchestrator.Operation) error

	// UpdateStatus обновляет статус операции.
	UpdateStatus(ctx context.Context, id uuid.UUID, status orchestrator.OperationStatus, result string, errorMsg string) error

	// AssignAgent назначает агента для выполнения операции.
	AssignAgent(ctx context.Context, operationID uuid.UUID, agentID string) error
}
