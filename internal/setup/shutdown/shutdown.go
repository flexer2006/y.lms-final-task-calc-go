package shutdown

import "time"

type Config struct {
	ShutdownTimeout time.Duration `env:"GRACEFUL_SHUTDOWN_TIMEOUT" env-default:"5s"`
}
