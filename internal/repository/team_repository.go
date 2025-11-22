package repository

import (
	"database/sql"
	"fmt"

	"github.com/user/pr-reviewer/internal/database"
	"github.com/user/pr-reviewer/internal/models"
)

// TeamRepository репозиторий для работы с командами
type TeamRepository struct {
	db *database.DB
}

// NewTeamRepository создаёт новый репозиторий команд
func NewTeamRepository(db *database.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

// Create создаёт новую команду
func (r *TeamRepository) Create(team *models.Team) error {
	query := `
		INSERT INTO teams (name) 
		VALUES ($1) 
		RETURNING id, created_at, updated_at`

	err := r.db.QueryRow(query, team.Name).Scan(&team.ID, &team.CreatedAt, &team.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create team: %w", err)
	}

	return nil
}

// GetByID возвращает команду по ID
func (r *TeamRepository) GetByID(id int) (*models.Team, error) {
	team := &models.Team{}
	query := `
		SELECT id, name, created_at, updated_at 
		FROM teams 
		WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(&team.ID, &team.Name, &team.CreatedAt, &team.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("team not found")
		}
		return nil, fmt.Errorf("failed to get team: %w", err)
	}

	return team, nil
}

// GetByName возвращает команду по имени
func (r *TeamRepository) GetByName(name string) (*models.Team, error) {
	team := &models.Team{}
	query := `
		SELECT id, name, created_at, updated_at 
		FROM teams 
		WHERE name = $1`

	err := r.db.QueryRow(query, name).Scan(&team.ID, &team.Name, &team.CreatedAt, &team.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get team by name: %w", err)
	}

	return team, nil
}

// GetAll возвращает все команды
func (r *TeamRepository) GetAll() ([]*models.Team, error) {
	query := `
		SELECT id, name, created_at, updated_at 
		FROM teams 
		ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get teams: %w", err)
	}
	defer rows.Close()

	var teams []*models.Team
	for rows.Next() {
		team := &models.Team{}
		if err := rows.Scan(&team.ID, &team.Name, &team.CreatedAt, &team.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan team: %w", err)
		}
		teams = append(teams, team)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate teams: %w", err)
	}

	return teams, nil
}

// AddUser добавляет пользователя в команду
func (r *TeamRepository) AddUser(teamID, userID int) error {
	query := `UPDATE users SET team_id = $1 WHERE id = $2`

	result, err := r.db.Exec(query, teamID, userID)
	if err != nil {
		return fmt.Errorf("failed to add user to team: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// RemoveUser удаляет пользователя из команды
func (r *TeamRepository) RemoveUser(teamID, userID int) error {
	query := `UPDATE users SET team_id = NULL WHERE id = $1 AND team_id = $2`

	result, err := r.db.Exec(query, userID, teamID)
	if err != nil {
		return fmt.Errorf("failed to remove user from team: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found in team")
	}

	return nil
}

// Delete удаляет команду по ID
func (r *TeamRepository) Delete(id int) error {
	// Сначала удаляем связь пользователей с командой
	query := `UPDATE users SET team_id = NULL WHERE team_id = $1`

	_, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to remove users from team: %w", err)
	}

	// Затем удаляем саму команду
	query = `DELETE FROM teams WHERE id = $1`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete team: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("team not found")
	}

	return nil
}
