// Package auth содержит модели для работы с аутентификацией.
package auth

import (
	"time"

	"github.com/google/uuid"
)

// User представляет пользователя системы.
type User struct {
	ID           uuid.UUID `json:"id"`
	Login        string    `json:"login"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}
