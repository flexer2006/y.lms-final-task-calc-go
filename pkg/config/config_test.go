package config_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/config"
	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type TestConfig struct {
	ServerHost string `yaml:"server_host" env:"SERVER_HOST" env-default:"localhost"`
	ServerPort int    `yaml:"server_port" env:"SERVER_PORT" env-default:"8080"`
	Debug      bool   `yaml:"debug" env:"DEBUG" env-default:"false"`
}

type InvalidEnvConfig struct {
	Channel chan int `env:"INVALID_CHANNEL"`
}

func TestLoad(t *testing.T) {
	ctx := context.Background()

	t.Run("LoadDefaultValues", func(t *testing.T) {
		cfg, err := config.Load[TestConfig](ctx)
		require.NoError(t, err)
		assert.Equal(t, "localhost", cfg.ServerHost)
		assert.Equal(t, 8080, cfg.ServerPort)
		assert.False(t, cfg.Debug)
	})

	t.Run("LoadFromEnvironment", func(t *testing.T) {
		t.Setenv("SERVER_HOST", "env-host")
		t.Setenv("SERVER_PORT", "9090")
		t.Setenv("DEBUG", "true")

		cfg, err := config.Load[TestConfig](ctx)
		require.NoError(t, err)
		assert.Equal(t, "env-host", cfg.ServerHost)
		assert.Equal(t, 9090, cfg.ServerPort)
		assert.True(t, cfg.Debug)
	})

	t.Run("LoadFromFile", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")
		configContent := []byte(`
server_host: file-host
server_port: 5000
debug: true
`)
		require.NoError(t, os.WriteFile(configPath, configContent, 0644)) //nolint:gosec

		cfg, err := config.Load[TestConfig](ctx, config.WithConfigPath(configPath))
		require.NoError(t, err)
		assert.Equal(t, "file-host", cfg.ServerHost)
		assert.Equal(t, 5000, cfg.ServerPort)
		assert.True(t, cfg.Debug)
	})

	t.Run("EnvOverridesFile", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")
		configContent := []byte(`
server_host: file-host
server_port: 5000
debug: false
`)
		require.NoError(t, os.WriteFile(configPath, configContent, 0644)) //nolint:gosec

		t.Setenv("SERVER_HOST", "override-host")

		cfg, err := config.Load[TestConfig](ctx, config.WithConfigPath(configPath))
		require.NoError(t, err)

		assert.Equal(t, "override-host", cfg.ServerHost)
		assert.Equal(t, 5000, cfg.ServerPort)
		assert.False(t, cfg.Debug)
	})

	t.Run("NonExistentConfigFile", func(t *testing.T) {
		cfg, err := config.Load[TestConfig](ctx, config.WithConfigPath("/path/to/nonexistent/config.yaml"))
		require.NoError(t, err)
		assert.Equal(t, "localhost", cfg.ServerHost)
	})

	t.Run("InvalidConfigFile", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "invalid.yaml")
		invalidContent := []byte(`
invalid: [unclosed array
server_host: broken-host
`)
		require.NoError(t, os.WriteFile(configPath, invalidContent, 0644)) //nolint:gosec

		_, err := config.Load[TestConfig](ctx, config.WithConfigPath(configPath))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load configuration from file")
	})

	t.Run("InvalidEnvironmentVariable", func(t *testing.T) {
		t.Setenv("INVALID_CHANNEL", "not-a-channel")

		_, err := config.Load[InvalidEnvConfig](ctx)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to load configuration from environment")
	})

	t.Run("CustomLogger", func(t *testing.T) {
		customLogger := logger.Console(logger.InfoLevel, false)
		ctx := logger.WithLogger(context.Background(), customLogger)

		_, err := config.Load[TestConfig](ctx)
		require.NoError(t, err)
	})

	t.Run("EmptyOptions", func(t *testing.T) {
		_, err := config.Load[TestConfig](ctx)
		require.NoError(t, err)
	})

	t.Run("MultipleOptions", func(t *testing.T) {
		tmpDir := t.TempDir()
		configPath := filepath.Join(tmpDir, "config.yaml")
		configContent := []byte(`
server_host: file-host
server_port: 5000
debug: true
`)
		require.NoError(t, os.WriteFile(configPath, configContent, 0644)) //nolint:gosec

		// Создаем логгер для теста
		testLogger, err := logger.Development()
		require.NoError(t, err)

		// Создаем контекст с логгером
		ctx = logger.WithLogger(context.Background(), testLogger)

		cfg, err := config.Load[TestConfig](ctx, config.WithConfigPath(configPath))
		require.NoError(t, err)
		assert.Equal(t, "file-host", cfg.ServerHost)
		assert.Equal(t, 5000, cfg.ServerPort)
		assert.True(t, cfg.Debug)
	})
}

func TestWithConfigPath(t *testing.T) {
	opt := config.WithConfigPath("test/path.yaml")
	require.NotNil(t, opt)
}
