package database_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/database"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/database/migrate"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/database/postgres"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	errMigration         = errors.New("migration error")
	errRollback          = errors.New("rollback error")
	errMigrateToVersion  = errors.New("migrate to version error")
	errMigrateSteps      = errors.New("steps error")
	errMigrateForce      = errors.New("force error")
	errMigrateVersion    = errors.New("version error")
	errAcquireConnection = errors.New("acquire error")
	errPingDatabase      = errors.New("ping error")
)

func TestErrorReExports(t *testing.T) {
	tests := []struct {
		name          string
		packageError  error
		reexportError error
	}{

		{"ErrHostRequired", postgres.ErrHostRequired, database.ErrHostRequired},
		{"ErrInvalidPort", postgres.ErrInvalidPort, database.ErrInvalidPort},
		{"ErrUserRequired", postgres.ErrUserRequired, database.ErrUserRequired},
		{"ErrDatabaseRequired", postgres.ErrDatabaseRequired, database.ErrDatabaseRequired},
		{"ErrConnectionPoolNil", postgres.ErrConnectionPoolNil, database.ErrConnectionPoolNil},

		{"ErrMigrationPathNotSpecified", migrate.ErrMigrationPathNotSpecified, database.ErrMigrationPathNotSpecified},
		{"ErrMigratorCreation", migrate.ErrMigratorCreation, database.ErrMigratorCreation},
		{"ErrApplyMigrations", migrate.ErrApplyMigrations, database.ErrApplyMigrations},
		{"ErrRollbackMigrations", migrate.ErrRollbackMigrations, database.ErrRollbackMigrations},
		{"ErrMigrateToVersion", migrate.ErrMigrateToVersion, database.ErrMigrateToVersion},
		{"ErrGetVersion", migrate.ErrGetVersion, database.ErrGetVersion},
		{"ErrForceVersion", migrate.ErrForceVersion, database.ErrForceVersion},
		{"ErrStepMigrations", migrate.ErrStepMigrations, database.ErrStepMigrations},
		{"ErrCloseMigrator", migrate.ErrCloseMigrator, database.ErrCloseMigrator},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.packageError, tc.reexportError, "Re-exported error should match the original")
		})
	}
}

func TestNewMigrator(t *testing.T) {
	migrator := database.NewMigrator()
	assert.NotNil(t, migrator, "NewMigrator should return a non-nil instance")
}

func TestErrorHandling(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name        string
		testFunc    func(context.Context) error
		errorPrefix string
	}{
		{
			name: "NewPostgres with invalid config",
			testFunc: func(ctx context.Context) error {
				invalidConfig := database.PostgresConfig{
					Port:     5432,
					User:     "user",
					Database: "db",
				}

				_, err := database.NewPostgres(ctx, invalidConfig)
				return fmt.Errorf("test wrapper: %w", err)
			},
			errorPrefix: "creating PostgreSQL connection",
		},
		{
			name: "NewPostgresWithDSN with invalid DSN",
			testFunc: func(ctx context.Context) error {
				invalidDSN := "not-a-valid-dsn"

				_, err := database.NewPostgresWithDSN(ctx, invalidDSN, 1, 5)
				return fmt.Errorf("test wrapper: %w", err)
			},
			errorPrefix: "creating PostgreSQL connection using DSN",
		},
		{
			name: "NewHandler with invalid config",
			testFunc: func(ctx context.Context) error {
				invalidConfig := database.PostgresConfig{
					Port:     5432,
					User:     "user",
					Database: "db",
				}
				migrateConfig := database.MigrateConfig{
					Path: "/some/path",
				}

				_, err := database.NewHandler(ctx, invalidConfig, migrateConfig)
				return fmt.Errorf("test wrapper: %w", err)
			},
			errorPrefix: "creating PostgreSQL connection",
		},
		{
			name: "NewHandlerWithDSN with invalid DSN",
			testFunc: func(ctx context.Context) error {
				invalidDSN := "not-a-valid-dsn"

				_, err := database.NewHandlerWithDSN(ctx, invalidDSN, 1, 5)
				return fmt.Errorf("test wrapper: %w", err)
			},
			errorPrefix: "creating PostgreSQL connection using DSN",
		},
		{
			name: "InitializeWithMigrations with invalid config",
			testFunc: func(ctx context.Context) error {
				invalidConfig := database.PostgresConfig{
					Port:     5432,
					User:     "user",
					Database: "db",
				}
				migrateConfig := database.MigrateConfig{
					Path: "/some/path",
				}

				_, err := database.InitializeWithMigrations(ctx, invalidConfig, migrateConfig)
				return fmt.Errorf("test wrapper: %w", err)
			},
			errorPrefix: "creating PostgreSQL connection",
		},
		{
			name: "InitializeWithMigrationsFromDSN with invalid DSN",
			testFunc: func(ctx context.Context) error {
				invalidDSN := "not-a-valid-dsn"
				migrateConfig := database.MigrateConfig{
					Path: "/some/path",
				}

				_, err := database.InitializeWithMigrationsFromDSN(ctx, invalidDSN, 1, 5, migrateConfig)
				return fmt.Errorf("test wrapper: %w", err)
			},
			errorPrefix: "creating PostgreSQL connection using DSN",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.testFunc(ctx)
			require.Error(t, err, "Function should return an error")
			assert.Contains(t, err.Error(), tc.errorPrefix, "Error should contain the expected prefix")
		})
	}
}

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      database.PostgresConfig
		expectedErr error
	}{
		{
			name: "Valid config",
			config: database.PostgresConfig{
				Host:     "localhost",
				Port:     5432,
				User:     "postgres",
				Database: "testdb",
			},
			expectedErr: nil,
		},
		{
			name: "Missing host",
			config: database.PostgresConfig{
				Port:     5432,
				User:     "postgres",
				Database: "testdb",
			},
			expectedErr: database.ErrHostRequired,
		},
		{
			name: "Invalid port - too low",
			config: database.PostgresConfig{
				Host:     "localhost",
				Port:     0,
				User:     "postgres",
				Database: "testdb",
			},
			expectedErr: database.ErrInvalidPort,
		},
		{
			name: "Missing user",
			config: database.PostgresConfig{
				Host:     "localhost",
				Port:     5432,
				Database: "testdb",
			},
			expectedErr: database.ErrUserRequired,
		},
		{
			name: "Missing database",
			config: database.PostgresConfig{
				Host: "localhost",
				Port: 5432,
				User: "postgres",
			},
			expectedErr: database.ErrDatabaseRequired,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.config.Validate()
			if tc.expectedErr != nil {
				assert.ErrorIs(t, err, tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

type MockDB struct {
	dsn         string
	closeCalled bool
	pingErr     error
	acquireErr  error
	pool        *pgxpool.Pool
}

func (m *MockDB) Close(ctx context.Context) {
	m.closeCalled = true
}

func (m *MockDB) GetDSN() string {
	return m.dsn
}

func (m *MockDB) Ping(ctx context.Context) error {
	return m.pingErr
}

func (m *MockDB) AcquireConn(ctx context.Context) (*pgxpool.Conn, error) {
	if m.acquireErr != nil {
		return nil, m.acquireErr
	}
	return nil, nil
}

func (m *MockDB) Pool() *pgxpool.Pool {
	return m.pool
}

type MockMigr struct {
	upErr       error
	downErr     error
	migrateErr  error
	stepsErr    error
	forceErr    error
	versionErr  error
	version     uint
	dirty       bool
	callHistory []string
}

func (m *MockMigr) Up(ctx context.Context, dsn string, cfg database.MigrateConfig) error {
	m.callHistory = append(m.callHistory, "Up")
	return m.upErr
}

func (m *MockMigr) Down(ctx context.Context, dsn string, cfg database.MigrateConfig) error {
	m.callHistory = append(m.callHistory, "Down")
	return m.downErr
}

func (m *MockMigr) Migrate(ctx context.Context, dsn string, version uint, cfg database.MigrateConfig) error {
	m.callHistory = append(m.callHistory, "Migrate")
	return m.migrateErr
}

func (m *MockMigr) Steps(ctx context.Context, dsn string, n int, cfg database.MigrateConfig) error {
	m.callHistory = append(m.callHistory, "Steps")
	return m.stepsErr
}

func (m *MockMigr) Force(ctx context.Context, dsn string, version int, cfg database.MigrateConfig) error {
	m.callHistory = append(m.callHistory, "Force")
	return m.forceErr
}

func (m *MockMigr) Version(ctx context.Context, dsn string, cfg database.MigrateConfig) (uint, bool, error) {
	m.callHistory = append(m.callHistory, "Version")
	return m.version, m.dirty, m.versionErr
}

type MockHandler struct {
	db       *MockDB
	migrator *MockMigr
}

func (h *MockHandler) MigrateUp(ctx context.Context, cfg database.MigrateConfig) error {
	if err := h.migrator.Up(ctx, h.db.GetDSN(), cfg); err != nil {
		return fmt.Errorf("applying migrations: %w", err)
	}
	return nil
}

func (h *MockHandler) MigrateDown(ctx context.Context, cfg database.MigrateConfig) error {
	if err := h.migrator.Down(ctx, h.db.GetDSN(), cfg); err != nil {
		return fmt.Errorf("rolling back migrations: %w", err)
	}
	return nil
}

func (h *MockHandler) MigrateToVersion(ctx context.Context, version uint, cfg database.MigrateConfig) error {
	if err := h.migrator.Migrate(ctx, h.db.GetDSN(), version, cfg); err != nil {
		return fmt.Errorf("migrating to version %d: %w", version, err)
	}
	return nil
}

func (h *MockHandler) MigrateSteps(ctx context.Context, steps int, cfg database.MigrateConfig) error {
	if err := h.migrator.Steps(ctx, h.db.GetDSN(), steps, cfg); err != nil {
		return fmt.Errorf("executing %d migration steps: %w", steps, err)
	}
	return nil
}

func (h *MockHandler) MigrateForce(ctx context.Context, version int, cfg database.MigrateConfig) error {
	if err := h.migrator.Force(ctx, h.db.GetDSN(), version, cfg); err != nil {
		return fmt.Errorf("forcing version %d: %w", version, err)
	}
	return nil
}

func (h *MockHandler) GetMigrationVersion(ctx context.Context, cfg database.MigrateConfig) (uint, bool, error) {
	return h.migrator.Version(ctx, h.db.GetDSN(), cfg)
}

func (h *MockHandler) Close(ctx context.Context) {
	h.db.Close(ctx)
}

func (h *MockHandler) AcquireConn(ctx context.Context) (*pgxpool.Conn, error) {
	return h.db.AcquireConn(ctx)
}

func (h *MockHandler) Pool() *pgxpool.Pool {
	return h.db.Pool()
}

func (h *MockHandler) Ping(ctx context.Context) error {
	return h.db.Ping(ctx)
}

func TestHandlerMethods(t *testing.T) {
	ctx := context.Background()

	t.Run("MigrateUp", func(t *testing.T) {
		testCases := []struct {
			name      string
			mockError error
			expectErr bool
		}{
			{"Success", nil, false},
			{"Error", errMigration, true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mockDB := &MockDB{dsn: "mock-dsn"}
				mockMigrator := &MockMigr{upErr: tc.mockError}
				mockHandler := &MockHandler{db: mockDB, migrator: mockMigrator}

				err := mockHandler.MigrateUp(ctx, database.MigrateConfig{Path: "/test"})

				if tc.expectErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
				assert.Contains(t, mockMigrator.callHistory, "Up")
			})
		}
	})

	t.Run("MigrateDown", func(t *testing.T) {
		testCases := []struct {
			name      string
			mockError error
			expectErr bool
		}{
			{"Success", nil, false},
			{"Error", errRollback, true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mockDB := &MockDB{dsn: "mock-dsn"}
				mockMigrator := &MockMigr{downErr: tc.mockError}
				mockHandler := &MockHandler{db: mockDB, migrator: mockMigrator}

				err := mockHandler.MigrateDown(ctx, database.MigrateConfig{Path: "/test"})

				if tc.expectErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
				assert.Contains(t, mockMigrator.callHistory, "Down")
			})
		}
	})

	t.Run("MigrateToVersion", func(t *testing.T) {
		testCases := []struct {
			name      string
			mockError error
			expectErr bool
			version   uint
		}{
			{"Success", nil, false, 5},
			{"Error", errMigrateToVersion, true, 3},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mockDB := &MockDB{dsn: "mock-dsn"}
				mockMigrator := &MockMigr{migrateErr: tc.mockError}
				mockHandler := &MockHandler{db: mockDB, migrator: mockMigrator}

				err := mockHandler.MigrateToVersion(ctx, tc.version, database.MigrateConfig{Path: "/test"})

				if tc.expectErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
				assert.Contains(t, mockMigrator.callHistory, "Migrate")
			})
		}
	})

	t.Run("MigrateSteps", func(t *testing.T) {
		testCases := []struct {
			name      string
			mockError error
			expectErr bool
			steps     int
		}{
			{"Success Forward", nil, false, 2},
			{"Success Backward", nil, false, -2},
			{"Error", errMigrateSteps, true, 1},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mockDB := &MockDB{dsn: "mock-dsn"}
				mockMigrator := &MockMigr{stepsErr: tc.mockError}
				mockHandler := &MockHandler{db: mockDB, migrator: mockMigrator}

				err := mockHandler.MigrateSteps(ctx, tc.steps, database.MigrateConfig{Path: "/test"})

				if tc.expectErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
				assert.Contains(t, mockMigrator.callHistory, "Steps")
			})
		}
	})

	t.Run("MigrateForce", func(t *testing.T) {
		testCases := []struct {
			name      string
			mockError error
			expectErr bool
			version   int
		}{
			{"Success", nil, false, 3},
			{"Error", errMigrateForce, true, 1},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mockDB := &MockDB{dsn: "mock-dsn"}
				mockMigrator := &MockMigr{forceErr: tc.mockError}
				mockHandler := &MockHandler{db: mockDB, migrator: mockMigrator}

				err := mockHandler.MigrateForce(ctx, tc.version, database.MigrateConfig{Path: "/test"})

				if tc.expectErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
				assert.Contains(t, mockMigrator.callHistory, "Force")
			})
		}
	})

	t.Run("GetMigrationVersion", func(t *testing.T) {
		testCases := []struct {
			name        string
			mockError   error
			expectErr   bool
			version     uint
			dirty       bool
			expectVer   uint
			expectDirty bool
		}{
			{"Success Clean", nil, false, 5, false, 5, false},
			{"Success Dirty", nil, false, 3, true, 3, true},
			{"Error", errMigrateVersion, true, 0, false, 0, false},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mockDB := &MockDB{dsn: "mock-dsn"}
				mockMigrator := &MockMigr{
					versionErr: tc.mockError,
					version:    tc.version,
					dirty:      tc.dirty,
				}
				mockHandler := &MockHandler{db: mockDB, migrator: mockMigrator}

				version, dirty, err := mockHandler.GetMigrationVersion(ctx, database.MigrateConfig{Path: "/test"})

				if tc.expectErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
					assert.Equal(t, tc.expectVer, version)
					assert.Equal(t, tc.expectDirty, dirty)
				}
				assert.Contains(t, mockMigrator.callHistory, "Version")
			})
		}
	})

	t.Run("Close", func(t *testing.T) {
		mockDB := &MockDB{}
		mockHandler := &MockHandler{db: mockDB, migrator: &MockMigr{}}

		mockHandler.Close(ctx)

		assert.True(t, mockDB.closeCalled, "Close method should have been called")
	})

	t.Run("AcquireConn", func(t *testing.T) {
		testCases := []struct {
			name      string
			mockError error
			expectErr bool
		}{
			{"Success", nil, false},
			{"Error", errAcquireConnection, true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mockDB := &MockDB{acquireErr: tc.mockError}
				mockHandler := &MockHandler{db: mockDB, migrator: &MockMigr{}}

				conn, err := mockHandler.AcquireConn(ctx)

				if tc.expectErr {
					require.Error(t, err)
					assert.Nil(t, conn)
				} else {
					require.NoError(t, err)
				}
			})
		}
	})

	t.Run("Pool", func(t *testing.T) {
		pool := &pgxpool.Pool{}
		mockDB := &MockDB{pool: pool}
		mockHandler := &MockHandler{db: mockDB, migrator: &MockMigr{}}

		result := mockHandler.Pool()

		assert.Equal(t, pool, result)
	})

	t.Run("Ping", func(t *testing.T) {
		testCases := []struct {
			name      string
			mockError error
			expectErr bool
		}{
			{"Success", nil, false},
			{"Error", errPingDatabase, true},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				mockDB := &MockDB{pingErr: tc.mockError}
				mockHandler := &MockHandler{db: mockDB, migrator: &MockMigr{}}

				err := mockHandler.Ping(ctx)

				if tc.expectErr {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
				}
			})
		}
	})
}

func TestHandlerCloseBehavior(t *testing.T) {
	t.Run("Handler.Close should not panic", func(t *testing.T) {
		ctx := context.Background()

		handler := &database.Handler{
			DB:       nil,
			Migrator: database.NewMigrator(),
		}

		assert.NotPanics(t, func() {
			if handler.DB != nil {
				handler.DB.Close(ctx)
			}
		})
	})
}

func TestDatabaseIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	ctx := context.Background()
	config := database.PostgresConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "postgres",
		Database:        "postgres",
		SSLMode:         "disable",
		MinConns:        1,
		MaxConns:        5,
		ConnTimeout:     5 * time.Second,
		HealthPeriod:    30 * time.Second,
		ApplicationName: "test-app",
	}
	migrateConfig := database.MigrateConfig{
		Path: "./testdata/migrations",
	}

	handler, err := database.NewHandler(ctx, config, migrateConfig)
	if err != nil {
		t.Skip("Integration test requires a working database connection")
		return
	}
	defer handler.Close(ctx)

	t.Run("Integration - Ping", func(t *testing.T) {
		err := handler.Ping(ctx)
		assert.NoError(t, err)
	})

	t.Run("Integration - AcquireConn", func(t *testing.T) {
		conn, err := handler.AcquireConn(ctx)
		require.NoError(t, err)
		assert.NotNil(t, conn)
		conn.Release()
	})

	t.Run("Integration - Pool", func(t *testing.T) {
		pool := handler.Pool()
		assert.NotNil(t, pool)
	})
}
