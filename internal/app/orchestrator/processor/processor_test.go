package processor_test

import (
	"context"
	"errors"
	"testing"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/app/orchestrator/processor"
	domainerrors "github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/errord"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/agent"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/orchestrator"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
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

type MockCalculationRepository struct {
	mock.Mock
}

func (m *MockCalculationRepository) Create(ctx context.Context, calculation *orchestrator.Calculation) (*orchestrator.Calculation, error) {
	args := m.Called(ctx, calculation)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*orchestrator.Calculation), args.Error(1)
}

func (m *MockCalculationRepository) FindByID(ctx context.Context, id uuid.UUID) (*orchestrator.Calculation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*orchestrator.Calculation), args.Error(1)
}

func (m *MockCalculationRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*orchestrator.Calculation, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*orchestrator.Calculation), args.Error(1)
}

func (m *MockCalculationRepository) Update(ctx context.Context, calculation *orchestrator.Calculation) error {
	args := m.Called(ctx, calculation)
	return args.Error(0)
}

func (m *MockCalculationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status orchestrator.CalculationStatus, result string, errorMsg string) error {
	args := m.Called(ctx, id, status, result, errorMsg)
	return args.Error(0)
}

func (m *MockCalculationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockCalcUseCase struct {
	mock.Mock
}

func (m *MockCalcUseCase) CalculateExpression(ctx context.Context, userID uuid.UUID, expression string) (*orchestrator.Calculation, error) {
	args := m.Called(ctx, userID, expression)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*orchestrator.Calculation), args.Error(1)
}

func (m *MockCalcUseCase) GetCalculation(ctx context.Context, calculationID uuid.UUID, userID uuid.UUID) (*orchestrator.Calculation, error) {
	args := m.Called(ctx, calculationID, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*orchestrator.Calculation), args.Error(1)
}

func (m *MockCalcUseCase) ListCalculations(ctx context.Context, userID uuid.UUID) ([]*orchestrator.Calculation, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*orchestrator.Calculation), args.Error(1)
}

func (m *MockCalcUseCase) UpdateCalculationStatus(ctx context.Context, calculationID uuid.UUID) error {
	args := m.Called(ctx, calculationID)
	return args.Error(0)
}

func (m *MockCalcUseCase) ProcessPendingOperations(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockCalcUseCase) Close() error {
	args := m.Called()
	return args.Error(0)
}

type MockOperationExecutor struct {
	mock.Mock
}

func (m *MockOperationExecutor) ExecuteOperation(ctx context.Context, operation *orchestrator.Operation) error {
	args := m.Called(ctx, operation)
	return args.Error(0)
}

func (m *MockOperationExecutor) ReleaseOperation(operationID uuid.UUID) {
	m.Called(operationID)
}

func (m *MockOperationExecutor) GetOperationAgent(operationID uuid.UUID) (string, bool) {
	args := m.Called(operationID)
	return args.String(0), args.Bool(1)
}

func (m *MockOperationExecutor) GetAgentsStatus(ctx context.Context) ([]*agent.Agent, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*agent.Agent), args.Error(1)
}

func (m *MockOperationExecutor) GetAssignedOperationsCount() int {
	args := m.Called()
	return args.Int(0)
}

type MockAgentPool struct {
	mock.Mock
}

func (m *MockAgentPool) Start(ctx context.Context) {
	m.Called(ctx)
}

func (m *MockAgentPool) Stop(ctx context.Context) {
	m.Called(ctx)
}

func (m *MockAgentPool) GetAvailableAgent(operationType int) (*agent.Agent, error) {
	args := m.Called(operationType)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*agent.Agent), args.Error(1)
}

func (m *MockAgentPool) AssignOperation(agentID string, operation *orchestrator.Operation) error {
	args := m.Called(agentID, operation)
	return args.Error(0)
}

func (m *MockAgentPool) GetAgentStatus(agentID string) (*agent.Agent, error) {
	args := m.Called(agentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*agent.Agent), args.Error(1)
}

func (m *MockAgentPool) ListAgents() ([]*agent.Agent, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*agent.Agent), args.Error(1)
}

func TestAssignOperationToAgent(t *testing.T) {
	operationID := uuid.New()

	tests := []struct {
		name          string
		agent         *agent.Agent
		operation     *orchestrator.Operation
		mockSetup     func(*MockOperationRepository, *MockAgentPool)
		expectedError error
	}{
		{
			name: "Successful operation assignment",
			agent: &agent.Agent{
				ID:          "agent-1",
				Status:      agent.AgentStatusOnline,
				CurrentLoad: 0,
				MaxCapacity: 5,
			},
			operation: &orchestrator.Operation{
				ID:            operationID,
				OperationType: orchestrator.OperationTypeAddition,
			},
			mockSetup: func(opRepo *MockOperationRepository, agentPool *MockAgentPool) {
				opRepo.On("UpdateStatus", mock.Anything, operationID, orchestrator.OperationStatusInProgress, "", "").Return(nil)
				agentPool.On("AssignOperation", "agent-1", mock.Anything).Return(nil)
			},
			expectedError: nil,
		},
		{
			name:      "Nil agent",
			agent:     nil,
			operation: &orchestrator.Operation{ID: operationID},
			mockSetup: func(opRepo *MockOperationRepository, agentPool *MockAgentPool) {
			},
			expectedError: domainerrors.ErrInvalidArgs,
		},
		{
			name: "Nil operation",
			agent: &agent.Agent{
				ID: "agent-1",
			},
			operation: nil,
			mockSetup: func(opRepo *MockOperationRepository, agentPool *MockAgentPool) {
			},
			expectedError: domainerrors.ErrInvalidArgs,
		},
		{
			name: "Agent at maximum load",
			agent: &agent.Agent{
				ID:          "agent-1",
				Status:      agent.AgentStatusBusy,
				CurrentLoad: 5,
				MaxCapacity: 5,
			},
			operation: &orchestrator.Operation{
				ID:            operationID,
				OperationType: orchestrator.OperationTypeAddition,
			},
			mockSetup: func(opRepo *MockOperationRepository, agentPool *MockAgentPool) {
			},
			expectedError: errors.New("agent agent-1 is at capacity (5/5)"),
		},
		{
			name: "Assignment error",
			agent: &agent.Agent{
				ID:          "agent-1",
				Status:      agent.AgentStatusOnline,
				CurrentLoad: 0,
				MaxCapacity: 5,
			},
			operation: &orchestrator.Operation{
				ID:            operationID,
				OperationType: orchestrator.OperationTypeAddition,
			},
			mockSetup: func(opRepo *MockOperationRepository, agentPool *MockAgentPool) {
				opRepo.On("UpdateStatus", mock.Anything, operationID, orchestrator.OperationStatusInProgress, "", "").Return(nil)
				agentPool.On("AssignOperation", "agent-1", mock.Anything).Return(errors.New("assignment error"))
			},
			expectedError: errors.New("failed to assign operation to agent agent-1: assignment error"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			opRepo := new(MockOperationRepository)
			calcRepo := new(MockCalculationRepository)
			calcUseCase := new(MockCalcUseCase)
			calcUseCase.On("Close").Return(nil)
			opExecutor := new(MockOperationExecutor)
			agentPool := new(MockAgentPool)

			tc.mockSetup(opRepo, agentPool)

			agentConfig := processor.AgentConfig{
				AgentID:       "test-agent",
				ComputerPower: 5,
			}

			proc := processor.NewProcessor(opRepo, calcRepo, calcUseCase, agentConfig, opExecutor, agentPool)

			err := proc.ExportAssignOperationToAgent(context.Background(), tc.agent, tc.operation)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			opRepo.AssertExpectations(t)
			agentPool.AssertExpectations(t)
		})
	}
}
