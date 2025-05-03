// Package config содержит утилиты для загрузки конфигурации из переменных окружения.
package config

import (
	"context"
	"fmt"
	"os"

	"github.com/flexer2006/y.lms-final-task-calc-go/pkg/logger"
	"github.com/ilyakaznacheev/cleanenv"
	"go.uber.org/zap"
)

// Option определяет функциональную опцию для настройки процесса загрузки конфигурации.
type Option func(*loadOptions)

// loadOptions содержит внутренние настройки для загрузки конфигурации.
type loadOptions struct {
	configPath string
}

// WithConfigPath задает путь к файлу конфигурации.
func WithConfigPath(path string) Option {
	return func(opts *loadOptions) {
		opts.configPath = path
	}
}

// Load загружает конфигурацию для любого типа T.
// Если указан путь к файлу конфигурации, сначала загружается из него.
// Затем загружаются переменные окружения, которые могут переопределить значения из файла.
func Load[T any](ctx context.Context, opts ...Option) (*T, error) {
	// Инициализация настроек
	options := loadOptions{}
	for _, opt := range opts {
		opt(&options)
	}

	log, ok := logger.FromContext(ctx)
	if !ok {
		dev, _ := logger.Development()
		log = dev
	}

	log.Info("loading configuration")
	var cfg T

	if options.configPath != "" {
		if _, err := os.Stat(options.configPath); err == nil {
			if err := cleanenv.ReadConfig(options.configPath, &cfg); err != nil {
				log.Error("failed to load configuration from file",
					zap.Error(err),
					zap.String("path", options.configPath),
				)
				return nil, fmt.Errorf("failed to load configuration from file %s: %w", options.configPath, err)
			}
		}
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Error("failed to load configuration from environment", zap.Error(err))
		return nil, fmt.Errorf("failed to load configuration from environment: %w", err)
	}

	log.Info("configuration loaded successfully")
	return &cfg, nil
}
