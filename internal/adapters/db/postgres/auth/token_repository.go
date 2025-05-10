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
	queryInsertToken = `
        INSERT INTO tokens (id, user_id, token, expires_at, created_at, is_revoked)
        VALUES ($1, $2, $3, $4, $5, $6)`

	queryFindTokenByString = `
        SELECT id, user_id, token, expires_at, created_at, is_revoked
        FROM tokens
        WHERE token = $1`

	queryFindTokenByID = `
        SELECT id, user_id, token, expires_at, created_at, is_revoked
        FROM tokens
        WHERE id = $1`

	queryRevokeToken = `
        UPDATE tokens
        SET is_revoked = true
        WHERE token = $1`

	queryRevokeAllUserTokens = `
        UPDATE tokens
        SET is_revoked = true
        WHERE user_id = $1 AND is_revoked = false`

	queryDeleteExpiredTokens = `
        DELETE FROM tokens
        WHERE expires_at < $1`
)

var ErrTokenNotFound = errors.New("token not found")

type PgTokenRepository struct {
	db *database.Handler
}

var _ authrepo.TokenRepository = (*PgTokenRepository)(nil)

func NewTokenRepository(db *database.Handler) *PgTokenRepository {
	return &PgTokenRepository{db: db}
}

func (r *PgTokenRepository) Store(ctx context.Context, token *authmodels.Token) error {
	const op = "PgTokenRepository.Store"

	if token.ID == uuid.Nil {
		token.ID = uuid.New()
	}

	if token.CreatedAt.IsZero() {
		token.CreatedAt = time.Now()
	}

	conn, err := r.acquireConn(ctx, op)
	if err != nil {
		return err
	}
	defer conn.Release()

	_, err = conn.Exec(ctx, queryInsertToken,
		token.ID,
		token.UserID,
		token.TokenStr,
		token.ExpiresAt,
		token.CreatedAt,
		token.IsRevoked,
	)

	if err != nil {
		return r.logError(ctx, op, "store token", err)
	}

	return nil
}

func (r *PgTokenRepository) FindByTokenString(ctx context.Context, tokenStr string) (*authmodels.Token, error) {
	const op = "PgTokenRepository.FindByTokenString"

	conn, err := r.acquireConn(ctx, op)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	var token authmodels.Token
	err = conn.QueryRow(ctx, queryFindTokenByString, tokenStr).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenStr,
		&token.ExpiresAt,
		&token.CreatedAt,
		&token.IsRevoked,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, r.logError(ctx, op, "find token by string", err)
	}

	return &token, nil
}

func (r *PgTokenRepository) FindByID(ctx context.Context, id uuid.UUID) (*authmodels.Token, error) {
	const op = "PgTokenRepository.FindByID"

	conn, err := r.acquireConn(ctx, op)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	var token authmodels.Token
	err = conn.QueryRow(ctx, queryFindTokenByID, id).Scan(
		&token.ID,
		&token.UserID,
		&token.TokenStr,
		&token.ExpiresAt,
		&token.CreatedAt,
		&token.IsRevoked,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, r.logError(ctx, op, "find token by ID", err)
	}

	return &token, nil
}

func (r *PgTokenRepository) RevokeToken(ctx context.Context, tokenStr string) error {
	const op = "PgTokenRepository.RevokeToken"

	conn, err := r.acquireConn(ctx, op)
	if err != nil {
		return err
	}
	defer conn.Release()

	result, err := conn.Exec(ctx, queryRevokeToken, tokenStr)
	if err != nil {
		return r.logError(ctx, op, "revoke token", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf("%s: %w", op, ErrTokenNotFound)
	}

	return nil
}

func (r *PgTokenRepository) RevokeAllUserTokens(ctx context.Context, userID uuid.UUID) error {
	const op = "PgTokenRepository.RevokeAllUserTokens"

	conn, err := r.acquireConn(ctx, op)
	if err != nil {
		return err
	}
	defer conn.Release()

	result, err := conn.Exec(ctx, queryRevokeAllUserTokens, userID)
	if err != nil {
		return r.logError(ctx, op, "revoke all user tokens", err)
	}

	logger.Info(ctx, nil, "User tokens revoked",
		zap.String("op", op),
		zap.String("userID", userID.String()),
		zap.Int64("count", result.RowsAffected()))

	return nil
}

func (r *PgTokenRepository) DeleteExpiredTokens(ctx context.Context, before time.Time) error {
	const op = "PgTokenRepository.DeleteExpiredTokens"

	conn, err := r.acquireConn(ctx, op)
	if err != nil {
		return err
	}
	defer conn.Release()

	result, err := conn.Exec(ctx, queryDeleteExpiredTokens, before)
	if err != nil {
		return r.logError(ctx, op, "delete expired tokens", err)
	}

	logger.Info(ctx, nil, "Expired tokens deleted",
		zap.String("op", op),
		zap.Time("before", before),
		zap.Int64("count", result.RowsAffected()))

	return nil
}

func (r *PgTokenRepository) acquireConn(ctx context.Context, op string) (*pgxpool.Conn, error) {
	conn, err := r.db.AcquireConn(ctx)
	if err != nil {
		logger.Error(ctx, nil, "Failed to acquire connection", zap.String("op", op), zap.Error(err))
		return nil, fmt.Errorf("%s: acquire connection: %w", op, err)
	}
	return conn, nil
}

func (r *PgTokenRepository) logError(ctx context.Context, op, action string, err error) error {
	logger.Error(ctx, nil, "Failed to "+action, zap.String("op", op), zap.Error(err))
	return fmt.Errorf("%s: %s: %w", op, action, err)
}
