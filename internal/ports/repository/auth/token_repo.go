package auth

import (
	"context"
	"time"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/auth"
	"github.com/google/uuid"
)

// TokenRepository определяет интерфейс для работы с хранилищем токенов.
type TokenRepository interface {
	// Store сохраняет токен.
	Store(ctx context.Context, token *auth.Token) error

	// FindByTokenString находит токен по его значению.
	FindByTokenString(ctx context.Context, tokenStr string) (*auth.Token, error)

	// FindByID находит токен по его ID.
	FindByID(ctx context.Context, id uuid.UUID) (*auth.Token, error)

	// RevokeToken аннулирует токен.
	RevokeToken(ctx context.Context, tokenStr string) error

	// RevokeAllUserTokens аннулирует все токены пользователя.
	RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error

	// DeleteExpiredTokens удаляет просроченные токены.
	DeleteExpiredTokens(ctx context.Context, before time.Time) error
}
