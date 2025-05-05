package grpc

type Config struct {
	Host string `yaml:"host" env:"ORCHESTRATOR_GRPC_HOST" env-default:"0.0.0.0"`
	Port int    `yaml:"port" env:"ORCHESTRATOR_GRPC_PORT" env-default:"50053"`
}
