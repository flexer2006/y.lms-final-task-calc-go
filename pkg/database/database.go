// Пакет database предоставляет унифицированный доступ к функциональности базы данных,
// объединяя управление соединениями и миграции.
package database

import (
	"context"
	"fmt"

	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/database/migrate"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/database/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Реэкспорт типов из подпакетов для удобства использования.
type (
	// PostgresConfig содержит конфигурацию для подключения к PostgreSQL.
	PostgresConfig = postgres.PostgresConfig
	// MigrateConfig содержит конфигурацию для миграций базы данных.
	MigrateConfig = migrate.Config
	// Database представляет соединение с базой данных PostgreSQL.
	Database = postgres.Database
	// Migrator представляет сервис миграции базы данных.
	Migrator = migrate.Migrator
)

// Реэкспорт ошибок из пакета postgres.
var (
	// Ошибки валидации конфигурации базы данных.
	ErrHostRequired      = postgres.ErrHostRequired
	ErrInvalidPort       = postgres.ErrInvalidPort
	ErrUserRequired      = postgres.ErrUserRequired
	ErrDatabaseRequired  = postgres.ErrDatabaseRequired
	ErrConnectionPoolNil = postgres.ErrConnectionPoolNil
)

// Реэкспорт ошибок из пакета migrate.
var (
	// Ошибки, связанные с миграциями.
	ErrMigrationPathNotSpecified = migrate.ErrMigrationPathNotSpecified
	ErrMigratorCreation          = migrate.ErrMigratorCreation
	ErrApplyMigrations           = migrate.ErrApplyMigrations
	ErrRollbackMigrations        = migrate.ErrRollbackMigrations
	ErrMigrateToVersion          = migrate.ErrMigrateToVersion
	ErrGetVersion                = migrate.ErrGetVersion
	ErrForceVersion              = migrate.ErrForceVersion
	ErrStepMigrations            = migrate.ErrStepMigrations
	ErrCloseMigrator             = migrate.ErrCloseMigrator
)

// NewPostgres создает новое соединение с базой данных PostgreSQL.
func NewPostgres(ctx context.Context, config PostgresConfig) (*Database, error) {
	db, err := postgres.New(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("creating PostgreSQL connection: %w", err)
	}
	return db, nil
}

// NewPostgresWithDSN создает новое соединение с базой данных PostgreSQL, используя DSN.
func NewPostgresWithDSN(ctx context.Context, dsn string, minConn, maxConn int) (*Database, error) {
	db, err := postgres.NewWithDSN(ctx, dsn, minConn, maxConn)
	if err != nil {
		return nil, fmt.Errorf("creating PostgreSQL connection using DSN: %w", err)
	}
	return db, nil
}

// NewMigrator создает новый экземпляр мигратора базы данных.
func NewMigrator() *Migrator {
	return migrate.NewMigrator()
}

// Handler объединяет возможности соединения с базой данных и миграции в одной структуре.
type Handler struct {
	DB       *Database
	Migrator *Migrator
}

// NewHandler создает новый обработчик базы данных с соединением и мигратором.
func NewHandler(ctx context.Context, dbConfig PostgresConfig, migrateConfig MigrateConfig) (*Handler, error) {
	db, err := NewPostgres(ctx, dbConfig)
	if err != nil {
		return nil, err
	}

	migrator := NewMigrator()

	return &Handler{
		DB:       db,
		Migrator: migrator,
	}, nil
}

// NewHandlerWithDSN создает новый обработчик базы данных, используя строку DSN.
func NewHandlerWithDSN(ctx context.Context, dsn string, minConn, maxConn int) (*Handler, error) {
	db, err := NewPostgresWithDSN(ctx, dsn, minConn, maxConn)
	if err != nil {
		return nil, err
	}

	migrator := NewMigrator()

	return &Handler{
		DB:       db,
		Migrator: migrator,
	}, nil
}

// MigrateUp применяет все доступные миграции.
func (h *Handler) MigrateUp(ctx context.Context, migrateConfig MigrateConfig) error {
	dsn := h.DB.GetDSN()
	if err := h.Migrator.Up(ctx, dsn, migrateConfig); err != nil {
		return fmt.Errorf("applying migrations: %w", err)
	}
	return nil
}

// MigrateDown откатывает все миграции.
func (h *Handler) MigrateDown(ctx context.Context, migrateConfig MigrateConfig) error {
	dsn := h.DB.GetDSN()
	if err := h.Migrator.Down(ctx, dsn, migrateConfig); err != nil {
		return fmt.Errorf("rolling back migrations: %w", err)
	}
	return nil
}

// MigrateToVersion выполняет миграцию до определенной версии.
func (h *Handler) MigrateToVersion(ctx context.Context, version uint, migrateConfig MigrateConfig) error {
	dsn := h.DB.GetDSN()
	if err := h.Migrator.Migrate(ctx, dsn, version, migrateConfig); err != nil {
		return fmt.Errorf("migrating to version %d: %w", version, err)
	}
	return nil
}

// MigrateSteps выполняет определенное количество миграций.
func (h *Handler) MigrateSteps(ctx context.Context, steps int, migrateConfig MigrateConfig) error {
	dsn := h.DB.GetDSN()
	if err := h.Migrator.Steps(ctx, dsn, steps, migrateConfig); err != nil {
		return fmt.Errorf("executing %d migration steps: %w", steps, err)
	}
	return nil
}

// MigrateForce принудительно устанавливает определенную версию миграции.
func (h *Handler) MigrateForce(ctx context.Context, version int, migrateConfig MigrateConfig) error {
	dsn := h.DB.GetDSN()
	if err := h.Migrator.Force(ctx, dsn, version, migrateConfig); err != nil {
		return fmt.Errorf("forcing version %d: %w", version, err)
	}
	return nil
}

// GetMigrationVersion возвращает текущую версию миграции и состояние "грязный".
func (h *Handler) GetMigrationVersion(ctx context.Context, migrateConfig MigrateConfig) (uint, bool, error) {
	dsn := h.DB.GetDSN()
	version, dirty, err := h.Migrator.Version(ctx, dsn, migrateConfig)
	if err != nil {
		return 0, false, fmt.Errorf("getting migration version: %w", err)
	}
	return version, dirty, nil
}

// Close закрывает соединение с базой данных.
func (h *Handler) Close(ctx context.Context) {
	h.DB.Close(ctx)
}

// AcquireConn получает соединение из пула.
func (h *Handler) AcquireConn(ctx context.Context) (*pgxpool.Conn, error) {
	conn, err := h.DB.AcquireConn(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquiring connection from pool: %w", err)
	}
	return conn, nil
}

// Pool возвращает базовый пул соединений.
func (h *Handler) Pool() *pgxpool.Pool {
	return h.DB.Pool()
}

// Ping проверяет доступность базы данных.
func (h *Handler) Ping(ctx context.Context) error {
	if err := h.DB.Ping(ctx); err != nil {
		return fmt.Errorf("checking database availability: %w", err)
	}
	return nil
}

// InitializeWithMigrations создает соединение с базой данных и применяет все миграции.
func InitializeWithMigrations(ctx context.Context, dbConfig PostgresConfig, migrateConfig MigrateConfig) (*Handler, error) {
	handler, err := NewHandler(ctx, dbConfig, migrateConfig)
	if err != nil {
		return nil, err
	}

	if err := handler.MigrateUp(ctx, migrateConfig); err != nil {
		handler.Close(ctx)
		return nil, err
	}

	return handler, nil
}

// InitializeWithMigrationsFromDSN создает соединение с базой данных из DSN и применяет все миграции.
func InitializeWithMigrationsFromDSN(ctx context.Context, dsn string, minConn, maxConn int, migrateConfig MigrateConfig) (*Handler, error) {
	handler, err := NewHandlerWithDSN(ctx, dsn, minConn, maxConn)
	if err != nil {
		return nil, err
	}

	if err := handler.MigrateUp(ctx, migrateConfig); err != nil {
		handler.Close(ctx)
		return nil, err
	}

	return handler, nil
}
