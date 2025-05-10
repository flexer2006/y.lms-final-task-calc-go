package handlers

import (
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/adapters/servers/http/handlers/auth"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/adapters/servers/http/handlers/orchestrator"
	authAPI "github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/api/auth"
	orchAPI "github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/api/orchestrator"
)

type Handlers struct {
	Auth         *auth.Handler
	Orchestrator *orchestrator.Handler
}

func NewHandlers(
	authUseCase authAPI.UseCaseUser,
	calcUseCase orchAPI.UseCaseCalculation,
) *Handlers {
	return &Handlers{
		Auth:         auth.NewHandler(authUseCase),
		Orchestrator: orchestrator.NewHandler(calcUseCase),
	}
}
