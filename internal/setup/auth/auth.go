package auth

import (
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/auth/db"
	"github.com/flexer2006/y.lms-final-task-calc-go/internal/setup/auth/grpc"
)

type Config struct {
	Db   db.Config
	Grpc grpc.Config
}
