-- Добавление поля updated_at в таблицу teams
ALTER TABLE teams 
ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- Добавление поля updated_at в таблицу users
ALTER TABLE users 
ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- Добавление поля updated_at в таблицу pull_requests
ALTER TABLE pull_requests 
ADD COLUMN updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP;

-- Создание функции для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

-- Создание триггера для teams
CREATE TRIGGER update_teams_updated_at 
BEFORE UPDATE ON teams 
FOR EACH ROW 
EXECUTE FUNCTION update_updated_at_column();

-- Создание триггера для users
CREATE TRIGGER update_users_updated_at 
BEFORE UPDATE ON users 
FOR EACH ROW 
EXECUTE FUNCTION update_updated_at_column();

-- Создание триггера для pull_requests
CREATE TRIGGER update_pull_requests_updated_at 
BEFORE UPDATE ON pull_requests 
FOR EACH ROW 
EXECUTE FUNCTION update_updated_at_column();

