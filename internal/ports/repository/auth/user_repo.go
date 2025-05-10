// Package auth содержит интерфейс для работы с хранилищем пользователей.
package auth

import (
	"context"

	"github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/auth"
	"github.com/google/uuid"
)

// UserRepository определяет интерфейс для работы с хранилищем пользователей.
type UserRepository interface {
	// Create создаёт нового пользователя.
	Create(ctx context.Context, user *auth.User) (*auth.User, error)

	// FindByID находит пользователя по ID.
	FindByID(ctx context.Context, id uuid.UUID) (*auth.User, error)

	// FindByLogin находит пользователя по логину.
	FindByLogin(ctx context.Context, login string) (*auth.User, error)

	// Update обновляет данные пользователя.
	Update(ctx context.Context, user *auth.User) error

	// Delete удаляет пользователя.
	Delete(ctx context.Context, id uuid.UUID) error
}
