package repository

import (
	"fmt"

	"github.com/user/pr-reviewer/internal/database"
	"github.com/user/pr-reviewer/internal/models"
)

// StatisticsRepository репозиторий для работы со статистикой
type StatisticsRepository struct {
	db *database.DB
}

// NewStatisticsRepository создаёт новый репозиторий статистики
func NewStatisticsRepository(db *database.DB) *StatisticsRepository {
	return &StatisticsRepository{db: db}
}

// GetStatistics возвращает общую статистику
func (r *StatisticsRepository) GetStatistics() (*models.Statistics, error) {
	stats := &models.Statistics{}

	// Получаем общие счётчики PR
	err := r.db.QueryRow(`
		SELECT 
			COUNT(*) as total,
			COUNT(*) FILTER (WHERE status = 'OPEN') as open,
			COUNT(*) FILTER (WHERE status = 'MERGED') as merged,
			COUNT(*) FILTER (WHERE status = 'CLOSED') as closed
		FROM pull_requests
	`).Scan(&stats.TotalPRs, &stats.OpenPRs, &stats.MergedPRs, &stats.ClosedPRs)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR counts: %w", err)
	}

	// Получаем статистику по пользователям
	userStats, err := r.getUserStatistics()
	if err != nil {
		return nil, err
	}
	stats.UserStats = userStats

	// Получаем статистику по командам
	teamStats, err := r.getTeamStatistics()
	if err != nil {
		return nil, err
	}
	stats.TeamStats = teamStats

	return stats, nil
}

// getUserStatistics возвращает статистику по пользователям
func (r *StatisticsRepository) getUserStatistics() ([]models.UserStatistic, error) {
	query := `
		SELECT 
			u.id as user_id,
			u.name as user_name,
			COUNT(pr.pr_id) as assignment_count
		FROM users u
		LEFT JOIN pr_reviewers pr ON u.id = pr.reviewer_id
		GROUP BY u.id, u.name
		HAVING COUNT(pr.pr_id) > 0
		ORDER BY assignment_count DESC, u.name
		LIMIT 20`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get user statistics: %w", err)
	}
	defer rows.Close()

	var stats []models.UserStatistic
	for rows.Next() {
		var stat models.UserStatistic
		if err := rows.Scan(&stat.UserID, &stat.UserName, &stat.AssignmentCount); err != nil {
			return nil, fmt.Errorf("failed to scan user statistic: %w", err)
		}
		stats = append(stats, stat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate user statistics: %w", err)
	}

	return stats, nil
}

// getTeamStatistics возвращает статистику по командам
func (r *StatisticsRepository) getTeamStatistics() ([]models.TeamStatistic, error) {
	query := `
		SELECT 
			t.id as team_id,
			t.name as team_name,
			COUNT(DISTINCT p.id) as pr_count
		FROM teams t
		LEFT JOIN users u ON t.id = u.team_id
		LEFT JOIN pull_requests p ON u.id = p.author_id
		GROUP BY t.id, t.name
		HAVING COUNT(DISTINCT p.id) > 0
		ORDER BY pr_count DESC, t.name
		LIMIT 20`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to get team statistics: %w", err)
	}
	defer rows.Close()

	var stats []models.TeamStatistic
	for rows.Next() {
		var stat models.TeamStatistic
		if err := rows.Scan(&stat.TeamID, &stat.TeamName, &stat.PRCount); err != nil {
			return nil, fmt.Errorf("failed to scan team statistic: %w", err)
		}
		stats = append(stats, stat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate team statistics: %w", err)
	}

	return stats, nil
}
