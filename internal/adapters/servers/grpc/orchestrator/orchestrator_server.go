package orchestrator

import (
	"context"
	"fmt"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/orchestrator"
	orchapi "github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/api/orchestrator"
	orchv1 "github.com/flexer2006/y.lms-final-task-calc-go/pkg/api/proto/v1/orchestrator"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	fieldOp            = "op"
	fieldCalculationID = "calculation_id"
	fieldCount         = "count"

	msgEmptyExpression      = "Empty expression provided"
	msgEmptyCalculationID   = "Empty calculation ID provided"
	msgInvalidCalculationID = "Invalid calculation ID"
	msgFailedGetUserID      = "Failed to get user ID"
	msgCalcNotFound         = "Calculation not found"
	msgCalcListSuccess      = "Calculations list retrieved successfully"

	errExpressionEmpty = "expression cannot be empty"
	errCalcIDEmpty     = "calculation ID cannot be empty"
	errInvalidCalcID   = "invalid calculation ID"
	errCalcNotFound    = "calculation not found"
	errCalcFailed      = "failed to calculate expression"
	errGetCalcFailed   = "failed to get calculation"
	errListCalcFailed  = "failed to list calculations"
	errMissingMetadata = "missing metadata"
	errMissingUserID   = "missing user ID"
	errInvalidUserID   = "invalid user ID"

	opCalculate        = "OrchestratorServer.Calculate"
	opGetCalculation   = "OrchestratorServer.GetCalculation"
	opListCalculations = "OrchestratorServer.ListCalculations"
)

type Server struct {
	orchv1.UnimplementedOrchestratorServiceServer
	calculationUseCase orchapi.UseCaseCalculation
}

func NewServer(calculationUseCase orchapi.UseCaseCalculation) *Server {
	return &Server{
		calculationUseCase: calculationUseCase,
	}
}

func newGRPCError(code codes.Code, msg string) error {
	return fmt.Errorf("gRPC error: %w", status.Error(code, msg))
}

func getUserID(ctx context.Context) (uuid.UUID, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return uuid.Nil, newGRPCError(codes.Unauthenticated, errMissingMetadata)
	}

	values := md.Get("user_id")
	if len(values) == 0 {
		return uuid.Nil, newGRPCError(codes.Unauthenticated, errMissingUserID)
	}

	userID, err := uuid.Parse(values[0])
	if err != nil {
		return uuid.Nil, newGRPCError(codes.Unauthenticated, errInvalidUserID)
	}

	return userID, nil
}

func (s *Server) Calculate(ctx context.Context, req *orchv1.CalculateRequest) (*orchv1.CalculateResponse, error) {
	log := logger.ContextLogger(ctx, nil).With(zap.String(fieldOp, opCalculate))

	if req.GetExpression() == "" {
		log.Warn(msgEmptyExpression)
		return nil, newGRPCError(codes.InvalidArgument, errExpressionEmpty)
	}

	userID, err := getUserID(ctx)
	if err != nil {
		log.Warn(msgFailedGetUserID, zap.Error(err))
		return nil, err
	}

	calculation, err := s.calculationUseCase.CalculateExpression(ctx, userID, req.GetExpression())
	if err != nil {
		log.Error(errCalcFailed, zap.Error(err))
		return nil, newGRPCError(codes.Internal, errCalcFailed)
	}

	return &orchv1.CalculateResponse{
		Id:           calculation.ID.String(),
		Status:       mapCalculationStatusToProto(calculation.Status),
		Result:       calculation.Result,
		ErrorMessage: calculation.ErrorMessage,
	}, nil
}

func (s *Server) GetCalculation(ctx context.Context, req *orchv1.GetCalculationRequest) (*orchv1.GetCalculationResponse, error) {
	log := logger.ContextLogger(ctx, nil).With(
		zap.String(fieldOp, opGetCalculation),
		zap.String(fieldCalculationID, req.GetId()),
	)

	if req.GetId() == "" {
		log.Warn(msgEmptyCalculationID)
		return nil, newGRPCError(codes.InvalidArgument, errCalcIDEmpty)
	}

	userID, err := getUserID(ctx)
	if err != nil {
		log.Warn(msgFailedGetUserID, zap.Error(err))
		return nil, err
	}

	calculationID, err := uuid.Parse(req.GetId())
	if err != nil {
		log.Warn(msgInvalidCalculationID, zap.Error(err))
		return nil, newGRPCError(codes.InvalidArgument, errInvalidCalcID)
	}

	calculation, err := s.calculationUseCase.GetCalculation(ctx, calculationID, userID)
	if err != nil {
		log.Error(errGetCalcFailed, zap.Error(err))
		return nil, newGRPCError(codes.Internal, errGetCalcFailed)
	}

	if calculation == nil {
		log.Warn(msgCalcNotFound)
		return nil, newGRPCError(codes.NotFound, errCalcNotFound)
	}

	return mapCalculationToProtoResponse(calculation), nil
}

func (s *Server) ListCalculations(ctx context.Context, _ *emptypb.Empty) (*orchv1.ListCalculationsResponse, error) {
	log := logger.ContextLogger(ctx, nil).With(zap.String(fieldOp, opListCalculations))

	userID, err := getUserID(ctx)
	if err != nil {
		log.Warn(msgFailedGetUserID, zap.Error(err))
		return nil, err
	}

	calculations, err := s.calculationUseCase.ListCalculations(ctx, userID)
	if err != nil {
		log.Error(errListCalcFailed, zap.Error(err))
		return nil, newGRPCError(codes.Internal, errListCalcFailed)
	}

	response := &orchv1.ListCalculationsResponse{
		Calculations: make([]*orchv1.GetCalculationResponse, len(calculations)),
	}

	for i, calc := range calculations {
		response.Calculations[i] = mapCalculationToProtoResponse(calc)
	}

	log.Info(msgCalcListSuccess, zap.Int(fieldCount, len(calculations)))
	return response, nil
}

func mapCalculationStatusToProto(status orchestrator.CalculationStatus) orchv1.CalculationStatus {
	switch status {
	case orchestrator.CalculationStatusPending:
		return orchv1.CalculationStatus_PENDING
	case orchestrator.CalculationStatusInProgress:
		return orchv1.CalculationStatus_IN_PROGRESS
	case orchestrator.CalculationStatusCompleted:
		return orchv1.CalculationStatus_COMPLETED
	case orchestrator.CalculationStatusError:
		return orchv1.CalculationStatus_ERROR
	default:
		return orchv1.CalculationStatus_PENDING
	}
}

func mapCalculationToProtoResponse(calculation *orchestrator.Calculation) *orchv1.GetCalculationResponse {
	if calculation == nil {
		return nil
	}

	return &orchv1.GetCalculationResponse{
		Id:           calculation.ID.String(),
		UserId:       calculation.UserID.String(),
		Expression:   calculation.Expression,
		Result:       calculation.Result,
		Status:       mapCalculationStatusToProto(calculation.Status),
		ErrorMessage: calculation.ErrorMessage,
		CreatedAt:    timestamppb.New(calculation.CreatedAt),
		UpdatedAt:    timestamppb.New(calculation.UpdatedAt),
	}
}
