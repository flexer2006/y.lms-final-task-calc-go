FROM golang:1.24.3-alpine AS builder

WORKDIR /app

# Установка зависимостей для сборки
RUN apk add --no-cache git

# Копирование и загрузка зависимостей
COPY go.mod go.sum ./
RUN go mod download

# Копирование исходного кода
COPY . .

# Компиляция приложения
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o auth-service ./cmd/auth

# Финальный образ
FROM alpine:3.21

WORKDIR /app

# Установка необходимых пакетов
RUN apk --no-cache add ca-certificates tzdata netcat-openbsd

# Копирование бинарного файла из предыдущего этапа
COPY --from=builder /app/auth-service .

# Копирование миграций
COPY --from=builder /app/migrations/auth /app/migrations/auth

# Проверка здоровья
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
  CMD nc -z localhost ${AUTH_GRPC_PORT:-50052} || exit 1

ENTRYPOINT ["./auth-service"]