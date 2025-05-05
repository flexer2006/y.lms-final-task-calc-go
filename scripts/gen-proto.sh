#!/bin/bash
set -e

# Проверка наличия protoc
if ! command -v protoc &> /dev/null; then
    echo "Ошибка: protoc не установлен"
    echo "Установите protoc с помощью инструкций: https://grpc.io/docs/protoc-installation/"
    exit 1
fi

# Проверка наличия плагинов protoc
if ! command -v protoc-gen-go &> /dev/null; then
    echo "Ошибка: protoc-gen-go не установлен"
    echo "Установите с помощью: go install google.golang.org/protobuf/cmd/protoc-gen-go@latest"
    exit 1
fi

if ! command -v protoc-gen-go-grpc &> /dev/null; then
    echo "Ошибка: protoc-gen-go-grpc не установлен"
    echo "Установите с помощью: go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest"
    exit 1
fi

# Основная рабочая директория проекта
PROJECT_ROOT=$(pwd)

# Создание директории для сгенерированных файлов
mkdir -p ${PROJECT_ROOT}/pkg/api/proto/v1/agent
mkdir -p ${PROJECT_ROOT}/pkg/api/proto/v1/auth
mkdir -p ${PROJECT_ROOT}/pkg/api/proto/v1/common
mkdir -p ${PROJECT_ROOT}/pkg/api/proto/v1/orchestrator

# Установка прав на исполнение скрипта
chmod +x "$0"

echo "Генерация Go-кода из proto файлов в pkg/api/proto/v1..."

# Генерация Go-кода для всех proto-файлов в нужную директорию
protoc \
    --proto_path="${PROJECT_ROOT}" \
    --go_out="${PROJECT_ROOT}" \
    --go_opt=paths=source_relative \
    --go-grpc_out="${PROJECT_ROOT}" \
    --go-grpc_opt=paths=source_relative \
    proto/v1/agent/agent.proto \
    proto/v1/auth/auth.proto \
    proto/v1/common/common.proto \
    proto/v1/orchestrator/orchestrator.proto

# Перемещаем сгенерированные файлы в нужную директорию
mkdir -p ${PROJECT_ROOT}/pkg/api/proto/v1/agent
mkdir -p ${PROJECT_ROOT}/pkg/api/proto/v1/auth
mkdir -p ${PROJECT_ROOT}/pkg/api/proto/v1/common
mkdir -p ${PROJECT_ROOT}/pkg/api/proto/v1/orchestrator

# Перемещаем сгенерированные файлы в нужную папку
mv ${PROJECT_ROOT}/proto/v1/agent/*.go ${PROJECT_ROOT}/pkg/api/proto/v1/agent/ 2>/dev/null || true
mv ${PROJECT_ROOT}/proto/v1/auth/*.go ${PROJECT_ROOT}/pkg/api/proto/v1/auth/ 2>/dev/null || true
mv ${PROJECT_ROOT}/proto/v1/common/*.go ${PROJECT_ROOT}/pkg/api/proto/v1/common/ 2>/dev/null || true
mv ${PROJECT_ROOT}/proto/v1/orchestrator/*.go ${PROJECT_ROOT}/pkg/api/proto/v1/orchestrator/ 2>/dev/null || true

echo "Код успешно сгенерирован в директории pkg/api/proto/v1"