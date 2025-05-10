package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	authmodels "github.com/flexer2006/y.lms-final-task-calc-go/internal/domain/models/auth"
	authrepo "github.com/flexer2006/y.lms-final-task-calc-go/internal/ports/repository/auth"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/database"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

const (
	queryInsertUser = `
        INSERT INTO users (id, login, password_hash, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id, login, password_hash, created_at, updated_at`

	queryFindUserByID = `
        SELECT id, login, password_hash, created_at, updated_at
        FROM users
        WHERE id = $1`

	queryFindUserByLogin = `
        SELECT id, login, password_hash, created_at, updated_at
        FROM users
        WHERE login = $1`

	queryUpdateUser = `
        UPDATE users
        SET login = $2, password_hash = $3, updated_at = $4
        WHERE id = $1`

	queryDeleteUser = `
        DELETE FROM users
        WHERE id = $1`
)

var (
	ErrInvalidUserID = errors.New("invalid user ID")
	ErrEmptyLogin    = errors.New("empty login provided")
	ErrInvalidUser   = errors.New("invalid user or ID")
	ErrUserNotFound  = errors.New("user not found")
)

type PgUserRepository struct {
	db *database.Handler
}

var _ authrepo.UserRepository = (*PgUserRepository)(nil)

func NewUserRepository(db *database.Handler) *PgUserRepository {
	return &PgUserRepository{db: db}
}

func (r *PgUserRepository) Create(ctx context.Context, user *authmodels.User) (*authmodels.User, error) {
	const op = "PgUserRepository.Create"

	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}

	now := time.Now()
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = now
	}

	conn, err := r.acquireConn(ctx, op)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	var createdUser authmodels.User
	err = conn.QueryRow(ctx, queryInsertUser,
		user.ID,
		user.Login,
		user.PasswordHash,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(
		&createdUser.ID,
		&createdUser.Login,
		&createdUser.PasswordHash,
		&createdUser.CreatedAt,
		&createdUser.UpdatedAt,
	)

	if err != nil {
		return nil, r.logError(ctx, op, "create user", err)
	}

	return &createdUser, nil
}

func (r *PgUserRepository) FindByID(ctx context.Context, id uuid.UUID) (*authmodels.User, error) {
	const op = "PgUserRepository.FindByID"

	if id == uuid.Nil {
		return nil, fmt.Errorf("%s: %w", op, ErrInvalidUserID)
	}

	return r.findUserByQuery(ctx, op, queryFindUserByID, id)
}

func (r *PgUserRepository) FindByLogin(ctx context.Context, login string) (*authmodels.User, error) {
	const op = "PgUserRepository.FindByLogin"

	if login == "" {
		return nil, fmt.Errorf("%s: %w", op, ErrEmptyLogin)
	}

	return r.findUserByQuery(ctx, op, queryFindUserByLogin, login)
}

func (r *PgUserRepository) Update(ctx context.Context, user *authmodels.User) error {
	const op = "PgUserRepository.Update"

	if user == nil || user.ID == uuid.Nil {
		return fmt.Errorf("%s: %w", op, ErrInvalidUser)
	}

	user.UpdatedAt = time.Now()

	conn, err := r.acquireConn(ctx, op)
	if err != nil {
		return err
	}
	defer conn.Release()

	result, err := conn.Exec(ctx, queryUpdateUser,
		user.ID,
		user.Login,
		user.PasswordHash,
		user.UpdatedAt,
	)

	if err != nil {
		return r.logError(ctx, op, "update user", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w: user with ID %s", op, ErrUserNotFound, user.ID)
	}

	return nil
}

func (r *PgUserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const op = "PgUserRepository.Delete"

	if id == uuid.Nil {
		return fmt.Errorf("%s: %w", op, ErrInvalidUserID)
	}

	conn, err := r.acquireConn(ctx, op)
	if err != nil {
		return err
	}
	defer conn.Release()

	result, err := conn.Exec(ctx, queryDeleteUser, id)
	if err != nil {
		return r.logError(ctx, op, "delete user", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w: user with ID %s", op, ErrUserNotFound, id)
	}

	return nil
}

func (r *PgUserRepository) acquireConn(ctx context.Context, op string) (*pgxpool.Conn, error) {
	conn, err := r.db.AcquireConn(ctx)
	if err != nil {
		logger.Error(ctx, nil, "Failed to acquire connection", zap.String("op", op), zap.Error(err))
		return nil, fmt.Errorf("%s: acquire connection: %w", op, err)
	}
	return conn, nil
}

func (r *PgUserRepository) logError(ctx context.Context, op, action string, err error) error {
	logger.Error(ctx, nil, "Failed to "+action, zap.String("op", op), zap.Error(err))
	return fmt.Errorf("%s: %s: %w", op, action, err)
}

func (r *PgUserRepository) findUserByQuery(ctx context.Context, op, query string, arg interface{}) (*authmodels.User, error) {
	conn, err := r.acquireConn(ctx, op)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	var user authmodels.User
	err = conn.QueryRow(ctx, query, arg).Scan(
		&user.ID,
		&user.Login,
		&user.PasswordHash,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, r.logError(ctx, op, "find user", err)
	}

	return &user, nil
}
