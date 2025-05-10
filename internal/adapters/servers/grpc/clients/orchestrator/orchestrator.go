package orchestrator

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/orchestrator"
	orchAPI "github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/api/orchestrator"
	orchv1 "github.com/flexer2006/y.lms-final-task-calc-go/pkg/api/proto/v1/orchestrator"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const (
	methodCalculate        = "CalculateExpression"
	methodGetCalculation   = "GetCalculation"
	methodListCalculations = "ListCalculations"

	fieldMethod        = "method"
	fieldUserID        = "user_id"
	fieldCalculationID = "calculation_id"
	fieldExpression    = "expression"
	fieldStatus        = "status"
	fieldCount         = "count"

	metadataUserID = "user_id"

	msgFailedCalculate        = "failed to calculate expression"
	msgFailedGetCalculation   = "failed to get calculation"
	msgFailedListCalculations = "failed to list calculations"
	msgInvalidCalculationID   = "invalid calculation ID"
	msgInvalidUserID          = "invalid user ID"
	msgEmptyExpression        = "expression cannot be empty"

	defaultDialTimeout = 5 * time.Second
)

var (
	ErrConnectionTimeout    = errors.New("connection timeout: failed to connect to orchestrator service")
	ErrInvalidResponse      = errors.New("invalid response from orchestrator service")
	ErrInvalidCalculationID = errors.New("invalid calculation ID format")
	ErrInvalidUserID        = errors.New("invalid user ID format")
	ErrCalculationNotFound  = errors.New("calculation not found")
	ErrUnauthorizedAccess   = errors.New("unauthorized access to calculation")
	ErrInvalidExpression    = errors.New("invalid expression")
	ErrInternalServerError  = errors.New("internal server error")
	ErrInvalidArgument      = errors.New("invalid argument") // Add this new error
)

type Client struct {
	client orchv1.OrchestratorServiceClient
	conn   *grpc.ClientConn
}

func NewCalculationUseCase(ctx context.Context, address string) (orchAPI.UseCaseCalculation, error) {
	dialCtx, cancel := context.WithTimeout(ctx, defaultDialTimeout)
	defer cancel()

	// NewClient takes a target string followed by options (not a context)
	conn, err := grpc.Dial(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to orchestrator service at %s: %w", address, err)
	}

	if !waitForConnection(dialCtx, conn) {
		if err := conn.Close(); err != nil {
			return nil, fmt.Errorf("failed to close connection: %w", err)
		}
		return nil, ErrConnectionTimeout
	}

	return &Client{
		client: orchv1.NewOrchestratorServiceClient(conn),
		conn:   conn,
	}, nil
}

func waitForConnection(ctx context.Context, conn *grpc.ClientConn) bool {
	for {
		if conn.GetState() == connectivity.Ready {
			return true
		}
		if !conn.WaitForStateChange(ctx, conn.GetState()) {
			return false
		}
	}
}

func (c *Client) CalculateExpression(ctx context.Context, userID uuid.UUID, expression string) (*orchestrator.Calculation, error) {
	log := logger.ContextLogger(ctx, nil).With(
		zap.String(fieldMethod, methodCalculate),
		zap.String(fieldUserID, userID.String()),
		zap.String(fieldExpression, expression),
	)

	ctx = metadata.AppendToOutgoingContext(ctx, metadataUserID, userID.String())

	resp, err := c.client.Calculate(ctx, &orchv1.CalculateRequest{
		Expression: expression,
	})
	if err != nil {
		log.Error("Failed to calculate expression", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", msgFailedCalculate, mapGRPCError(err))
	}

	calculationID, err := uuid.Parse(resp.GetId())
	if err != nil {
		log.Error("Invalid calculation ID received",
			zap.String(fieldCalculationID, resp.GetId()),
			zap.Error(err))
		return nil, ErrInvalidCalculationID
	}

	status := mapProtoStatusToDomain(resp.GetStatus())

	calculation := &orchestrator.Calculation{
		ID:           calculationID,
		UserID:       userID,
		Expression:   expression,
		Result:       resp.GetResult(),
		Status:       status,
		ErrorMessage: resp.GetErrorMessage(),
	}

	log.Info("Expression calculation initiated successfully",
		zap.String(fieldCalculationID, calculationID.String()),
		zap.String(fieldStatus, string(status)))

	return calculation, nil
}

func (c *Client) GetCalculation(ctx context.Context, calculationID uuid.UUID, userID uuid.UUID) (*orchestrator.Calculation, error) {
	log := logger.ContextLogger(ctx, nil).With(
		zap.String(fieldMethod, methodGetCalculation),
		zap.String(fieldCalculationID, calculationID.String()),
		zap.String(fieldUserID, userID.String()),
	)

	ctx = metadata.AppendToOutgoingContext(ctx, metadataUserID, userID.String())

	resp, err := c.client.GetCalculation(ctx, &orchv1.GetCalculationRequest{
		Id: calculationID.String(),
	})
	if err != nil {
		log.Error("Failed to get calculation", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", msgFailedGetCalculation, mapGRPCError(err))
	}

	calcID, err := uuid.Parse(resp.GetId())
	if err != nil {
		log.Error("Invalid calculation ID received",
			zap.String(fieldCalculationID, resp.GetId()),
			zap.Error(err))
		return nil, ErrInvalidCalculationID
	}

	respUserID, err := uuid.Parse(resp.GetUserId())
	if err != nil {
		log.Error("Invalid user ID received",
			zap.String(fieldUserID, resp.GetUserId()),
			zap.Error(err))
		return nil, ErrInvalidUserID
	}

	status := mapProtoStatusToDomain(resp.GetStatus())

	calculation := &orchestrator.Calculation{
		ID:           calcID,
		UserID:       respUserID,
		Expression:   resp.GetExpression(),
		Result:       resp.GetResult(),
		Status:       status,
		ErrorMessage: resp.GetErrorMessage(),
		CreatedAt:    resp.GetCreatedAt().AsTime(),
		UpdatedAt:    resp.GetUpdatedAt().AsTime(),
	}

	log.Debug("Calculation retrieved successfully", zap.String(fieldStatus, string(status)))
	return calculation, nil
}

func (c *Client) ListCalculations(ctx context.Context, userID uuid.UUID) ([]*orchestrator.Calculation, error) {
	log := logger.ContextLogger(ctx, nil).With(
		zap.String(fieldMethod, methodListCalculations),
		zap.String(fieldUserID, userID.String()),
	)

	ctx = metadata.AppendToOutgoingContext(ctx, metadataUserID, userID.String())

	resp, err := c.client.ListCalculations(ctx, &emptypb.Empty{})
	if err != nil {
		log.Error("Failed to list calculations", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", msgFailedListCalculations, mapGRPCError(err))
	}

	calculations := make([]*orchestrator.Calculation, 0, len(resp.GetCalculations()))

	for _, calc := range resp.GetCalculations() {
		calcID, err := uuid.Parse(calc.GetId())
		if err != nil {
			log.Warn("Skipping calculation with invalid ID",
				zap.String(fieldCalculationID, calc.GetId()),
				zap.Error(err))
			continue
		}

		respUserID, err := uuid.Parse(calc.GetUserId())
		if err != nil {
			log.Warn("Skipping calculation with invalid user ID",
				zap.String(fieldUserID, calc.GetUserId()),
				zap.Error(err))
			continue
		}

		status := mapProtoStatusToDomain(calc.GetStatus())

		calculation := &orchestrator.Calculation{
			ID:           calcID,
			UserID:       respUserID,
			Expression:   calc.GetExpression(),
			Result:       calc.GetResult(),
			Status:       status,
			ErrorMessage: calc.GetErrorMessage(),
			CreatedAt:    calc.GetCreatedAt().AsTime(),
			UpdatedAt:    calc.GetUpdatedAt().AsTime(),
		}

		calculations = append(calculations, calculation)
	}

	log.Info("User calculations retrieved successfully", zap.Int(fieldCount, len(calculations)))
	return calculations, nil
}

func (c *Client) ProcessPendingOperations(ctx context.Context) error {
	return nil
}

func (c *Client) UpdateCalculationStatus(ctx context.Context, calculationID uuid.UUID) error {
	return nil
}

func (c *Client) Close() error {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return fmt.Errorf("failed to close gRPC connection: %w", err)
		}
	}
	return nil
}

func mapProtoStatusToDomain(status orchv1.CalculationStatus) orchestrator.CalculationStatus {
	switch status {
	case orchv1.CalculationStatus_PENDING:
		return orchestrator.CalculationStatusPending
	case orchv1.CalculationStatus_IN_PROGRESS:
		return orchestrator.CalculationStatusInProgress
	case orchv1.CalculationStatus_COMPLETED:
		return orchestrator.CalculationStatusCompleted
	case orchv1.CalculationStatus_ERROR:
		return orchestrator.CalculationStatusError
	default:
		return orchestrator.CalculationStatusPending
	}
}

func mapGRPCError(err error) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return err
	}

	switch st.Code() {
	case codes.NotFound:
		return ErrCalculationNotFound
	case codes.PermissionDenied, codes.Unauthenticated:
		return ErrUnauthorizedAccess
	case codes.InvalidArgument:
		if st.Message() == msgEmptyExpression {
			return ErrInvalidExpression
		}
		// Use static error instead of dynamic error
		return fmt.Errorf("%w: %s", ErrInvalidArgument, st.Message())
	case codes.Internal:
		return ErrInternalServerError
	default:
		return err
	}
}
