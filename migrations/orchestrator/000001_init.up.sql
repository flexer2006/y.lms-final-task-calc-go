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

-- Таблица операций для хранения шагов вычислений.
CREATE TABLE operations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    calculation_id UUID NOT NULL REFERENCES calculations(id) ON DELETE CASCADE,
    operation_type INT NOT NULL,
    operand1 TEXT NOT NULL,
    operand2 TEXT NOT NULL,
    result TEXT,
    status VARCHAR(50) NOT NULL,
    error_message TEXT,
    processing_time_ms BIGINT DEFAULT 0,
    agent_id TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Индекс для поиска операций по ID вычисления.
CREATE INDEX idx_operations_calculation_id ON operations(calculation_id);

-- Индекс для поиска операций по статусу.
CREATE INDEX idx_operations_status ON operations(status);

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

-- Триггер для обновления временной метки при обновлении операций.
CREATE TRIGGER set_operation_timestamp
    BEFORE UPDATE ON operations
    FOR EACH ROW
    EXECUTE FUNCTION trigger_set_timestamp();