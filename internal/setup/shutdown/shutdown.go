package shutdown

type Config struct {
	ShutdownTimeout string `env:"GRACEFUL_SHUTDOWN_TIMEOUT" env-default:"5s"`
}
