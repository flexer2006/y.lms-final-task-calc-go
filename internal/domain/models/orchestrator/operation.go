// Package orchestrator содержит модели для работы с операциями.
package orchestrator

import (
	"github.com/google/uuid"
)

// OperationType определяет тип арифметической операции.
type OperationType int

const (
	// OperationTypeUnspecified - неопределенная операция.
	OperationTypeUnspecified OperationType = 0
	// OperationTypeAddition - сложение.
	OperationTypeAddition OperationType = 1
	// OperationTypeSubtraction - вычитание.
	OperationTypeSubtraction OperationType = 2
	// OperationTypeMultiplication - умножение.
	OperationTypeMultiplication OperationType = 3
	// OperationTypeDivision - деление.
	OperationTypeDivision OperationType = 4
)

// OperationStatus определяет статус выполнения операции.
type OperationStatus string

const (
	// OperationStatusPending - операция ожидает выполнения.
	OperationStatusPending OperationStatus = "PENDING"
	// OperationStatusInProgress - операция выполняется.
	OperationStatusInProgress OperationStatus = "IN_PROGRESS"
	// OperationStatusCompleted - операция успешно выполнена.
	OperationStatusCompleted OperationStatus = "COMPLETED"
	// OperationStatusError - ошибка выполнения операции.
	OperationStatusError OperationStatus = "ERROR"
)

// Operation представляет одну арифметическую операцию.
type Operation struct {
	ID             uuid.UUID       `json:"id"`
	CalculationID  uuid.UUID       `json:"calculation_id"`
	OperationType  OperationType   `json:"operation_type"`
	Operand1       string          `json:"operand1"`
	Operand2       string          `json:"operand2"`
	Result         string          `json:"result"`
	Status         OperationStatus `json:"status"`
	ErrorMessage   string          `json:"error_message"`
	ProcessingTime int64           `json:"processing_time_ms"`
	AgentID        string          `json:"agent_id,omitempty"`
}
