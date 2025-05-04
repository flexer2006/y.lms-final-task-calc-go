package logger

type Config struct {
	Level        string `env:"LOGGER_LEVEL" env-default:"info"`
	Format       string `env:"LOGGER_FORMAT" env-default:"json"`
	Output       string `env:"LOGGER_OUTPUT" env-default:"stdout"`
	TimeEncoding string `env:"LOGGER_TIME_FORMAT" env-default:"iso8601"`
	Caller       bool   `env:"LOGGER_CALLER" env-default:"true"`
	Stacktrace   bool   `env:"LOGGER_STACKTRACE" env-default:"true"`
	Model        string `env:"LOGGER_MODEL" env-default:"development"`
}
