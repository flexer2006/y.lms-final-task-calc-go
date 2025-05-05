package orchestrator

import (
	"context"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/agent"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/orchestrator"
)

// AgentPool определяет интерфейс для управления пулом агентов-горутин.
type AgentPool interface {
	// Start запускает пул агентов.
	Start(ctx context.Context)

	// Stop останавливает все агенты.
	Stop()

	// GetAvailableAgent находит доступного агента для операции.
	GetAvailableAgent(operationType int) (*agent.Agent, error)

	// AssignOperation назначает операцию агенту.
	AssignOperation(agentID string, operation *orchestrator.Operation) error

	// GetAgentStatus получает статус конкретного агента.
	GetAgentStatus(agentID string) (*agent.Agent, error)

	// ListAgents возвращает список всех агентов.
	ListAgents() ([]*agent.Agent, error)
}
