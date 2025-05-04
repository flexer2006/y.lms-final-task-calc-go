-- Включение расширения UUID.
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Таблица пользователей для сервиса аутентификации.
CREATE TABLE users (
                       id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                       login VARCHAR(255) NOT NULL UNIQUE,
                       password_hash VARCHAR(255) NOT NULL,
                       created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                       updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индекс для более быстрого поиска логинов.
CREATE INDEX idx_users_login ON users(login);

-- Таблица токенов JWT (для отслеживания/аннулирования токенов при необходимости).
CREATE TABLE tokens (
                        id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                        user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
                        token TEXT NOT NULL UNIQUE,
                        expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
                        created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                        is_revoked BOOLEAN NOT NULL DEFAULT FALSE
);

-- Индексы для операций с токенами.
CREATE INDEX idx_tokens_user_id ON tokens(user_id);
CREATE INDEX idx_tokens_token ON tokens(token);