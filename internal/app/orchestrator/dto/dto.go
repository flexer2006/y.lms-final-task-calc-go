// Package dto предоставляет DTO для взаимодействия с API.
package dto

import (
	"time"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/orchestrator"
	"github.com/google/uuid"
)

// CalculateRequest представляет запрос на вычисление выражения.
type CalculateRequest struct {
	Expression string    `json:"expression" validate:"required"`
	UserID     uuid.UUID `json:"user_id" validate:"required"`
}

// CalculationResponse представляет ответ с информацией о вычислении.
type CalculationResponse struct {
	ID           string              `json:"id"`
	UserID       string              `json:"user_id"`
	Expression   string              `json:"expression"`
	Result       string              `json:"result,omitempty"`
	Status       string              `json:"status"`
	ErrorMessage string              `json:"error_message,omitempty"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedAt    time.Time           `json:"updated_at"`
	Operations   []OperationResponse `json:"operations,omitempty"`
}

// OperationResponse представляет информацию об одной операции.
type OperationResponse struct {
	ID             string `json:"id"`
	OperationType  string `json:"operation_type"`
	Operand1       string `json:"operand1"`
	Operand2       string `json:"operand2"`
	Result         string `json:"result,omitempty"`
	Status         string `json:"status"`
	ErrorMessage   string `json:"error_message,omitempty"`
	ProcessingTime int64  `json:"processing_time_ms,omitempty"`
	AgentID        string `json:"agent_id,omitempty"`
}

// FromCalculation конвертирует доменную модель в DTO.
func FromCalculation(calc *orchestrator.Calculation) *CalculationResponse {
	if calc == nil {
		return nil
	}

	resp := &CalculationResponse{
		ID:           calc.ID.String(),
		UserID:       calc.UserID.String(),
		Expression:   calc.Expression,
		Result:       calc.Result,
		Status:       string(calc.Status),
		ErrorMessage: calc.ErrorMessage,
		CreatedAt:    calc.CreatedAt,
		UpdatedAt:    calc.UpdatedAt,
	}

	// Если в модели есть операции, конвертируем их тоже
	if len(calc.Operations) > 0 {
		resp.Operations = make([]OperationResponse, len(calc.Operations))
		for i, op := range calc.Operations {
			resp.Operations[i] = FromOperation(&op)
		}
	}

	return resp
}

// FromCalculationList конвертирует список вычислений в список DTO.
func FromCalculationList(calcs []*orchestrator.Calculation) []*CalculationResponse {
	result := make([]*CalculationResponse, len(calcs))
	for i, calc := range calcs {
		result[i] = FromCalculation(calc)
	}
	return result
}

// GetOperationTypeString возвращает строковое представление типа операции.
func GetOperationTypeString(opType orchestrator.OperationType) string {
	switch opType {
	case orchestrator.OperationTypeAddition:
		return "ADDITION"
	case orchestrator.OperationTypeSubtraction:
		return "SUBTRACTION"
	case orchestrator.OperationTypeMultiplication:
		return "MULTIPLICATION"
	case orchestrator.OperationTypeDivision:
		return "DIVISION"
	default:
		return "UNSPECIFIED"
	}
}

// FromOperation конвертирует доменную модель операции в DTO.
func FromOperation(op *orchestrator.Operation) OperationResponse {
	if op == nil {
		return OperationResponse{}
	}

	return OperationResponse{
		ID:             op.ID.String(),
		OperationType:  GetOperationTypeString(op.OperationType),
		Operand1:       op.Operand1,
		Operand2:       op.Operand2,
		Result:         op.Result,
		Status:         string(op.Status),
		ErrorMessage:   op.ErrorMessage,
		ProcessingTime: op.ProcessingTime,
		AgentID:        op.AgentID,
	}
}
