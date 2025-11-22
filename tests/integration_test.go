//go:build integration

package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/pr-reviewer/internal/config"
	"github.com/user/pr-reviewer/internal/database"
	"github.com/user/pr-reviewer/internal/handler"
	"github.com/user/pr-reviewer/internal/models"
	"github.com/user/pr-reviewer/internal/service"
)

var (
	testRouter *mux.Router
	testDB     *database.DB
)

func TestMain(m *testing.M) {
	// Настройка тестового окружения
	os.Setenv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/pr_reviewer_test?sslmode=disable")

	// Определяем путь к миграциям
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		// По умолчанию ищем в родительской директории (относительно tests/)
		migrationsPath = "file://../migrations"
	} else if migrationsPath == "file://migrations" {
		// Если указан относительный путь от корня проекта, преобразуем его
		// относительно директории tests/
		migrationsPath = "file://../migrations"
	}
	os.Setenv("MIGRATIONS_PATH", migrationsPath)

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Подключение к тестовой БД
	testDB, err = database.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to test database: %v", err)
	}
	defer testDB.Close()

	// Выполнение миграций
	if err := testDB.Migrate(cfg.MigrationsPath); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Инициализация сервисов и роутера
	svc := service.New(testDB)
	logger := log.New(os.Stdout, "[TEST] ", log.LstdFlags)
	h := handler.New(svc, logger)

	testRouter = mux.NewRouter()
	h.RegisterRoutes(testRouter)

	// Запуск тестов
	code := m.Run()

	// Очистка
	cleanupTestData()

	os.Exit(code)
}

func cleanupTestData() {
	// Очистка тестовых данных
	testDB.Exec("TRUNCATE TABLE pr_reviewers, pull_requests, users, teams RESTART IDENTITY CASCADE")
}

func TestHealthCheck(t *testing.T) {
	req, _ := http.NewRequest("GET", "/health", nil)
	response := executeRequest(req)

	assert.Equal(t, http.StatusOK, response.Code)

	var result map[string]string
	err := json.NewDecoder(response.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "healthy", result["status"])
}

func TestTeamCRUD(t *testing.T) {
	// Создание команды
	teamData := models.CreateTeamRequest{Name: "Test Team"}
	body, _ := json.Marshal(teamData)

	req, _ := http.NewRequest("POST", "/teams", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	response := executeRequest(req)

	assert.Equal(t, http.StatusCreated, response.Code)

	var team models.Team
	err := json.NewDecoder(response.Body).Decode(&team)
	require.NoError(t, err)
	assert.Equal(t, "Test Team", team.Name)
	assert.NotZero(t, team.ID)

	// Получение команды по ID
	req, _ = http.NewRequest("GET", "/teams/"+strconv.Itoa(int(team.ID)), nil)
	response = executeRequest(req)
	assert.Equal(t, http.StatusOK, response.Code)

	// Получение всех команд
	req, _ = http.NewRequest("GET", "/teams", nil)
	response = executeRequest(req)
	assert.Equal(t, http.StatusOK, response.Code)

	var teams []models.Team
	err = json.NewDecoder(response.Body).Decode(&teams)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(teams), 1)
}

func TestUserCRUD(t *testing.T) {
	// Сначала создаём команду
	teamData := models.CreateTeamRequest{Name: "User Test Team"}
	body, _ := json.Marshal(teamData)
	req, _ := http.NewRequest("POST", "/teams", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	response := executeRequest(req)

	var team models.Team
	json.NewDecoder(response.Body).Decode(&team)

	// Создание пользователя
	userData := models.CreateUserRequest{
		Username: "testuser",
		Name:     "Test User",
		TeamID:   &team.ID,
	}
	body, _ = json.Marshal(userData)

	req, _ = http.NewRequest("POST", "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	response = executeRequest(req)

	assert.Equal(t, http.StatusCreated, response.Code)

	var user models.User
	err := json.NewDecoder(response.Body).Decode(&user)
	require.NoError(t, err)
	assert.Equal(t, "Test User", user.Name)
	assert.True(t, user.IsActive)
	assert.Equal(t, team.ID, *user.TeamID)

	// Обновление пользователя
	isActive := false
	updateData := models.UpdateUserRequest{
		IsActive: &isActive,
	}
	body, _ = json.Marshal(updateData)

	req, _ = http.NewRequest("PATCH", "/users/"+strconv.Itoa(int(user.ID)), bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	response = executeRequest(req)

	assert.Equal(t, http.StatusOK, response.Code)

	var updatedUser models.User
	err = json.NewDecoder(response.Body).Decode(&updatedUser)
	require.NoError(t, err)
	assert.False(t, updatedUser.IsActive)
}

func TestPullRequestFlow(t *testing.T) {
	// Создание команды
	teamData := models.CreateTeamRequest{Name: "PR Test Team"}
	body, _ := json.Marshal(teamData)
	req, _ := http.NewRequest("POST", "/teams", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	response := executeRequest(req)

	var team models.Team
	json.NewDecoder(response.Body).Decode(&team)

	// Создание автора
	authorData := models.CreateUserRequest{
		Username: "prauthor",
		Name:     "PR Author",
		TeamID:   &team.ID,
	}
	body, _ = json.Marshal(authorData)
	req, _ = http.NewRequest("POST", "/users", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	response = executeRequest(req)

	var author models.User
	json.NewDecoder(response.Body).Decode(&author)

	// Создание потенциальных рецензентов
	for i := 1; i <= 3; i++ {
		reviewerData := models.CreateUserRequest{
			Username: fmt.Sprintf("reviewer%d", i),
			Name:     fmt.Sprintf("Reviewer %d", i),
			TeamID:   &team.ID,
		}
		body, _ = json.Marshal(reviewerData)
		req, _ = http.NewRequest("POST", "/users", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		executeRequest(req)
	}

	// Создание Pull Request
	prData := models.CreatePullRequestRequest{
		Title:    "Test PR",
		AuthorID: author.ID,
	}
	body, _ = json.Marshal(prData)

	req, _ = http.NewRequest("POST", "/pull-requests", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	response = executeRequest(req)

	assert.Equal(t, http.StatusCreated, response.Code)

	var pr models.PullRequest
	err := json.NewDecoder(response.Body).Decode(&pr)
	require.NoError(t, err)
	assert.Equal(t, "Test PR", pr.Title)
	assert.Equal(t, models.PRStatusOpen, pr.Status)
	assert.LessOrEqual(t, len(pr.Reviewers), 2) // Максимум 2 рецензента

	// Merge Pull Request
	req, _ = http.NewRequest("POST", "/pull-requests/"+strconv.Itoa(int(pr.ID))+"/merge", nil)
	response = executeRequest(req)

	assert.Equal(t, http.StatusOK, response.Code)

	var mergedPR models.PullRequest
	err = json.NewDecoder(response.Body).Decode(&mergedPR)
	require.NoError(t, err)
	assert.Equal(t, models.PRStatusMerged, mergedPR.Status)
	assert.NotNil(t, mergedPR.MergedAt)

	// Проверка идемпотентности merge
	req, _ = http.NewRequest("POST", "/pull-requests/"+strconv.Itoa(int(pr.ID))+"/merge", nil)
	response = executeRequest(req)
	assert.Equal(t, http.StatusOK, response.Code)
}

func TestStatistics(t *testing.T) {
	req, _ := http.NewRequest("GET", "/statistics", nil)
	response := executeRequest(req)

	assert.Equal(t, http.StatusOK, response.Code)

	var stats models.Statistics
	err := json.NewDecoder(response.Body).Decode(&stats)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, stats.TotalPRs, 0)
}

func executeRequest(req *http.Request) *httptest.ResponseRecorder {
	rr := httptest.NewRecorder()
	testRouter.ServeHTTP(rr, req)
	return rr
}
