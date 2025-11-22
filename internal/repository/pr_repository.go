package repository

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/user/pr-reviewer/internal/database"
	"github.com/user/pr-reviewer/internal/models"
)

// PRRepository репозиторий для работы с Pull Requests
type PRRepository struct {
	db *database.DB
}

// NewPRRepository создаёт новый репозиторий PR
func NewPRRepository(db *database.DB) *PRRepository {
	return &PRRepository{db: db}
}

// Create создаёт новый PR
func (r *PRRepository) Create(pr *models.PullRequest) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Создаём PR
	query := `
		INSERT INTO pull_requests (title, author_id, status) 
		VALUES ($1, $2, $3) 
		RETURNING id, created_at, updated_at`

	err = tx.QueryRow(query, pr.Title, pr.AuthorID, pr.Status).
		Scan(&pr.ID, &pr.CreatedAt, &pr.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create PR: %w", err)
	}

	// Добавляем рецензентов
	if len(pr.Reviewers) > 0 {
		if err := r.addReviewersTx(tx, pr.ID, pr.Reviewers); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetByID возвращает PR по ID с рецензентами
func (r *PRRepository) GetByID(id int) (*models.PullRequest, error) {
	pr := &models.PullRequest{}
	query := `
		SELECT id, title, author_id, status, created_at, merged_at, updated_at 
		FROM pull_requests 
		WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status,
		&pr.CreatedAt, &pr.MergedAt, &pr.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("PR not found")
		}
		return nil, fmt.Errorf("failed to get PR: %w", err)
	}

	// Получаем рецензентов
	reviewers, err := r.getReviewers(pr.ID)
	if err != nil {
		return nil, err
	}
	pr.Reviewers = reviewers

	return pr, nil
}

// GetAll возвращает все PR с фильтрами
func (r *PRRepository) GetAll(userID *int, authorID *int, status *string) ([]*models.PullRequest, error) {
	baseQuery := `
		SELECT DISTINCT p.id, p.title, p.author_id, p.status, p.created_at, p.merged_at, p.updated_at 
		FROM pull_requests p`

	whereClauses := []string{}
	args := []interface{}{}
	argNum := 1
	needJoin := false

	if userID != nil {
		needJoin = true
		whereClauses = append(whereClauses, fmt.Sprintf("pr.reviewer_id = $%d", argNum))
		args = append(args, *userID)
		argNum++
	}

	if authorID != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("p.author_id = $%d", argNum))
		args = append(args, *authorID)
		argNum++
	}

	if status != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("p.status = $%d", argNum))
		args = append(args, *status)
		argNum++
	}

	query := baseQuery
	if needJoin {
		query += " LEFT JOIN pr_reviewers pr ON p.id = pr.pr_id"
	}

	if len(whereClauses) > 0 {
		query += " WHERE " + strings.Join(whereClauses, " AND ")
	}

	query += " ORDER BY p.created_at DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get PRs: %w", err)
	}
	defer rows.Close()

	var prs []*models.PullRequest
	for rows.Next() {
		pr := &models.PullRequest{}
		if err := rows.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt, &pr.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan PR: %w", err)
		}

		// Получаем рецензентов для каждого PR
		reviewers, err := r.getReviewers(pr.ID)
		if err != nil {
			return nil, err
		}
		pr.Reviewers = reviewers

		prs = append(prs, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate PRs: %w", err)
	}

	return prs, nil
}

// Merge переводит PR в состояние MERGED
func (r *PRRepository) Merge(id int) (*models.PullRequest, error) {
	now := time.Now()
	query := `
		UPDATE pull_requests 
		SET status = $1, merged_at = $2 
		WHERE id = $3 AND (status = 'OPEN' OR status = 'MERGED')
		RETURNING id, title, author_id, status, created_at, merged_at, updated_at`

	pr := &models.PullRequest{}
	err := r.db.QueryRow(query, models.PRStatusMerged, now, id).Scan(
		&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status,
		&pr.CreatedAt, &pr.MergedAt, &pr.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("PR not found")
		}
		return nil, fmt.Errorf("failed to merge PR: %w", err)
	}

	// Получаем рецензентов
	reviewers, err := r.getReviewers(pr.ID)
	if err != nil {
		return nil, err
	}
	pr.Reviewers = reviewers

	return pr, nil
}

// Close переводит PR в состояние CLOSED (закрыт без мерджа)
func (r *PRRepository) Close(id int) (*models.PullRequest, error) {
	query := `
		UPDATE pull_requests 
		SET status = $1
		WHERE id = $2 AND status = 'OPEN'
		RETURNING id, title, author_id, status, created_at, merged_at, updated_at`

	pr := &models.PullRequest{}
	err := r.db.QueryRow(query, models.PRStatusClosed, id).Scan(
		&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status,
		&pr.CreatedAt, &pr.MergedAt, &pr.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("PR not found or already closed/merged")
		}
		return nil, fmt.Errorf("failed to close PR: %w", err)
	}

	// Получаем рецензентов
	reviewers, err := r.getReviewers(pr.ID)
	if err != nil {
		return nil, err
	}
	pr.Reviewers = reviewers

	return pr, nil
}

// ReplaceReviewer заменяет рецензента
func (r *PRRepository) ReplaceReviewer(prID, oldReviewerID, newReviewerID int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Проверяем, что PR не в статусе MERGED
	var status models.PRStatus
	err = tx.QueryRow("SELECT status FROM pull_requests WHERE id = $1", prID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return fmt.Errorf("PR not found")
		}
		return fmt.Errorf("failed to check PR status: %w", err)
	}

	if status == models.PRStatusMerged {
		return fmt.Errorf("cannot change reviewers of merged PR")
	}

	// Удаляем старого рецензента
	deleteQuery := `DELETE FROM pr_reviewers WHERE pr_id = $1 AND reviewer_id = $2`
	result, err := tx.Exec(deleteQuery, prID, oldReviewerID)
	if err != nil {
		return fmt.Errorf("failed to remove old reviewer: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("reviewer not found in PR")
	}

	// Добавляем нового рецензента
	insertQuery := `INSERT INTO pr_reviewers (pr_id, reviewer_id) VALUES ($1, $2)`
	_, err = tx.Exec(insertQuery, prID, newReviewerID)
	if err != nil {
		return fmt.Errorf("failed to add new reviewer: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetOpenPRsWithReviewer возвращает открытые PR с указанным рецензентом
func (r *PRRepository) GetOpenPRsWithReviewer(reviewerID int) ([]*models.PullRequest, error) {
	query := `
		SELECT DISTINCT p.id, p.title, p.author_id, p.status, p.created_at, p.merged_at, p.updated_at
		FROM pull_requests p
		JOIN pr_reviewers pr ON p.id = pr.pr_id
		WHERE pr.reviewer_id = $1 AND p.status = 'OPEN'
		ORDER BY p.created_at DESC`

	rows, err := r.db.Query(query, reviewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get open PRs with reviewer: %w", err)
	}
	defer rows.Close()

	var prs []*models.PullRequest
	for rows.Next() {
		pr := &models.PullRequest{}
		if err := rows.Scan(&pr.ID, &pr.Title, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt, &pr.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan PR: %w", err)
		}

		// Получаем рецензентов
		reviewers, err := r.getReviewers(pr.ID)
		if err != nil {
			return nil, err
		}
		pr.Reviewers = reviewers

		prs = append(prs, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate PRs: %w", err)
	}

	return prs, nil
}

// getReviewers возвращает рецензентов для PR
func (r *PRRepository) getReviewers(prID int) ([]models.User, error) {
	query := `
		SELECT u.id, u.username, u.name, u.is_active, u.team_id, u.created_at, u.updated_at
		FROM users u
		JOIN pr_reviewers pr ON u.id = pr.reviewer_id
		WHERE pr.pr_id = $1
		ORDER BY u.id`

	rows, err := r.db.Query(query, prID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviewers: %w", err)
	}
	defer rows.Close()

	var reviewers []models.User
	for rows.Next() {
		var reviewer models.User
		if err := rows.Scan(&reviewer.ID, &reviewer.Username, &reviewer.Name, &reviewer.IsActive, &reviewer.TeamID, &reviewer.CreatedAt, &reviewer.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan reviewer: %w", err)
		}
		reviewers = append(reviewers, reviewer)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate reviewers: %w", err)
	}

	return reviewers, nil
}

// addReviewersTx добавляет рецензентов в транзакции
func (r *PRRepository) addReviewersTx(tx *sql.Tx, prID int, reviewers []models.User) error {
	stmt, err := tx.Prepare(`INSERT INTO pr_reviewers (pr_id, reviewer_id) VALUES ($1, $2)`)
	if err != nil {
		return fmt.Errorf("failed to prepare statement: %w", err)
	}
	defer stmt.Close()

	for _, reviewer := range reviewers {
		if _, err := stmt.Exec(prID, reviewer.ID); err != nil {
			return fmt.Errorf("failed to add reviewer %d: %w", reviewer.ID, err)
		}
	}

	return nil
}

// AddReviewers добавляет рецензентов к PR
func (r *PRRepository) AddReviewers(prID int, reviewers []models.User) error {
	if len(reviewers) == 0 {
		return nil
	}

	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	if err := r.addReviewersTx(tx, prID, reviewers); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
