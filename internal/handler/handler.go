package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/user/pr-reviewer/internal/logger"
	"github.com/user/pr-reviewer/internal/models"
	"github.com/user/pr-reviewer/internal/service"
)

// Handler обрабатывает HTTP запросы
type Handler struct {
	service *service.Service
	logger  interface{} // Can be either *log.Logger or *logger.Logger
}

// New создаёт новый HTTP handler
// logger can be either *log.Logger (stdlib) or *logger.Logger (custom)
func New(service *service.Service, logger interface{}) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// logf is a helper to handle both standard and custom loggers
func (h *Handler) logf(format string, args ...interface{}) {
	switch l := h.logger.(type) {
	case *log.Logger:
		l.Printf(format, args...)
	case *logger.Logger:
		l.Infof(format, args...)
	}
}

// RegisterRoutes регистрирует все маршруты
func (h *Handler) RegisterRoutes(router *mux.Router) {
	// Middleware
	router.Use(h.loggingMiddleware)

	// Health check
	router.HandleFunc("/health", h.HealthCheck).Methods("GET")

	// Teams
	router.HandleFunc("/teams", h.GetTeams).Methods("GET")
	router.HandleFunc("/teams", h.CreateTeam).Methods("POST")
	router.HandleFunc("/teams/{teamId}", h.GetTeam).Methods("GET")
	router.HandleFunc("/teams/{teamId}", h.DeleteTeam).Methods("DELETE")
	router.HandleFunc("/teams/{teamId}/users", h.AddUserToTeam).Methods("POST")
	router.HandleFunc("/teams/{teamId}/users", h.RemoveUserFromTeam).Methods("DELETE")
	router.HandleFunc("/teams/{teamId}/users/deactivate", h.BulkDeactivateUsers).Methods("POST")

	// Users
	router.HandleFunc("/users", h.GetUsers).Methods("GET")
	router.HandleFunc("/users", h.CreateUser).Methods("POST")
	router.HandleFunc("/users/{userId}", h.GetUser).Methods("GET")
	router.HandleFunc("/users/{userId}", h.UpdateUser).Methods("PATCH")

	// Pull Requests
	router.HandleFunc("/pull-requests", h.GetPullRequests).Methods("GET")
	router.HandleFunc("/pull-requests", h.CreatePullRequest).Methods("POST")
	router.HandleFunc("/pull-requests/{prId}", h.GetPullRequest).Methods("GET")
	router.HandleFunc("/pull-requests/{prId}/reviewers", h.AddReviewer).Methods("POST")
	router.HandleFunc("/pull-requests/{prId}/reviewers", h.ReassignReviewer).Methods("PUT")
	router.HandleFunc("/pull-requests/{prId}/merge", h.MergePullRequest).Methods("POST")
	router.HandleFunc("/pull-requests/{prId}/close", h.ClosePullRequest).Methods("POST")

	// Statistics
	router.HandleFunc("/statistics", h.GetStatistics).Methods("GET")
}

// loggingMiddleware логирует все запросы
func (h *Handler) loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.logf("[%s] %s %s", r.Method, r.URL.Path, r.RemoteAddr)
		next.ServeHTTP(w, r)
	})
}

// HealthCheck обрабатывает запрос проверки здоровья
func (h *Handler) HealthCheck(w http.ResponseWriter, _ *http.Request) {
	response := models.HealthResponse{
		Status: "healthy",
	}
	h.sendJSON(w, http.StatusOK, response)
}

// GetTeams возвращает все команды
func (h *Handler) GetTeams(w http.ResponseWriter, _ *http.Request) {
	teams, err := h.service.GetAllTeams()
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to get teams")
		return
	}
	h.sendJSON(w, http.StatusOK, teams)
}

// CreateTeam создаёт новую команду
func (h *Handler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTeamRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	team, err := h.service.CreateTeam(&req)
	if err != nil {
		if err.Error() == "team with name '"+req.Name+"' already exists" {
			h.sendError(w, http.StatusConflict, err.Error())
		} else {
			h.sendError(w, http.StatusInternalServerError, "Failed to create team")
		}
		return
	}

	h.sendJSON(w, http.StatusCreated, team)
}

// GetTeam возвращает команду по ID
func (h *Handler) GetTeam(w http.ResponseWriter, r *http.Request) {
	h.handleGetByID(w, r, "teamId", func(id int) (interface{}, error) {
		return h.service.GetTeam(id)
	}, "Team not found")
}

// DeleteTeam удаляет команду по ID
func (h *Handler) DeleteTeam(w http.ResponseWriter, r *http.Request) {
	teamID, err := h.getIntParam(r, "teamId")
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid team ID")
		return
	}

	if err := h.service.DeleteTeam(teamID); err != nil {
		if err.Error() == "team not found" {
			h.sendError(w, http.StatusNotFound, "Team not found")
		} else {
			h.sendError(w, http.StatusInternalServerError, "Failed to delete team")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// AddUserToTeam добавляет пользователя в команду
func (h *Handler) AddUserToTeam(w http.ResponseWriter, r *http.Request) {
	teamID, err := h.getIntParam(r, "teamId")
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid team ID")
		return
	}

	var req struct {
		UserID int `json:"userId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if err := h.service.AddUserToTeam(teamID, req.UserID); err != nil {
		if err.Error() == "user already in team" {
			h.sendError(w, http.StatusConflict, err.Error())
		} else if err.Error() == "team not found" || err.Error() == "user not found" {
			h.sendError(w, http.StatusNotFound, err.Error())
		} else {
			h.sendError(w, http.StatusInternalServerError, "Failed to add user to team")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveUserFromTeam удаляет пользователя из команды
func (h *Handler) RemoveUserFromTeam(w http.ResponseWriter, r *http.Request) {
	teamID, err := h.getIntParam(r, "teamId")
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid team ID")
		return
	}

	userID, err := h.getIntQuery(r, "userId")
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	if err := h.service.RemoveUserFromTeam(teamID, userID); err != nil {
		if err.Error() == "user not found in team" {
			h.sendError(w, http.StatusNotFound, err.Error())
		} else {
			h.sendError(w, http.StatusInternalServerError, "Failed to remove user from team")
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// BulkDeactivateUsers массово деактивирует пользователей
func (h *Handler) BulkDeactivateUsers(w http.ResponseWriter, r *http.Request) {
	teamID, err := h.getIntParam(r, "teamId")
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid team ID")
		return
	}

	var req models.BulkDeactivateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	response, err := h.service.BulkDeactivateUsers(teamID, &req)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to deactivate users")
		return
	}

	h.sendJSON(w, http.StatusOK, response)
}

// GetUsers возвращает всех пользователей
func (h *Handler) GetUsers(w http.ResponseWriter, r *http.Request) {
	var teamID *int
	var isActive *bool

	if id, err := h.getIntQuery(r, "teamId"); err == nil {
		teamID = &id
	}

	if active, err := h.getBoolQuery(r, "isActive"); err == nil {
		isActive = &active
	}

	users, err := h.service.GetAllUsers(teamID, isActive)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to get users")
		return
	}

	h.sendJSON(w, http.StatusOK, users)
}

// CreateUser создаёт нового пользователя
func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req models.CreateUserRequest
	h.handleCreateEntity(w, r, &req, func() (interface{}, error) {
		return h.service.CreateUser(&req)
	}, map[string]int{"not found": http.StatusNotFound})
}

// GetUser возвращает пользователя по ID
func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) {
	h.handleGetByID(w, r, "userId", func(id int) (interface{}, error) {
		return h.service.GetUser(id)
	}, "User not found")
}

// UpdateUser обновляет пользователя
func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID, err := h.getIntParam(r, "userId")
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid user ID")
		return
	}

	var req models.UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	user, err := h.service.UpdateUser(userID, &req)
	if err != nil {
		if err.Error() == "user not found" {
			h.sendError(w, http.StatusNotFound, "User not found")
		} else {
			h.sendError(w, http.StatusInternalServerError, "Failed to update user")
		}
		return
	}

	h.sendJSON(w, http.StatusOK, user)
}

// GetPullRequests возвращает все PR
func (h *Handler) GetPullRequests(w http.ResponseWriter, r *http.Request) {
	var userID *int
	var authorID *int
	var status *string

	if id, err := h.getIntQuery(r, "userId"); err == nil {
		userID = &id
	}

	if id, err := h.getIntQuery(r, "authorId"); err == nil {
		authorID = &id
	}

	if s := r.URL.Query().Get("status"); s != "" {
		// Преобразуем статус в uppercase для совместимости с БД
		uppercaseStatus := strings.ToUpper(s)
		status = &uppercaseStatus
	}

	prs, err := h.service.GetAllPullRequests(userID, authorID, status)
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to get pull requests")
		return
	}

	h.sendJSON(w, http.StatusOK, prs)
}

// CreatePullRequest создаёт новый PR
func (h *Handler) CreatePullRequest(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePullRequestRequest
	h.handleCreateEntity(w, r, &req, func() (interface{}, error) {
		return h.service.CreatePullRequest(&req)
	}, map[string]int{"not found": http.StatusNotFound})
}

// GetPullRequest возвращает PR по ID
func (h *Handler) GetPullRequest(w http.ResponseWriter, r *http.Request) {
	h.handleGetByID(w, r, "prId", func(id int) (interface{}, error) {
		return h.service.GetPullRequest(id)
	}, "Pull request not found")
}

// AddReviewer добавляет нового рецензента к PR
func (h *Handler) AddReviewer(w http.ResponseWriter, r *http.Request) {
	prID, err := h.getIntParam(r, "prId")
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid PR ID")
		return
	}

	var req struct {
		ReviewerID int `json:"reviewerId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	pr, err := h.service.AddReviewer(prID, req.ReviewerID)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.sendError(w, http.StatusNotFound, err.Error())
		} else if strings.Contains(err.Error(), "cannot") || strings.Contains(err.Error(), "already") || strings.Contains(err.Error(), "author") {
			h.sendError(w, http.StatusBadRequest, err.Error())
		} else {
			h.sendError(w, http.StatusInternalServerError, "Failed to add reviewer")
		}
		return
	}

	h.sendJSON(w, http.StatusOK, pr)
}

// ReassignReviewer переназначает рецензента
func (h *Handler) ReassignReviewer(w http.ResponseWriter, r *http.Request) {
	prID, err := h.getIntParam(r, "prId")
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid PR ID")
		return
	}

	var req models.ReassignReviewerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	pr, err := h.service.ReassignReviewer(prID, &req)
	if err != nil {
		if err.Error() == "PR not found" || err.Error() == "reviewer not found in PR" {
			h.sendError(w, http.StatusNotFound, err.Error())
		} else if err.Error() == "cannot change reviewers of merged PR" {
			h.sendError(w, http.StatusBadRequest, err.Error())
		} else {
			h.sendError(w, http.StatusInternalServerError, "Failed to reassign reviewer")
		}
		return
	}

	h.sendJSON(w, http.StatusOK, pr)
}

// MergePullRequest переводит PR в состояние MERGED
func (h *Handler) MergePullRequest(w http.ResponseWriter, r *http.Request) {
	h.handleUpdateEntity(w, r, "prId", func(id int) (interface{}, error) {
		return h.service.MergePullRequest(id)
	}, "Pull request not found", "Failed to merge pull request")
}

// ClosePullRequest переводит PR в состояние CLOSED (закрыт без мерджа)
func (h *Handler) ClosePullRequest(w http.ResponseWriter, r *http.Request) {
	h.handleUpdateEntity(w, r, "prId", func(id int) (interface{}, error) {
		return h.service.ClosePullRequest(id)
	}, "Pull request not found or already closed/merged", "Failed to close pull request")
}

// GetStatistics возвращает статистику
func (h *Handler) GetStatistics(w http.ResponseWriter, _ *http.Request) {
	stats, err := h.service.GetStatistics()
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "Failed to get statistics")
		return
	}

	h.sendJSON(w, http.StatusOK, stats)
}

// Helper methods

func (h *Handler) getIntParam(r *http.Request, name string) (int, error) {
	vars := mux.Vars(r)
	return strconv.Atoi(vars[name])
}

func (h *Handler) getIntQuery(r *http.Request, name string) (int, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return 0, http.ErrNotSupported
	}
	return strconv.Atoi(value)
}

func (h *Handler) getBoolQuery(r *http.Request, name string) (bool, error) {
	value := r.URL.Query().Get(name)
	if value == "" {
		return false, http.ErrNotSupported
	}
	return strconv.ParseBool(value)
}

func (h *Handler) sendJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logf("Failed to encode response: %v", err)
	}
}

func (h *Handler) sendError(w http.ResponseWriter, status int, message string) {
	response := map[string]string{"error": message}
	h.sendJSON(w, status, response)
}

// handleGetByID обрабатывает запросы получения сущности по ID
func (h *Handler) handleGetByID(w http.ResponseWriter, r *http.Request, paramName string, getFunc func(int) (interface{}, error), notFoundMsg string) {
	id, err := h.getIntParam(r, paramName)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid "+paramName)
		return
	}

	entity, err := getFunc(id)
	if err != nil {
		if err.Error() == "team not found" || err.Error() == "user not found" || err.Error() == "PR not found" {
			h.sendError(w, http.StatusNotFound, notFoundMsg)
		} else {
			h.sendError(w, http.StatusInternalServerError, "Failed to get "+paramName)
		}
		return
	}

	h.sendJSON(w, http.StatusOK, entity)
}

// handleCreateEntity обрабатывает запросы создания сущности
func (h *Handler) handleCreateEntity(w http.ResponseWriter, r *http.Request, req interface{}, createFunc func() (interface{}, error), errorMap map[string]int) {
	if err := json.NewDecoder(r.Body).Decode(req); err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	entity, err := createFunc()
	if err != nil {
		for errMsg, status := range errorMap {
			if strings.Contains(err.Error(), errMsg) {
				h.sendError(w, status, err.Error())
				return
			}
		}
		h.sendError(w, http.StatusInternalServerError, "Failed to create entity")
		return
	}

	h.sendJSON(w, http.StatusCreated, entity)
}

// handleUpdateEntity обрабатывает запросы обновления PR
func (h *Handler) handleUpdateEntity(w http.ResponseWriter, r *http.Request, idParamName string, updateFunc func(int) (interface{}, error), notFoundMsg, errorMsg string) {
	id, err := h.getIntParam(r, idParamName)
	if err != nil {
		h.sendError(w, http.StatusBadRequest, "Invalid "+idParamName)
		return
	}

	entity, err := updateFunc(id)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			h.sendError(w, http.StatusNotFound, notFoundMsg)
		} else {
			h.sendError(w, http.StatusInternalServerError, errorMsg)
		}
		return
	}

	h.sendJSON(w, http.StatusOK, entity)
}
