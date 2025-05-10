// Package pool реализует пул агентов для обработки операций.
package pool

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/app/agent/worker"
	domainerrors "github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/errord"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/agent"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/orchestrator"
	agentRepo "github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/repository/agent"
	orchestratorRepo "github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/repository/orchestrator"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// AgentPool управляет пулом агентов-воркеров для выполнения вычислительных операций.
type AgentPool struct {
	workers        map[string]*worker.Worker            // карта активных воркеров
	storage        agentRepo.AgentStorage               // хранилище агентов
	operationTimes map[string]time.Duration             // время выполнения различных операций
	operationRepo  orchestratorRepo.OperationRepository // репозиторий операций
	capacity       int                                  // максимальное количество агентов
	mu             sync.RWMutex                         // мьютекс для безопасного доступа к полям
	ctx            context.Context                      // контекст для отмены операций
	cancel         context.CancelFunc                   // функция для отмены контекста
	running        bool                                 // флаг работы пула
}

// NewAgentPool создает новый пул агентов с заданными параметрами.
func NewAgentPool(storage agentRepo.AgentStorage, operationRepo orchestratorRepo.OperationRepository, operationTimes map[string]time.Duration, capacity int) (*AgentPool, error) {
	if storage == nil {
		return nil, domainerrors.ErrNilStorage
	}
	if operationRepo == nil {
		return nil, domainerrors.ErrNilOperationRepo
	}

	// Устанавливаем значения по умолчанию
	if capacity <= 0 {
		capacity = 4
	}
	if operationTimes == nil {
		operationTimes = map[string]time.Duration{
			"addition":       1 * time.Second,
			"subtraction":    1 * time.Second,
			"multiplication": 2 * time.Second,
			"division":       2 * time.Second,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	return &AgentPool{
		workers:        make(map[string]*worker.Worker),
		storage:        storage,
		operationRepo:  operationRepo,
		operationTimes: operationTimes,
		capacity:       capacity,
		ctx:            ctx,
		cancel:         cancel,
	}, nil
}

// Start запускает пул агентов с использованием переданного контекста.
func (p *AgentPool) Start(parentCtx context.Context) { //nolint:contextcheck
	if parentCtx == nil {
		parentCtx = p.ctx
	}
	log := logger.ContextLogger(parentCtx, nil)
	log.Info("Starting agent pool", zap.Int("capacity", p.capacity))

	// Проверяем, не запущен ли уже пул.
	p.mu.Lock()
	if p.running {
		p.mu.Unlock()
		log.Warn("Agent pool is already running")
		return
	}
	p.running = true
	p.mu.Unlock()

	// Создаем и запускаем воркеров.
	for i := range p.capacity {
		agentID := fmt.Sprintf("agent-%s-%d", uuid.New().String()[:8], i)
		w, err := worker.NewWorker(agentID, 3, p.operationTimes, p.operationRepo)
		if err != nil {
			log.Error("Failed to create worker", zap.String("agent_id", agentID), zap.Error(err))
			continue
		}

		p.mu.Lock()
		p.workers[agentID] = w
		p.mu.Unlock()

		w.Start(parentCtx)

		// Регистрируем агента в хранилище.
		agentStatus := w.GetStatus()
		if agentStatus == nil {
			log.Error("Failed to get agent status, using default values", zap.String("agent_id", agentID))
			agentStatus = &agent.Agent{
				ID:          agentID,
				Status:      agent.AgentStatusOnline,
				MaxCapacity: 3,
			}
		}
		p.storage.Add(agentStatus)
		log.Info("Started agent worker", zap.String("agent_id", agentID), zap.Int("capacity", agentStatus.MaxCapacity), zap.String("status", string(agentStatus.Status)))
	}

	// Запускаем фоновое обновление статусов.
	go p.updateAgentStatuses(parentCtx)
	log.Info("Agent pool started successfully", zap.Int("worker_count", p.capacity), zap.Int("operation_types", len(p.operationTimes)))
}

// Stop останавливает пул агентов и освобождает ресурсы.
func (p *AgentPool) Stop(ctx context.Context) {
	log := logger.ContextLogger(ctx, nil)
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		log.Debug("Agent pool is already stopped")
		return
	}

	log.Info("Stopping agent pool")
	stopCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if p.cancel != nil {
		p.cancel()
	}

	// Останавливаем всех воркеров и удаляем из хранилища.
	var stopErrors []error
	for id, w := range p.workers {
		if w != nil {
			w.Stop()
			if err := p.storage.Remove(id); err != nil {
				stopErrors = append(stopErrors, fmt.Errorf("failed to remove agent %s: %w", id, err))
				log.Warn("Failed to remove agent from storage", zap.String("agent_id", id), zap.Error(err))
			} else {
				log.Debug("Agent removed successfully", zap.String("agent_id", id))
			}
		}
	}

	p.workers = make(map[string]*worker.Worker)
	p.running = false

	// Логируем результат остановки.
	if len(stopErrors) > 0 {
		log.Warn("Agent pool stopped with errors", zap.Int("error_count", len(stopErrors)), zap.Error(fmt.Errorf("first error: %w", stopErrors[0])))
	} else {
		log.Info("Agent pool stopped successfully")
	}

	// Ждем завершения очистки ресурсов.
	select {
	case <-stopCtx.Done():
		if errors.Is(stopCtx.Err(), context.DeadlineExceeded) {
			log.Warn("Timeout waiting for all agent pool resources to clean up")
		}
	case <-time.After(100 * time.Millisecond):
	}
}

// GetAvailableAgent возвращает агента с наименьшей текущей нагрузкой для выполнения операции.
func (p *AgentPool) GetAvailableAgent(operationType int) (*agent.Agent, error) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if !p.running {
		return nil, domainerrors.ErrPoolNotRunning
	}

	if len(p.workers) == 0 {
		return nil, domainerrors.ErrNoAgentsAvailable
	}

	// Ищем воркера с наименьшей нагрузкой.
	var bestWorker *worker.Worker
	var lowestLoad = -1
	for _, w := range p.workers {
		if w == nil || !w.IsRunning() {
			continue
		}

		load := w.CurrentLoad()
		status := w.GetStatus()
		if status == nil {
			continue
		}

		if load >= status.MaxCapacity {
			continue
		}

		if lowestLoad == -1 || load < lowestLoad {
			bestWorker = w
			lowestLoad = load
		}
	}

	if bestWorker == nil {
		return nil, fmt.Errorf("%w: no workers available", domainerrors.ErrNoAgentsAvailable)
	}

	status := bestWorker.GetStatus()
	if status == nil {
		return nil, fmt.Errorf("%w: worker returned nil status", domainerrors.ErrNoAgentsAvailable)
	}

	return status, nil
}

// AssignOperation назначает операцию агенту с указанным ID.
func (p *AgentPool) AssignOperation(agentID string, operation *orchestrator.Operation) error {
	if operation == nil {
		return domainerrors.ErrNilOperation
	}

	if agentID == "" {
		return fmt.Errorf("%w: empty agent ID", domainerrors.ErrAgentNotFound)
	}

	// Генерируем ID если не задан.
	if operation.ID == uuid.Nil {
		operation.ID = uuid.New()
	}

	// Находим воркера по ID.
	p.mu.RLock()
	w, exists := p.workers[agentID]
	p.mu.RUnlock()

	if !exists || w == nil {
		return fmt.Errorf("%w: agent %s", domainerrors.ErrAgentNotFound, agentID)
	}

	if !w.IsRunning() {
		return fmt.Errorf("%w: agent %s is not running", domainerrors.ErrOperationAssignment, agentID)
	}

	log := logger.ContextLogger(context.Background(), nil).With(
		zap.String("operation_id", operation.ID.String()),
		zap.String("agent_id", agentID),
	)
	log.Info("Assigning operation to agent")

	// Выполняем операцию.
	_, err := w.PerformOperation(operation)
	if err != nil {
		log.Error("Failed to assign operation to agent", zap.Error(err))
		return fmt.Errorf("%w: %w", domainerrors.ErrOperationAssignment, err)
	}

	return nil
}

// GetAgentStatus возвращает текущий статус агента по его ID.
func (p *AgentPool) GetAgentStatus(agentID string) (*agent.Agent, error) {
	if agentID == "" {
		return nil, fmt.Errorf("%w: empty agent ID", domainerrors.ErrAgentNotFound)
	}

	p.mu.RLock()
	w, exists := p.workers[agentID]
	p.mu.RUnlock()

	if !exists || w == nil {
		return nil, fmt.Errorf("%w: agent %s", domainerrors.ErrAgentNotFound, agentID)
	}

	status := w.GetStatus()
	if status == nil {
		return nil, fmt.Errorf("%w: agent %s", domainerrors.ErrNilWorkerStatus, agentID)
	}

	return status, nil
}

// ListAgents возвращает список всех агентов.
func (p *AgentPool) ListAgents() ([]*agent.Agent, error) {
	agents := p.storage.List()
	if agents == nil {
		return make([]*agent.Agent, 0), nil
	}
	return agents, nil
}

// IsRunning возвращает состояние пула агентов (запущен или нет).
func (p *AgentPool) IsRunning() bool {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.running
}

// GetWorkerCount возвращает текущее количество воркеров в пуле.
func (p *AgentPool) GetWorkerCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.workers)
}

// GetCapacity возвращает максимальную емкость пула агентов.
func (p *AgentPool) GetCapacity() int {
	return p.capacity
}

// updateAgentStatuses запускает периодическое обновление статусов агентов в хранилище.
func (p *AgentPool) updateAgentStatuses(ctx context.Context) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	log := logger.ContextLogger(ctx, nil)
	log.Debug("Starting agent status update routine")

	for {
		select {
		case <-ctx.Done():
			log.Debug("Context done, stopping agent status updates")
			return
		case <-ticker.C:
			func() {
				p.mu.RLock()
				defer p.mu.RUnlock()

				if !p.running {
					return
				}

				// Обновляем статус каждого агента.
				for id, worker := range p.workers {
					if worker == nil {
						continue
					}

					status := worker.GetStatus()
					if status == nil {
						log.Warn("Worker returned nil status", zap.String("agent_id", id))
						continue
					}

					if err := p.storage.UpdateStatus(id, status.Status, status.CurrentLoad, status.MaxCapacity); err != nil {
						log.Warn("Failed to update agent status", zap.String("agent_id", id), zap.Error(err))
					}
				}
			}()
		}
	}
}
