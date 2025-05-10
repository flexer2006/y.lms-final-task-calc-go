package agent_test

import (
	"testing"
	"time"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/adapters/db/memory/agent"
	agentModel "github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/agent"
)

func TestNewAgentStorage(t *testing.T) {
	storage := agent.NewAgentStorage()
	if storage == nil {
		t.Error("Expected creation of a new storage, got nil")
	}

	agents := storage.List()
	if len(agents) != 0 {
		t.Errorf("New storage should be empty, but contains %d agents", len(agents))
	}
}

func TestAdd(t *testing.T) {
	storage := agent.NewAgentStorage()

	t.Run("AddValidAgent", func(t *testing.T) {
		testAgent := createTestAgent("agent1", agentModel.AgentStatusOnline, 0, 5)
		storage.Add(testAgent)

		agents := storage.List()
		if len(agents) != 1 {
			t.Errorf("Expected 1 agent in storage, got: %d", len(agents))
		}

		retrievedAgent, err := storage.GetByID("agent1")
		if err != nil {
			t.Errorf("Failed to get added agent: %v", err)
		}
		if retrievedAgent.ID != "agent1" {
			t.Errorf("Expected ID agent1, got: %s", retrievedAgent.ID)
		}
	})

	t.Run("AddNilAgent", func(t *testing.T) {
		initialCount := len(storage.List())
		storage.Add(nil)
		if len(storage.List()) != initialCount {
			t.Error("nil agent should not be added to storage")
		}
	})

	t.Run("AddAgentWithEmptyID", func(t *testing.T) {
		initialCount := len(storage.List())
		emptyIDAgent := createTestAgent("", agentModel.AgentStatusOnline, 0, 5)
		storage.Add(emptyIDAgent)
		if len(storage.List()) != initialCount {
			t.Error("Agent with empty ID should not be added to storage")
		}
	})

	t.Run("AddOfflineAgent", func(t *testing.T) {
		offlineAgent := createTestAgent("offline1", agentModel.AgentStatusOffline, 0, 5)
		storage.Add(offlineAgent)

		// Check that agent is added to the general list
		_, err := storage.GetByID("offline1")
		if err != nil {
			t.Errorf("Failed to get added offline agent: %v", err)
		}

		// Check that GetAvailable() does not return this agent
		availableAgent, err := storage.GetAvailable()
		if err != nil && availableAgent != nil && availableAgent.ID == "offline1" {
			t.Error("Offline agent should not be available via GetAvailable()")
		}
	})

	t.Run("AddAgentWithSameID", func(t *testing.T) {
		initialAgent := createTestAgent("duplicate", agentModel.AgentStatusOnline, 0, 5)
		storage.Add(initialAgent)

		updatedAgent := createTestAgent("duplicate", agentModel.AgentStatusOffline, 2, 10)
		storage.Add(updatedAgent)

		retrievedAgent, _ := storage.GetByID("duplicate")
		if retrievedAgent.Status != agentModel.AgentStatusOffline {
			t.Errorf("Agent status should be updated to %s, got: %s",
				agentModel.AgentStatusOffline, retrievedAgent.Status)
		}
		if retrievedAgent.MaxCapacity != 10 {
			t.Errorf("Agent max capacity should be updated to 10, got: %d",
				retrievedAgent.MaxCapacity)
		}
	})
}

func TestGetByID(t *testing.T) {
	storage := agent.NewAgentStorage()
	testAgent := createTestAgent("test_id", agentModel.AgentStatusOnline, 1, 5)
	storage.Add(testAgent)

	t.Run("GetExistingAgent", func(t *testing.T) {
		agent, err := storage.GetByID("test_id")
		if err != nil {
			t.Errorf("Error when getting existing agent: %v", err)
		}
		if agent.ID != "test_id" {
			t.Errorf("Expected ID test_id, got: %s", agent.ID)
		}
	})

	t.Run("GetNonExistentAgent", func(t *testing.T) {
		_, err := storage.GetByID("nonexistent")
		if err == nil {
			t.Error("Expected error when getting non-existent agent, but no error occurred")
		}
		if err != agent.ErrAgentNotFound {
			t.Errorf("Expected error %v, got: %v", agent.ErrAgentNotFound, err)
		}
	})

	t.Run("GetAgentWithEmptyID", func(t *testing.T) {
		_, err := storage.GetByID("")
		if err == nil {
			t.Error("Expected error when getting agent with empty ID, but no error occurred")
		}
		if err != agent.ErrAgentNotFound {
			t.Errorf("Expected error %v, got: %v", agent.ErrAgentNotFound, err)
		}
	})
}

func TestGetAvailable(t *testing.T) {
	storage := agent.NewAgentStorage()

	t.Run("EmptyStorage", func(t *testing.T) {
		_, err := storage.GetAvailable()
		if err == nil {
			t.Error("Expected error when getting available agent from empty storage")
		}
		if err != agent.ErrNoAgentAvailable {
			t.Errorf("Expected error %v, got: %v", agent.ErrNoAgentAvailable, err)
		}
	})

	t.Run("OnlyOfflineAgents", func(t *testing.T) {
		offlineAgent := createTestAgent("offline", agentModel.AgentStatusOffline, 0, 5)
		storage.Add(offlineAgent)

		_, err := storage.GetAvailable()
		if err == nil {
			t.Error("Expected error when getting available agent when all agents are offline")
		}
		if err != agent.ErrNoAgentAvailable {
			t.Errorf("Expected error %v, got: %v", agent.ErrNoAgentAvailable, err)
		}
	})

	t.Run("AllAgentsLoaded", func(t *testing.T) {
		busyAgent := createTestAgent("busy", agentModel.AgentStatusOnline, 5, 5)
		storage.Add(busyAgent)

		_, err := storage.GetAvailable()
		if err == nil {
			t.Error("Expected error when getting available agent when all agents are loaded")
		}
		if err != agent.ErrNoAgentAvailable {
			t.Errorf("Expected error %v, got: %v", agent.ErrNoAgentAvailable, err)
		}
	})

	t.Run("AvailableAgents", func(t *testing.T) {
		// Clear storage
		for _, a := range storage.List() {
			storage.Remove(a.ID)
		}

		agent1 := createTestAgent("agent1", agentModel.AgentStatusOnline, 2, 5)
		agent2 := createTestAgent("agent2", agentModel.AgentStatusOnline, 1, 5)
		agent3 := createTestAgent("agent3", agentModel.AgentStatusOnline, 3, 5)

		storage.Add(agent1)
		storage.Add(agent2)
		storage.Add(agent3)

		availableAgent, err := storage.GetAvailable()
		if err != nil {
			t.Errorf("Failed to get available agent: %v", err)
		}
		if availableAgent.ID != "agent2" {
			t.Errorf("Expected agent with lowest load (agent2), got: %s", availableAgent.ID)
		}
	})
}

func TestUpdateStatus(t *testing.T) {
	storage := agent.NewAgentStorage()
	testAgent := createTestAgent("agent1", agentModel.AgentStatusOffline, 0, 5)
	storage.Add(testAgent)

	t.Run("UpdateStatusOfExistingAgent", func(t *testing.T) {
		err := storage.UpdateStatus("agent1", agentModel.AgentStatusOnline, 2, 10)
		if err != nil {
			t.Errorf("Error when updating status: %v", err)
		}

		agent, _ := storage.GetByID("agent1")
		if agent.Status != agentModel.AgentStatusOnline {
			t.Errorf("Status not updated, expected %s, got: %s",
				agentModel.AgentStatusOnline, agent.Status)
		}
		if agent.CurrentLoad != 2 {
			t.Errorf("Current load not updated, expected 2, got: %d", agent.CurrentLoad)
		}
		if agent.MaxCapacity != 10 {
			t.Errorf("Max capacity not updated, expected 10, got: %d", agent.MaxCapacity)
		}

		// Check agent availability
		availableAgent, err := storage.GetAvailable()
		if err != nil {
			t.Errorf("Failed to get available agent after changing status to ONLINE: %v", err)
		}
		if availableAgent.ID != "agent1" {
			t.Errorf("Expected agent agent1, got: %s", availableAgent.ID)
		}
	})

	t.Run("UpdateStatusToOffline", func(t *testing.T) {
		err := storage.UpdateStatus("agent1", agentModel.AgentStatusOffline, 0, 10)
		if err != nil {
			t.Errorf("Error when updating status to offline: %v", err)
		}

		_, err = storage.GetAvailable()
		if err == nil {
			t.Error("Expected error when getting available agent after changing status to OFFLINE")
		}
	})

	t.Run("NonExistentAgent", func(t *testing.T) {
		err := storage.UpdateStatus("nonexistent", agentModel.AgentStatusOnline, 0, 5)
		if err == nil {
			t.Error("Expected error when updating status of non-existent agent")
		}
		if err != agent.ErrAgentNotFound {
			t.Errorf("Expected error %v, got: %v", agent.ErrAgentNotFound, err)
		}
	})

	t.Run("EmptyID", func(t *testing.T) {
		err := storage.UpdateStatus("", agentModel.AgentStatusOnline, 0, 5)
		if err == nil {
			t.Error("Expected error when updating status with empty ID")
		}
		if err != agent.ErrAgentNotFound {
			t.Errorf("Expected error %v, got: %v", agent.ErrAgentNotFound, err)
		}
	})
}

func TestUpdateStats(t *testing.T) {
	storage := agent.NewAgentStorage()
	testAgent := createTestAgent("agent1", agentModel.AgentStatusOnline, 0, 5)
	storage.Add(testAgent)

	t.Run("UpdateCompletedOperation", func(t *testing.T) {
		err := storage.UpdateStats("agent1", true, false)
		if err != nil {
			t.Errorf("Error when updating statistics: %v", err)
		}

		agent, _ := storage.GetByID("agent1")
		if agent.OperationsStats.Total != 1 {
			t.Errorf("Total operations count should be 1, got: %d", agent.OperationsStats.Total)
		}
		if agent.OperationsStats.Completed != 1 {
			t.Errorf("Completed operations count should be 1, got: %d", agent.OperationsStats.Completed)
		}
		if agent.OperationsStats.Failed != 0 {
			t.Errorf("Failed operations count should be 0, got: %d", agent.OperationsStats.Failed)
		}
	})

	t.Run("UpdateFailedOperation", func(t *testing.T) {
		err := storage.UpdateStats("agent1", false, true)
		if err != nil {
			t.Errorf("Error when updating statistics: %v", err)
		}

		agent, _ := storage.GetByID("agent1")
		if agent.OperationsStats.Total != 2 {
			t.Errorf("Total operations count should be 2, got: %d", agent.OperationsStats.Total)
		}
		if agent.OperationsStats.Completed != 1 {
			t.Errorf("Completed operations count should be 1, got: %d", agent.OperationsStats.Completed)
		}
		if agent.OperationsStats.Failed != 1 {
			t.Errorf("Failed operations count should be 1, got: %d", agent.OperationsStats.Failed)
		}
	})

	t.Run("NonExistentAgent", func(t *testing.T) {
		err := storage.UpdateStats("nonexistent", true, false)
		if err == nil {
			t.Error("Expected error when updating statistics of non-existent agent")
		}
		if err != agent.ErrAgentNotFound {
			t.Errorf("Expected error %v, got: %v", agent.ErrAgentNotFound, err)
		}
	})

	t.Run("EmptyID", func(t *testing.T) {
		err := storage.UpdateStats("", true, false)
		if err == nil {
			t.Error("Expected error when updating statistics with empty ID")
		}
		if err != agent.ErrAgentNotFound {
			t.Errorf("Expected error %v, got: %v", agent.ErrAgentNotFound, err)
		}
	})
}

func TestList(t *testing.T) {
	storage := agent.NewAgentStorage()

	t.Run("EmptyStorage", func(t *testing.T) {
		agents := storage.List()
		if len(agents) != 0 {
			t.Errorf("Expected empty list, got list with size %d", len(agents))
		}
	})

	t.Run("WithMultipleAgents", func(t *testing.T) {
		storage.Add(createTestAgent("agent1", agentModel.AgentStatusOnline, 0, 5))
		storage.Add(createTestAgent("agent2", agentModel.AgentStatusOffline, 0, 5))
		storage.Add(createTestAgent("agent3", agentModel.AgentStatusBusy, 0, 5))

		agents := storage.List()
		if len(agents) != 3 {
			t.Errorf("Expected 3 agents, got: %d", len(agents))
		}

		// Check that a copy of agents is returned
		agents[0].Status = agentModel.AgentStatusOffline
		retrievedAgent, _ := storage.GetByID("agent1")
		if retrievedAgent.Status != agentModel.AgentStatusOnline {
			t.Error("Modification of returned agent should not affect the storage")
		}
	})
}

func TestRemove(t *testing.T) {
	storage := agent.NewAgentStorage()
	storage.Add(createTestAgent("agent1", agentModel.AgentStatusOnline, 0, 5))
	storage.Add(createTestAgent("agent2", agentModel.AgentStatusOffline, 0, 5))

	t.Run("RemoveExistingAgent", func(t *testing.T) {
		err := storage.Remove("agent1")
		if err != nil {
			t.Errorf("Error when removing existing agent: %v", err)
		}

		_, err = storage.GetByID("agent1")
		if err == nil {
			t.Error("Agent was not removed from storage")
		}

		agents := storage.List()
		if len(agents) != 1 {
			t.Errorf("After removal, 1 agent should remain, found: %d", len(agents))
		}
	})

	t.Run("RemoveNonExistentAgent", func(t *testing.T) {
		err := storage.Remove("nonexistent")
		if err == nil {
			t.Error("Expected error when removing non-existent agent")
		}
		if err != agent.ErrAgentNotFound {
			t.Errorf("Expected error %v, got: %v", agent.ErrAgentNotFound, err)
		}
	})

	t.Run("RemoveWithEmptyID", func(t *testing.T) {
		err := storage.Remove("")
		if err == nil {
			t.Error("Expected error when removing agent with empty ID")
		}
		if err != agent.ErrAgentNotFound {
			t.Errorf("Expected error %v, got: %v", agent.ErrAgentNotFound, err)
		}
	})
}

// Helper function to create a test agent
func createTestAgent(id string, status agentModel.AgentStatus, currentLoad, maxCapacity int) *agentModel.Agent {
	return &agentModel.Agent{
		ID:          id,
		Status:      status,
		CurrentLoad: currentLoad,
		MaxCapacity: maxCapacity,
		OperationCosts: map[string]int{
			"addition":       1,
			"subtraction":    1,
			"multiplication": 2,
			"division":       2,
		},
		OperationsStats: agentModel.OperationsStats{
			Completed: 0,
			Failed:    0,
			Total:     0,
		},
		StartedAt:       time.Now(),
		LastOperationAt: time.Now(),
		UptimeSeconds:   0,
	}
}
