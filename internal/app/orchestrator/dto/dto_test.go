package dto_test

import (
	"testing"
	"time"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/app/orchestrator/dto"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/orchestrator"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestFromCalculation(t *testing.T) {
	fixedTime := time.Date(2023, 5, 15, 10, 30, 0, 0, time.UTC)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	calcID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	opID := uuid.MustParse("00000000-0000-0000-0000-000000000003")

	tests := []struct {
		name     string
		input    *orchestrator.Calculation
		expected *dto.CalculationResponse
	}{
		{
			name:     "Nil calculation",
			input:    nil,
			expected: nil,
		},
		{
			name: "Simple calculation without operations",
			input: &orchestrator.Calculation{
				ID:           calcID,
				UserID:       userID,
				Expression:   "2+2",
				Result:       "4",
				Status:       orchestrator.CalculationStatusCompleted,
				ErrorMessage: "",
				CreatedAt:    fixedTime,
				UpdatedAt:    fixedTime,
			},
			expected: &dto.CalculationResponse{
				ID:           calcID.String(),
				UserID:       userID.String(),
				Expression:   "2+2",
				Result:       "4",
				Status:       "COMPLETED",
				ErrorMessage: "",
				CreatedAt:    fixedTime,
				UpdatedAt:    fixedTime,
			},
		},
		{
			name: "Calculation with error",
			input: &orchestrator.Calculation{
				ID:           calcID,
				UserID:       userID,
				Expression:   "2/0",
				Result:       "",
				Status:       orchestrator.CalculationStatusError,
				ErrorMessage: "division by zero",
				CreatedAt:    fixedTime,
				UpdatedAt:    fixedTime,
			},
			expected: &dto.CalculationResponse{
				ID:           calcID.String(),
				UserID:       userID.String(),
				Expression:   "2/0",
				Result:       "",
				Status:       "ERROR",
				ErrorMessage: "division by zero",
				CreatedAt:    fixedTime,
				UpdatedAt:    fixedTime,
			},
		},
		{
			name: "Calculation with operations",
			input: &orchestrator.Calculation{
				ID:           calcID,
				UserID:       userID,
				Expression:   "2+3*4",
				Result:       "14",
				Status:       orchestrator.CalculationStatusCompleted,
				ErrorMessage: "",
				CreatedAt:    fixedTime,
				UpdatedAt:    fixedTime,
				Operations: []orchestrator.Operation{
					{
						ID:             opID,
						CalculationID:  calcID,
						OperationType:  orchestrator.OperationTypeAddition,
						Operand1:       "2",
						Operand2:       "12",
						Result:         "14",
						Status:         orchestrator.OperationStatusCompleted,
						ErrorMessage:   "",
						ProcessingTime: 100,
						AgentID:        "agent-1",
					},
				},
			},
			expected: &dto.CalculationResponse{
				ID:           calcID.String(),
				UserID:       userID.String(),
				Expression:   "2+3*4",
				Result:       "14",
				Status:       "COMPLETED",
				ErrorMessage: "",
				CreatedAt:    fixedTime,
				UpdatedAt:    fixedTime,
				Operations: []dto.OperationResponse{
					{
						ID:             opID.String(),
						OperationType:  "ADDITION",
						Operand1:       "2",
						Operand2:       "12",
						Result:         "14",
						Status:         "COMPLETED",
						ErrorMessage:   "",
						ProcessingTime: 100,
						AgentID:        "agent-1",
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := dto.FromCalculation(tt.input)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestFromCalculationList(t *testing.T) {
	fixedTime := time.Date(2023, 5, 15, 10, 30, 0, 0, time.UTC)
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	calc1ID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	calc2ID := uuid.MustParse("00000000-0000-0000-0000-000000000003")

	tests := []struct {
		name     string
		input    []*orchestrator.Calculation
		expected []*dto.CalculationResponse
	}{
		{
			name:     "Empty list",
			input:    []*orchestrator.Calculation{},
			expected: []*dto.CalculationResponse{},
		},
		{
			name: "List with multiple calculations",
			input: []*orchestrator.Calculation{
				{
					ID:           calc1ID,
					UserID:       userID,
					Expression:   "1+2",
					Result:       "3",
					Status:       orchestrator.CalculationStatusCompleted,
					ErrorMessage: "",
					CreatedAt:    fixedTime,
					UpdatedAt:    fixedTime,
				},
				{
					ID:           calc2ID,
					UserID:       userID,
					Expression:   "3*4",
					Result:       "12",
					Status:       orchestrator.CalculationStatusCompleted,
					ErrorMessage: "",
					CreatedAt:    fixedTime,
					UpdatedAt:    fixedTime,
				},
			},
			expected: []*dto.CalculationResponse{
				{
					ID:           calc1ID.String(),
					UserID:       userID.String(),
					Expression:   "1+2",
					Result:       "3",
					Status:       "COMPLETED",
					ErrorMessage: "",
					CreatedAt:    fixedTime,
					UpdatedAt:    fixedTime,
				},
				{
					ID:           calc2ID.String(),
					UserID:       userID.String(),
					Expression:   "3*4",
					Result:       "12",
					Status:       "COMPLETED",
					ErrorMessage: "",
					CreatedAt:    fixedTime,
					UpdatedAt:    fixedTime,
				},
			},
		},
		{
			name:     "Nil list",
			input:    nil,
			expected: []*dto.CalculationResponse{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := dto.FromCalculationList(tt.input)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestGetOperationTypeString(t *testing.T) {
	tests := []struct {
		name     string
		input    orchestrator.OperationType
		expected string
	}{
		{
			name:     "Addition",
			input:    orchestrator.OperationTypeAddition,
			expected: "ADDITION",
		},
		{
			name:     "Subtraction",
			input:    orchestrator.OperationTypeSubtraction,
			expected: "SUBTRACTION",
		},
		{
			name:     "Multiplication",
			input:    orchestrator.OperationTypeMultiplication,
			expected: "MULTIPLICATION",
		},
		{
			name:     "Division",
			input:    orchestrator.OperationTypeDivision,
			expected: "DIVISION",
		},
		{
			name:     "Invalid",
			input:    orchestrator.OperationType(99),
			expected: "UNSPECIFIED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := dto.GetOperationTypeString(tt.input)
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestFromOperation(t *testing.T) {
	opID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	tests := []struct {
		name     string
		input    *orchestrator.Operation
		expected dto.OperationResponse
	}{
		{
			name:  "Nil operation",
			input: nil,
			expected: dto.OperationResponse{
				ID:             "",
				OperationType:  "",
				Operand1:       "",
				Operand2:       "",
				Result:         "",
				Status:         "",
				ErrorMessage:   "",
				ProcessingTime: 0,
				AgentID:        "",
			},
		},
		{
			name: "Simple addition operation",
			input: &orchestrator.Operation{
				ID:             opID,
				OperationType:  orchestrator.OperationTypeAddition,
				Operand1:       "5",
				Operand2:       "3",
				Result:         "8",
				Status:         orchestrator.OperationStatusCompleted,
				ErrorMessage:   "",
				ProcessingTime: 150,
				AgentID:        "agent-1",
			},
			expected: dto.OperationResponse{
				ID:             opID.String(),
				OperationType:  "ADDITION",
				Operand1:       "5",
				Operand2:       "3",
				Result:         "8",
				Status:         "COMPLETED",
				ErrorMessage:   "",
				ProcessingTime: 150,
				AgentID:        "agent-1",
			},
		},
		{
			name: "Division by zero",
			input: &orchestrator.Operation{
				ID:             opID,
				OperationType:  orchestrator.OperationTypeDivision,
				Operand1:       "5",
				Operand2:       "0",
				Result:         "",
				Status:         orchestrator.OperationStatusError,
				ErrorMessage:   "division by zero",
				ProcessingTime: 50,
				AgentID:        "agent-2",
			},
			expected: dto.OperationResponse{
				ID:             opID.String(),
				OperationType:  "DIVISION",
				Operand1:       "5",
				Operand2:       "0",
				Result:         "",
				Status:         "ERROR",
				ErrorMessage:   "division by zero",
				ProcessingTime: 50,
				AgentID:        "agent-2",
			},
		},
		{
			name: "Multiplication without agent",
			input: &orchestrator.Operation{
				ID:             opID,
				OperationType:  orchestrator.OperationTypeMultiplication,
				Operand1:       "2",
				Operand2:       "3",
				Result:         "6",
				Status:         orchestrator.OperationStatusCompleted,
				ErrorMessage:   "",
				ProcessingTime: 80,
				AgentID:        "",
			},
			expected: dto.OperationResponse{
				ID:             opID.String(),
				OperationType:  "MULTIPLICATION",
				Operand1:       "2",
				Operand2:       "3",
				Result:         "6",
				Status:         "COMPLETED",
				ErrorMessage:   "",
				ProcessingTime: 80,
				AgentID:        "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := dto.FromOperation(tt.input)
			assert.Equal(t, tt.expected, actual)
		})
	}
}
