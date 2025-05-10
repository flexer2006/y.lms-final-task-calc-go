// Package auth содержит модели для работы с аутентификацией.
package auth

import (
	"time"

	"github.com/google/uuid"
)

// Token представляет JWT токен для аутентификации.
type Token struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	TokenStr  string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
	IsRevoked bool      `json:"is_revoked"`
}

// TokenPair содержит пару токенов доступа и обновления.
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
	UserID       uuid.UUID `json:"user_id"`
}
