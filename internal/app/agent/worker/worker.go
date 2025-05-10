// Package worker реализует функционал рабочего агента, выполняющего вычислительные операции.
package worker

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	domainerrors "github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/errord"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/agent"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/orchestrator"
	orchestratorRepo "github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/repository/orchestrator"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Worker представляет исполнителя операций с собственным состоянием и очередью заданий.
type Worker struct {
	agent           *agent.Agent                         // состояние агента
	operationTimes  map[string]time.Duration             // время выполнения различных типов операций
	operationsQueue chan *orchestrator.Operation         // очередь операций для обработки
	stopCh          chan struct{}                        // канал для сигнала остановки
	running         int32                                // флаг работы (используется атомарно)
	mu              sync.RWMutex                         // мьютекс для безопасного доступа к полям
	operationRepo   orchestratorRepo.OperationRepository // репозиторий для сохранения операций
}

// NewWorker создает нового воркера с указанными параметрами.
// Возвращает ошибку, если operationRepo равен nil.
func NewWorker(id string, capacity int, operationTimes map[string]time.Duration, operationRepo orchestratorRepo.OperationRepository) (*Worker, error) {
	if operationRepo == nil {
		return nil, fmt.Errorf("operation repository cannot be nil: %w", domainerrors.ErrNilOperationRepo)
	}

	if capacity <= 0 {
		capacity = 3
	}

	if operationTimes == nil {
		operationTimes = map[string]time.Duration{
			"addition":       time.Second,
			"subtraction":    time.Second,
			"multiplication": 2 * time.Second,
			"division":       2 * time.Second,
		}
	}

	queueSize := capacity * 2

	return &Worker{
		agent: &agent.Agent{
			ID:          id,
			Status:      agent.AgentStatusOffline,
			CurrentLoad: 0,
			MaxCapacity: capacity,
			OperationCosts: map[string]int{
				"addition":       1,
				"subtraction":    1,
				"multiplication": 1,
				"division":       1,
			},
			OperationsStats: agent.OperationsStats{
				Completed: 0,
				Failed:    0,
				Total:     0,
			},
			StartedAt:       time.Now(),
			LastOperationAt: time.Now(),
			UptimeSeconds:   0,
		},
		operationTimes:  operationTimes,
		operationsQueue: make(chan *orchestrator.Operation, queueSize),
		stopCh:          make(chan struct{}),
		operationRepo:   operationRepo,
	}, nil
}

// Start запускает обработку операций в фоновом режиме.
// Переводит агента в статус Online.
func (w *Worker) Start(ctx context.Context) {
	if w == nil || ctx == nil {
		return
	}

	// Атомарно устанавливаем флаг работы (предотвращает двойной запуск)
	if !atomic.CompareAndSwapInt32(&w.running, 0, 1) {
		return
	}

	w.mu.Lock()
	if w.agent != nil {
		w.agent.Status = agent.AgentStatusOnline
	}
	w.mu.Unlock()

	var log *zap.Logger
	ctxLogger := logger.ContextLogger(ctx, nil)
	if ctxLogger != nil {
		if w.agent != nil {
			loggerWithID := ctxLogger.With(zap.String("agent_id", w.agent.ID))
			log = logger.GetZapLogger(loggerWithID)
		} else {
			log = ctxLogger.RawLogger()
		}
		log.Info("Starting agent worker")
	}

	// Запускаем обработку в фоновой горутине
	go w.processOperations(ctx)
}

// Stop останавливает обработку операций и переводит агента в статус Offline.
func (w *Worker) Stop() {
	if w == nil {
		return
	}

	// Атомарно сбрасываем флаг работы (предотвращает двойную остановку)
	if !atomic.CompareAndSwapInt32(&w.running, 1, 0) {
		return
	}

	close(w.stopCh)

	w.mu.Lock()
	if w.agent != nil {
		w.agent.Status = agent.AgentStatusOffline
	}
	w.mu.Unlock()
}

// PerformOperation добавляет операцию в очередь на выполнение.
// Возвращает ошибку, если агент недоступен или перегружен.
func (w *Worker) PerformOperation(operation *orchestrator.Operation) (*orchestrator.Operation, error) {
	if w == nil {
		return nil, fmt.Errorf("worker is nil")
	}

	if operation == nil {
		return nil, domainerrors.ErrNilOperation
	}

	if operation.ID == uuid.Nil {
		operation.ID = uuid.New()
	}

	// Проверяем, запущен ли агент
	if atomic.LoadInt32(&w.running) != 1 {
		agentID := "unknown"
		if w.agent != nil {
			agentID = w.agent.ID
		}
		return nil, fmt.Errorf("%w: agent %s", domainerrors.ErrAgentNotRunning, agentID)
	}

	// Проверяем статус и нагрузку агента
	w.mu.RLock()
	agentID := "unknown"
	isOnline := false
	atCapacity := true
	if w.agent != nil {
		agentID = w.agent.ID
		isOnline = w.agent.Status == agent.AgentStatusOnline
		atCapacity = w.agent.CurrentLoad >= w.agent.MaxCapacity
	}
	w.mu.RUnlock()

	if !isOnline {
		return nil, fmt.Errorf("%w: agent %s", domainerrors.ErrAgentNotRunning, agentID)
	}

	if atCapacity {
		return nil, fmt.Errorf("%w: agent %s", domainerrors.ErrAgentAtCapacity, agentID)
	}

	// Пытаемся добавить операцию в очередь с таймаутом
	select {
	case w.operationsQueue <- operation:
		w.mu.Lock()
		if w.agent != nil {
			w.agent.CurrentLoad++
		}

		operationID := operation.ID.String()
		ctx := context.Background()
		ctxLogger := logger.ContextLogger(ctx, nil)
		if ctxLogger != nil && w.agent != nil {
			ctxLogger.Debug("Agent capacity updated",
				zap.String("agent_id", w.agent.ID),
				zap.String("operation", operationID),
				zap.Int("current_load", w.agent.CurrentLoad),
				zap.Int("max_capacity", w.agent.MaxCapacity))
		}
		w.mu.Unlock()
		return operation, nil
	case <-time.After(100 * time.Millisecond):
		return nil, fmt.Errorf("%w: agent %s", domainerrors.ErrQueueFull, agentID)
	}
}

// GetStatus возвращает копию текущего состояния агента с актуальными данными.
func (w *Worker) GetStatus() *agent.Agent {
	if w == nil {
		return nil
	}

	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.agent == nil {
		return nil
	}

	// Создаем копию для потокобезопасности
	agentCopy := *w.agent

	// Обновляем динамические поля
	agentCopy.UptimeSeconds = int64(time.Since(w.agent.StartedAt).Seconds())

	// Определяем актуальный статус на основе текущей нагрузки
	if atomic.LoadInt32(&w.running) == 1 {
		if agentCopy.CurrentLoad >= agentCopy.MaxCapacity {
			agentCopy.Status = agent.AgentStatusBusy
		} else {
			agentCopy.Status = agent.AgentStatusOnline
		}
	} else {
		agentCopy.Status = agent.AgentStatusOffline
	}

	return &agentCopy
}

// UpdateStatus обновляет статус и нагрузку агента.
// Отрицательная нагрузка будет скорректирована до нуля.
func (w *Worker) UpdateStatus(status agent.AgentStatus, load int) {
	if w == nil {
		return
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	if w.agent == nil {
		return
	}

	w.agent.Status = status

	if load < 0 {
		load = 0
	}
	w.agent.CurrentLoad = load

	w.agent.LastOperationAt = time.Now()
}

// IsRunning возвращает true, если воркер запущен и обрабатывает операции.
func (w *Worker) IsRunning() bool {
	if w == nil {
		return false
	}
	return atomic.LoadInt32(&w.running) == 1
}

// CurrentLoad возвращает текущую нагрузку агента (количество обрабатываемых операций).
func (w *Worker) CurrentLoad() int {
	if w == nil {
		return 0
	}

	w.mu.RLock()
	defer w.mu.RUnlock()

	if w.agent == nil {
		return 0
	}

	return w.agent.CurrentLoad
}

// processOperations - основной цикл обработки операций из очереди.
// Выполняется в отдельной горутине до получения сигнала остановки.
func (w *Worker) processOperations(ctx context.Context) {
	if w == nil || ctx == nil {
		return
	}

	var log *zap.Logger
	agentID := "unknown"

	ctxLogger := logger.ContextLogger(ctx, nil)
	if ctxLogger != nil {
		if w.agent != nil {
			agentID = w.agent.ID
			loggerWithID := ctxLogger.With(zap.String("agent_id", agentID))
			log = logger.GetZapLogger(loggerWithID)
		} else {
			log = ctxLogger.RawLogger()
		}
		log.Debug("Starting operation processing loop")
	}

	for {
		select {
		case <-ctx.Done():
			if log != nil {
				log.Debug("Context canceled, stopping operation processing")
			}
			return
		case <-w.stopCh:
			if log != nil {
				log.Debug("Stop signal received, stopping operation processing")
			}
			return
		case op := <-w.operationsQueue:
			if op == nil {
				if log != nil {
					log.Warn("Received nil operation, skipping")
				}
				continue
			}

			if op.ID == uuid.Nil {
				op.ID = uuid.New()
			}

			opID := op.ID.String()

			if log != nil {
				log.Debug("Processing operation",
					zap.String("operation_id", opID),
					zap.Int("operation_type", int(op.OperationType)))
			}

			var result string
			var err error

			// Выполняем операцию
			result, err = w.executeOperation(ctx, op)

			// Определяем статус операции после выполнения
			opStatus := orchestrator.OperationStatusCompleted
			errMsg := ""
			if err != nil {
				opStatus = orchestrator.OperationStatusError
				errMsg = err.Error()
			}

			// Обновляем статус операции в репозитории
			if w.operationRepo != nil {
				if updateErr := w.operationRepo.UpdateStatus(ctx, op.ID, opStatus, result, errMsg); updateErr != nil && log != nil {
					log.Error("Failed to update operation status",
						zap.String("operation_id", opID),
						zap.Error(updateErr))
				}
			}

			// Обновляем статистику агента
			w.mu.Lock()
			if w.agent != nil {
				w.agent.CurrentLoad--
				if w.agent.CurrentLoad < 0 {
					w.agent.CurrentLoad = 0
					if log != nil {
						log.Warn("Corrected negative agent load", zap.String("agent_id", agentID))
					}
				}

				w.agent.LastOperationAt = time.Now()
				w.agent.OperationsStats.Total++

				if err != nil {
					w.agent.OperationsStats.Failed++
				} else {
					w.agent.OperationsStats.Completed++
				}
			}
			w.mu.Unlock()

			// Логируем результат выполнения
			if err != nil && log != nil {
				log.Error("Failed to execute operation",
					zap.String("operation_id", opID),
					zap.Error(err))
			} else if log != nil {
				log.Debug("Operation executed successfully",
					zap.String("operation_id", opID),
					zap.String("result", result))
			}
		}
	}
}

// resolveReference разрешает ссылки на результаты других операций.
// Поддерживает формат "ref:UUID" для получения результата предыдущей операции.
func (w *Worker) resolveReference(ctx context.Context, refStr string, log *zap.Logger) (string, error) {
	if w == nil || ctx == nil {
		return "", fmt.Errorf("worker or context is nil")
	}

	refID := strings.TrimPrefix(refStr, "ref:")

	if w.operationRepo == nil {
		return "", domainerrors.ErrRepoNotInitialized
	}

	// Парсим UUID из ссылки
	uid, err := uuid.Parse(refID)
	if err != nil {
		if log != nil {
			log.Error("Failed to parse reference ID",
				zap.String("ref_id", refID), zap.Error(err))
		}
		return "", fmt.Errorf("%w: %s", domainerrors.ErrInvalidReferenceID, refID)
	}

	// Ищем связанную операцию в репозитории
	refOp, err := w.operationRepo.FindByID(ctx, uid)
	if err != nil {
		if log != nil {
			log.Error("Failed to lookup referenced operation",
				zap.String("ref_id", refID), zap.Error(err))
		}
		return "", fmt.Errorf("reference lookup failed: %w", err)
	}

	if refOp == nil {
		return "", fmt.Errorf("%w: %s", domainerrors.ErrReferenceNotFound, refID)
	}

	// Проверяем, что связанная операция завершена успешно
	if refOp.Status != orchestrator.OperationStatusCompleted {
		return "", fmt.Errorf("%w: %s", domainerrors.ErrRefNotCompleted, refID)
	}

	if log != nil {
		log.Debug("Resolved operation reference",
			zap.String("ref_id", refID),
			zap.String("result", refOp.Result))
	}

	return refOp.Result, nil
}

// executeOperation выполняет конкретную математическую операцию.
// Поддерживает базовые операции: сложение, вычитание, умножение и деление.
func (w *Worker) executeOperation(ctx context.Context, op *orchestrator.Operation) (string, error) {
	if w == nil || ctx == nil {
		return "", fmt.Errorf("worker or context is nil")
	}

	if op == nil {
		return "", domainerrors.ErrNilOperation
	}

	if op.ID == uuid.Nil {
		op.ID = uuid.New()
	}

	opID := op.ID.String()

	// Настраиваем логгер
	var zapLog *zap.Logger
	agentID := "unknown"
	if w.agent != nil {
		agentID = w.agent.ID
	}

	ctxLogger := logger.ContextLogger(ctx, nil)
	if ctxLogger != nil {
		loggerWithFields := ctxLogger.With(
			zap.String("agent_id", agentID),
			zap.String("operation_id", opID),
		)
		zapLog = logger.GetZapLogger(loggerWithFields)
	}

	operand1Str := op.Operand1
	operand2Str := op.Operand2

	// Разрешаем ссылки на результаты других операций
	var err error
	if strings.HasPrefix(operand1Str, "ref:") {
		operand1Str, err = w.resolveReference(ctx, operand1Str, zapLog)
		if err != nil {
			return "", err
		}
	}

	if strings.HasPrefix(operand2Str, "ref:") {
		operand2Str, err = w.resolveReference(ctx, operand2Str, zapLog)
		if err != nil {
			return "", err
		}
	}

	// Преобразуем строковые операнды в числа
	operand1, err := strconv.ParseFloat(operand1Str, 64)
	if err != nil {
		return "", fmt.Errorf("%w: %s", domainerrors.ErrInvalidOperand, operand1Str)
	}

	operand2, err := strconv.ParseFloat(operand2Str, 64)
	if err != nil {
		return "", fmt.Errorf("%w: %s", domainerrors.ErrInvalidOperand, operand2Str)
	}

	var operationTime time.Duration
	var result float64

	// Выполняем математическую операцию в зависимости от типа
	switch op.OperationType {
	case orchestrator.OperationTypeAddition:
		if zapLog != nil {
			zapLog.Debug("Performing addition",
				zap.Float64("operand1", operand1),
				zap.Float64("operand2", operand2))
		}
		operationTime = w.getOperationTime("addition")
		result = operand1 + operand2
	case orchestrator.OperationTypeSubtraction:
		if zapLog != nil {
			zapLog.Debug("Performing subtraction",
				zap.Float64("operand1", operand1),
				zap.Float64("operand2", operand2))
		}
		operationTime = w.getOperationTime("subtraction")
		result = operand1 - operand2
	case orchestrator.OperationTypeMultiplication:
		if zapLog != nil {
			zapLog.Debug("Performing multiplication",
				zap.Float64("operand1", operand1),
				zap.Float64("operand2", operand2))
		}
		operationTime = w.getOperationTime("multiplication")
		result = operand1 * operand2
	case orchestrator.OperationTypeDivision:
		if zapLog != nil {
			zapLog.Debug("Performing division",
				zap.Float64("operand1", operand1),
				zap.Float64("operand2", operand2))
		}
		operationTime = w.getOperationTime("division")

		if operand2 == 0 {
			return "", domainerrors.ErrDivisionByZero
		}

		result = operand1 / operand2
	default:
		return "", fmt.Errorf("%w: %d", domainerrors.ErrUnsupportedOp, op.OperationType)
	}

	// Эмулируем время выполнения операции
	select {
	case <-ctx.Done():
		return "", fmt.Errorf("%w: %w", domainerrors.ErrContextCanceled, ctx.Err())
	case <-time.After(operationTime):
	}

	return formatNumericResult(result), nil
}

// getOperationTime возвращает время выполнения операции указанного типа.
// Для неизвестных типов операций возвращает 1 секунду.
func (w *Worker) getOperationTime(operation string) time.Duration {
	if w == nil || w.operationTimes == nil {
		return time.Second
	}

	if duration, ok := w.operationTimes[operation]; ok {
		return duration
	}

	return time.Second
}

// formatNumericResult форматирует числовой результат в удобочитаемую строку.
// Если результат целочисленный, убирает десятичную часть.
func formatNumericResult(result float64) string {
	if result == math.Trunc(result) {
		return fmt.Sprintf("%.0f", result)
	}

	return strconv.FormatFloat(result, 'f', -1, 64)
}
