-- Создание таблицы для webhook подписок
CREATE TABLE IF NOT EXISTS webhook_subscriptions (
    id BIGSERIAL PRIMARY KEY,
    url VARCHAR(500) NOT NULL,
    events TEXT[] NOT NULL,
    secret VARCHAR(255),
    active BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Создание таблицы для webhook delivery attempts
CREATE TABLE IF NOT EXISTS webhook_deliveries (
    id BIGSERIAL PRIMARY KEY,
    subscription_id BIGINT NOT NULL REFERENCES webhook_subscriptions(id) ON DELETE CASCADE,
    event VARCHAR(50) NOT NULL,
    payload JSONB NOT NULL,
    status VARCHAR(20) NOT NULL, -- pending, success, failed
    response_code INT,
    response_body TEXT,
    error TEXT,
    attempts INT NOT NULL DEFAULT 0,
    next_retry_at TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    delivered_at TIMESTAMP
);

-- Индексы
CREATE INDEX idx_webhook_subscriptions_active ON webhook_subscriptions(active) WHERE active = true;
CREATE INDEX idx_webhook_deliveries_subscription ON webhook_deliveries(subscription_id);
CREATE INDEX idx_webhook_deliveries_status ON webhook_deliveries(status);
CREATE INDEX idx_webhook_deliveries_next_retry ON webhook_deliveries(next_retry_at) 
    WHERE status = 'pending' AND next_retry_at IS NOT NULL;
CREATE INDEX idx_webhook_deliveries_created ON webhook_deliveries(created_at DESC);

-- Комментарии
COMMENT ON TABLE webhook_subscriptions IS 'Webhook подписки для уведомлений о событиях';
COMMENT ON TABLE webhook_deliveries IS 'История доставки webhook';
COMMENT ON COLUMN webhook_subscriptions.events IS 'Массив событий: pr.created, pr.merged, reviewer.assigned и т.д.';
COMMENT ON COLUMN webhook_subscriptions.secret IS 'Секрет для HMAC подписи';
COMMENT ON COLUMN webhook_deliveries.attempts IS 'Количество попыток доставки';

