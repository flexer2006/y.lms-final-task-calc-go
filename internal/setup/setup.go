package setup

import (
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/auth"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/logger"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/orchestrator"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/server"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/shutdown"
)

type Config struct {
	Logger           logger.Config
	Auth             auth.Config
	Orchestrator     orchestrator.Config
	GracefulShutdown shutdown.Config
	Server           server.Config
}
