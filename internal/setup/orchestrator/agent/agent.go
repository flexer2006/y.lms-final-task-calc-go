package agent

import "time"

type Config struct {
	ComputerPower       int           `env:"COMPUTING_POWER" env-default:"4"`
	TimeAddition        time.Duration `env:"TIME_ADDITION" env-default:"1s"`
	TimeSubtraction     time.Duration `env:"TIME_SUBTRACTION" env-default:"1s"`
	TimeMultiplications time.Duration `env:"TIME_MULTIPLICATIONS" env-default:"2s"`
	TimeDivisions       time.Duration `env:"TIME_DIVISIONS" env-default:"2s"`
}
