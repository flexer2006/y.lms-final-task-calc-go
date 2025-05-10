// Package password содержит интерфейс для работы с паролями.
package password

import (
	"context"
)

// Service определяет интерфейс для работы с паролями.
type Service interface {
	// Hash хеширует пароль
	Hash(ctx context.Context, password string) (string, error)

	// Verify проверяет соответствие пароля хешу.
	Verify(ctx context.Context, password, hashedPassword string) (bool, error)

	// GenerateRandom генерирует случайный пароль.
	GenerateRandom(ctx context.Context, length int) (string, error)
}
