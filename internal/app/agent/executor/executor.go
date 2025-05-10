// Package executor предоставляет функционал для выполнения операций через агенты.
// Включает механизмы распределения, повторных попыток и мониторинга выполнения операций.
package executor

import (
	"context"
	"fmt"
	"sync"
	"time"

	errors "github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/errord"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/agent"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/orchestrator"
	agentPool "github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/api/orchestrator"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// OperationExecutor управляет распределением операций между доступными агентами
// и предоставляет механизм повторных попыток при неудачах.
type OperationExecutor struct {
	pool           agentPool.AgentPool
	maxRetries     int
	retryDelay     time.Duration
	mu             sync.RWMutex
	assignedAgents map[uuid.UUID]string
}

// NewOperationExecutor создает новый экземпляр OperationExecutor с указанными параметрами.
// Возвращает nil, если pool равен nil.
func NewOperationExecutor(pool agentPool.AgentPool, maxRetries int, retryDelay time.Duration) *OperationExecutor {
	if pool == nil {
		return nil
	}

	// Ensure sane defaults
	if maxRetries < 0 {
		maxRetries = 0
	}
	if retryDelay <= 0 {
		retryDelay = 100 * time.Millisecond
	}

	return &OperationExecutor{
		pool:           pool,
		maxRetries:     maxRetries,
		retryDelay:     retryDelay,
		assignedAgents: make(map[uuid.UUID]string),
	}
}

// ExecuteOperation выполняет операцию, подбирая подходящий агент и переназначая
// в случае неудачи в пределах допустимого количества попыток.
func (e *OperationExecutor) ExecuteOperation(ctx context.Context, operation *orchestrator.Operation) error {
	if operation == nil {
		return errors.ErrNilOperation
	}

	if operation.ID == uuid.Nil {
		return fmt.Errorf("%w: operation must have a valid ID", errors.ErrInvalidOperationID)
	}

	log := logger.ContextLogger(ctx, nil).With(
		zap.String("operation_id", operation.ID.String()),
		zap.Int("operation_type", int(operation.OperationType)),
	)

	var lastError error

	for attempt := 0; attempt <= e.maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			e.removeAgentAssignment(operation.ID)
			return fmt.Errorf("%w: %w", errors.ErrContextCanceled, ctx.Err())
		default:
		}

		if attempt > 0 {
			log.Info("Retrying operation execution",
				zap.Int("attempt", attempt),
				zap.Error(lastError))

			select {
			case <-ctx.Done():
				e.removeAgentAssignment(operation.ID)
				return fmt.Errorf("%w: %w", errors.ErrContextCanceled, ctx.Err())
			case <-time.After(e.retryDelay):
			}
		}

		agent, err := e.pool.GetAvailableAgent(int(operation.OperationType))
		if err != nil {
			lastError = fmt.Errorf("%w: %w", errors.ErrNoAgentsAvailable, err)
			continue
		}

		if agent == nil {
			lastError = fmt.Errorf("%w: agent pool returned nil agent", errors.ErrNoAgentsAvailable)
			continue
		}

		log.Debug("Found available agent", zap.String("agent_id", agent.ID))

		e.recordAgentAssignment(operation.ID, agent.ID)

		err = e.pool.AssignOperation(agent.ID, operation)
		if err != nil {
			log.Warn("Failed to assign operation to agent",
				zap.String("agent_id", agent.ID),
				zap.Error(err))
			lastError = fmt.Errorf("%w: %w", errors.ErrOperationFailed, err)

			e.removeAgentAssignment(operation.ID)
			continue
		}

		log.Info("Operation assigned to agent successfully",
			zap.String("agent_id", agent.ID))
		return nil
	}

	log.Error("All attempts to execute operation failed",
		zap.Int("max_retries", e.maxRetries),
		zap.Error(lastError))
	return fmt.Errorf("%w: %w", errors.ErrMaxRetriesExceeded, lastError)
}

// recordAgentAssignment регистрирует связь между операцией и агентом.
func (e *OperationExecutor) recordAgentAssignment(operationID uuid.UUID, agentID string) {
	if operationID == uuid.Nil || agentID == "" {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	e.assignedAgents[operationID] = agentID
}

// removeAgentAssignment удаляет связь между операцией и агентом.
func (e *OperationExecutor) removeAgentAssignment(operationID uuid.UUID) {
	if operationID == uuid.Nil {
		return
	}

	e.mu.Lock()
	defer e.mu.Unlock()
	delete(e.assignedAgents, operationID)
}

// GetOperationAgent возвращает ID агента, назначенного на выполнение операции,
// и флаг, указывающий, было ли найдено назначение.
func (e *OperationExecutor) GetOperationAgent(operationID uuid.UUID) (string, bool) {
	if operationID == uuid.Nil {
		return "", false
	}

	e.mu.RLock()
	defer e.mu.RUnlock()
	agentID, found := e.assignedAgents[operationID]
	return agentID, found
}

// ReleaseOperation освобождает операцию, удаляя информацию о ее назначении на агента.
func (e *OperationExecutor) ReleaseOperation(operationID uuid.UUID) {
	if operationID == uuid.Nil {
		return
	}
	e.removeAgentAssignment(operationID)
}

// GetAgentsStatus возвращает список всех агентов и их текущие статусы.
func (e *OperationExecutor) GetAgentsStatus(ctx context.Context) ([]*agent.Agent, error) {
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	default:
		agents, err := e.pool.ListAgents()
		if err != nil {
			return nil, fmt.Errorf("failed to list agents: %w", err)
		}
		return agents, nil
	}
}

// GetAssignedOperationsCount возвращает количество операций, назначенных на агентов.
func (e *OperationExecutor) GetAssignedOperationsCount() int {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return len(e.assignedAgents)
}
