-- Удаление триггеров
DROP TRIGGER IF EXISTS update_pull_requests_updated_at ON pull_requests;
DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP TRIGGER IF EXISTS update_teams_updated_at ON teams;

-- Удаление функции
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Удаление столбцов
ALTER TABLE pull_requests DROP COLUMN IF EXISTS updated_at;
ALTER TABLE users DROP COLUMN IF EXISTS updated_at;
ALTER TABLE teams DROP COLUMN IF EXISTS updated_at;

