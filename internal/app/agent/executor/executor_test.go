// Package executor_test содержит тесты для пакета executor.
package executor

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/agent"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/orchestrator"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

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

func TestNewOperationExecutor(t *testing.T) {
	t.Run("Valid parameters", func(t *testing.T) {
		pool := &MockAgentPool{}
		executor := NewOperationExecutor(pool, 3, 200*time.Millisecond)

		assert.NotNil(t, executor)
		assert.Equal(t, pool, executor.pool)
		assert.Equal(t, 3, executor.maxRetries)
		assert.Equal(t, 200*time.Millisecond, executor.retryDelay)
		assert.NotNil(t, executor.assignedAgents)
	})

	t.Run("Nil pool", func(t *testing.T) {
		executor := NewOperationExecutor(nil, 3, 200*time.Millisecond)
		assert.Nil(t, executor)
	})

	t.Run("Negative parameters", func(t *testing.T) {
		pool := &MockAgentPool{}
		executor := NewOperationExecutor(pool, -1, -100*time.Millisecond)

		assert.NotNil(t, executor)
		assert.Equal(t, 0, executor.maxRetries)
		assert.Equal(t, 100*time.Millisecond, executor.retryDelay)
	})
}

func TestOperationAgentMapping(t *testing.T) {
	t.Run("Assign and retrieve agent", func(t *testing.T) {
		pool := new(MockAgentPool)
		executor := NewOperationExecutor(pool, 3, 100*time.Millisecond)

		operationID := uuid.New()
		agentID := "test-agent-1"

		executor.recordAgentAssignment(operationID, agentID)

		retrievedAgentID, found := executor.GetOperationAgent(operationID)
		assert.True(t, found)
		assert.Equal(t, agentID, retrievedAgentID)
	})

	t.Run("Search for non-existent operation", func(t *testing.T) {
		pool := new(MockAgentPool)
		executor := NewOperationExecutor(pool, 3, 100*time.Millisecond)

		_, found := executor.GetOperationAgent(uuid.New())
		assert.False(t, found)
	})

	t.Run("Search for nil UUID", func(t *testing.T) {
		pool := new(MockAgentPool)
		executor := NewOperationExecutor(pool, 3, 100*time.Millisecond)

		agentID, found := executor.GetOperationAgent(uuid.Nil)
		assert.False(t, found)
		assert.Equal(t, "", agentID)
	})

	t.Run("Remove assignment", func(t *testing.T) {
		pool := new(MockAgentPool)
		executor := NewOperationExecutor(pool, 3, 100*time.Millisecond)

		operationID := uuid.New()
		agentID := "test-agent-1"

		executor.recordAgentAssignment(operationID, agentID)
		executor.removeAgentAssignment(operationID)

		_, found := executor.GetOperationAgent(operationID)
		assert.False(t, found)
	})

	t.Run("ReleaseOperation removes assignment", func(t *testing.T) {
		pool := new(MockAgentPool)
		executor := NewOperationExecutor(pool, 3, 100*time.Millisecond)

		operationID := uuid.New()
		agentID := "test-agent-1"

		executor.recordAgentAssignment(operationID, agentID)
		executor.ReleaseOperation(operationID)

		_, found := executor.GetOperationAgent(operationID)
		assert.False(t, found)
	})

	t.Run("Concurrent access to assignedAgents", func(t *testing.T) {
		pool := new(MockAgentPool)
		executor := NewOperationExecutor(pool, 3, 100*time.Millisecond)

		concurrentOperations := 100
		var wg sync.WaitGroup
		wg.Add(concurrentOperations)

		for i := 0; i < concurrentOperations; i++ {
			go func() {
				defer wg.Done()
				operationID := uuid.New()
				agentID := "test-agent-" + uuid.New().String()

				executor.recordAgentAssignment(operationID, agentID)

				retrievedAgentID, found := executor.GetOperationAgent(operationID)
				assert.True(t, found)
				assert.Equal(t, agentID, retrievedAgentID)

				executor.removeAgentAssignment(operationID)

				_, found = executor.GetOperationAgent(operationID)
				assert.False(t, found)
			}()
		}

		wg.Wait()
	})
}

func TestGetAgentsStatus(t *testing.T) {
	t.Run("Successfully retrieve agents list", func(t *testing.T) {
		pool := new(MockAgentPool)
		executor := NewOperationExecutor(pool, 3, 100*time.Millisecond)

		mockAgents := []*agent.Agent{
			{ID: "agent-1", Status: agent.AgentStatusOnline},
			{ID: "agent-2", Status: agent.AgentStatusBusy},
		}

		pool.On("ListAgents").Return(mockAgents, nil)

		agents, err := executor.GetAgentsStatus(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, mockAgents, agents)
		pool.AssertExpectations(t)
	})

	t.Run("Error retrieving agents list", func(t *testing.T) {
		pool := new(MockAgentPool)
		executor := NewOperationExecutor(pool, 3, 100*time.Millisecond)

		expectedErr := errors.New("database error")
		pool.On("ListAgents").Return(nil, expectedErr)

		agents, err := executor.GetAgentsStatus(context.Background())

		assert.Error(t, err)
		assert.Nil(t, agents)
		assert.Contains(t, err.Error(), "failed to list agents")
		pool.AssertExpectations(t)
	})

	t.Run("Context cancellation", func(t *testing.T) {
		pool := new(MockAgentPool)
		executor := NewOperationExecutor(pool, 3, 100*time.Millisecond)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		agents, err := executor.GetAgentsStatus(ctx)

		assert.Error(t, err)
		assert.Nil(t, agents)
		assert.Contains(t, err.Error(), "context cancelled")
	})
}

func TestGetAssignedOperationsCount(t *testing.T) {
	t.Run("Count assigned operations", func(t *testing.T) {
		pool := new(MockAgentPool)
		executor := NewOperationExecutor(pool, 3, 100*time.Millisecond)

		count := executor.GetAssignedOperationsCount()
		assert.Equal(t, 0, count)

		executor.recordAgentAssignment(uuid.New(), "agent-1")
		executor.recordAgentAssignment(uuid.New(), "agent-2")
		executor.recordAgentAssignment(uuid.New(), "agent-3")

		count = executor.GetAssignedOperationsCount()
		assert.Equal(t, 3, count)

		opID := uuid.New()
		executor.recordAgentAssignment(opID, "agent-4")
		count = executor.GetAssignedOperationsCount()
		assert.Equal(t, 4, count)

		executor.removeAgentAssignment(opID)
		count = executor.GetAssignedOperationsCount()
		assert.Equal(t, 3, count)
	})
}
