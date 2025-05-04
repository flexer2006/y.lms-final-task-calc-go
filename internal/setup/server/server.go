package server

import "time"

type Config struct {
	Host         string        `env:"HTTP_HOST" env-default:"0.0.0.0"`
	Port         int           `env:"HTTP_PORT" env-default:"8080"`
	ReadTimeout  time.Duration `env:"HTTP_READ_TIMEOUT" env-default:"5s"`
	WriteTimeout time.Duration `env:"HTTP_WRITE_TIMEOUT" env-default:"10s"`
}
