package parser

import (
	"context"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/orchestrator"
)

// ExpressionParser определяет интерфейс для парсинга арифметических выражений.
type ExpressionParser interface {
	// Parse разбирает выражение на операции.
	Parse(ctx context.Context, expression string) ([]*orchestrator.Operation, error)

	// Validate проверяет корректность выражения.
	Validate(ctx context.Context, expression string) error
}
