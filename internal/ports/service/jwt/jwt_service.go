package jwt

import (
	"context"
	"time"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/auth"
	"github.com/google/uuid"
)

// Service определяет интерфейс для работы с JWT токенами.
type Service interface {
	// GenerateTokens генерирует пару токенов (access и refresh).
	GenerateTokens(ctx context.Context, userID uuid.UUID, login string) (*auth.TokenPair, error)

	// ValidateToken проверяет токен и возвращает ID пользователя.
	ValidateToken(ctx context.Context, token string) (uuid.UUID, error)

	// ParseToken разбирает токен без проверки подписи.
	ParseToken(ctx context.Context, token string) (map[string]interface{}, error)

	// GetTokenTTL возвращает время жизни токена доступа.
	GetTokenTTL() time.Duration

	// GetRefreshTokenTTL возвращает время жизни refresh токена.
	GetRefreshTokenTTL() time.Duration
}
