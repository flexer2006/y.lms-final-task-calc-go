package orchestrator

import (
	"context"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/orchestrator"
	"github.com/google/uuid"
)

// UseCaseCalculation определяет основной порт для операций вычисления.
type UseCaseCalculation interface {
	// CalculateExpression создаёт новое вычисление для выражения.
	CalculateExpression(ctx context.Context, userID uuid.UUID, expression string) (*orchestrator.Calculation, error)

	// GetCalculation возвращает вычисление по ID.
	GetCalculation(ctx context.Context, calculationID uuid.UUID, userID uuid.UUID) (*orchestrator.Calculation, error)

	// ListCalculations возвращает список вычислений пользователя.
	ListCalculations(ctx context.Context, userID uuid.UUID) ([]*orchestrator.Calculation, error)

	// ProcessPendingOperations запускает обработку ожидающих операций.
	ProcessPendingOperations(ctx context.Context) error

	// UpdateCalculationStatus обновляет статус вычисления после обработки операций.
	UpdateCalculationStatus(ctx context.Context, calculationID uuid.UUID) error
}
