package pool

import (
	"context"
	"testing"
	"time"

	domainerrors "github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/errord"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/agent"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/orchestrator"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAgentStorage struct {
	mock.Mock
}

func (m *MockAgentStorage) Add(agent *agent.Agent) {
	m.Called(agent)
}

func (m *MockAgentStorage) GetByID(id string) (*agent.Agent, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*agent.Agent), args.Error(1)
}

func (m *MockAgentStorage) GetAvailable() (*agent.Agent, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*agent.Agent), args.Error(1)
}

func (m *MockAgentStorage) UpdateStatus(id string, status agent.AgentStatus, load int, capacity int) error {
	args := m.Called(id, status, load, capacity)
	return args.Error(0)
}

func (m *MockAgentStorage) UpdateStats(id string, completed bool, failed bool) error {
	args := m.Called(id, completed, failed)
	return args.Error(0)
}

func (m *MockAgentStorage) List() []*agent.Agent {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]*agent.Agent)
}

func (m *MockAgentStorage) Remove(id string) error {
	args := m.Called(id)
	return args.Error(0)
}

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

type MockWorker struct {
	mock.Mock
}

func (m *MockWorker) Start(ctx context.Context) {
	m.Called(ctx)
}

func (m *MockWorker) Stop() {
	m.Called()
}

func (m *MockWorker) PerformOperation(operation *orchestrator.Operation) (*orchestrator.Operation, error) {
	args := m.Called(operation)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*orchestrator.Operation), args.Error(1)
}

func (m *MockWorker) GetStatus() *agent.Agent {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*agent.Agent)
}

func (m *MockWorker) IsRunning() bool {
	args := m.Called()
	return args.Bool(0)
}

func (m *MockWorker) CurrentLoad() int {
	args := m.Called()
	return args.Int(0)
}

func TestNewAgentPool(t *testing.T) {
	t.Run("Successful creation", func(t *testing.T) {
		storage := new(MockAgentStorage)
		operationRepo := new(MockOperationRepository)
		operationTimes := map[string]time.Duration{
			"addition":       1 * time.Second,
			"subtraction":    1 * time.Second,
			"multiplication": 2 * time.Second,
			"division":       2 * time.Second,
		}
		capacity := 5

		pool, err := NewAgentPool(storage, operationRepo, operationTimes, capacity)

		assert.NoError(t, err)
		assert.NotNil(t, pool)
		assert.Equal(t, storage, pool.storage)
		assert.Equal(t, operationRepo, pool.operationRepo)
		assert.Equal(t, operationTimes, pool.operationTimes)
		assert.Equal(t, capacity, pool.capacity)
		assert.NotNil(t, pool.workers)
		assert.NotNil(t, pool.ctx)
		assert.NotNil(t, pool.cancel)
		assert.False(t, pool.running)
	})

	t.Run("Missing storage", func(t *testing.T) {
		operationRepo := new(MockOperationRepository)
		pool, err := NewAgentPool(nil, operationRepo, nil, 5)

		assert.Error(t, err)
		assert.Nil(t, pool)
		assert.ErrorIs(t, err, domainerrors.ErrNilStorage)
	})

	t.Run("Missing operation repository", func(t *testing.T) {
		storage := new(MockAgentStorage)
		pool, err := NewAgentPool(storage, nil, nil, 5)

		assert.Error(t, err)
		assert.Nil(t, pool)
		assert.ErrorIs(t, err, domainerrors.ErrNilOperationRepo)
	})

	t.Run("Negative capacity", func(t *testing.T) {
		storage := new(MockAgentStorage)
		operationRepo := new(MockOperationRepository)
		capacity := -1

		pool, err := NewAgentPool(storage, operationRepo, nil, capacity)

		assert.NoError(t, err)
		assert.NotNil(t, pool)
		assert.Equal(t, 4, pool.capacity)
	})

	t.Run("Zero capacity", func(t *testing.T) {
		storage := new(MockAgentStorage)
		operationRepo := new(MockOperationRepository)
		capacity := 0

		pool, err := NewAgentPool(storage, operationRepo, nil, capacity)

		assert.NoError(t, err)
		assert.NotNil(t, pool)
		assert.Equal(t, 4, pool.capacity)
	})

	t.Run("Missing operation times", func(t *testing.T) {
		storage := new(MockAgentStorage)
		operationRepo := new(MockOperationRepository)
		capacity := 5

		pool, err := NewAgentPool(storage, operationRepo, nil, capacity)

		assert.NoError(t, err)
		assert.NotNil(t, pool)
		assert.NotNil(t, pool.operationTimes)
		assert.Len(t, pool.operationTimes, 4)
	})
}

func TestGetAvailableAgent(t *testing.T) {
	t.Run("Pool not running", func(t *testing.T) {
		storage := new(MockAgentStorage)
		operationRepo := new(MockOperationRepository)
		pool, _ := NewAgentPool(storage, operationRepo, nil, 5)

		agent, err := pool.GetAvailableAgent(1)

		assert.Error(t, err)
		assert.Nil(t, agent)
		assert.ErrorIs(t, err, domainerrors.ErrPoolNotRunning)
	})

	t.Run("No available agents", func(t *testing.T) {
		storage := new(MockAgentStorage)
		operationRepo := new(MockOperationRepository)
		pool, _ := NewAgentPool(storage, operationRepo, nil, 5)

		pool.running = true

		agent, err := pool.GetAvailableAgent(1)

		assert.Error(t, err)
		assert.Nil(t, agent)
		assert.ErrorIs(t, err, domainerrors.ErrNoAgentsAvailable)
	})
}

func TestAssignOperation(t *testing.T) {
	t.Run("Nil operation", func(t *testing.T) {
		storage := new(MockAgentStorage)
		operationRepo := new(MockOperationRepository)
		pool, _ := NewAgentPool(storage, operationRepo, nil, 5)

		err := pool.AssignOperation("agent1", nil)

		assert.Error(t, err)
		assert.ErrorIs(t, err, domainerrors.ErrNilOperation)
	})

	t.Run("Empty agent ID", func(t *testing.T) {
		storage := new(MockAgentStorage)
		operationRepo := new(MockOperationRepository)
		pool, _ := NewAgentPool(storage, operationRepo, nil, 5)

		operation := &orchestrator.Operation{ID: uuid.New()}
		err := pool.AssignOperation("", operation)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "empty agent ID")
	})

	t.Run("Agent not found", func(t *testing.T) {
		storage := new(MockAgentStorage)
		operationRepo := new(MockOperationRepository)
		pool, _ := NewAgentPool(storage, operationRepo, nil, 5)

		operation := &orchestrator.Operation{ID: uuid.New()}
		err := pool.AssignOperation("non-existent-agent", operation)

		assert.Error(t, err)
		assert.ErrorContains(t, err, "agent not found")
	})
}

func TestGetAgentStatus(t *testing.T) {
	t.Run("Empty agent ID", func(t *testing.T) {
		storage := new(MockAgentStorage)
		operationRepo := new(MockOperationRepository)
		pool, _ := NewAgentPool(storage, operationRepo, nil, 5)

		status, err := pool.GetAgentStatus("")

		assert.Error(t, err)
		assert.Nil(t, status)
		assert.ErrorContains(t, err, "empty agent ID")
	})

	t.Run("Agent not found", func(t *testing.T) {
		storage := new(MockAgentStorage)
		operationRepo := new(MockOperationRepository)
		pool, _ := NewAgentPool(storage, operationRepo, nil, 5)

		status, err := pool.GetAgentStatus("non-existent-agent")

		assert.Error(t, err)
		assert.Nil(t, status)
		assert.ErrorContains(t, err, "agent not found")
	})
}

func TestListAgents(t *testing.T) {
	t.Run("Empty list", func(t *testing.T) {
		storage := new(MockAgentStorage)
		operationRepo := new(MockOperationRepository)

		storage.On("List").Return(nil)

		pool, _ := NewAgentPool(storage, operationRepo, nil, 5)

		agents, err := pool.ListAgents()

		assert.NoError(t, err)
		assert.NotNil(t, agents)
		assert.Empty(t, agents)
		storage.AssertExpectations(t)
	})

	t.Run("List with agents", func(t *testing.T) {
		storage := new(MockAgentStorage)
		operationRepo := new(MockOperationRepository)

		agentList := []*agent.Agent{
			{ID: "agent1", Status: agent.AgentStatusOnline},
			{ID: "agent2", Status: agent.AgentStatusBusy},
		}

		storage.On("List").Return(agentList)

		pool, _ := NewAgentPool(storage, operationRepo, nil, 5)

		agents, err := pool.ListAgents()

		assert.NoError(t, err)
		assert.Equal(t, agentList, agents)
		storage.AssertExpectations(t)
	})
}

func TestHelperMethods(t *testing.T) {
	t.Run("IsRunning", func(t *testing.T) {
		storage := new(MockAgentStorage)
		operationRepo := new(MockOperationRepository)
		pool, _ := NewAgentPool(storage, operationRepo, nil, 5)

		assert.False(t, pool.IsRunning())

		pool.running = true
		assert.True(t, pool.IsRunning())
	})

	t.Run("GetWorkerCount", func(t *testing.T) {
		storage := new(MockAgentStorage)
		operationRepo := new(MockOperationRepository)
		pool, _ := NewAgentPool(storage, operationRepo, nil, 5)

		assert.Equal(t, 0, pool.GetWorkerCount())

		pool.workers["worker1"] = nil
		pool.workers["worker2"] = nil
		assert.Equal(t, 2, pool.GetWorkerCount())
	})

	t.Run("GetCapacity", func(t *testing.T) {
		storage := new(MockAgentStorage)
		operationRepo := new(MockOperationRepository)
		pool, _ := NewAgentPool(storage, operationRepo, nil, 5)

		assert.Equal(t, 5, pool.GetCapacity())
	})
}
