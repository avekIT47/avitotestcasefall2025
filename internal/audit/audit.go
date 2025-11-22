package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"github.com/user/pr-reviewer/internal/logger"
)

// Action тип действия в audit log
type Action string

const (
	ActionCreate Action = "create"
	ActionUpdate Action = "update"
	ActionDelete Action = "delete"
	ActionRead   Action = "read"
	ActionLogin  Action = "login"
	ActionLogout Action = "logout"
)

// Entity тип сущности
type Entity string

const (
	EntityUser        Entity = "user"
	EntityTeam        Entity = "team"
	EntityPullRequest Entity = "pull_request"
	EntityReviewer    Entity = "reviewer"
)

// Entry запись в audit log
type Entry struct {
	ID          int64                  `json:"id"`
	Timestamp   time.Time              `json:"timestamp"`
	Action      Action                 `json:"action"`
	Entity      Entity                 `json:"entity"`
	EntityID    int64                  `json:"entity_id"`
	UserID      int64                  `json:"user_id,omitempty"`
	UserEmail   string                 `json:"user_email,omitempty"`
	IP          string                 `json:"ip"`
	UserAgent   string                 `json:"user_agent,omitempty"`
	RequestID   string                 `json:"request_id,omitempty"`
	Changes     map[string]interface{} `json:"changes,omitempty"`
	Description string                 `json:"description,omitempty"`
}

// Logger логирует действия пользователей
type Logger struct {
	db     *sql.DB
	logger *logger.Logger
}

// NewLogger создает новый audit logger
func NewLogger(db *sql.DB, log *logger.Logger) *Logger {
	return &Logger{
		db:     db,
		logger: log,
	}
}

// Log записывает действие в audit log
func (l *Logger) Log(ctx context.Context, entry *Entry) error {
	entry.Timestamp = time.Now()

	// Сериализуем changes
	changesJSON, err := json.Marshal(entry.Changes)
	if err != nil {
		l.logger.Errorw("Failed to marshal audit changes", "error", err)
		changesJSON = []byte("{}")
	}

	query := `
		INSERT INTO audit_logs (
			timestamp, action, entity, entity_id, user_id, user_email,
			ip, user_agent, request_id, changes, description
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`

	err = l.db.QueryRowContext(ctx, query,
		entry.Timestamp,
		entry.Action,
		entry.Entity,
		entry.EntityID,
		sql.NullInt64{Int64: entry.UserID, Valid: entry.UserID > 0},
		entry.UserEmail,
		entry.IP,
		entry.UserAgent,
		entry.RequestID,
		changesJSON,
		entry.Description,
	).Scan(&entry.ID)

	if err != nil {
		l.logger.Errorw("Failed to write audit log",
			"action", entry.Action,
			"entity", entry.Entity,
			"error", err,
		)
		return err
	}

	l.logger.Debugw("Audit log entry created",
		"id", entry.ID,
		"action", entry.Action,
		"entity", entry.Entity,
		"entity_id", entry.EntityID,
		"user_id", entry.UserID,
	)

	return nil
}

// Query возвращает записи audit log
func (l *Logger) Query(ctx context.Context, filter Filter) ([]*Entry, error) {
	query := `
		SELECT id, timestamp, action, entity, entity_id, 
		       COALESCE(user_id, 0), user_email, ip, user_agent, 
		       request_id, changes, description
		FROM audit_logs
		WHERE 1=1
	`
	args := []interface{}{}
	argNum := 1

	if filter.UserID > 0 {
		query += ` AND user_id = $` + string(rune(argNum+'0'))
		args = append(args, filter.UserID)
		argNum++
	}

	if filter.Entity != "" {
		query += ` AND entity = $` + string(rune(argNum+'0'))
		args = append(args, filter.Entity)
		argNum++
	}

	if filter.Action != "" {
		query += ` AND action = $` + string(rune(argNum+'0'))
		args = append(args, filter.Action)
		argNum++
	}

	if !filter.From.IsZero() {
		query += ` AND timestamp >= $` + string(rune(argNum+'0'))
		args = append(args, filter.From)
		argNum++
	}

	if !filter.To.IsZero() {
		query += ` AND timestamp <= $` + string(rune(argNum+'0'))
		args = append(args, filter.To)
		argNum++
	}

	query += ` ORDER BY timestamp DESC LIMIT $` + string(rune(argNum+'0'))
	args = append(args, filter.Limit)

	rows, err := l.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := []*Entry{}
	for rows.Next() {
		entry := &Entry{}
		var changesJSON []byte
		var userID sql.NullInt64

		err := rows.Scan(
			&entry.ID,
			&entry.Timestamp,
			&entry.Action,
			&entry.Entity,
			&entry.EntityID,
			&userID,
			&entry.UserEmail,
			&entry.IP,
			&entry.UserAgent,
			&entry.RequestID,
			&changesJSON,
			&entry.Description,
		)
		if err != nil {
			return nil, err
		}

		if userID.Valid {
			entry.UserID = userID.Int64
		}

		if len(changesJSON) > 0 {
			json.Unmarshal(changesJSON, &entry.Changes)
		}

		entries = append(entries, entry)
	}

	return entries, rows.Err()
}

// Filter фильтр для запросов audit log
type Filter struct {
	UserID int64
	Entity Entity
	Action Action
	From   time.Time
	To     time.Time
	Limit  int
}

// Helper функции для логирования различных действий

// LogUserCreated логирует создание пользователя
func (l *Logger) LogUserCreated(ctx context.Context, userID int64, actorID int64, ip string) error {
	return l.Log(ctx, &Entry{
		Action:      ActionCreate,
		Entity:      EntityUser,
		EntityID:    userID,
		UserID:      actorID,
		IP:          ip,
		Description: "User created",
	})
}

// LogUserUpdated логирует обновление пользователя
func (l *Logger) LogUserUpdated(ctx context.Context, userID int64, actorID int64, changes map[string]interface{}, ip string) error {
	return l.Log(ctx, &Entry{
		Action:      ActionUpdate,
		Entity:      EntityUser,
		EntityID:    userID,
		UserID:      actorID,
		IP:          ip,
		Changes:     changes,
		Description: "User updated",
	})
}

// LogUserDeactivated логирует деактивацию пользователя
func (l *Logger) LogUserDeactivated(ctx context.Context, userID int64, actorID int64, ip string) error {
	return l.Log(ctx, &Entry{
		Action:      ActionUpdate,
		Entity:      EntityUser,
		EntityID:    userID,
		UserID:      actorID,
		IP:          ip,
		Changes:     map[string]interface{}{"is_active": false},
		Description: "User deactivated",
	})
}

// LogPRCreated логирует создание PR
func (l *Logger) LogPRCreated(ctx context.Context, prID int64, authorID int64, ip string) error {
	return l.Log(ctx, &Entry{
		Action:      ActionCreate,
		Entity:      EntityPullRequest,
		EntityID:    prID,
		UserID:      authorID,
		IP:          ip,
		Description: "Pull request created",
	})
}

// LogReviewerAssigned логирует назначение рецензента
func (l *Logger) LogReviewerAssigned(ctx context.Context, prID int64, reviewerID int64, actorID int64, ip string) error {
	return l.Log(ctx, &Entry{
		Action:   ActionUpdate,
		Entity:   EntityPullRequest,
		EntityID: prID,
		UserID:   actorID,
		IP:       ip,
		Changes: map[string]interface{}{
			"reviewer_id": reviewerID,
		},
		Description: "Reviewer assigned",
	})
}
