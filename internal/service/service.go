package service

import (
	"fmt"
	"math/rand"

	"github.com/user/pr-reviewer/internal/database"
	"github.com/user/pr-reviewer/internal/models"
	"github.com/user/pr-reviewer/internal/repository"
)

const (
	errTeamNotFound = "team not found"
	errUserNotFound = "user not found"
	errPRNotFound   = "PR not found"
)

// Service предоставляет бизнес-логику приложения
type Service struct {
	teamRepo  *repository.TeamRepository
	userRepo  *repository.UserRepository
	prRepo    *repository.PRRepository
	statsRepo *repository.StatisticsRepository
}

// New создаёт новый экземпляр сервиса
func New(db *database.DB) *Service {
	return &Service{
		teamRepo:  repository.NewTeamRepository(db),
		userRepo:  repository.NewUserRepository(db),
		prRepo:    repository.NewPRRepository(db),
		statsRepo: repository.NewStatisticsRepository(db),
	}
}

// CreateTeam создаёт новую команду
func (s *Service) CreateTeam(req *models.CreateTeamRequest) (*models.Team, error) {
	// Проверяем, не существует ли уже команда с таким именем
	existing, err := s.teamRepo.GetByName(req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check team existence: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("team with name '%s' already exists", req.Name)
	}

	team := &models.Team{
		Name: req.Name,
	}

	if err := s.teamRepo.Create(team); err != nil {
		return nil, err
	}

	return team, nil
}

// GetTeam возвращает команду по ID
func (s *Service) GetTeam(id int) (*models.Team, error) {
	return s.teamRepo.GetByID(id)
}

// GetAllTeams возвращает все команды
func (s *Service) GetAllTeams() ([]*models.Team, error) {
	return s.teamRepo.GetAll()
}

// DeleteTeam удаляет команду по ID
func (s *Service) DeleteTeam(id int) error {
	// Проверяем существование команды
	if _, err := s.teamRepo.GetByID(id); err != nil {
		return fmt.Errorf(errTeamNotFound)
	}

	return s.teamRepo.Delete(id)
}

// CreateUser создаёт нового пользователя
func (s *Service) CreateUser(req *models.CreateUserRequest) (*models.User, error) {
	user := &models.User{
		Username: req.Username,
		Name:     req.Name,
		IsActive: true,
		TeamID:   req.TeamID,
	}

	// Проверяем существование команды, если указана
	var team *models.Team
	if req.TeamID != nil {
		t, err := s.teamRepo.GetByID(*req.TeamID)
		if err != nil {
			return nil, fmt.Errorf(errTeamNotFound)
		}
		team = t
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Обогащаем пользователя информацией о команде
	if team != nil {
		user.Teams = []models.Team{*team}
	}

	return user, nil
}

// GetUser возвращает пользователя по ID
func (s *Service) GetUser(id int) (*models.User, error) {
	user, err := s.userRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Обогащаем пользователя информацией о команде
	if user.TeamID != nil {
		team, err := s.teamRepo.GetByID(*user.TeamID)
		if err == nil {
			user.Teams = []models.Team{*team}
		}
	}

	return user, nil
}

// GetAllUsers возвращает всех пользователей с фильтрами
func (s *Service) GetAllUsers(teamID *int, isActive *bool) ([]*models.User, error) {
	users, err := s.userRepo.GetAll(teamID, isActive)
	if err != nil {
		return nil, err
	}

	// Обогащаем пользователей информацией о командах
	for i, user := range users {
		if user.TeamID != nil {
			team, err := s.teamRepo.GetByID(*user.TeamID)
			if err == nil {
				users[i].Teams = []models.Team{*team}
			}
		}
	}

	return users, nil
}

// UpdateUser обновляет пользователя
func (s *Service) UpdateUser(id int, req *models.UpdateUserRequest) (*models.User, error) {
	user, err := s.userRepo.Update(id, req)
	if err != nil {
		return nil, err
	}

	// Обогащаем пользователя информацией о команде
	if user.TeamID != nil {
		team, err := s.teamRepo.GetByID(*user.TeamID)
		if err == nil {
			user.Teams = []models.Team{*team}
		}
	}

	return user, nil
}

// AddUserToTeam добавляет пользователя в команду
func (s *Service) AddUserToTeam(teamID, userID int) error {
	// Проверяем существование команды
	if _, err := s.teamRepo.GetByID(teamID); err != nil {
		return fmt.Errorf(errTeamNotFound)
	}

	// Проверяем существование пользователя
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return fmt.Errorf(errUserNotFound)
	}

	// Проверяем, не находится ли пользователь уже в команде
	if user.TeamID != nil && *user.TeamID == teamID {
		return fmt.Errorf("user already in team")
	}

	return s.teamRepo.AddUser(teamID, userID)
}

// RemoveUserFromTeam удаляет пользователя из команды
func (s *Service) RemoveUserFromTeam(teamID, userID int) error {
	return s.teamRepo.RemoveUser(teamID, userID)
}

// CreatePullRequest создаёт новый PR и автоматически назначает рецензентов
func (s *Service) CreatePullRequest(req *models.CreatePullRequestRequest) (*models.PullRequest, error) {
	// Проверяем существование автора
	author, err := s.userRepo.GetByID(req.AuthorID)
	if err != nil {
		return nil, fmt.Errorf("author not found")
	}

	pr := &models.PullRequest{
		Title:    req.Title,
		AuthorID: req.AuthorID,
		Status:   models.PRStatusOpen,
	}

	// Автоматически назначаем рецензентов, если автор в команде
	if author.TeamID != nil {
		reviewers, err := s.selectReviewers(*author.TeamID, author.ID, 2)
		if err != nil {
			// Не блокируем создание PR, если не удалось выбрать рецензентов
			// Просто создаём PR без рецензентов
		} else {
			pr.Reviewers = reviewers
		}

		// Загружаем команду
		team, err := s.teamRepo.GetByID(*author.TeamID)
		if err == nil {
			pr.Team = team
			author.Teams = []models.Team{*team}
		}
	}

	if err := s.prRepo.Create(pr); err != nil {
		return nil, err
	}

	// Обогащаем PR автором
	pr.Author = author

	// Обогащаем рецензентов информацией о командах
	for i := range pr.Reviewers {
		if pr.Reviewers[i].TeamID != nil {
			team, err := s.teamRepo.GetByID(*pr.Reviewers[i].TeamID)
			if err == nil {
				pr.Reviewers[i].Teams = []models.Team{*team}
			}
		}
	}

	return pr, nil
}

// GetPullRequest возвращает PR по ID
func (s *Service) GetPullRequest(id int) (*models.PullRequest, error) {
	pr, err := s.prRepo.GetByID(id)
	if err != nil {
		return nil, err
	}

	s.enrichPR(pr)
	return pr, nil
}

// GetAllPullRequests возвращает все PR с фильтрами
func (s *Service) GetAllPullRequests(userID *int, authorID *int, status *string) ([]*models.PullRequest, error) {
	prs, err := s.prRepo.GetAll(userID, authorID, status)
	if err != nil {
		return nil, err
	}

	// Обогащаем каждый PR информацией об авторе и команде
	for i, pr := range prs {
		// Загружаем автора
		author, err := s.userRepo.GetByID(pr.AuthorID)
		if err == nil {
			// Загружаем команду автора
			if author.TeamID != nil {
				team, err := s.teamRepo.GetByID(*author.TeamID)
				if err == nil {
					author.Teams = []models.Team{*team}
					prs[i].Team = team
				}
			}
			prs[i].Author = author
		}

		// Обогащаем рецензентов информацией о командах
		for j := range prs[i].Reviewers {
			if prs[i].Reviewers[j].TeamID != nil {
				team, err := s.teamRepo.GetByID(*prs[i].Reviewers[j].TeamID)
				if err == nil {
					prs[i].Reviewers[j].Teams = []models.Team{*team}
				}
			}
		}
	}

	return prs, nil
}

// MergePullRequest переводит PR в состояние MERGED (идемпотентная операция)
func (s *Service) MergePullRequest(id int) (*models.PullRequest, error) {
	pr, err := s.prRepo.Merge(id)
	if err != nil {
		return nil, err
	}

	s.enrichPR(pr)
	return pr, nil
}

// ClosePullRequest переводит PR в состояние CLOSED (закрыт без мерджа)
func (s *Service) ClosePullRequest(id int) (*models.PullRequest, error) {
	pr, err := s.prRepo.Close(id)
	if err != nil {
		return nil, err
	}

	s.enrichPR(pr)
	return pr, nil
}

// AddReviewer добавляет нового рецензента к PR
func (s *Service) AddReviewer(prID int, reviewerID int) (*models.PullRequest, error) {
	// Получаем PR
	pr, err := s.prRepo.GetByID(prID)
	if err != nil {
		return nil, err
	}

	// Проверяем, что PR не в статусе MERGED или CLOSED
	if pr.Status == models.PRStatusMerged || pr.Status == models.PRStatusClosed {
		return nil, fmt.Errorf("cannot add reviewers to merged or closed PR")
	}

	// Проверяем, что этот рецензент уже не назначен
	for _, reviewer := range pr.Reviewers {
		if reviewer.ID == reviewerID {
			return nil, fmt.Errorf("reviewer already assigned to this PR")
		}
	}

	// Проверяем, что рецензент не является автором
	if pr.AuthorID == reviewerID {
		return nil, fmt.Errorf("author cannot be a reviewer")
	}

	// Получаем пользователя
	user, err := s.userRepo.GetByID(reviewerID)
	if err != nil {
		return nil, fmt.Errorf("reviewer not found")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("reviewer is not active")
	}

	// Добавляем рецензента
	if err := s.prRepo.AddReviewers(prID, []models.User{*user}); err != nil {
		return nil, err
	}

	// Возвращаем обновлённый PR
	return s.GetPullRequest(prID)
}

// ReassignReviewer переназначает рецензента
func (s *Service) ReassignReviewer(prID int, req *models.ReassignReviewerRequest) (*models.PullRequest, error) {
	// Получаем PR
	pr, err := s.prRepo.GetByID(prID)
	if err != nil {
		return nil, err
	}

	// Проверяем, что PR не в статусе MERGED
	if pr.Status == models.PRStatusMerged {
		return nil, fmt.Errorf("cannot change reviewers of merged PR")
	}

	// Находим старого рецензента среди текущих
	var oldReviewer *models.User
	for i := range pr.Reviewers {
		if pr.Reviewers[i].ID == req.OldReviewerID {
			oldReviewer = &pr.Reviewers[i]
			break
		}
	}

	if oldReviewer == nil {
		return nil, fmt.Errorf("reviewer not found in PR")
	}

	// Если старый рецензент не в команде, не можем выбрать замену
	if oldReviewer.TeamID == nil {
		return nil, fmt.Errorf("reviewer is not in a team")
	}

	// Выбираем нового рецензента из той же команды
	newReviewer, err := s.selectRandomReviewer(*oldReviewer.TeamID, pr.AuthorID, getReviewerIDs(pr.Reviewers))
	if err != nil {
		return nil, fmt.Errorf("failed to select new reviewer: %w", err)
	}

	// Заменяем рецензента
	if err := s.prRepo.ReplaceReviewer(prID, req.OldReviewerID, newReviewer.ID); err != nil {
		return nil, err
	}

	// Возвращаем обновлённый PR
	return s.prRepo.GetByID(prID)
}

// BulkDeactivateUsers массово деактивирует пользователей и переназначает их PR
func (s *Service) BulkDeactivateUsers(teamID int, req *models.BulkDeactivateRequest) (*models.BulkDeactivateResponse, error) {
	// Деактивируем пользователей
	deactivatedCount, err := s.userRepo.BulkDeactivate(teamID, req.UserIDs)
	if err != nil {
		return nil, err
	}

	reassignedCount := 0
	// Для каждого деактивированного пользователя переназначаем открытые PR
	for _, userID := range req.UserIDs {
		// Получаем открытые PR, где пользователь является рецензентом
		prs, err := s.prRepo.GetOpenPRsWithReviewer(userID)
		if err != nil {
			continue // Продолжаем даже если ошибка
		}

		for _, pr := range prs {
			// Пытаемся найти замену из команды пользователя
			user, err := s.userRepo.GetByID(userID)
			if err != nil || user.TeamID == nil {
				continue
			}

			// Выбираем нового рецензента
			newReviewer, err := s.selectRandomReviewer(*user.TeamID, pr.AuthorID, getReviewerIDs(pr.Reviewers))
			if err != nil {
				continue // Если не можем найти замену, пропускаем
			}

			// Заменяем рецензента
			if err := s.prRepo.ReplaceReviewer(pr.ID, userID, newReviewer.ID); err == nil {
				reassignedCount++
			}
		}
	}

	return &models.BulkDeactivateResponse{
		DeactivatedCount:  deactivatedCount,
		ReassignedPRCount: reassignedCount,
	}, nil
}

// GetStatistics возвращает статистику
func (s *Service) GetStatistics() (*models.Statistics, error) {
	return s.statsRepo.GetStatistics()
}

// selectReviewers выбирает до maxCount рецензентов из команды
func (s *Service) selectReviewers(teamID, authorID, maxCount int) ([]models.User, error) {
	// Получаем активных пользователей из команды, исключая автора
	candidates, err := s.userRepo.GetActiveUsersFromTeam(teamID, authorID)
	if err != nil {
		return nil, err
	}

	// Если кандидатов меньше, чем нужно, возвращаем всех
	if len(candidates) <= maxCount {
		reviewers := make([]models.User, len(candidates))
		for i, c := range candidates {
			reviewers[i] = *c
		}
		return reviewers, nil
	}

	// Случайно выбираем maxCount рецензентов
	// #nosec G404 - не криптографическая операция, случайность для выбора ревьюеров
	rand.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	reviewers := make([]models.User, maxCount)
	for i := 0; i < maxCount; i++ {
		reviewers[i] = *candidates[i]
	}

	return reviewers, nil
}

// selectRandomReviewer выбирает случайного рецензента из команды, исключая указанных пользователей
func (s *Service) selectRandomReviewer(teamID, authorID int, excludeIDs []int) (*models.User, error) {
	// Получаем активных пользователей из команды
	candidates, err := s.userRepo.GetActiveUsersFromTeam(teamID, authorID)
	if err != nil {
		return nil, err
	}

	// Фильтруем кандидатов, исключая тех, кто уже рецензент
	filtered := make([]*models.User, 0)
	excludeMap := make(map[int]bool)
	for _, id := range excludeIDs {
		excludeMap[id] = true
	}

	for _, c := range candidates {
		if !excludeMap[c.ID] {
			filtered = append(filtered, c)
		}
	}

	if len(filtered) == 0 {
		return nil, fmt.Errorf("no available reviewers in team")
	}

	// Случайно выбираем одного
	// #nosec G404 - не криптографическая операция, случайность для выбора ревьюеров
	return filtered[rand.Intn(len(filtered))], nil
}

// enrichPR обогащает PR информацией об авторе, команде и рецензентах
func (s *Service) enrichPR(pr *models.PullRequest) {
	// Обогащаем PR информацией об авторе
	author, err := s.userRepo.GetByID(pr.AuthorID)
	if err == nil {
		// Загружаем команду автора
		if author.TeamID != nil {
			team, err := s.teamRepo.GetByID(*author.TeamID)
			if err == nil {
				author.Teams = []models.Team{*team}
				pr.Team = team
			}
		}
		pr.Author = author
	}

	// Обогащаем рецензентов информацией о командах
	for i := range pr.Reviewers {
		if pr.Reviewers[i].TeamID != nil {
			team, err := s.teamRepo.GetByID(*pr.Reviewers[i].TeamID)
			if err == nil {
				pr.Reviewers[i].Teams = []models.Team{*team}
			}
		}
	}
}

// getReviewerIDs извлекает ID рецензентов
func getReviewerIDs(reviewers []models.User) []int {
	ids := make([]int, len(reviewers))
	for i, r := range reviewers {
		ids[i] = r.ID
	}
	return ids
}
