// Package shutdown содержит конфигурацию для завершения работы приложения.
package shutdown

import "time"

// Config содержит конфигурацию для завершения работы приложения.
type Config struct {
	ShutdownTimeout time.Duration `env:"GRACEFUL_SHUTDOWN_TIMEOUT" env-default:"5s"`
}
