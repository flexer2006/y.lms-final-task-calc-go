-- Включение расширения UUID.
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Таблица выражений для хранения вычислений с пользовательским контекстом.
CREATE TABLE calculations (
                              id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
                              user_id UUID NOT NULL,
                              expression TEXT NOT NULL,
                              result TEXT,
                              status VARCHAR(50) NOT NULL,
                              error_message TEXT,
                              created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
                              updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индекс для поиска выражений пользователем.
CREATE INDEX idx_calculations_user_id ON calculations(user_id);

-- Функция автоматического обновления временных меток.
CREATE OR REPLACE FUNCTION trigger_set_timestamp()
RETURNS TRIGGER AS $$
BEGIN
  NEW.updated_at = NOW();
RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Триггер для обновления временной метки при обновлении вычислений.
CREATE TRIGGER set_calculation_timestamp
    BEFORE UPDATE ON calculations
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_timestamp();