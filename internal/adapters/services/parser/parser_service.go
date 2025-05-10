package parser

import (
	"context"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/orchestrator"
	parserPort "github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/service/parser"
	"github.com/google/uuid"
)

var (
	ErrEmptyExpression        = errors.New("expression is empty")
	ErrInvalidExpression      = errors.New("invalid expression")
	ErrUnsupportedOperator    = errors.New("unsupported operator")
	ErrParsingExpression      = errors.New("error parsing expression")
	ErrInvalidBinaryOperation = errors.New("invalid binary operation")
	ErrInvalidParenExpression = errors.New("invalid parenthesized expression")
	ErrDivisionByZero         = errors.New("division by zero")
	ErrExpressionTooComplex   = errors.New("expression too complex")
)

type Service struct {
	maxOperations int
}

var _ parserPort.ExpressionParser = (*Service)(nil)

func NewService(maxOperations int) *Service {
	if maxOperations <= 0 {
		maxOperations = 100
	}
	return &Service{maxOperations: maxOperations}
}

func (s *Service) Validate(ctx context.Context, expression string) error {
	if strings.TrimSpace(expression) == "" {
		return ErrEmptyExpression
	}

	if _, err := parser.ParseExpr(expression); err != nil {
		return fmt.Errorf("%w: %s", ErrInvalidExpression, err.Error())
	}

	return nil
}

func (s *Service) Parse(ctx context.Context, expression string) ([]*orchestrator.Operation, error) {
	if err := s.Validate(ctx, expression); err != nil {
		return nil, err
	}

	expr, err := parser.ParseExpr(expression)
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrParsingExpression, err.Error())
	}

	operations := make([]*orchestrator.Operation, 0, 16)
	if _, err = s.processExpression(ctx, expr, &operations, nil); err != nil {
		return nil, err
	}

	if len(operations) > s.maxOperations {
		return nil, ErrExpressionTooComplex
	}

	return operations, nil
}

func (s *Service) processExpression(
	ctx context.Context,
	expr ast.Expr,
	operations *[]*orchestrator.Operation,
	calculationID *uuid.UUID,
) (string, error) {
	var calcID uuid.UUID
	if calculationID != nil {
		calcID = *calculationID
	}

	switch e := expr.(type) {
	case *ast.BinaryExpr:
		return s.processBinaryExpr(ctx, e, operations, calculationID)

	case *ast.BasicLit:
		return e.Value, nil

	case *ast.ParenExpr:
		return s.processExpression(ctx, e.X, operations, calculationID)

	case *ast.UnaryExpr:
		if e.Op == token.SUB {
			val, err := s.processExpression(ctx, e.X, operations, calculationID)
			if err != nil {
				return "", err
			}

			if _, err := strconv.ParseFloat(val, 64); err == nil {
				return "-" + val, nil
			}

			op := &orchestrator.Operation{
				ID:            uuid.New(),
				CalculationID: calcID,
				OperationType: orchestrator.OperationTypeSubtraction,
				Operand1:      "0",
				Operand2:      val,
				Status:        orchestrator.OperationStatusPending,
			}

			*operations = append(*operations, op)
			return op.ID.String(), nil
		}
		return "", ErrUnsupportedOperator

	default:
		return "", ErrInvalidExpression
	}
}

func (s *Service) processBinaryExpr(
	ctx context.Context,
	expr *ast.BinaryExpr,
	operations *[]*orchestrator.Operation,
	calculationID *uuid.UUID,
) (string, error) {
	leftVal, err := s.processExpression(ctx, expr.X, operations, calculationID)
	if err != nil {
		return "", err
	}

	rightVal, err := s.processExpression(ctx, expr.Y, operations, calculationID)
	if err != nil {
		return "", err
	}

	// Check if operands are numeric or reference IDs
	leftIsUUID := isUUIDReference(leftVal)
	rightIsUUID := isUUIDReference(rightVal)

	// If division by zero check is needed, make sure to parse non-UUID values
	if expr.Op == token.QUO && !rightIsUUID {
		if rightVal == "0" {
			return "", ErrDivisionByZero
		}
	}

	var operType orchestrator.OperationType
	switch expr.Op {
	case token.ADD:
		operType = orchestrator.OperationTypeAddition
	case token.SUB:
		operType = orchestrator.OperationTypeSubtraction
	case token.MUL:
		operType = orchestrator.OperationTypeMultiplication
	case token.QUO:
		operType = orchestrator.OperationTypeDivision
	default:
		return "", ErrUnsupportedOperator
	}

	var calcID uuid.UUID
	if calculationID != nil {
		calcID = *calculationID
	}

	// Store metadata about operand types
	var metadataLeft, metadataRight string
	if leftIsUUID {
		metadataLeft = "ref:"
	}
	if rightIsUUID {
		metadataRight = "ref:"
	}

	op := &orchestrator.Operation{
		ID:            uuid.New(),
		CalculationID: calcID,
		OperationType: operType,
		Operand1:      metadataLeft + leftVal,
		Operand2:      metadataRight + rightVal,
		Status:        orchestrator.OperationStatusPending,
	}

	*operations = append(*operations, op)
	return op.ID.String(), nil
}

func isUUIDReference(val string) bool {
	_, err := uuid.Parse(val)
	return err == nil && len(val) == 36 // Standard UUID length
}

func (s *Service) SetCalculationID(operations []*orchestrator.Operation, calculationID uuid.UUID) {
	for i := range operations {
		operations[i].CalculationID = calculationID
	}
}
