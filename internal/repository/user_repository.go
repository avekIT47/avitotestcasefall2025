package repository

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/user/pr-reviewer/internal/database"
	"github.com/user/pr-reviewer/internal/models"
)

// UserRepository репозиторий для работы с пользователями
type UserRepository struct {
	db *database.DB
}

// NewUserRepository создаёт новый репозиторий пользователей
func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create создаёт нового пользователя
func (r *UserRepository) Create(user *models.User) error {
	query := `
		INSERT INTO users (username, name, is_active, team_id) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(query, user.Username, user.Name, user.IsActive, user.TeamID).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}

	return nil
}

// GetByID возвращает пользователя по ID
func (r *UserRepository) GetByID(id int) (*models.User, error) {
	user := &models.User{}
	query := `
		SELECT id, username, name, is_active, team_id, created_at, updated_at 
		FROM users 
		WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Name, &user.IsActive, &user.TeamID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// GetAll возвращает всех пользователей с фильтрами
func (r *UserRepository) GetAll(teamID *int, isActive *bool) ([]*models.User, error) {
	query := `
		SELECT id, username, name, is_active, team_id, created_at, updated_at 
		FROM users 
		WHERE 1=1`

	args := []interface{}{}
	argNum := 1

	if teamID != nil {
		query += fmt.Sprintf(" AND team_id = $%d", argNum)
		args = append(args, *teamID)
		argNum++
	}

	if isActive != nil {
		query += fmt.Sprintf(" AND is_active = $%d", argNum)
		args = append(args, *isActive)
		argNum++
	}

	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(&user.ID, &user.Username, &user.Name, &user.IsActive, &user.TeamID, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate users: %w", err)
	}

	return users, nil
}

// GetByIDs возвращает пользователей по списку ID
func (r *UserRepository) GetByIDs(ids []int) ([]*models.User, error) {
	if len(ids) == 0 {
		return []*models.User{}, nil
	}

	placeholders := make([]string, len(ids))
	args := make([]interface{}, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}

	query := fmt.Sprintf(`
		SELECT id, username, name, is_active, team_id, created_at, updated_at 
		FROM users 
		WHERE id IN (%s)
		ORDER BY id`, strings.Join(placeholders, ","))

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by IDs: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(&user.ID, &user.Username, &user.Name, &user.IsActive, &user.TeamID, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate users: %w", err)
	}

	return users, nil
}

// Update обновляет пользователя
func (r *UserRepository) Update(id int, req *models.UpdateUserRequest) (*models.User, error) {
	// Сначала проверяем существование пользователя
	user, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Строим динамический запрос
	setClauses := []string{}
	args := []interface{}{}
	argNum := 1

	if req.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argNum))
		args = append(args, *req.Name)
		argNum++
	}

	if req.IsActive != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argNum))
		args = append(args, *req.IsActive)
		argNum++
	}

	if len(setClauses) == 0 {
		return user, nil // Нечего обновлять
	}

	args = append(args, id)
	query := fmt.Sprintf(`
		UPDATE users 
		SET %s 
		WHERE id = $%d
		RETURNING id, username, name, is_active, team_id, created_at, updated_at`,
		strings.Join(setClauses, ", "), argNum)

	err = r.db.QueryRow(query, args...).Scan(
		&user.ID, &user.Username, &user.Name, &user.IsActive, &user.TeamID, &user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user, nil
}

// GetActiveUsersFromTeam возвращает активных пользователей из команды
func (r *UserRepository) GetActiveUsersFromTeam(teamID int, excludeUserID int) ([]*models.User, error) {
	query := `
		SELECT id, username, name, is_active, team_id, created_at, updated_at 
		FROM users 
		WHERE team_id = $1 AND is_active = true AND id != $2
		ORDER BY RANDOM()`

	rows, err := r.db.Query(query, teamID, excludeUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active team users: %w", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{}
		if err := rows.Scan(&user.ID, &user.Username, &user.Name, &user.IsActive, &user.TeamID, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate users: %w", err)
	}

	return users, nil
}

// BulkDeactivate деактивирует несколько пользователей
func (r *UserRepository) BulkDeactivate(teamID int, userIDs []int) (int, error) {
	if len(userIDs) == 0 {
		return 0, nil
	}

	placeholders := make([]string, len(userIDs))
	args := make([]interface{}, len(userIDs)+1)
	args[0] = teamID

	for i, id := range userIDs {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}

	query := fmt.Sprintf(`
		UPDATE users 
		SET is_active = false 
		WHERE team_id = $1 AND id IN (%s) AND is_active = true`,
		strings.Join(placeholders, ","))

	result, err := r.db.Exec(query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to bulk deactivate users: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get affected rows: %w", err)
	}

	return int(affected), nil
}
