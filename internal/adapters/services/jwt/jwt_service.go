package jwt

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/auth"
	jwtPort "github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/service/jwt"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type TokenType string

const (
	TokenTypeAccess TokenType = "access"

	TokenTypeRefresh TokenType = "refresh"

	minSecretKeyLength = 16
)

var (
	ErrInvalidSigningMethod = errors.New("invalid signing method")
	ErrEmptyToken           = errors.New("empty token")
	ErrInvalidToken         = errors.New("invalid token format")
	ErrTokenExpired         = errors.New("token expired")
	ErrInvalidUserID        = errors.New("invalid user ID in token")
	ErrInvalidClaims        = errors.New("invalid token claims")
	ErrGeneratingToken      = errors.New("failed to generate token")
	ErrInsecureSecretKey    = errors.New("provided secret key is too short")
)

type Claims struct {
	UserID string    `json:"user_id"`
	Login  string    `json:"login,omitempty"`
	Type   TokenType `json:"type"`
	jwt.RegisteredClaims
}

type Service struct {
	secretKey       []byte
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

var _ jwtPort.Service = (*Service)(nil)

func NewService(secretKey string, accessTokenTTL, refreshTokenTTL time.Duration) *Service {
	if accessTokenTTL <= 0 {
		accessTokenTTL = 15 * time.Minute
	}
	if refreshTokenTTL <= 0 {
		refreshTokenTTL = 24 * time.Hour
	}

	return &Service{
		secretKey:       []byte(secretKey),
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (s *Service) GetTokenTTL() time.Duration {
	return s.accessTokenTTL
}

func (s *Service) GetRefreshTokenTTL() time.Duration {
	return s.refreshTokenTTL
}

func (s *Service) GenerateTokens(ctx context.Context, userID uuid.UUID, login string) (*auth.TokenPair, error) {
	const op = "JWTService.GenerateTokens"
	log := logger.ContextLogger(ctx, nil).With(zap.String("op", op), zap.String("userID", userID.String()))

	if userID == uuid.Nil {
		log.Error("Invalid user ID provided")
		return nil, fmt.Errorf("%w: user ID cannot be nil", ErrGeneratingToken)
	}

	if len(s.secretKey) == 0 {
		log.Error("Empty secret key")
		return nil, fmt.Errorf("%w: empty secret key", ErrGeneratingToken)
	}

	if len(s.secretKey) < minSecretKeyLength {
		log.Warn("Secret key is too short", zap.Int("minLength", minSecretKeyLength))
	}

	now := time.Now()
	userIDStr := userID.String()

	accessTokenString, err := s.generateToken(ctx, userIDStr, login, TokenTypeAccess, now, s.accessTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshTokenString, err := s.generateToken(ctx, userIDStr, "", TokenTypeRefresh, now, s.refreshTokenTTL)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	expiresAt := now.Add(s.accessTokenTTL)
	log.Debug("Generated token pair", zap.Time("expiresAt", expiresAt))

	return &auth.TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    expiresAt,
		UserID:       userID,
	}, nil
}

func (s *Service) generateToken(
	ctx context.Context,
	userID string,
	login string,
	tokenType TokenType,
	now time.Time,
	expiration time.Duration,
) (string, error) {
	expiresAt := now.Add(expiration)

	claims := Claims{
		UserID: userID,
		Login:  login,
		Type:   tokenType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			Subject:   userID,
			ID:        uuid.New().String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		logger.Error(ctx, nil, "Failed to sign token",
			zap.String("type", string(tokenType)),
			zap.Error(err))
		return "", fmt.Errorf("%w: %s", ErrGeneratingToken, err.Error())
	}

	return tokenString, nil
}

func (s *Service) ValidateToken(ctx context.Context, tokenString string) (uuid.UUID, error) {
	const op = "JWTService.ValidateToken"
	log := logger.ContextLogger(ctx, nil).With(zap.String("op", op))

	if tokenString == "" {
		return uuid.Nil, ErrEmptyToken
	}

	var claims Claims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("%w: expected HS256, got %v", ErrInvalidSigningMethod, token.Header["alg"])
		}
		return s.secretKey, nil
	})

	if err != nil {
		if strings.Contains(err.Error(), "token is expired") {
			return uuid.Nil, ErrTokenExpired
		}

		log.Debug("Failed to parse token",
			zap.Error(err),
			zap.String("token_prefix", tokenString[:min(10, len(tokenString))]))

		return uuid.Nil, fmt.Errorf("%w: %s", ErrInvalidToken, err.Error())
	}

	if !token.Valid {
		return uuid.Nil, ErrInvalidToken
	}

	if claims.Type != TokenTypeAccess {
		log.Debug("Invalid token type", zap.String("expected", string(TokenTypeAccess)), zap.String("got", string(claims.Type)))
		return uuid.Nil, ErrInvalidToken
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: %s", ErrInvalidUserID, err.Error())
	}

	return userID, nil
}

func (s *Service) ParseToken(ctx context.Context, tokenString string) (map[string]interface{}, error) {
	const op = "JWTService.ParseToken"
	log := logger.ContextLogger(ctx, nil).With(zap.String("op", op))

	if tokenString == "" {
		return nil, ErrEmptyToken
	}

	token, _, err := new(jwt.Parser).ParseUnverified(tokenString, &Claims{})
	if err != nil {
		log.Debug("Failed to parse token", zap.Error(err))
		return nil, fmt.Errorf("%w: %w", ErrInvalidToken, err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, ErrInvalidClaims
	}

	expectedSize := 5
	claimsMap := make(map[string]any, expectedSize)

	claimsMap["user_id"] = claims.UserID
	claimsMap["type"] = claims.Type

	if claims.Login != "" {
		claimsMap["login"] = claims.Login
	}

	if claims.ExpiresAt != nil {
		claimsMap["expires_at"] = claims.ExpiresAt.Time
	}

	if claims.IssuedAt != nil {
		claimsMap["issued_at"] = claims.IssuedAt.Time
	}

	if claims.Subject != "" {
		claimsMap["sub"] = claims.Subject
	}

	if claims.ID != "" {
		claimsMap["jti"] = claims.ID
	}

	return claimsMap, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
