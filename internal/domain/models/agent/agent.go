// Package agent содержит модели для работы с агентами.
package agent

import (
	"time"
)

// AgentStatus определяет текущий статус агента.
type AgentStatus string

const (
	// AgentStatusOnline - агент активен и обрабатывает задачи.
	AgentStatusOnline AgentStatus = "ONLINE"
	// AgentStatusOffline - агент недоступен.
	AgentStatusOffline AgentStatus = "OFFLINE"
	// AgentStatusBusy - агент занят и не может принимать новые задачи.
	AgentStatusBusy AgentStatus = "BUSY"
)

// Agent представляет горутину-агент для выполнения вычислений.
type Agent struct {
	ID              string          `json:"id"`
	Status          AgentStatus     `json:"status"`
	CurrentLoad     int             `json:"current_load"`
	MaxCapacity     int             `json:"max_capacity"`
	OperationCosts  map[string]int  `json:"operation_costs"`
	OperationsStats OperationsStats `json:"operations_stats"`
	StartedAt       time.Time       `json:"started_at"`
	LastOperationAt time.Time       `json:"last_operation_at"`
	UptimeSeconds   int64           `json:"uptime_seconds"`
}

// OperationsStats содержит статистику выполненных операций агентом.
type OperationsStats struct {
	Completed int64 `json:"completed"`
	Failed    int64 `json:"failed"`
	Total     int64 `json:"total"`
}
