package processor

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"sync/atomic"
	"time"

	domainerrors "github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/errord"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/agent"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/orchestrator"
	orchapi "github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/api/orchestrator"
	orchrepo "github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/repository/orchestrator"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AgentConfig struct {
	AgentID             string
	ComputerPower       int
	TimeAddition        time.Duration
	TimeSubtraction     time.Duration
	TimeMultiplications time.Duration
	TimeDivisions       time.Duration
}

type OperationProcessor struct {
	operationRepo     orchrepo.OperationRepository
	calculationRepo   orchrepo.CalculationRepository
	calcUseCase       orchapi.UseCaseCalculation
	agentConfig       AgentConfig
	workerSem         chan struct{}
	agentID           string
	running           int32
	operationExecutor orchapi.OperationExecutor
	agentPool         orchapi.AgentPool
}

func NewProcessor(
	operationRepo orchrepo.OperationRepository,
	calculationRepo orchrepo.CalculationRepository,
	calcUseCase orchapi.UseCaseCalculation,
	agentConfig AgentConfig,
	operationExecutor orchapi.OperationExecutor,
	agentPool orchapi.AgentPool,
) *OperationProcessor {
	if operationRepo == nil {
		panic(fmt.Sprintf("%v: operation repository", domainerrors.ErrNilDependency))
	}
	if calculationRepo == nil {
		panic(fmt.Sprintf("%v: calculation repository", domainerrors.ErrNilDependency))
	}
	if calcUseCase == nil {
		panic(fmt.Sprintf("%v: calculation use case", domainerrors.ErrNilDependency))
	}
	if operationExecutor == nil {
		panic(fmt.Sprintf("%v: operation executor", domainerrors.ErrNilDependency))
	}
	if agentPool == nil {
		panic(fmt.Sprintf("%v: agent pool", domainerrors.ErrNilDependency))
	}

	if agentConfig.AgentID == "" {
		agentConfig.AgentID = fmt.Sprintf("processor-%s", uuid.New().String()[:8])
	}

	setDefaultIfZero(&agentConfig.ComputerPower, 5)
	setDefaultIfZero(&agentConfig.TimeAddition, 100*time.Millisecond)
	setDefaultIfZero(&agentConfig.TimeSubtraction, 150*time.Millisecond)
	setDefaultIfZero(&agentConfig.TimeMultiplications, 200*time.Millisecond)
	setDefaultIfZero(&agentConfig.TimeDivisions, 300*time.Millisecond)

	return &OperationProcessor{
		operationRepo:     operationRepo,
		calculationRepo:   calculationRepo,
		calcUseCase:       calcUseCase,
		agentConfig:       agentConfig,
		workerSem:         make(chan struct{}, agentConfig.ComputerPower),
		agentID:           agentConfig.AgentID,
		operationExecutor: operationExecutor,
		agentPool:         agentPool,
		running:           0,
	}
}

func setDefaultIfZero[T comparable](value *T, defaultValue T) {
	var zero T
	if *value == zero {
		*value = defaultValue
	}
}

func (p *OperationProcessor) Start(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("cannot start processor with nil context")
	}

	if p.operationRepo == nil || p.calculationRepo == nil || p.calcUseCase == nil ||
		p.operationExecutor == nil || p.agentPool == nil {
		return fmt.Errorf("cannot start processor: dependency is nil")
	}

	if !atomic.CompareAndSwapInt32(&p.running, 0, 1) {
		return nil
	}

	log := logger.ContextLogger(ctx, nil).With(zap.String("agent_id", p.agentID))
	log.Info("Starting operation processor", zap.Int("computer_power", p.agentConfig.ComputerPower))

	processorCtx, cancel := context.WithCancel(ctx)

	go func() {
		defer cancel()
		defer func() {
			if r := recover(); r != nil {
				log.Error("Recovered from panic in operation processor",
					zap.Any("panic", r),
					zap.String("stack", string(debug.Stack())))
				atomic.StoreInt32(&p.running, 0)
			}
		}()

		p.processOperations(processorCtx)
		atomic.StoreInt32(&p.running, 0)
	}()

	return nil
}

func (p *OperationProcessor) Stop() {
	atomic.StoreInt32(&p.running, 0)
}

func (p *OperationProcessor) IsRunning() bool {
	return atomic.LoadInt32(&p.running) == 1
}

func (p *OperationProcessor) processOperations(ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			log := logger.ContextLogger(ctx, nil).With(zap.String("agent_id", p.agentID))
			log.Error("Recovered from panic in operation processing loop",
				zap.Any("error", r),
				zap.String("stack", string(debug.Stack())))

			if atomic.LoadInt32(&p.running) == 1 {
				time.Sleep(1 * time.Second)
				go p.processOperations(ctx)
			}
		}
	}()

	log := logger.ContextLogger(ctx, nil).With(zap.String("agent_id", p.agentID))
	log.Debug("Starting operation processing loop")

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	// Создаем отдельный тикер для проверки зависших вычислений
	statusCheckTicker := time.NewTicker(5 * time.Second)
	defer statusCheckTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Info("Context cancelled, stopping processor")
			p.Stop()
			return
		case <-statusCheckTicker.C:
			// Периодически проверяем статусы незавершенных вычислений
			if p.IsRunning() {
				zapLogger := logger.GetZapLogger(log)
				if zapLogger != nil {
					go p.checkPendingCalculations(ctx, zapLogger)
				}
			}
		case <-ticker.C:
			if !p.IsRunning() {
				log.Info("Operation processor stopped")
				return
			}

			zapLogger := logger.GetZapLogger(log)
			if zapLogger != nil {
				batchCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
				p.processPendingBatch(batchCtx, zapLogger)
				cancel()
			} else {
				log.Warn("Failed to get zap logger for processing batch")
			}
		}
	}
}

func (p *OperationProcessor) processPendingBatch(ctx context.Context, log *zap.Logger) {
	if !p.IsRunning() {
		return
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	operations, err := p.operationRepo.GetPendingOperations(ctxWithTimeout, p.agentConfig.ComputerPower)
	if err != nil {
		log.Error("Failed to get pending operations", zap.Error(err))
		return
	}

	if len(operations) == 0 {
		return
	}

	log.Debug("Processing batch of operations", zap.Int("count", len(operations)))

	for _, op := range operations {
		select {
		case <-ctx.Done():
			log.Debug("Context cancelled during batch processing")
			return
		default:
			if op == nil {
				log.Warn("Skipping nil operation in pending batch")
				continue
			}

			operation := *op

			if operation.ID == uuid.Nil {
				operation.ID = uuid.New()
				log.Debug("Generated new ID for operation with nil ID")
			}

			p.processOperation(ctx, &operation, log)
		}
	}
}

func (p *OperationProcessor) processOperation(ctx context.Context, operation *orchestrator.Operation, log *zap.Logger) {
	if operation == nil {
		log.Warn("Attempted to process nil operation")
		return
	}

	if !p.IsRunning() {
		return
	}

	if operation.ID == uuid.Nil {
		operation.ID = uuid.New()
	}

	if operation.CalculationID == uuid.Nil {
		log.Error("Invalid operation with nil calculation ID",
			zap.String("operation_id", operation.ID.String()))
		return
	}

	select {
	case <-ctx.Done():
		return
	case p.workerSem <- struct{}{}:
	}

	go func() {
		defer func() { <-p.workerSem }()

		defer func() {
			if r := recover(); r != nil {
				opLog := log.With(zap.String("operation_id", operation.ID.String()))
				opLog.Error("Recovered from panic while processing operation",
					zap.Any("error", r),
					zap.String("stack", string(debug.Stack())))

				panicErr := fmt.Errorf("%w: %v", domainerrors.ErrPanic, r)
				p.handleOperationError(ctx, operation, panicErr, opLog)
			}
		}()

		opLog := log.With(
			zap.String("operation_id", operation.ID.String()),
			zap.String("calculation_id", operation.CalculationID.String()),
		)

		opCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		err := p.executeWithRetry(opCtx, operation, opLog)
		if err != nil {
			opLog.Error("Failed to execute operation after retries", zap.Error(err))
			p.handleOperationError(ctx, operation, err, opLog)
			return
		}

		statusCtx, statusCancel := context.WithTimeout(ctx, 5*time.Second)
		defer statusCancel()

		if err := p.calcUseCase.UpdateCalculationStatus(statusCtx, operation.CalculationID); err != nil {
			opLog.Error("Failed to update calculation status",
				zap.Error(err),
				zap.String("calculation_id", operation.CalculationID.String()))
		} else {
			opLog.Debug("Operation completed and calculation status updated successfully")
		}
	}()
}

func (p *OperationProcessor) executeWithRetry(ctx context.Context, operation *orchestrator.Operation, log *zap.Logger) error {
	if operation == nil {
		return domainerrors.ErrNilOperation
	}

	if log == nil {
		log = getDefaultLogger()
	}

	const maxRetries = 3
	var lastErr error

	opLogger := log.With(
		zap.String("operation_id", operation.ID.String()),
		zap.Int("operation_type", int(operation.OperationType)),
		zap.String("calculation_id", operation.CalculationID.String()),
	)

	for attempt := 0; attempt < maxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return fmt.Errorf("%w: %w", domainerrors.ErrContextDone, ctx.Err())
		default:
		}

		if attempt > 0 {
			backoffDuration := time.Duration(50*(1<<attempt)) * time.Millisecond
			opLogger.Debug("Retrying operation execution",
				zap.Int("attempt", attempt+1),
				zap.Duration("backoff", backoffDuration),
				zap.Error(lastErr))

			select {
			case <-ctx.Done():
				return fmt.Errorf("%w: %w", domainerrors.ErrContextDone, ctx.Err())
			case <-time.After(backoffDuration):
			}
		}

		execCtx, execCancel := context.WithTimeout(ctx, 5*time.Second)
		startTime := time.Now()

		err := func() error {
			defer execCancel()

			agent, agentErr := p.getAgentForOperation(execCtx, operation, opLogger)
			if agentErr != nil {
				return agentErr
			}

			if agent == nil {
				return domainerrors.ErrNoAgentAvailable
			}

			assignErr := p.assignOperationToAgent(execCtx, agent, operation, opLogger)
			if assignErr != nil {
				return assignErr
			}

			return nil
		}()

		if err == nil {
			opLogger.Debug("Operation successfully assigned to agent",
				zap.Duration("duration", time.Since(startTime)))
			return nil
		}

		if errors.Is(err, domainerrors.ErrContextDone) || errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return fmt.Errorf("context error during execution: %w", err)
		}

		lastErr = err
		opLogger.Warn("Failed attempt to execute operation",
			zap.Int("attempt", attempt+1),
			zap.Error(err))
	}

	return fmt.Errorf("operation execution failed after %d retries: %w", maxRetries, lastErr)
}

func (p *OperationProcessor) getAgentForOperation(ctx context.Context, operation *orchestrator.Operation, log *zap.Logger) (*agent.Agent, error) {
	if operation == nil {
		return nil, domainerrors.ErrNilOperation
	}

	if log == nil {
		log = getDefaultLogger()
	}

	if ctx.Err() != nil {
		return nil, fmt.Errorf("context error before getting agent: %w", ctx.Err())
	}

	if p.agentPool == nil {
		log.Error("Agent pool is nil", zap.String("operation_id", operation.ID.String()))
		return nil, fmt.Errorf("%w: agent pool is nil", domainerrors.ErrPoolNotRunning)
	}

	if !p.IsRunning() {
		log.Error("Operation processor is not running", zap.String("operation_id", operation.ID.String()))
		return nil, domainerrors.ErrPoolNotRunning
	}

	operationType := int(operation.OperationType)
	agentEntity, err := p.agentPool.GetAvailableAgent(operationType)
	if err != nil {
		log.Warn("Failed to get available agent",
			zap.String("operation_id", operation.ID.String()),
			zap.Int("operation_type", operationType),
			zap.Error(err))
		return nil, fmt.Errorf("failed to get available agent: %w", err)
	}

	if agentEntity == nil {
		log.Warn("No agent available for operation",
			zap.String("operation_id", operation.ID.String()),
			zap.Int("operation_type", operationType))
		return nil, domainerrors.ErrNoAgentAvailable
	}

	log.Debug("Found available agent",
		zap.String("agent_id", agentEntity.ID),
		zap.String("agent_status", string(agentEntity.Status)),
		zap.Int("current_load", agentEntity.CurrentLoad),
		zap.Int("max_capacity", agentEntity.MaxCapacity))

	return agentEntity, nil
}

func (p *OperationProcessor) assignOperationToAgent(ctx context.Context, agent *agent.Agent, operation *orchestrator.Operation, log *zap.Logger) error {
	if agent == nil || operation == nil {
		return domainerrors.ErrInvalidArgs
	}

	if log == nil {
		log = getDefaultLogger()
	}

	if ctx.Err() != nil {
		return fmt.Errorf("context error before assigning operation: %w", ctx.Err())
	}

	if agent.CurrentLoad >= agent.MaxCapacity {
		log.Warn("Agent is at capacity",
			zap.String("agent_id", agent.ID),
			zap.Int("current_load", agent.CurrentLoad),
			zap.Int("max_capacity", agent.MaxCapacity),
			zap.String("operation_id", operation.ID.String()))
		return fmt.Errorf("agent %s is at capacity (%d/%d)", agent.ID, agent.CurrentLoad, agent.MaxCapacity)
	}

	opLog := log.With(
		zap.String("operation_id", operation.ID.String()),
		zap.String("agent_id", agent.ID))

	updateCtx, updateCancel := context.WithTimeout(ctx, 3*time.Second)
	defer updateCancel()

	updateErr := p.operationRepo.UpdateStatus(
		updateCtx,
		operation.ID,
		orchestrator.OperationStatusInProgress,
		"",
		"",
	)

	if updateErr != nil {
		opLog.Warn("Failed to update operation status to IN_PROGRESS, continuing anyway",
			zap.Error(updateErr))
	}

	err := p.agentPool.AssignOperation(agent.ID, operation)
	if err != nil {
		opLog.Error("Failed to assign operation to agent",
			zap.Error(err))
		return fmt.Errorf("failed to assign operation to agent %s: %w", agent.ID, err)
	}

	opLog.Info("Operation assigned to agent successfully",
		zap.Int("agent_current_load", agent.CurrentLoad),
		zap.Int("agent_max_capacity", agent.MaxCapacity))

	return nil
}

func (p *OperationProcessor) handleOperationError(ctx context.Context, operation *orchestrator.Operation, execErr error, log *zap.Logger) {
	if operation == nil || operation.ID == uuid.Nil {
		if log != nil {
			log.Error("Cannot handle error for nil or invalid operation")
		}
		return
	}

	if ctx == nil {
		ctx = context.Background()
	}

	localLog := getLoggerOrDefault(log)
	localLog = localLog.With(
		zap.String("operation_id", operation.ID.String()),
		zap.String("calculation_id", operation.CalculationID.String()),
		zap.String("error", execErr.Error()),
	)

	errorMsg := "Failed to assign operation to agent: " + execErr.Error()

	updateCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if p.operationRepo == nil {
		localLog.Error("Cannot update operation status - operation repository is nil")
		return
	}

	err := p.operationRepo.UpdateStatus(
		updateCtx,
		operation.ID,
		orchestrator.OperationStatusError,
		"",
		errorMsg,
	)

	if err != nil {
		localLog.Error("Failed to update operation status", zap.Error(err))
	}

	if operation.CalculationID == uuid.Nil || p.calcUseCase == nil {
		localLog.Warn("Not updating calculation status - invalid calculation ID or nil calcUseCase")
		return
	}

	calcCtx, calcCancel := context.WithTimeout(ctx, 10*time.Second)
	defer calcCancel()

	safeUpdateStatus(calcCtx, p.calcUseCase, operation.CalculationID, localLog)
}

func safeUpdateStatus(ctx context.Context, calcUseCase orchapi.UseCaseCalculation, calculationID uuid.UUID, logger *zap.Logger) {
	logger = getLoggerOrDefault(logger)

	if calcUseCase == nil || calculationID == uuid.Nil {
		logger.Error("Cannot update status: invalid parameters")
		return
	}

	ctxToUse := ctx
	if ctxToUse == nil {
		ctxToUse = context.Background()
	}

	defer func() {
		if r := recover(); r != nil {
			logger.Error("Panic recovered in safeUpdateStatus",
				zap.Any("panic", r),
				zap.String("calculation_id", calculationID.String()),
				zap.String("stack", string(debug.Stack())))
		}
	}()

	var calcErr error
	func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Error("Panic during UpdateCalculationStatus",
					zap.Any("panic", r),
					zap.String("calculation_id", calculationID.String()))
				calcErr = fmt.Errorf("panic in UpdateCalculationStatus: %v", r)
			}
		}()

		if ctxToUse.Err() != nil {
			calcErr = fmt.Errorf("context error before update: %w", ctxToUse.Err())
			return
		}

		calcErr = calcUseCase.UpdateCalculationStatus(ctxToUse, calculationID)
	}()

	if calcErr != nil {
		logger.Error("Failed to update calculation status",
			zap.String("calculation_id", calculationID.String()),
			zap.Error(calcErr))
	} else {
		logger.Debug("Calculation status updated successfully",
			zap.String("calculation_id", calculationID.String()))
	}
}

func getDefaultLogger() *zap.Logger {
	logger := zap.L()
	if logger == nil {
		logger = zap.NewExample()
	}
	return logger
}

func getLoggerOrDefault(log *zap.Logger) *zap.Logger {
	if log == nil {
		return getDefaultLogger()
	}
	return log
}

// checkPendingCalculations проверяет и обновляет статусы зависших вычислений
func (p *OperationProcessor) checkPendingCalculations(ctx context.Context, log *zap.Logger) {
	if !p.IsRunning() || p.calculationRepo == nil || p.calcUseCase == nil {
		return
	}

	// Создаем контекст с таймаутом для операции проверки
	ctxWithTimeout, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	log.Debug("Checking for stuck calculations")

	// Получаем список операций, которые в процессе обработки
	pendingOperations, err := p.operationRepo.GetPendingOperations(ctxWithTimeout, 50)
	if err != nil {
		log.Error("Failed to fetch pending operations", zap.Error(err))
		return
	}

	if len(pendingOperations) == 0 {
		log.Debug("No pending operations found")
		return
	}

	// Создаем карту уникальных ID вычислений из операций
	calculationIDs := make(map[uuid.UUID]bool)
	for _, op := range pendingOperations {
		if op != nil && op.CalculationID != uuid.Nil {
			calculationIDs[op.CalculationID] = true
		}
	}

	if len(calculationIDs) == 0 {
		return
	}

	log.Info("Found calculations to check", zap.Int("count", len(calculationIDs)))

	// Обрабатываем каждое вычисление
	for calcID := range calculationIDs {
		updateCtx, updateCancel := context.WithTimeout(ctx, 5*time.Second)

		// Принудительно обновляем статус каждого расчета
		err := p.calcUseCase.UpdateCalculationStatus(updateCtx, calcID)
		if err != nil {
			log.Warn("Failed to update calculation status during check",
				zap.String("calculation_id", calcID.String()),
				zap.Error(err))
		} else {
			log.Debug("Successfully updated calculation status during check",
				zap.String("calculation_id", calcID.String()))
		}

		updateCancel()
	}
}

func (p *OperationProcessor) ExportGetAgentForOperation(ctx context.Context, operation *orchestrator.Operation) (*agent.Agent, error) {
	return p.getAgentForOperation(ctx, operation, zap.NewNop())
}

func (p *OperationProcessor) ExportAssignOperationToAgent(ctx context.Context, agent *agent.Agent, operation *orchestrator.Operation) error {
	return p.assignOperationToAgent(ctx, agent, operation, zap.NewNop())
}

func (p *OperationProcessor) ExportHandleOperationError(ctx context.Context, operation *orchestrator.Operation, execErr error) {
	p.handleOperationError(ctx, operation, execErr, zap.NewNop())
}

func (p *OperationProcessor) ExportCheckPendingCalculations(ctx context.Context) {
	p.checkPendingCalculations(ctx, zap.NewNop())
}
