package agent

import (
	"errors"
	"sync"

	agentModel "github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/agent"
	agentRepo "github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/repository/agent"
)

var (
	ErrAgentNotFound    = errors.New("agent not found")
	ErrNoAgentAvailable = errors.New("no agent available")
)

type MemoryAgentStorage struct {
	agents       map[string]*agentModel.Agent
	onlineAgents map[string]*agentModel.Agent
	mu           sync.RWMutex
}

var _ agentRepo.AgentStorage = (*MemoryAgentStorage)(nil)

func NewAgentStorage() *MemoryAgentStorage {
	return &MemoryAgentStorage{
		agents:       make(map[string]*agentModel.Agent),
		onlineAgents: make(map[string]*agentModel.Agent),
	}
}

func (s *MemoryAgentStorage) Add(agent *agentModel.Agent) {
	if agent == nil || agent.ID == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	agentCopy := *agent
	s.agents[agent.ID] = &agentCopy

	if agent.Status == agentModel.AgentStatusOnline {
		s.onlineAgents[agent.ID] = &agentCopy
	}
}

func (s *MemoryAgentStorage) GetByID(id string) (*agentModel.Agent, error) {
	if id == "" {
		return nil, ErrAgentNotFound
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	a, exists := s.agents[id]
	if !exists {
		return nil, ErrAgentNotFound
	}

	agentCopy := *a
	return &agentCopy, nil
}

func (s *MemoryAgentStorage) GetAvailable() (*agentModel.Agent, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var bestAgent *agentModel.Agent
	lowestLoad := -1

	for _, a := range s.onlineAgents {
		if a.CurrentLoad >= a.MaxCapacity {
			continue
		}

		if lowestLoad == -1 || a.CurrentLoad < lowestLoad {
			bestAgent = a
			lowestLoad = a.CurrentLoad
		}
	}

	if bestAgent == nil {
		return nil, ErrNoAgentAvailable
	}

	agentCopy := *bestAgent
	return &agentCopy, nil
}

func (s *MemoryAgentStorage) UpdateStatus(id string, status agentModel.AgentStatus, load int, capacity int) error {
	if id == "" {
		return ErrAgentNotFound
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	a, exists := s.agents[id]
	if !exists {
		return ErrAgentNotFound
	}

	wasOnline := a.Status == agentModel.AgentStatusOnline
	isOnline := status == agentModel.AgentStatusOnline

	a.Status = status
	a.CurrentLoad = load
	a.MaxCapacity = capacity

	if wasOnline != isOnline {
		if isOnline {
			s.onlineAgents[id] = a
		} else {
			delete(s.onlineAgents, id)
		}
	}

	return nil
}

func (s *MemoryAgentStorage) UpdateStats(id string, completed bool, failed bool) error {
	if id == "" {
		return ErrAgentNotFound
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	a, exists := s.agents[id]
	if !exists {
		return ErrAgentNotFound
	}

	a.OperationsStats.Total++

	if completed {
		a.OperationsStats.Completed++
	}
	if failed {
		a.OperationsStats.Failed++
	}

	return nil
}

func (s *MemoryAgentStorage) List() []*agentModel.Agent {
	s.mu.RLock()
	defer s.mu.RUnlock()

	agents := make([]*agentModel.Agent, 0, len(s.agents))

	for _, a := range s.agents {
		agentCopy := *a
		agents = append(agents, &agentCopy)
	}

	return agents
}

func (s *MemoryAgentStorage) Remove(id string) error {
	if id == "" {
		return ErrAgentNotFound
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	a, exists := s.agents[id]
	if !exists {
		return ErrAgentNotFound
	}

	delete(s.agents, id)

	if a.Status == agentModel.AgentStatusOnline {
		delete(s.onlineAgents, id)
	}

	return nil
}
