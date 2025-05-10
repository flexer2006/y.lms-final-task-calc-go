// Package agent содержит интерфейс для работы с агентами.
package agent

import (
	"context"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/agent"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/orchestrator"
)

// AgentWorker определяет интерфейс для горутины-агента, выполняющей операции.
type AgentWorker interface {
	// Start запускает горутину агента.
	Start(ctx context.Context)

	// Stop останавливает горутину агента.
	Stop()

	// PerformOperation выполняет арифметическую операцию.
	PerformOperation(operation *orchestrator.Operation) (*orchestrator.Operation, error)

	// GetStatus возвращает текущий статус агента.
	GetStatus() *agent.Agent

	// UpdateStatus обновляет статус агента.
	UpdateStatus(status agent.AgentStatus, load int)
}
