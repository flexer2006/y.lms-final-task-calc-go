package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/auth"
	authAPI "github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/api/auth"
	authv1 "github.com/flexer2006/y.lms-final-task-calc-go/pkg/api/proto/v1/auth"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/connectivity"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

const (
	methodRegister      = "Register"
	methodLogin         = "Login"
	methodValidateToken = "ValidateToken"
	methodRefreshToken  = "RefreshToken"
	methodLogout        = "Logout"

	fieldMethod = "method"
	fieldLogin  = "login"
	fieldUserID = "user_id"

	errMsgRegister      = "failed to register user"
	errMsgLogin         = "failed to login"
	errMsgValidateToken = "failed to validate token"

	defaultDialTimeout = 5 * time.Second
	defaultTokenExpiry = 15 * time.Minute
)

var (
	ErrConnectionTimeout = errors.New("connection timeout: failed to connect to auth service")
	ErrInvalidResponse   = errors.New("invalid response from auth service")
	ErrInvalidUserID     = errors.New("invalid user ID format")
	ErrNotImplemented    = errors.New("method not implemented")
	ErrInvalidToken      = errors.New("invalid token")
	ErrEmptyUserID       = errors.New("empty user ID") // Added static error instead of dynamic one

	errUserExists       = errors.New("user already exists")
	errUserNotFound     = errors.New("user not found")
	errInvalidArgument  = errors.New("invalid argument")
	errAuthFailed       = errors.New("authentication failed")
	errPermissionDenied = errors.New("permission denied")
)

type Client struct {
	client authv1.AuthServiceClient
	conn   *grpc.ClientConn
}

func NewAuthUseCase(ctx context.Context, address string) (authAPI.UseCaseUser, error) {
	dialCtx, cancel := context.WithTimeout(ctx, defaultDialTimeout)
	defer cancel()

	// Updated to use recommended approach
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to auth service at %s: %w", address, err)
	}

	if !waitForConnection(dialCtx, conn) {
		if err := conn.Close(); err != nil {
			return nil, fmt.Errorf("failed to close connection: %w", err)
		}
		return nil, ErrConnectionTimeout
	}

	return &Client{
		client: authv1.NewAuthServiceClient(conn),
		conn:   conn,
	}, nil
}

func waitForConnection(ctx context.Context, conn *grpc.ClientConn) bool {
	for {
		state := conn.GetState()
		if state == connectivity.Ready {
			return true
		}
		if !conn.WaitForStateChange(ctx, state) {
			return false
		}
	}
}

func (c *Client) Register(ctx context.Context, login, password string) (uuid.UUID, error) {
	log := logger.ContextLogger(ctx, nil).With(
		zap.String(fieldMethod, methodRegister),
		zap.String(fieldLogin, login),
	)

	resp, err := c.client.Register(ctx, &authv1.RegisterRequest{
		Login:    login,
		Password: password,
	})
	if err != nil {
		log.Error("Failed to register user", zap.Error(err))
		return uuid.Nil, fmt.Errorf("%s: %w", errMsgRegister, mapGRPCError(err))
	}

	userID, err := parseUserID(resp.GetUserId())
	if err != nil {
		log.Error("Invalid user ID received", zap.String(fieldUserID, resp.GetUserId()), zap.Error(err))
		return uuid.Nil, ErrInvalidUserID
	}

	log.Info("User registered successfully", zap.String(fieldUserID, userID.String()))
	return userID, nil
}

func (c *Client) Login(ctx context.Context, login, password string) (*auth.TokenPair, error) {
	log := logger.ContextLogger(ctx, nil).With(
		zap.String(fieldMethod, methodLogin),
		zap.String(fieldLogin, login),
	)

	resp, err := c.client.Login(ctx, &authv1.LoginRequest{
		Login:    login,
		Password: password,
	})
	if err != nil {
		log.Error("Failed to login user", zap.Error(err))
		return nil, fmt.Errorf("%s: %w", errMsgLogin, mapGRPCError(err))
	}

	userID, err := parseUserID(resp.GetUserId())
	if err != nil {
		log.Error("Invalid user ID received", zap.String(fieldUserID, resp.GetUserId()), zap.Error(err))
		return nil, ErrInvalidUserID
	}

	var expiresAt time.Time
	if resp.ExpiresAt != nil {
		expiresAt = resp.ExpiresAt.AsTime()
	} else {
		expiresAt = time.Now().Add(defaultTokenExpiry)
	}

	tokenPair := &auth.TokenPair{
		UserID:       userID,
		AccessToken:  resp.GetAccessToken(),
		RefreshToken: resp.GetRefreshToken(),
		ExpiresAt:    expiresAt,
	}

	log.Info("User logged in successfully", zap.String(fieldUserID, userID.String()))
	return tokenPair, nil
}

func (c *Client) ValidateToken(ctx context.Context, token string) (uuid.UUID, error) {
	log := logger.ContextLogger(ctx, nil).With(zap.String(fieldMethod, methodValidateToken))

	resp, err := c.client.ValidateToken(ctx, &authv1.ValidateTokenRequest{
		Token: token,
	})
	if err != nil {
		log.Error("Failed to validate token", zap.Error(err))
		return uuid.Nil, fmt.Errorf("%s: %w", errMsgValidateToken, mapGRPCError(err))
	}

	if !resp.GetValid() {
		log.Debug("Token is not valid")
		return uuid.Nil, ErrInvalidToken
	}

	userID, err := parseUserID(resp.GetUserId())
	if err != nil {
		log.Error("Invalid user ID received", zap.String(fieldUserID, resp.GetUserId()), zap.Error(err))
		return uuid.Nil, ErrInvalidUserID
	}

	log.Debug("Token validated successfully", zap.String(fieldUserID, userID.String()))
	return userID, nil
}

func parseUserID(id string) (uuid.UUID, error) {
	if id == "" {
		return uuid.Nil, ErrEmptyUserID // Using static error instead of dynamic one
	}

	// Wrapping external error
	parsedID, err := uuid.Parse(id)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse user ID: %w", err)
	}
	return parsedID, nil
}

func (c *Client) RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenPair, error) {
	log := logger.ContextLogger(ctx, nil).With(zap.String(fieldMethod, methodRefreshToken))
	log.Warn("RefreshToken not implemented in gRPC client")
	return nil, ErrNotImplemented
}

func (c *Client) Logout(ctx context.Context, token string) error {
	log := logger.ContextLogger(ctx, nil).With(zap.String(fieldMethod, methodLogout))
	log.Warn("Logout not implemented in gRPC client")
	return ErrNotImplemented
}

func (c *Client) Close() error {
	if c.conn != nil {
		// Wrapping the external error
		if err := c.conn.Close(); err != nil {
			return fmt.Errorf("failed to close gRPC connection: %w", err)
		}
	}
	return nil
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
	case codes.AlreadyExists:
		return errUserExists
	case codes.NotFound:
		return errUserNotFound
	case codes.InvalidArgument:
		if st.Message() != "" {
			return fmt.Errorf("%w: %s", errInvalidArgument, st.Message())
		}
		return errInvalidArgument
	case codes.Unauthenticated:
		if st.Message() != "" {
			return fmt.Errorf("%w: %s", errAuthFailed, st.Message())
		}
		return errAuthFailed
	case codes.PermissionDenied:
		if st.Message() != "" {
			return fmt.Errorf("%w: %s", errPermissionDenied, st.Message())
		}
		return errPermissionDenied
	default:
		return err
	}
}
