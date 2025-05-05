package orchestrator

import (
	"context"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/orchestrator"
	"github.com/google/uuid"
)

// CalculationRepository определяет интерфейс для работы с хранилищем вычислений.
type CalculationRepository interface {
	// Create создаёт новое вычисление.
	Create(ctx context.Context, calculation *orchestrator.Calculation) (*orchestrator.Calculation, error)

	// FindByID находит вычисление по ID.
	FindByID(ctx context.Context, id uuid.UUID) (*orchestrator.Calculation, error)

	// FindByUserID находит вычисления пользователя.
	FindByUserID(ctx context.Context, userID uuid.UUID) ([]*orchestrator.Calculation, error)

	// Update обновляет вычисление.
	Update(ctx context.Context, calculation *orchestrator.Calculation) error

	// UpdateStatus обновляет статус вычисления.
	UpdateStatus(ctx context.Context, id uuid.UUID, status orchestrator.CalculationStatus, result string, errorMsg string) error

	// Delete удаляет вычисление.
	Delete(ctx context.Context, id uuid.UUID) error
}
