// Package agent содержит интерфейс для работы с хранилищем агентов.
package agent

import (
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/agent"
)

// AgentStorage определяет интерфейс для in-memory хранения агентов.
type AgentStorage interface {
	// Add добавляет агента в хранилище.
	Add(agent *agent.Agent)

	// GetByID находит агента по ID.
	GetByID(id string) (*agent.Agent, error)

	// GetAvailable находит доступного агента для операции.
	GetAvailable() (*agent.Agent, error)

	// UpdateStatus обновляет статус агента.
	UpdateStatus(id string, status agent.AgentStatus, load int, capacity int) error

	// UpdateStats обновляет статистику выполненных операций агента.
	UpdateStats(id string, completed bool, failed bool) error

	// List возвращает список всех агентов.
	List() []*agent.Agent

	// Remove удаляет агента из хранилища.
	Remove(id string) error
}
