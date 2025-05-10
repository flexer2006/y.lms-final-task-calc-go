package worker

import (
	"context"
	"errors"
	"testing"
	"time"

	domainerrors "github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/errord"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/agent"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/orchestrator"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

type MockOperationRepository struct {
	mock.Mock
}

func (m *MockOperationRepository) Create(ctx context.Context, operation *orchestrator.Operation) (*orchestrator.Operation, error) {
	args := m.Called(ctx, operation)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*orchestrator.Operation), args.Error(1)
}

func (m *MockOperationRepository) CreateBatch(ctx context.Context, operations []*orchestrator.Operation) error {
	args := m.Called(ctx, operations)
	return args.Error(0)
}

func (m *MockOperationRepository) FindByID(ctx context.Context, id uuid.UUID) (*orchestrator.Operation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*orchestrator.Operation), args.Error(1)
}

func (m *MockOperationRepository) FindByCalculationID(ctx context.Context, calculationID uuid.UUID) ([]*orchestrator.Operation, error) {
	args := m.Called(ctx, calculationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*orchestrator.Operation), args.Error(1)
}

func (m *MockOperationRepository) GetPendingOperations(ctx context.Context, limit int) ([]*orchestrator.Operation, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*orchestrator.Operation), args.Error(1)
}

func (m *MockOperationRepository) Update(ctx context.Context, operation *orchestrator.Operation) error {
	args := m.Called(ctx, operation)
	return args.Error(0)
}

func (m *MockOperationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status orchestrator.OperationStatus, result string, errorMsg string) error {
	args := m.Called(ctx, id, status, result, errorMsg)
	return args.Error(0)
}

func (m *MockOperationRepository) AssignAgent(ctx context.Context, operationID uuid.UUID, agentID string) error {
	args := m.Called(ctx, operationID, agentID)
	return args.Error(0)
}

func TestStartStop(t *testing.T) {
	repo := new(MockOperationRepository)
	w, err := NewWorker("agent-test", 3, nil, repo)
	require.NoError(t, err)

	t.Run("Start changes status to online", func(t *testing.T) {
		ctx := context.Background()

		assert.Equal(t, agent.AgentStatusOffline, w.agent.Status)
		assert.Equal(t, false, w.IsRunning())

		repo.On("UpdateStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

		w.Start(ctx)

		assert.Equal(t, agent.AgentStatusOnline, w.agent.Status)
		assert.Equal(t, true, w.IsRunning())
	})

	t.Run("Start is idempotent", func(t *testing.T) {
		ctx := context.Background()
		initialStatus := w.agent.Status

		w.Start(ctx)

		assert.Equal(t, initialStatus, w.agent.Status)
		assert.Equal(t, true, w.IsRunning())
	})

	t.Run("Stop changes status to offline", func(t *testing.T) {
		w.Stop()

		assert.Equal(t, agent.AgentStatusOffline, w.agent.Status)
		assert.Equal(t, false, w.IsRunning())
	})

	t.Run("Stop is idempotent", func(t *testing.T) {
		initialStatus := w.agent.Status

		w.Stop()

		assert.Equal(t, initialStatus, w.agent.Status)
		assert.Equal(t, false, w.IsRunning())
	})
}

func TestPerformOperation(t *testing.T) {
	tests := []struct {
		name          string
		operation     *orchestrator.Operation
		isRunning     bool
		agentStatus   agent.AgentStatus
		currentLoad   int
		maxCapacity   int
		expectError   bool
		expectedError error
	}{
		{
			name:          "Nil operation",
			operation:     nil,
			isRunning:     true,
			agentStatus:   agent.AgentStatusOnline,
			currentLoad:   0,
			maxCapacity:   3,
			expectError:   true,
			expectedError: domainerrors.ErrNilOperation,
		},
		{
			name:          "Agent not running",
			operation:     &orchestrator.Operation{},
			isRunning:     false,
			agentStatus:   agent.AgentStatusOffline,
			currentLoad:   0,
			maxCapacity:   3,
			expectError:   true,
			expectedError: domainerrors.ErrAgentNotRunning,
		},
		{
			name:          "Agent not online",
			operation:     &orchestrator.Operation{},
			isRunning:     true,
			agentStatus:   agent.AgentStatusOffline,
			currentLoad:   0,
			maxCapacity:   3,
			expectError:   true,
			expectedError: domainerrors.ErrAgentNotRunning,
		},
		{
			name:          "Agent at capacity",
			operation:     &orchestrator.Operation{},
			isRunning:     true,
			agentStatus:   agent.AgentStatusOnline,
			currentLoad:   3,
			maxCapacity:   3,
			expectError:   true,
			expectedError: domainerrors.ErrAgentAtCapacity,
		},
		{
			name:        "Valid operation",
			operation:   &orchestrator.Operation{ID: uuid.New()},
			isRunning:   true,
			agentStatus: agent.AgentStatusOnline,
			currentLoad: 0,
			maxCapacity: 3,
			expectError: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(MockOperationRepository)
			w, err := NewWorker("agent-test", tc.maxCapacity, nil, repo)
			require.NoError(t, err)

			if tc.isRunning {
				ctx := context.Background()
				repo.On("UpdateStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				w.Start(ctx)
			}

			w.UpdateStatus(tc.agentStatus, tc.currentLoad)

			result, err := w.PerformOperation(tc.operation)

			if tc.expectError {
				assert.Error(t, err)
				if tc.expectedError != nil {
					assert.True(t, errors.Is(err, tc.expectedError) ||
						errors.Is(errors.Unwrap(err), tc.expectedError),
						"expected error containing %v, got %v", tc.expectedError, err)
				}
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.currentLoad+1, w.CurrentLoad())
			}
		})
	}
}

func TestGetStatus(t *testing.T) {
	tests := []struct {
		name           string
		isRunning      bool
		currentLoad    int
		maxCapacity    int
		expectedStatus agent.AgentStatus
	}{
		{
			name:           "Offline when not running",
			isRunning:      false,
			currentLoad:    0,
			maxCapacity:    3,
			expectedStatus: agent.AgentStatusOffline,
		},
		{
			name:           "Online when running and not at capacity",
			isRunning:      true,
			currentLoad:    2,
			maxCapacity:    3,
			expectedStatus: agent.AgentStatusOnline,
		},
		{
			name:           "Busy when running and at capacity",
			isRunning:      true,
			currentLoad:    3,
			maxCapacity:    3,
			expectedStatus: agent.AgentStatusBusy,
		},
		{
			name:           "Busy when running and over capacity",
			isRunning:      true,
			currentLoad:    4,
			maxCapacity:    3,
			expectedStatus: agent.AgentStatusBusy,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(MockOperationRepository)
			w, err := NewWorker("agent-test", tc.maxCapacity, nil, repo)
			require.NoError(t, err)

			if tc.isRunning {
				ctx := context.Background()
				repo.On("UpdateStatus", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)
				w.Start(ctx)
			}

			w.mu.Lock()
			w.agent.CurrentLoad = tc.currentLoad
			w.mu.Unlock()

			status := w.GetStatus()

			assert.NotNil(t, status)
			assert.Equal(t, tc.expectedStatus, status.Status)
			assert.Equal(t, tc.currentLoad, status.CurrentLoad)
			assert.Equal(t, tc.maxCapacity, status.MaxCapacity)
			assert.True(t, status.UptimeSeconds >= 0)
		})
	}
}

func TestUpdateStatus(t *testing.T) {
	tests := []struct {
		name         string
		newStatus    agent.AgentStatus
		newLoad      int
		expectedLoad int
	}{
		{
			name:         "Update to online",
			newStatus:    agent.AgentStatusOnline,
			newLoad:      2,
			expectedLoad: 2,
		},
		{
			name:         "Update to offline",
			newStatus:    agent.AgentStatusOffline,
			newLoad:      0,
			expectedLoad: 0,
		},
		{
			name:         "Update to busy",
			newStatus:    agent.AgentStatusBusy,
			newLoad:      3,
			expectedLoad: 3,
		},
		{
			name:         "Negative load corrected to zero",
			newStatus:    agent.AgentStatusOnline,
			newLoad:      -1,
			expectedLoad: 0,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(MockOperationRepository)
			w, err := NewWorker("agent-test", 3, nil, repo)
			require.NoError(t, err)

			oldLastOperation := w.agent.LastOperationAt
			time.Sleep(1 * time.Millisecond)

			w.UpdateStatus(tc.newStatus, tc.newLoad)

			assert.Equal(t, tc.newStatus, w.agent.Status)
			assert.Equal(t, tc.expectedLoad, w.agent.CurrentLoad)
			assert.True(t, w.agent.LastOperationAt.After(oldLastOperation))
		})
	}
}

func TestExecuteOperation(t *testing.T) {
	tests := []struct {
		name            string
		operation       *orchestrator.Operation
		setupRepo       func(*MockOperationRepository)
		expectedResult  string
		expectError     bool
		expectedErrorIs error
	}{
		{
			name:            "Nil operation",
			operation:       nil,
			expectedResult:  "",
			expectError:     true,
			expectedErrorIs: domainerrors.ErrNilOperation,
		},
		{
			name: "Addition operation",
			operation: &orchestrator.Operation{
				ID:            uuid.New(),
				OperationType: orchestrator.OperationTypeAddition,
				Operand1:      "5",
				Operand2:      "3",
			},
			expectedResult: "8",
			expectError:    false,
		},
		{
			name: "Subtraction operation",
			operation: &orchestrator.Operation{
				ID:            uuid.New(),
				OperationType: orchestrator.OperationTypeSubtraction,
				Operand1:      "5",
				Operand2:      "3",
			},
			expectedResult: "2",
			expectError:    false,
		},
		{
			name: "Multiplication operation",
			operation: &orchestrator.Operation{
				ID:            uuid.New(),
				OperationType: orchestrator.OperationTypeMultiplication,
				Operand1:      "5",
				Operand2:      "3",
			},
			expectedResult: "15",
			expectError:    false,
		},
		{
			name: "Division operation",
			operation: &orchestrator.Operation{
				ID:            uuid.New(),
				OperationType: orchestrator.OperationTypeDivision,
				Operand1:      "6",
				Operand2:      "3",
			},
			expectedResult: "2",
			expectError:    false,
		},
		{
			name: "Division by zero",
			operation: &orchestrator.Operation{
				ID:            uuid.New(),
				OperationType: orchestrator.OperationTypeDivision,
				Operand1:      "5",
				Operand2:      "0",
			},
			expectedResult:  "",
			expectError:     true,
			expectedErrorIs: domainerrors.ErrDivisionByZero,
		},
		{
			name: "Invalid operand",
			operation: &orchestrator.Operation{
				ID:            uuid.New(),
				OperationType: orchestrator.OperationTypeAddition,
				Operand1:      "five",
				Operand2:      "3",
			},
			expectedResult:  "",
			expectError:     true,
			expectedErrorIs: domainerrors.ErrInvalidOperand,
		},
		{
			name: "Unsupported operation",
			operation: &orchestrator.Operation{
				ID:            uuid.New(),
				OperationType: 99,
				Operand1:      "5",
				Operand2:      "3",
			},
			expectedResult:  "",
			expectError:     true,
			expectedErrorIs: domainerrors.ErrUnsupportedOp,
		},
		{
			name: "Reference operand",
			operation: &orchestrator.Operation{
				ID:            uuid.New(),
				OperationType: orchestrator.OperationTypeAddition,
				Operand1:      "ref:12345678-1234-1234-1234-123456789abc",
				Operand2:      "3",
			},
			setupRepo: func(repo *MockOperationRepository) {
				refID, _ := uuid.Parse("12345678-1234-1234-1234-123456789abc")
				repo.On("FindByID", mock.Anything, refID).Return(
					&orchestrator.Operation{
						ID:     refID,
						Result: "5",
						Status: orchestrator.OperationStatusCompleted,
					}, nil)
			},
			expectedResult: "8",
			expectError:    false,
		},
		{
			name: "Reference not found",
			operation: &orchestrator.Operation{
				ID:            uuid.New(),
				OperationType: orchestrator.OperationTypeAddition,
				Operand1:      "ref:12345678-1234-1234-1234-123456789abc",
				Operand2:      "3",
			},
			setupRepo: func(repo *MockOperationRepository) {
				refID, _ := uuid.Parse("12345678-1234-1234-1234-123456789abc")
				repo.On("FindByID", mock.Anything, refID).Return(nil, errors.New("not found"))
			},
			expectedResult: "",
			expectError:    true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			repo := new(MockOperationRepository)
			if tc.setupRepo != nil {
				tc.setupRepo(repo)
			}

			w, err := NewWorker("agent-test", 3, nil, repo)
			require.NoError(t, err)

			ctx := context.Background()
			result, err := w.executeOperation(ctx, tc.operation)

			if tc.expectError {
				assert.Error(t, err)
				if tc.expectedErrorIs != nil {
					assert.True(t, errors.Is(err, tc.expectedErrorIs) ||
						errors.Is(errors.Unwrap(err), tc.expectedErrorIs),
						"expected error containing %v, got %v", tc.expectedErrorIs, err)
				}
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResult, result)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestFormatNumericResult(t *testing.T) {
	tests := []struct {
		name           string
		input          float64
		expectedOutput string
	}{
		{
			name:           "Integer result",
			input:          5.0,
			expectedOutput: "5",
		},
		{
			name:           "Decimal result",
			input:          5.5,
			expectedOutput: "5.5",
		},
		{
			name:           "Large number",
			input:          1000000.0,
			expectedOutput: "1000000",
		},
		{
			name:           "Small decimal",
			input:          0.123,
			expectedOutput: "0.123",
		},
		{
			name:           "Negative integer",
			input:          -5.0,
			expectedOutput: "-5",
		},
		{
			name:           "Negative decimal",
			input:          -5.5,
			expectedOutput: "-5.5",
		},
		{
			name:           "Zero",
			input:          0.0,
			expectedOutput: "0",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatNumericResult(tc.input)
			assert.Equal(t, tc.expectedOutput, result)
		})
	}
}

func TestIsRunningAndCurrentLoad(t *testing.T) {
	repo := new(MockOperationRepository)
	w, err := NewWorker("agent-test", 3, nil, repo)
	require.NoError(t, err)

	t.Run("IsRunning returns false when not started", func(t *testing.T) {
		assert.False(t, w.IsRunning())
	})

	t.Run("IsRunning returns true after start", func(t *testing.T) {
		ctx := context.Background()
		w.Start(ctx)
		assert.True(t, w.IsRunning())
	})

	t.Run("IsRunning returns false after stop", func(t *testing.T) {
		w.Stop()
		assert.False(t, w.IsRunning())
	})

	t.Run("CurrentLoad reflects agent load", func(t *testing.T) {
		w.mu.Lock()
		w.agent.CurrentLoad = 2
		w.mu.Unlock()

		assert.Equal(t, 2, w.CurrentLoad())

		w.mu.Lock()
		w.agent.CurrentLoad = 0
		w.mu.Unlock()

		assert.Equal(t, 0, w.CurrentLoad())
	})
}
