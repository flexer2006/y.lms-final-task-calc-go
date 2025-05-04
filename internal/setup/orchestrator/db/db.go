package db

import (
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/orchestrator/db/pgxx"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/orchestrator/db/postgres"
)

type Config struct {
	Postgres postgres.Config
	Pgx      pgxx.Config
}
