-- Добавляем поле username в таблицу users
ALTER TABLE users ADD COLUMN username VARCHAR(100);

-- Копируем данные из name в username для существующих записей
UPDATE users SET username = name;

-- Делаем поле username обязательным и уникальным
ALTER TABLE users ALTER COLUMN username SET NOT NULL;
ALTER TABLE users ADD CONSTRAINT users_username_unique UNIQUE (username);

-- Создаем индекс для оптимизации поиска по username
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);

