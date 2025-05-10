// Package dto предоставляет DTO (Data Transfer Object) для аутентификации и авторизации пользователей.
package dto

// RegisterRequest представляет данные для регистрации пользователя.
type RegisterRequest struct {
	Login    string `json:"login" validate:"required,min=3,max=50"`
	Password string `json:"password" validate:"required,min=8"`
}

// LoginRequest представляет данные для входа пользователя.
type LoginRequest struct {
	Login    string `json:"login" validate:"required"`
	Password string `json:"password" validate:"required"`
}

// TokenValidateRequest представляет данные для проверки токена.
type TokenValidateRequest struct {
	Token string `json:"token" validate:"required"`
}

// RefreshTokenRequest представляет данные для обновления токенов.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LogoutRequest представляет данные для выхода пользователя.
type LogoutRequest struct {
	Token string `json:"token" validate:"required"`
}
