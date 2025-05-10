package calculation_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/app/orchestrator/calculation"
	domainerrors "github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/errord"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/orchestrator"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

type MockCalculationRepository struct {
	mock.Mock
}

func (m *MockCalculationRepository) Create(ctx context.Context, calculation *orchestrator.Calculation) (*orchestrator.Calculation, error) {
	args := m.Called(ctx, calculation)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*orchestrator.Calculation), args.Error(1)
}

func (m *MockCalculationRepository) FindByID(ctx context.Context, id uuid.UUID) (*orchestrator.Calculation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*orchestrator.Calculation), args.Error(1)
}

func (m *MockCalculationRepository) FindByUserID(ctx context.Context, userID uuid.UUID) ([]*orchestrator.Calculation, error) {
	args := m.Called(ctx, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*orchestrator.Calculation), args.Error(1)
}

func (m *MockCalculationRepository) Update(ctx context.Context, calculation *orchestrator.Calculation) error {
	args := m.Called(ctx, calculation)
	return args.Error(0)
}

func (m *MockCalculationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status orchestrator.CalculationStatus, result string, errorMsg string) error {
	args := m.Called(ctx, id, status, result, errorMsg)
	return args.Error(0)
}

func (m *MockCalculationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockOperationRepository struct {
	mock.Mock
}

func (m *MockOperationRepository) Create(ctx context.Context, operation *orchestrator.Operation) (*orchestrator.Operation, error) {
	args := m.Called(ctx, operation)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*orchestrator.Operation), args.Error(1)
}

func (m *MockOperationRepository) CreateBatch(ctx context.Context, operations []*orchestrator.Operation) error {
	args := m.Called(ctx, operations)
	return args.Error(0)
}

func (m *MockOperationRepository) FindByID(ctx context.Context, id uuid.UUID) (*orchestrator.Operation, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*orchestrator.Operation), args.Error(1)
}

func (m *MockOperationRepository) FindByCalculationID(ctx context.Context, calculationID uuid.UUID) ([]*orchestrator.Operation, error) {
	args := m.Called(ctx, calculationID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*orchestrator.Operation), args.Error(1)
}

func (m *MockOperationRepository) GetPendingOperations(ctx context.Context, limit int) ([]*orchestrator.Operation, error) {
	args := m.Called(ctx, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*orchestrator.Operation), args.Error(1)
}

func (m *MockOperationRepository) Update(ctx context.Context, operation *orchestrator.Operation) error {
	args := m.Called(ctx, operation)
	return args.Error(0)
}

func (m *MockOperationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status orchestrator.OperationStatus, result string, errorMsg string) error {
	args := m.Called(ctx, id, status, result, errorMsg)
	return args.Error(0)
}

func (m *MockOperationRepository) AssignAgent(ctx context.Context, operationID uuid.UUID, agentID string) error {
	args := m.Called(ctx, operationID, agentID)
	return args.Error(0)
}

type MockExpressionParser struct {
	mock.Mock
}

func (m *MockExpressionParser) Parse(ctx context.Context, expression string) ([]*orchestrator.Operation, error) {
	args := m.Called(ctx, expression)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*orchestrator.Operation), args.Error(1)
}

func (m *MockExpressionParser) Validate(ctx context.Context, expression string) error {
	args := m.Called(ctx, expression)
	return args.Error(0)
}

func (m *MockExpressionParser) SetCalculationID(operations []*orchestrator.Operation, calculationID uuid.UUID) {
	m.Called(operations, calculationID)
}

type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, fields ...logger.Field) {
	var args []interface{}
	args = append(args, msg)
	for _, field := range fields {
		args = append(args, field)
	}
	m.Called(args...)
}

func (m *MockLogger) Info(msg string, fields ...logger.Field) {
	var args []interface{}
	args = append(args, msg)
	for _, field := range fields {
		args = append(args, field)
	}
	m.Called(args...)
}

func (m *MockLogger) Warn(msg string, fields ...logger.Field) {
	var args []interface{}
	args = append(args, msg)
	for _, field := range fields {
		args = append(args, field)
	}
	m.Called(args...)
}

func (m *MockLogger) Error(msg string, fields ...logger.Field) {
	var args []interface{}
	args = append(args, msg)
	for _, field := range fields {
		args = append(args, field)
	}
	m.Called(args...)
}

func (m *MockLogger) Fatal(msg string, fields ...logger.Field) {
	var args []interface{}
	args = append(args, msg)
	for _, field := range fields {
		args = append(args, field)
	}
	m.Called(args...)
}

func (m *MockLogger) With(fields ...logger.Field) logger.Logger {
	args := m.Called(fields)
	return args.Get(0).(logger.Logger)
}

func (m *MockLogger) SetLevel(lvl logger.LogLevel) {
	m.Called(lvl)
}

func (m *MockLogger) GetLevel() logger.LogLevel {
	args := m.Called()
	return args.Get(0).(logger.LogLevel)
}

func (m *MockLogger) Sync() error {
	args := m.Called()
	return args.Error(0)
}

func (m *MockLogger) RawLogger() *zap.Logger {
	args := m.Called()
	return args.Get(0).(*zap.Logger)
}

func setupTestContext() context.Context {
	mockLog := new(MockLogger)
	mockLog.On("With", mock.Anything).Return(mockLog).Maybe()
	mockLog.On("Debug", mock.Anything, mock.Anything).Maybe()
	mockLog.On("Info", mock.Anything, mock.Anything).Maybe()
	mockLog.On("Info", mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLog.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLog.On("Info", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Maybe()
	mockLog.On("Warn", mock.Anything, mock.Anything).Maybe()
	mockLog.On("Error", mock.Anything, mock.Anything).Maybe()
	mockLog.On("RawLogger").Return(zap.NewNop()).Maybe()

	ctx := logger.WithLogger(context.Background(), mockLog)
	return ctx
}

func TestCalculateExpression(t *testing.T) {
	testCases := []struct {
		name           string
		userID         uuid.UUID
		expression     string
		setupMocks     func(*MockCalculationRepository, *MockOperationRepository, *MockExpressionParser)
		expectedError  error
		expectedStatus orchestrator.CalculationStatus
	}{
		{
			name:       "Success case",
			userID:     uuid.New(),
			expression: "1+2",
			setupMocks: func(calcRepo *MockCalculationRepository, opRepo *MockOperationRepository, parser *MockExpressionParser) {
				parser.On("Validate", mock.Anything, "1+2").Return(nil)

				calcRepo.On("Create", mock.Anything, mock.MatchedBy(func(calc *orchestrator.Calculation) bool {
					return calc.Expression == "1+2" &&
						calc.Status == orchestrator.CalculationStatusPending
				})).Return(&orchestrator.Calculation{
					ID:         uuid.New(),
					UserID:     uuid.New(),
					Expression: "1+2",
					Status:     orchestrator.CalculationStatusPending,
				}, nil)

				operations := []*orchestrator.Operation{
					{
						ID:            uuid.New(),
						OperationType: orchestrator.OperationTypeAddition,
						Operand1:      "1",
						Operand2:      "2",
						Status:        orchestrator.OperationStatusPending,
					},
				}

				parser.On("Parse", mock.Anything, "1+2").Return(operations, nil)
				parser.On("SetCalculationID", operations, mock.Anything).Return()
				opRepo.On("CreateBatch", mock.Anything, operations).Return(nil)

				calcRepo.On("UpdateStatus", mock.Anything, mock.Anything, orchestrator.CalculationStatusInProgress, "", "").Return(nil)
				calcRepo.On("FindByID", mock.Anything, mock.Anything).Return(&orchestrator.Calculation{
					ID:         uuid.New(),
					UserID:     uuid.New(),
					Expression: "1+2",
					Status:     orchestrator.CalculationStatusInProgress,
				}, nil)
			},
			expectedError:  nil,
			expectedStatus: orchestrator.CalculationStatusInProgress,
		},
		{
			name:       "Invalid user ID",
			userID:     uuid.Nil,
			expression: "1+2",
			setupMocks: func(calcRepo *MockCalculationRepository, opRepo *MockOperationRepository, parser *MockExpressionParser) {
			},
			expectedError:  domainerrors.ErrInvalidUserID,
			expectedStatus: "",
		},
		{
			name:       "Empty expression",
			userID:     uuid.New(),
			expression: "",
			setupMocks: func(calcRepo *MockCalculationRepository, opRepo *MockOperationRepository, parser *MockExpressionParser) {
			},
			expectedError:  domainerrors.ErrInvalidExpression,
			expectedStatus: "",
		},
		{
			name:       "Invalid expression",
			userID:     uuid.New(),
			expression: "1++2",
			setupMocks: func(calcRepo *MockCalculationRepository, opRepo *MockOperationRepository, parser *MockExpressionParser) {
				parser.On("Validate", mock.Anything, "1++2").Return(errors.New("syntax error"))
			},
			expectedError:  domainerrors.ErrInvalidExpression,
			expectedStatus: "",
		},
		{
			name:       "Repository error",
			userID:     uuid.New(),
			expression: "1+2",
			setupMocks: func(calcRepo *MockCalculationRepository, opRepo *MockOperationRepository, parser *MockExpressionParser) {
				parser.On("Validate", mock.Anything, "1+2").Return(nil)

				calcRepo.On("Create", mock.Anything, mock.Anything).Return(nil, errors.New("database error"))
			},
			expectedError:  domainerrors.ErrInternalError,
			expectedStatus: "",
		},
		{
			name:       "Parser error",
			userID:     uuid.New(),
			expression: "1+2",
			setupMocks: func(calcRepo *MockCalculationRepository, opRepo *MockOperationRepository, parser *MockExpressionParser) {
				parser.On("Validate", mock.Anything, "1+2").Return(nil)

				calcRepo.On("Create", mock.Anything, mock.Anything).Return(&orchestrator.Calculation{
					ID:         uuid.New(),
					UserID:     uuid.New(),
					Expression: "1+2",
					Status:     orchestrator.CalculationStatusPending,
				}, nil)

				parser.On("Parse", mock.Anything, "1+2").Return(nil, errors.New("parsing error"))

				calcRepo.On("UpdateStatus", mock.Anything, mock.Anything, orchestrator.CalculationStatusError, "", "parsing error").Return(nil)
				calcRepo.On("FindByID", mock.Anything, mock.Anything).Return(&orchestrator.Calculation{
					ID:           uuid.New(),
					UserID:       uuid.New(),
					Expression:   "1+2",
					Status:       orchestrator.CalculationStatusError,
					ErrorMessage: "parsing error",
				}, nil)
			},
			expectedError:  nil,
			expectedStatus: orchestrator.CalculationStatusError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := setupTestContext()

			calcRepo := new(MockCalculationRepository)
			opRepo := new(MockOperationRepository)
			parser := new(MockExpressionParser)

			tc.setupMocks(calcRepo, opRepo, parser)

			uc := calculation.NewUseCase(calcRepo, opRepo, parser)

			result, err := uc.CalculateExpression(ctx, tc.userID, tc.expression)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tc.expectedError) ||
					strings.Contains(err.Error(), tc.expectedError.Error()),
					"expected error containing %v, got %v", tc.expectedError, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tc.expectedStatus != "" {
					assert.Equal(t, tc.expectedStatus, result.Status)
				}
			}

			calcRepo.AssertExpectations(t)
			opRepo.AssertExpectations(t)
			parser.AssertExpectations(t)
		})
	}
}

func TestGetCalculation(t *testing.T) {
	calculationID := uuid.New()
	userID := uuid.New()
	otherUserID := uuid.New()

	testCases := []struct {
		name          string
		calculationID uuid.UUID
		userID        uuid.UUID
		setupMocks    func(*MockCalculationRepository, *MockOperationRepository)
		expectedError error
	}{
		{
			name:          "Success case",
			calculationID: calculationID,
			userID:        userID,
			setupMocks: func(calcRepo *MockCalculationRepository, opRepo *MockOperationRepository) {
				calcRepo.On("FindByID", mock.Anything, calculationID).Return(&orchestrator.Calculation{
					ID:         calculationID,
					UserID:     userID,
					Expression: "1+2",
					Result:     "3",
					Status:     orchestrator.CalculationStatusCompleted,
				}, nil)

				operations := []*orchestrator.Operation{
					{
						ID:            uuid.New(),
						CalculationID: calculationID,
						OperationType: orchestrator.OperationTypeAddition,
						Operand1:      "1",
						Operand2:      "2",
						Result:        "3",
						Status:        orchestrator.OperationStatusCompleted,
					},
				}

				opRepo.On("FindByCalculationID", mock.Anything, calculationID).Return(operations, nil)
			},
			expectedError: nil,
		},
		{
			name:          "Calculation not found",
			calculationID: calculationID,
			userID:        userID,
			setupMocks: func(calcRepo *MockCalculationRepository, opRepo *MockOperationRepository) {
				calcRepo.On("FindByID", mock.Anything, calculationID).Return(nil, nil)
			},
			expectedError: domainerrors.ErrCalculationNotFound,
		},
		{
			name:          "Repository error",
			calculationID: calculationID,
			userID:        userID,
			setupMocks: func(calcRepo *MockCalculationRepository, opRepo *MockOperationRepository) {
				calcRepo.On("FindByID", mock.Anything, calculationID).Return(nil, errors.New("database error"))
			},
			expectedError: domainerrors.ErrInternalError,
		},
		{
			name:          "Unauthorized access",
			calculationID: calculationID,
			userID:        otherUserID,
			setupMocks: func(calcRepo *MockCalculationRepository, opRepo *MockOperationRepository) {
				calcRepo.On("FindByID", mock.Anything, calculationID).Return(&orchestrator.Calculation{
					ID:     calculationID,
					UserID: userID,
				}, nil)
			},
			expectedError: domainerrors.ErrUnauthorizedAccess,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := setupTestContext()

			calcRepo := new(MockCalculationRepository)
			opRepo := new(MockOperationRepository)
			parser := new(MockExpressionParser)

			tc.setupMocks(calcRepo, opRepo)

			uc := calculation.NewUseCase(calcRepo, opRepo, parser)

			result, err := uc.GetCalculation(ctx, tc.calculationID, tc.userID)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tc.expectedError) ||
					strings.Contains(err.Error(), tc.expectedError.Error()),
					"expected error containing %v, got %v", tc.expectedError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tc.calculationID, result.ID)
				assert.Equal(t, tc.userID, result.UserID)
			}

			calcRepo.AssertExpectations(t)
			opRepo.AssertExpectations(t)
		})
	}
}

func TestListCalculations(t *testing.T) {
	userID := uuid.New()

	testCases := []struct {
		name          string
		userID        uuid.UUID
		setupMocks    func(*MockCalculationRepository)
		expectedCount int
		expectedError error
	}{
		{
			name:   "Success case with calculations",
			userID: userID,
			setupMocks: func(calcRepo *MockCalculationRepository) {
				calculations := []*orchestrator.Calculation{
					{
						ID:         uuid.New(),
						UserID:     userID,
						Expression: "1+2",
						Result:     "3",
						Status:     orchestrator.CalculationStatusCompleted,
					},
					{
						ID:         uuid.New(),
						UserID:     userID,
						Expression: "3*4",
						Result:     "12",
						Status:     orchestrator.CalculationStatusCompleted,
					},
				}

				calcRepo.On("FindByUserID", mock.Anything, userID).Return(calculations, nil)
			},
			expectedCount: 2,
			expectedError: nil,
		},
		{
			name:   "Success case no calculations",
			userID: userID,
			setupMocks: func(calcRepo *MockCalculationRepository) {
				calcRepo.On("FindByUserID", mock.Anything, userID).Return([]*orchestrator.Calculation{}, nil)
			},
			expectedCount: 0,
			expectedError: nil,
		},
		{
			name:   "Invalid user ID",
			userID: uuid.Nil,
			setupMocks: func(calcRepo *MockCalculationRepository) {
			},
			expectedCount: 0,
			expectedError: domainerrors.ErrInvalidUserID,
		},
		{
			name:   "Repository error",
			userID: userID,
			setupMocks: func(calcRepo *MockCalculationRepository) {
				calcRepo.On("FindByUserID", mock.Anything, userID).Return(nil, errors.New("database error"))
			},
			expectedCount: 0,
			expectedError: domainerrors.ErrInternalError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := setupTestContext()

			calcRepo := new(MockCalculationRepository)
			opRepo := new(MockOperationRepository)
			parser := new(MockExpressionParser)

			tc.setupMocks(calcRepo)

			uc := calculation.NewUseCase(calcRepo, opRepo, parser)

			result, err := uc.ListCalculations(ctx, tc.userID)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tc.expectedError) ||
					strings.Contains(err.Error(), tc.expectedError.Error()),
					"expected error containing %v, got %v", tc.expectedError, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, tc.expectedCount)
			}

			calcRepo.AssertExpectations(t)
		})
	}
}

func TestUpdateCalculationStatus(t *testing.T) {
	calculationID := uuid.New()

	testCases := []struct {
		name          string
		calculationID uuid.UUID
		setupMocks    func(*MockCalculationRepository, *MockOperationRepository)
		expectedError error
	}{
		{
			name:          "Success case - completed",
			calculationID: calculationID,
			setupMocks: func(calcRepo *MockCalculationRepository, opRepo *MockOperationRepository) {
				calcRepo.On("FindByID", mock.Anything, calculationID).Return(&orchestrator.Calculation{
					ID: calculationID,
				}, nil)

				operations := []*orchestrator.Operation{
					{
						ID:            uuid.New(),
						CalculationID: calculationID,
						Result:        "3",
						Status:        orchestrator.OperationStatusCompleted,
					},
				}

				opRepo.On("FindByCalculationID", mock.Anything, calculationID).Return(operations, nil)

				calcRepo.On("UpdateStatus", mock.Anything, calculationID,
					orchestrator.CalculationStatusCompleted, "3", "").Return(nil)
			},
			expectedError: nil,
		},
		{
			name:          "Success case - in progress",
			calculationID: calculationID,
			setupMocks: func(calcRepo *MockCalculationRepository, opRepo *MockOperationRepository) {
				calcRepo.On("FindByID", mock.Anything, calculationID).Return(&orchestrator.Calculation{
					ID: calculationID,
				}, nil)

				operations := []*orchestrator.Operation{
					{
						ID:            uuid.New(),
						CalculationID: calculationID,
						Status:        orchestrator.OperationStatusInProgress,
					},
				}

				opRepo.On("FindByCalculationID", mock.Anything, calculationID).Return(operations, nil)

				calcRepo.On("UpdateStatus", mock.Anything, calculationID,
					orchestrator.CalculationStatusInProgress, "", "").Return(nil)
			},
			expectedError: nil,
		},
		{
			name:          "Success case - error",
			calculationID: calculationID,
			setupMocks: func(calcRepo *MockCalculationRepository, opRepo *MockOperationRepository) {
				calcRepo.On("FindByID", mock.Anything, calculationID).Return(&orchestrator.Calculation{
					ID: calculationID,
				}, nil)

				operations := []*orchestrator.Operation{
					{
						ID:            uuid.New(),
						CalculationID: calculationID,
						Status:        orchestrator.OperationStatusError,
						ErrorMessage:  "calculation error",
					},
				}

				opRepo.On("FindByCalculationID", mock.Anything, calculationID).Return(operations, nil)

				calcRepo.On("UpdateStatus", mock.Anything, calculationID,
					orchestrator.CalculationStatusError, "", "calculation error").Return(nil)
			},
			expectedError: nil,
		},
		{
			name:          "Invalid calculation ID",
			calculationID: uuid.Nil,
			setupMocks:    func(calcRepo *MockCalculationRepository, opRepo *MockOperationRepository) {},
			expectedError: domainerrors.ErrSpecificCalcNotFound,
		},
		{
			name:          "No operations",
			calculationID: calculationID,
			setupMocks: func(calcRepo *MockCalculationRepository, opRepo *MockOperationRepository) {
				calcRepo.On("FindByID", mock.Anything, calculationID).Return(&orchestrator.Calculation{
					ID: calculationID,
				}, nil)

				opRepo.On("FindByCalculationID", mock.Anything, calculationID).Return([]*orchestrator.Operation{}, nil)

				calcRepo.On("UpdateStatus", mock.Anything, calculationID,
					orchestrator.CalculationStatusError, "", "No operations found").Return(nil)
			},
			expectedError: nil,
		},
		{
			name:          "Calculation not found",
			calculationID: calculationID,
			setupMocks: func(calcRepo *MockCalculationRepository, opRepo *MockOperationRepository) {
				calcRepo.On("FindByID", mock.Anything, calculationID).Return(nil, domainerrors.ErrCalculationNotFound)
			},
			expectedError: domainerrors.ErrCalculationNotFound,
		},
		{
			name:          "Error fetching operations",
			calculationID: calculationID,
			setupMocks: func(calcRepo *MockCalculationRepository, opRepo *MockOperationRepository) {
				calcRepo.On("FindByID", mock.Anything, calculationID).Return(&orchestrator.Calculation{
					ID: calculationID,
				}, nil)

				opRepo.On("FindByCalculationID", mock.Anything, calculationID).Return(nil, errors.New("database error"))
			},
			expectedError: errors.New("failed to fetch operations"),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := setupTestContext()

			calcRepo := new(MockCalculationRepository)
			opRepo := new(MockOperationRepository)
			parser := new(MockExpressionParser)

			tc.setupMocks(calcRepo, opRepo)

			uc := calculation.NewUseCase(calcRepo, opRepo, parser)

			err := uc.UpdateCalculationStatus(ctx, tc.calculationID)

			if tc.expectedError != nil {
				assert.Error(t, err)
				assert.True(t, errors.Is(err, tc.expectedError) ||
					strings.Contains(err.Error(), tc.expectedError.Error()),
					"expected error containing %v, got %v", tc.expectedError, err)
			} else {
				assert.NoError(t, err)
			}

			calcRepo.AssertExpectations(t)
			opRepo.AssertExpectations(t)
		})
	}
}
