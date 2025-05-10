// Package auth содержит интерфейс для работы с аутентификацией.
package auth

import (
	"context"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/auth"
	"github.com/google/uuid"
)

// UseCaseUser определяет основной порт для операций аутентификации.
type UseCaseUser interface {
	// Register регистрирует нового пользователя.
	Register(ctx context.Context, login, password string) (uuid.UUID, error)

	// Login выполняет вход пользователя и возвращает токены.
	Login(ctx context.Context, login, password string) (*auth.TokenPair, error)

	// ValidateToken проверяет валидность токена и возвращает ID пользователя.
	ValidateToken(ctx context.Context, token string) (uuid.UUID, error)

	// RefreshToken обновляет пару токенов.
	RefreshToken(ctx context.Context, refreshToken string) (*auth.TokenPair, error)

	// Logout завершает сессию пользователя, аннулируя токен.
	Logout(ctx context.Context, token string) error

	// Close closes any resources used by this interface implementation
	Close() error
}
