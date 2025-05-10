// Package jwt содержит конфигурацию для JWT.
package jwt

import "time"

// Config содержит конфигурацию для JWT.
type Config struct {
	SecretKey       string        `yaml:"secret_key" env:"JWT_SECRET_KEY" env-default:"2hlsdwbzmv7yGxbQ4sIah/MuvvNoe889pbEzZql0SU8n3U1gYi29gZnFQKxiUdGH"`
	AccessTokenTTL  time.Duration `yaml:"access_token_ttl" env:"JWT_ACCESS_TOKEN_TTL" env-default:"15m"`
	RefreshTokenTTL time.Duration `yaml:"refresh_token_ttl" env:"JWT_REFRESH_TOKEN_TTL" env-default:"24h"`
	BCryptCost      int           `yaml:"bcrypt_cost" env:"JWT_BCRYPT_COST" env-default:"10"`
}
