-- Удаление индексов
DROP INDEX IF EXISTS idx_pr_reviewers_reviewer_id;
DROP INDEX IF EXISTS idx_pull_requests_status;
DROP INDEX IF EXISTS idx_pull_requests_author_id;
DROP INDEX IF EXISTS idx_users_is_active;
DROP INDEX IF EXISTS idx_users_team_id;

-- Удаление таблиц
DROP TABLE IF EXISTS pr_reviewers;
DROP TABLE IF EXISTS pull_requests;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS teams;
