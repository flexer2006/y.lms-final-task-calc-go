// Package orchestrator содержит интерфейс для выполнения операций.
package orchestrator

import (
	"context"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/agent"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/orchestrator"
	"github.com/google/uuid"
)

// OperationExecutor defines the interface for executing operations using the agent pool.
type OperationExecutor interface {
	// ExecuteOperation assigns an operation to an available agent.
	ExecuteOperation(ctx context.Context, operation *orchestrator.Operation) error

	// GetOperationAgent returns the agent assigned to a specific operation.
	GetOperationAgent(operationID uuid.UUID) (string, bool)

	// ReleaseOperation removes the operation-agent mapping after completion.
	ReleaseOperation(operationID uuid.UUID)

	// GetAgentsStatus returns the status of all agents in the pool.
	GetAgentsStatus(ctx context.Context) ([]*agent.Agent, error)
}
