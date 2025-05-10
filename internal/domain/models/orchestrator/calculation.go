// Package orchestrator содержит модели для работы с вычислениями.
package orchestrator

import (
	"time"

	"github.com/google/uuid"
)

// CalculationStatus определяет статус вычисления.
type CalculationStatus string

const (
	// CalculationStatusPending - ожидает выполнения.
	CalculationStatusPending CalculationStatus = "PENDING"
	// CalculationStatusInProgress - в процессе выполнения.
	CalculationStatusInProgress CalculationStatus = "IN_PROGRESS"
	// CalculationStatusCompleted - выполнено успешно.
	CalculationStatusCompleted CalculationStatus = "COMPLETED"
	// CalculationStatusError - ошибка выполнения.
	CalculationStatusError CalculationStatus = "ERROR"
)

// Calculation представляет собой вычисление арифметического выражения.
type Calculation struct {
	ID           uuid.UUID         `json:"id"`
	UserID       uuid.UUID         `json:"user_id"`
	Expression   string            `json:"expression"`
	Result       string            `json:"result"`
	Status       CalculationStatus `json:"status"`
	ErrorMessage string            `json:"error_message"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	Operations   []Operation       `json:"operations,omitempty"`
}
