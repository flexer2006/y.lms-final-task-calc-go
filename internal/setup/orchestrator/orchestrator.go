package orchestrator

import (
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/orchestrator/db"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/orchestrator/grpc"
)

type Config struct {
	Db   db.Config
	Grpc grpc.Config
}
