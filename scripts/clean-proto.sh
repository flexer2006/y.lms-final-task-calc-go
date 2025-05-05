#!/bin/bash
set -e

PROJECT_ROOT=$(pwd)

echo "Очистка сгенерированных proto файлов..."

if [ -d "${PROJECT_ROOT}/pkg/api/proto/v1" ]; then
    rm -rf "${PROJECT_ROOT}/pkg/api/proto/v1"
    echo "Удалены файлы в pkg/api/proto/v1"
    
    mkdir -p ${PROJECT_ROOT}/pkg/api/proto/v1
    echo "Директория pkg/api/proto/v1 пересоздана"
else
    echo "Директория pkg/api/proto/v1 не существует"
fi

echo "Очистка завершена"