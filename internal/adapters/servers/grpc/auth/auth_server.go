package auth

import (
	"context"
	"fmt"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/api/auth"
	authv1 "github.com/flexer2006/y.lms-final-task-calc-go/pkg/api/proto/v1/auth"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	fieldOp    = "op"
	fieldLogin = "login"

	msgEmptyLogin    = "Empty login provided"
	msgEmptyPassword = "Empty password provided"
	msgNoToken       = "Empty token provided" //nolint:gosec
	msgTokenFailed   = "Token validation failed"

	errLoginEmpty     = "login cannot be empty"
	errPasswordEmpty  = "password cannot be empty"
	errTokenEmpty     = "token cannot be empty"
	errRegisterFailed = "failed to register user"
	errLoginFailed    = "failed to login user"

	opRegister        = "AuthServer.Register"
	opLogin           = "AuthServer.Login"
	opTokenValidation = "AuthServer.ValidateToken" //nolint:gosec
)

func wrapError(code codes.Code, msg string) error {
	return fmt.Errorf("grpc error: %w", status.Error(code, msg))
}

type Server struct {
	authv1.UnimplementedAuthServiceServer
	authUseCase auth.UseCaseUser
}

func NewServer(authUseCase auth.UseCaseUser) *Server {
	return &Server{
		authUseCase: authUseCase,
	}
}

func (s *Server) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	login, password := req.GetLogin(), req.GetPassword()
	log := logger.ContextLogger(ctx, nil).With(zap.String(fieldOp, opRegister), zap.String(fieldLogin, login))

	if login == "" {
		log.Warn(msgEmptyLogin)
		return nil, wrapError(codes.InvalidArgument, errLoginEmpty)
	}

	if password == "" {
		log.Warn(msgEmptyPassword)
		return nil, wrapError(codes.InvalidArgument, errPasswordEmpty)
	}

	userID, err := s.authUseCase.Register(ctx, login, password)
	if err != nil {
		log.Error(errRegisterFailed, zap.Error(err))
		return nil, wrapError(codes.Internal, errRegisterFailed)
	}

	return &authv1.RegisterResponse{
		UserId: userID.String(),
	}, nil
}

func (s *Server) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	login, password := req.GetLogin(), req.GetPassword()
	log := logger.ContextLogger(ctx, nil).With(zap.String(fieldOp, opLogin), zap.String(fieldLogin, login))

	if login == "" {
		log.Warn(msgEmptyLogin)
		return nil, wrapError(codes.InvalidArgument, errLoginEmpty)
	}

	if password == "" {
		log.Warn(msgEmptyPassword)
		return nil, wrapError(codes.InvalidArgument, errPasswordEmpty)
	}

	tokenPair, err := s.authUseCase.Login(ctx, login, password)
	if err != nil {
		log.Error(errLoginFailed, zap.Error(err))
		return nil, wrapError(codes.Unauthenticated, errLoginFailed)
	}

	return &authv1.LoginResponse{
		UserId:       tokenPair.UserID.String(),
		Login:        login,
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		ExpiresAt:    timestamppb.New(tokenPair.ExpiresAt),
	}, nil
}

func (s *Server) ValidateToken(ctx context.Context, req *authv1.ValidateTokenRequest) (*authv1.ValidateTokenResponse, error) {
	token := req.GetToken()
	log := logger.ContextLogger(ctx, nil).With(zap.String(fieldOp, opTokenValidation))

	if token == "" {
		log.Warn(msgNoToken)
		return nil, wrapError(codes.InvalidArgument, errTokenEmpty)
	}

	userID, err := s.authUseCase.ValidateToken(ctx, token)
	if err != nil {
		log.Debug(msgTokenFailed, zap.Error(err))
		return &authv1.ValidateTokenResponse{
			UserId: "",
			Valid:  false,
		}, nil
	}

	return &authv1.ValidateTokenResponse{
		UserId: userID.String(),
		Valid:  true,
	}, nil
}
