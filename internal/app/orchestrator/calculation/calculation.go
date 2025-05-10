// Package calculation реализует логику вычисления математических выражений
package calculation

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	domainerrors "github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/errord"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/orchestrator"
	orchapi "github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/api/orchestrator"
	orchrepo "github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/repository/orchestrator"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/service/parser"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Константы для настройки таймаутов и лимитов вычислений
const (
	defaultTimeout    = 10 * time.Second
	validationTimeout = 5 * time.Second
	parsingTimeout    = 30 * time.Second
	statusTimeout     = 5 * time.Second
	maxRetries        = 3
	maxErrorLength    = 500
	maxOperations     = 500
)

// UseCaseImpl реализует логику вычисления математических выражений
type UseCaseImpl struct {
	calculationRepo orchrepo.CalculationRepository
	operationRepo   orchrepo.OperationRepository
	parser          parser.ExpressionParser
}

// Проверка соответствия интерфейсу
var _ orchapi.UseCaseCalculation = (*UseCaseImpl)(nil)

// NewUseCase создает новый экземпляр сервиса вычислений
func NewUseCase(
	calculationRepo orchrepo.CalculationRepository,
	operationRepo orchrepo.OperationRepository,
	parser parser.ExpressionParser,
) *UseCaseImpl {
	return &UseCaseImpl{
		calculationRepo: calculationRepo,
		operationRepo:   operationRepo,
		parser:          parser,
	}
}

// CalculateExpression вычисляет математическое выражение
// Создает запись вычисления, разбирает выражение на операции и запускает их выполнение
func (uc *UseCaseImpl) CalculateExpression(ctx context.Context, userID uuid.UUID, expression string) (*orchestrator.Calculation, error) {
	log := logger.ContextLogger(ctx, nil).With(
		zap.String("op", "CalculationUseCase.CalculateExpression"),
		zap.String("user_id", userID.String()),
		zap.String("expression", expression),
	)

	// Проверка корректности входных данных
	if userID == uuid.Nil {
		return nil, domainerrors.ErrInvalidUserID
	}

	if expression == "" {
		return nil, fmt.Errorf("%w: expression cannot be empty", domainerrors.ErrInvalidExpression)
	}

	// Валидация выражения
	validationCtx, cancel := context.WithTimeout(ctx, validationTimeout)
	defer cancel()

	if err := uc.parser.Validate(validationCtx, expression); err != nil {
		return nil, fmt.Errorf("%w: %v", domainerrors.ErrInvalidExpression, err)
	}

	// Создание записи вычисления
	calc := &orchestrator.Calculation{
		ID:         uuid.New(),
		UserID:     userID,
		Expression: expression,
		Status:     orchestrator.CalculationStatusPending,
	}

	createCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	savedCalc, err := uc.calculationRepo.Create(createCtx, calc)
	if err != nil {
		log.Error("Failed to create calculation", zap.Error(err))
		return nil, fmt.Errorf("%w: %v", domainerrors.ErrInternalError, err)
	}

	// Разбор выражения на операции
	parseCtx, cancel := context.WithTimeout(ctx, parsingTimeout)
	defer cancel()

	zapLogger := logger.GetZapLogger(log)
	if zapLogger == nil {
		zapLogger = zap.L()
	}

	_, err = uc.parseExpression(parseCtx, zapLogger, savedCalc.ID, expression)
	if err != nil {
		// Возвращаем результат с ошибкой, если она есть
		updatedCalc, findErr := uc.calculationRepo.FindByID(ctx, savedCalc.ID)
		if findErr == nil && updatedCalc != nil {
			return updatedCalc, nil
		}
		return savedCalc, nil
	}

	// Обновляем статус на "в процессе"
	updateCtx, cancel := context.WithTimeout(ctx, statusTimeout)
	defer cancel()

	if err = uc.calculationRepo.UpdateStatus(updateCtx, savedCalc.ID, orchestrator.CalculationStatusInProgress, "", ""); err != nil {
		log.Error("Failed to update calculation status", zap.Error(err))
	}

	// Получаем обновленный расчет
	result, err := uc.calculationRepo.FindByID(ctx, savedCalc.ID)
	if err != nil {
		return savedCalc, nil
	}

	return result, nil
}

// parseExpression разбирает выражение на операции и сохраняет их в БД
func (uc *UseCaseImpl) parseExpression(ctx context.Context, log *zap.Logger, calculationID uuid.UUID, expression string) ([]*orchestrator.Operation, error) {
	if log == nil {
		log = zap.L()
	}

	// Парсинг выражения в операции
	operations, err := uc.parser.Parse(ctx, expression)
	if err != nil {
		updateErr := uc.calculationRepo.UpdateStatus(ctx, calculationID, orchestrator.CalculationStatusError, "", err.Error())
		if updateErr != nil {
			log.Error("Failed to update calculation status", zap.Error(updateErr))
		}
		return nil, fmt.Errorf("%w: %v", domainerrors.ErrInvalidExpression, err)
	}

	if operations == nil {
		operations = []*orchestrator.Operation{}
	}

	// Проверка на превышение лимита операций
	if len(operations) > maxOperations {
		errMsg := "Expression too complex, too many operations"
		updateErr := uc.calculationRepo.UpdateStatus(ctx, calculationID, orchestrator.CalculationStatusError, "", errMsg)
		if updateErr != nil {
			log.Error("Failed to update calculation status", zap.Error(updateErr))
		}
		return nil, domainerrors.ErrTooManyOps
	}

	// Привязка операций к расчету
	uc.parser.SetCalculationID(operations, calculationID)

	// Сохранение операций
	if err = uc.operationRepo.CreateBatch(ctx, operations); err != nil {
		errMsg := "Failed to create operations"
		updateErr := uc.calculationRepo.UpdateStatus(ctx, calculationID, orchestrator.CalculationStatusError, "", errMsg)
		if updateErr != nil {
			log.Error("Failed to update calculation status", zap.Error(updateErr))
		}
		return nil, fmt.Errorf("%w: %v", domainerrors.ErrOperationCreationFailed, err)
	}

	return operations, nil
}

// GetCalculation получает информацию о вычислении с указанным ID
// Проверяет права доступа и обогащает результат данными об операциях
func (uc *UseCaseImpl) GetCalculation(ctx context.Context, calculationID uuid.UUID, userID uuid.UUID) (*orchestrator.Calculation, error) {
	log := logger.ContextLogger(ctx, nil).With(
		zap.String("op", "CalculationUseCase.GetCalculation"),
		zap.String("calculation_id", calculationID.String()),
		zap.String("user_id", userID.String()),
	)

	// Получение вычисления из репозитория
	calc, err := uc.calculationRepo.FindByID(ctx, calculationID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domainerrors.ErrInternalError, err)
	}

	if calc == nil {
		return nil, domainerrors.ErrCalculationNotFound
	}

	// Проверка прав доступа
	if calc.UserID != userID {
		return nil, domainerrors.ErrUnauthorizedAccess
	}

	// Обогащение данными об операциях
	zapLogger := logger.GetZapLogger(log)
	calc, err = uc.enrichCalculationWithOperations(ctx, zapLogger, calc)
	if err != nil {
		log.Warn("Unable to fetch operations", zap.Error(err))
	}

	return calc, nil
}

// enrichCalculationWithOperations добавляет данные об операциях в объект вычисления
func (uc *UseCaseImpl) enrichCalculationWithOperations(ctx context.Context, log *zap.Logger, calc *orchestrator.Calculation) (*orchestrator.Calculation, error) {
	operations, err := uc.operationRepo.FindByCalculationID(ctx, calc.ID)
	if err != nil {
		if log != nil {
			log.Error("Failed to fetch operations", zap.String("calculation_id", calc.ID.String()), zap.Error(err))
		}
		return calc, fmt.Errorf("failed to fetch operations: %w", err)
	}

	if len(operations) > 0 {
		calc.Operations = make([]orchestrator.Operation, len(operations))
		for i, op := range operations {
			calc.Operations[i] = *op
		}
	}

	return calc, nil
}

// ListCalculations возвращает список всех вычислений пользователя
func (uc *UseCaseImpl) ListCalculations(ctx context.Context, userID uuid.UUID) ([]*orchestrator.Calculation, error) {
	log := logger.ContextLogger(ctx, nil).With(
		zap.String("op", "CalculationUseCase.ListCalculations"),
		zap.String("user_id", userID.String()),
	)

	if userID == uuid.Nil {
		return nil, domainerrors.ErrInvalidUserID
	}

	calculations, err := uc.calculationRepo.FindByUserID(ctx, userID)
	if err != nil {
		log.Error("Failed to fetch user calculations", zap.Error(err))
		return nil, fmt.Errorf("%w: %v", domainerrors.ErrInternalError, err)
	}

	return calculations, nil
}

// ProcessPendingOperations заглушка для обработки ожидающих операций
func (uc *UseCaseImpl) ProcessPendingOperations(ctx context.Context) error {
	return nil
}

// UpdateCalculationStatus обновляет статус вычисления на основе статусов его операций
func (uc *UseCaseImpl) UpdateCalculationStatus(ctx context.Context, calculationID uuid.UUID) error {
	if calculationID == uuid.Nil {
		return fmt.Errorf("%w: %s", domainerrors.ErrSpecificCalcNotFound, calculationID)
	}

	timeoutCtx, cancel := context.WithTimeout(ctx, defaultTimeout)
	defer cancel()

	log := logger.ContextLogger(timeoutCtx, nil).With(
		zap.String("op", "CalculationUseCase.UpdateCalculationStatus"),
		zap.String("calculation_id", calculationID.String()),
	)

	// Проверка инициализации компонентов
	if uc == nil {
		return domainerrors.ErrUseCaseNil
	}

	if uc.calculationRepo == nil {
		return domainerrors.ErrCalcRepoNil
	}

	if uc.operationRepo == nil {
		return domainerrors.ErrOpRepoNil
	}

	// Получение вычисления с повторными попытками
	_, err := uc.getCalculationWithRetry(timeoutCtx, calculationID, log)
	if err != nil {
		return err
	}

	// Получение операций с повторными попытками
	operations, err := uc.getOperationsWithRetry(timeoutCtx, calculationID, log)
	if err != nil {
		return fmt.Errorf("failed to fetch operations: %w", err)
	}

	// Проверка наличия операций
	if len(operations) == 0 {
		updateErr := uc.calculationRepo.UpdateStatus(
			timeoutCtx,
			calculationID,
			orchestrator.CalculationStatusError,
			"",
			"No operations found",
		)
		if updateErr != nil {
			return fmt.Errorf("failed to update calculation status: %w", updateErr)
		}
		return nil
	}

	// Определение статуса вычисления на основе статусов операций
	status, result, errorMsg := uc.determineCalculationStatus(operations)
	log.Info("Determined calculation status",
		zap.String("status", string(status)),
		zap.String("result", result),
		zap.String("error_message", errorMsg))

	// Обновление статуса вычисления
	return uc.updateCalculationStatusWithRetry(timeoutCtx, calculationID, status, result, errorMsg, log)
}

// getCalculationWithRetry получает вычисление с повторными попытками при ошибках
func (uc *UseCaseImpl) getCalculationWithRetry(ctx context.Context, calculationID uuid.UUID, _ logger.Logger) (*orchestrator.Calculation, error) {
	var calculation *orchestrator.Calculation
	var err error
	var lastErr error

	// Повторные попытки с экспоненциальной задержкой
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			backoffDuration := time.Duration(100*(1<<attempt)) * time.Millisecond
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			case <-time.After(backoffDuration):
			}
		}

		calcCtx, calcCancel := context.WithTimeout(ctx, statusTimeout)
		calculation, err = uc.calculationRepo.FindByID(calcCtx, calculationID)
		calcCancel()

		if err == nil {
			break
		}

		lastErr = err
		if !isTransientError(err) {
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("failed to fetch calculation: %w", lastErr)
	}

	if calculation == nil {
		return nil, fmt.Errorf("%w: %s", domainerrors.ErrSpecificCalcNotFound, calculationID)
	}

	return calculation, nil
}

// getOperationsWithRetry получает операции вычисления с повторными попытками при ошибках
func (uc *UseCaseImpl) getOperationsWithRetry(ctx context.Context, calculationID uuid.UUID, log logger.Logger) ([]*orchestrator.Operation, error) {
	if calculationID == uuid.Nil {
		return []*orchestrator.Operation{}, fmt.Errorf("invalid calculation ID (nil UUID)")
	}

	safeLog := log
	if safeLog == nil {
		safeLog = logger.ContextLogger(ctx, nil)
	}

	if uc.operationRepo == nil {
		return []*orchestrator.Operation{}, domainerrors.ErrOpRepoNil
	}

	var operations []*orchestrator.Operation
	var lastErr error

	// Повторные попытки с экспоненциальной задержкой
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			backoffDuration := time.Duration(100*(1<<attempt)) * time.Millisecond
			select {
			case <-ctx.Done():
				return []*orchestrator.Operation{}, fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			case <-time.After(backoffDuration):
			}
		}

		opCtx, cancel := context.WithTimeout(ctx, statusTimeout)
		ops, err := uc.operationRepo.FindByCalculationID(opCtx, calculationID)
		cancel()

		if err == nil {
			operations = ops
			break
		}

		lastErr = err
		if !isTransientError(err) {
			break
		}
	}

	if lastErr != nil && operations == nil {
		return []*orchestrator.Operation{}, lastErr
	}

	if operations == nil {
		return []*orchestrator.Operation{}, nil
	}

	// Фильтрация нулевых операций
	validOps := make([]*orchestrator.Operation, 0, len(operations))
	for _, op := range operations {
		if op != nil {
			validOps = append(validOps, op)
		}
	}

	return validOps, nil
}

// determineCalculationStatus определяет статус вычисления на основе статусов операций
func (uc *UseCaseImpl) determineCalculationStatus(operations []*orchestrator.Operation) (orchestrator.CalculationStatus, string, string) {
	if len(operations) == 0 {
		return orchestrator.CalculationStatusError, "", "No operations found"
	}

	// Фильтрация нулевых операций
	validOps := make([]*orchestrator.Operation, 0, len(operations))
	for _, op := range operations {
		if op != nil {
			validOps = append(validOps, op)
		}
	}

	if len(validOps) == 0 {
		return orchestrator.CalculationStatusError, "", "No valid operations found"
	}

	// Подсчет операций по статусам
	totalOps := len(validOps)
	completedOps := 0
	errorOps := 0
	pendingOps := 0
	inProgressOps := 0
	var finalResult string
	var errorMessages []string

	for _, op := range validOps {
		switch op.Status {
		case orchestrator.OperationStatusCompleted:
			completedOps++
			finalResult = op.Result
		case orchestrator.OperationStatusError:
			errorOps++
			if op.ErrorMessage != "" {
				errorMessages = append(errorMessages, op.ErrorMessage)
			}
		case orchestrator.OperationStatusPending:
			pendingOps++
		case orchestrator.OperationStatusInProgress:
			inProgressOps++
		}
	}

	// Определение итогового статуса
	if completedOps == totalOps {
		return orchestrator.CalculationStatusCompleted, finalResult, ""
	}

	if pendingOps > 0 || inProgressOps > 0 {
		return orchestrator.CalculationStatusInProgress, "", ""
	}

	if errorOps > 0 {
		var errorMsg string
		if len(errorMessages) > 0 {
			fullError := strings.Join(errorMessages, "; ")
			if len(fullError) > maxErrorLength {
				errorMsg = fullError[:maxErrorLength] + "... (truncated)"
			} else {
				errorMsg = fullError
			}
		} else {
			errorMsg = "Calculation failed due to operation errors"
		}
		return orchestrator.CalculationStatusError, "", errorMsg
	}

	return orchestrator.CalculationStatusError, "", "Unknown calculation state"
}

// updateCalculationStatusWithRetry обновляет статус вычисления с повторными попытками при ошибках
func (uc *UseCaseImpl) updateCalculationStatusWithRetry(
	ctx context.Context,
	calculationID uuid.UUID,
	status orchestrator.CalculationStatus,
	result string,
	errorMsg string,
	_ logger.Logger,
) error {
	var lastErr error

	// Повторные попытки с экспоненциальной задержкой
	for attempt := range maxRetries {
		if attempt > 0 {
			backoffDuration := time.Duration(100*(1<<attempt)) * time.Millisecond
			select {
			case <-ctx.Done():
				return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			case <-time.After(backoffDuration):
			}
		}

		err := uc.calculationRepo.UpdateStatus(ctx, calculationID, status, result, errorMsg)
		if err == nil {
			return nil
		}

		lastErr = err
		if !isTransientError(err) {
			break
		}
	}

	return fmt.Errorf("failed to update calculation status after %d attempts: %w", maxRetries, lastErr)
}

// isTransientError определяет, является ли ошибка временной и подходящей для повторной попытки
func isTransientError(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}

	errMsg := err.Error()
	return strings.Contains(errMsg, "timeout") ||
		strings.Contains(errMsg, "temporary") ||
		strings.Contains(errMsg, "retriable") ||
		strings.Contains(errMsg, "agent busy")
}

// Close освобождает ресурсы сервиса вычислений
func (uc *UseCaseImpl) Close() error {
	return nil
}
