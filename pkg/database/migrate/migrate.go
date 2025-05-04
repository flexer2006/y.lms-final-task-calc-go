// Package migrate предоставляет функциональность для миграции базы данных.
package migrate

import (
	"context"
	"errors"
	"fmt"

	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/golang-migrate/migrate/v4"

	// Импортируем драйвер для работы с Postgres.
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	// Импортируем драйвер для работы с PGX v5.
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	// Импортируем драйвер для чтения миграций из файлов.
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"go.uber.org/zap"
)

// Определение ошибок пакета.
var (
	ErrMigrationPathNotSpecified = errors.New("migration path not specified")
	ErrMigratorCreation          = errors.New("failed to create migrator")
	ErrApplyMigrations           = errors.New("failed to apply migrations")
	ErrRollbackMigrations        = errors.New("failed to rollback migrations")
	ErrMigrateToVersion          = errors.New("failed to migrate to version")
	ErrGetVersion                = errors.New("failed to get migration version")
	ErrForceVersion              = errors.New("failed to force migration version")
	ErrStepMigrations            = errors.New("failed to perform step migrations")
	ErrCloseMigrator             = errors.New("failed to close migrator")
)

// Config содержит настройки для миграций.
type Config struct {
	// Path - путь к директории с файлами миграций.
	Path string
}

// Migrator представляет сервис для выполнения миграций базы данных.
type Migrator struct{}

// NewMigrator создает новый сервис миграций.
func NewMigrator() *Migrator {
	return &Migrator{}
}

// migrateOperation представляет операцию миграции, которую можно выполнить.
type migrateOperation func(*migrate.Migrate) error

// executeOperation выполняет операцию миграции с общей логикой подготовки и логирования.
func (m *Migrator) executeOperation(
	ctx context.Context,
	dsn string,
	cfg Config,
	operation migrateOperation,
	errType error,
	successMsg string,
	fields ...zap.Field,
) error {
	if cfg.Path == "" {
		return ErrMigrationPathNotSpecified
	}

	path := fmt.Sprintf("file://%s", cfg.Path)
	migrator, err := m.createMigrator(ctx, dsn, path)
	if err != nil {
		return err
	}
	defer m.closeMigrator(ctx, migrator)

	if err := operation(migrator); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		logger.Error(ctx, nil, "migration operation failed",
			zap.Error(err),
			zap.String("path", cfg.Path))

		for _, field := range fields {
			logger.Error(ctx, nil, "", field)
		}

		return fmt.Errorf("%w: %w", errType, err)
	}

	logger.Info(ctx, nil, successMsg, zap.String("path", cfg.Path))

	for _, field := range fields {
		logger.Info(ctx, nil, "", field)
	}

	return nil
}

// Up выполняет все доступные миграции.
func (m *Migrator) Up(ctx context.Context, dsn string, cfg Config) error {
	return m.executeOperation(
		ctx,
		dsn,
		cfg,
		func(migrator *migrate.Migrate) error {
			return migrator.Up()
		},
		ErrApplyMigrations,
		"database migrations applied successfully",
	)
}

// Down откатывает все миграции.
func (m *Migrator) Down(ctx context.Context, dsn string, cfg Config) error {
	return m.executeOperation(
		ctx,
		dsn,
		cfg,
		func(migrator *migrate.Migrate) error {
			return migrator.Down()
		},
		ErrRollbackMigrations,
		"database migrations rolled back successfully",
	)
}

// Migrate выполняет миграцию к определенной версии.
// Положительное значение означает миграцию вверх до указанной версии.
// Отрицательное значение означает миграцию вниз до указанной версии.
func (m *Migrator) Migrate(ctx context.Context, dsn string, version uint, cfg Config) error {
	return m.executeOperation(
		ctx,
		dsn,
		cfg,
		func(migrator *migrate.Migrate) error {
			return migrator.Migrate(version)
		},
		ErrMigrateToVersion,
		"database migrated to specific version successfully",
		zap.Uint("version", version),
	)
}

// Steps выполняет указанное количество миграций.
// Положительное значение применяет n миграций вверх.
// Отрицательное значение откатывает n миграций вниз.
func (m *Migrator) Steps(ctx context.Context, dsn string, n int, cfg Config) error {
	return m.executeOperation(
		ctx,
		dsn,
		cfg,
		func(migrator *migrate.Migrate) error {
			return migrator.Steps(n)
		},
		ErrStepMigrations,
		"database step migrations performed successfully",
		zap.Int("steps", n),
	)
}

// Version возвращает текущую версию миграции и статус "грязный".
// Грязная миграция означает, что миграция была прервана и база данных может быть в неконсистентном состоянии.
func (m *Migrator) Version(ctx context.Context, dsn string, cfg Config) (uint, bool, error) {
	if cfg.Path == "" {
		return 0, false, ErrMigrationPathNotSpecified
	}

	path := fmt.Sprintf("file://%s", cfg.Path)
	migrator, err := m.createMigrator(ctx, dsn, path)
	if err != nil {
		return 0, false, err
	}
	defer m.closeMigrator(ctx, migrator)

	version, dirty, err := migrator.Version()
	if err != nil {
		if errors.Is(err, migrate.ErrNilVersion) {
			logger.Debug(ctx, nil, "no migrations found",
				zap.String("path", cfg.Path),
			)
			return 0, false, nil
		}

		logger.Error(ctx, nil, "failed to get migration version",
			zap.Error(err),
			zap.String("path", cfg.Path),
		)
		return 0, false, fmt.Errorf("%w: %w", ErrGetVersion, err)
	}

	logger.Debug(ctx, nil, "current migration version",
		zap.Uint("version", version),
		zap.Bool("dirty", dirty),
		zap.String("path", cfg.Path),
	)
	return version, dirty, nil
}

// Force устанавливает определенную версию миграции принудительно,
// не выполняя никаких миграционных файлов.
func (m *Migrator) Force(ctx context.Context, dsn string, version int, cfg Config) error {
	return m.executeOperation(
		ctx,
		dsn,
		cfg,
		func(migrator *migrate.Migrate) error {
			return migrator.Force(version)
		},
		ErrForceVersion,
		"forced migration version successfully",
		zap.Int("version", version),
	)
}

// createMigrator создает новый экземпляр мигратора.
func (m *Migrator) createMigrator(ctx context.Context, dsn string, path string) (*migrate.Migrate, error) {
	migrator, err := migrate.New(path, dsn)
	if err != nil {
		logger.Error(ctx, nil, "failed to create migration instance",
			zap.Error(err),
			zap.String("path", path),
		)
		return nil, fmt.Errorf("%w: %w", ErrMigratorCreation, err)
	}

	migrator.Log = &migrationLogger{ctx: ctx}

	return migrator, nil
}

// closeMigrator безопасно закрывает мигратор.
func (m *Migrator) closeMigrator(ctx context.Context, migrator *migrate.Migrate) {
	sourceErr, dbErr := migrator.Close()
	if sourceErr != nil {
		logger.Error(ctx, nil, "failed to close migration source",
			zap.Error(sourceErr),
		)
	}
	if dbErr != nil {
		logger.Error(ctx, nil, "failed to close migration database",
			zap.Error(dbErr),
		)
	}
}

// migrationLogger реализует интерфейс логгера для golang-migrate.
type migrationLogger struct {
	ctx context.Context
}

// Printf записывает логи миграций через наш логгер.
func (m *migrationLogger) Printf(format string, v ...any) {
	msg := fmt.Sprintf(format, v...)
	logger.Debug(m.ctx, nil, msg)
}

// Verbose всегда возвращает true, чтобы видеть подробные логи миграций.
func (m *migrationLogger) Verbose() bool {
	return true
}
