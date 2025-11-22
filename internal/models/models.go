package models

import (
	"database/sql/driver"
	"strings"
	"time"
)

// User представляет пользователя системы
type User struct {
	ID        int       `json:"id" db:"id"`
	Username  string    `json:"username" db:"username"`
	Name      string    `json:"name" db:"name"`
	IsActive  bool      `json:"isActive" db:"is_active"`
	TeamID    *int      `json:"teamId,omitempty" db:"team_id"`
	Teams     []Team    `json:"teams,omitempty"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// Team представляет команду
type Team struct {
	ID        int       `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt time.Time `json:"updatedAt" db:"updated_at"`
}

// PRStatus представляет статус Pull Request
type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
	PRStatusClosed PRStatus = "CLOSED"
)

// Scan implements the Scanner interface for PRStatus
func (s *PRStatus) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	switch v := value.(type) {
	case string:
		*s = PRStatus(v)
	case []byte:
		*s = PRStatus(v)
	default:
		*s = PRStatus("")
	}
	return nil
}

// Value implements the driver Valuer interface for PRStatus
func (s PRStatus) Value() (driver.Value, error) {
	return string(s), nil
}

// MarshalJSON converts PRStatus to lowercase JSON
func (s PRStatus) MarshalJSON() ([]byte, error) {
	return []byte(`"` + strings.ToLower(string(s)) + `"`), nil
}

// UnmarshalJSON converts lowercase JSON to uppercase PRStatus
func (s *PRStatus) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)
	*s = PRStatus(strings.ToUpper(str))
	return nil
}

// PullRequest представляет Pull Request
type PullRequest struct {
	ID        int        `json:"id" db:"id"`
	Title     string     `json:"title" db:"title"`
	AuthorID  int        `json:"authorId" db:"author_id"`
	Author    *User      `json:"author,omitempty"`
	Team      *Team      `json:"team,omitempty"`
	Status    PRStatus   `json:"status" db:"status"`
	Reviewers []User     `json:"reviewers"`
	CreatedAt time.Time  `json:"createdAt" db:"created_at"`
	MergedAt  *time.Time `json:"mergedAt,omitempty" db:"merged_at"`
	UpdatedAt time.Time  `json:"updatedAt" db:"updated_at"`
}

// PRReviewer представляет связь между PR и рецензентом
type PRReviewer struct {
	PRID       int `db:"pr_id"`
	ReviewerID int `db:"reviewer_id"`
}

// CreateTeamRequest запрос на создание команды
type CreateTeamRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

// CreateUserRequest запрос на создание пользователя
type CreateUserRequest struct {
	Username string `json:"username" validate:"required,min=1,max=100"`
	Name     string `json:"name" validate:"required,min=1,max=100"`
	TeamID   *int   `json:"teamId,omitempty"`
}

// UpdateUserRequest запрос на обновление пользователя
type UpdateUserRequest struct {
	Name     *string `json:"name,omitempty" validate:"omitempty,min=1,max=100"`
	IsActive *bool   `json:"isActive,omitempty"`
}

// CreatePullRequestRequest запрос на создание PR
type CreatePullRequestRequest struct {
	Title    string `json:"title" validate:"required,min=1,max=255"`
	AuthorID int    `json:"authorId" validate:"required,min=1"`
}

// ReassignReviewerRequest запрос на переназначение рецензента
type ReassignReviewerRequest struct {
	OldReviewerID int `json:"oldReviewerId" validate:"required,min=1"`
}

// BulkDeactivateRequest запрос на массовую деактивацию пользователей
type BulkDeactivateRequest struct {
	UserIDs []int `json:"userIds" validate:"required,min=1"`
}

// BulkDeactivateResponse ответ на массовую деактивацию
type BulkDeactivateResponse struct {
	DeactivatedCount  int `json:"deactivatedCount"`
	ReassignedPRCount int `json:"reassignedPRCount"`
}

// Statistics статистика назначений
type Statistics struct {
	TotalPRs  int             `json:"totalPRs"`
	OpenPRs   int             `json:"openPRs"`
	MergedPRs int             `json:"mergedPRs"`
	ClosedPRs int             `json:"closedPRs"`
	UserStats []UserStatistic `json:"userStats"`
	TeamStats []TeamStatistic `json:"teamStats"`
}

// UserStatistic статистика по пользователю
type UserStatistic struct {
	UserID          int    `json:"userId" db:"user_id"`
	UserName        string `json:"userName" db:"user_name"`
	AssignmentCount int    `json:"assignmentCount" db:"assignment_count"`
}

// TeamStatistic статистика по команде
type TeamStatistic struct {
	TeamID   int    `json:"teamId" db:"team_id"`
	TeamName string `json:"teamName" db:"team_name"`
	PRCount  int    `json:"prCount" db:"pr_count"`
}

// HealthResponse ответ health check
type HealthResponse struct {
	Status string `json:"status"`
}
