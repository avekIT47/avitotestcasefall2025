-- Создание таблицы для audit logs
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    timestamp TIMESTAMP NOT NULL DEFAULT NOW(),
    action VARCHAR(50) NOT NULL,
    entity VARCHAR(50) NOT NULL,
    entity_id BIGINT NOT NULL,
    user_id BIGINT,
    user_email VARCHAR(255),
    ip VARCHAR(45),
    user_agent TEXT,
    request_id VARCHAR(100),
    changes JSONB,
    description TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Индексы для быстрого поиска
CREATE INDEX idx_audit_logs_timestamp ON audit_logs(timestamp DESC);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX idx_audit_logs_entity ON audit_logs(entity, entity_id);
CREATE INDEX idx_audit_logs_action ON audit_logs(action);
CREATE INDEX idx_audit_logs_request_id ON audit_logs(request_id) WHERE request_id IS NOT NULL;

-- Партиционирование по месяцам (для больших объемов данных)
-- Раскомментируйте если нужно:
-- CREATE TABLE audit_logs_2024_01 PARTITION OF audit_logs
--     FOR VALUES FROM ('2024-01-01') TO ('2024-02-01');

-- Комментарии
COMMENT ON TABLE audit_logs IS 'Audit log для отслеживания всех действий пользователей';
COMMENT ON COLUMN audit_logs.action IS 'Тип действия: create, update, delete, read, login, logout';
COMMENT ON COLUMN audit_logs.entity IS 'Тип сущности: user, team, pull_request, reviewer';
COMMENT ON COLUMN audit_logs.changes IS 'JSON объект с изменениями';

