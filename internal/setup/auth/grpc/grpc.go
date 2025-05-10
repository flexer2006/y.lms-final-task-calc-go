// Package grpc содержит конфигурацию для gRPC.
package grpc

// Config содержит конфигурацию для gRPC.
type Config struct {
	Host string `yaml:"host" env:"AUTH_GRPC_HOST" env-default:"0.0.0.0"`
	Port int    `yaml:"port" env:"AUTH_GRPC_PORT" env-default:"50052"`
}
