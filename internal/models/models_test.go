package models

import (
	"testing"
)

func TestPRStatus(t *testing.T) {
	tests := []struct {
		name   string
		status PRStatus
		want   string
	}{
		{"open status", PRStatusOpen, "OPEN"},
		{"merged status", PRStatusMerged, "MERGED"},
		{"closed status", PRStatusClosed, "CLOSED"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.status) != tt.want {
				t.Errorf("expected %s, got %s", tt.want, string(tt.status))
			}
		})
	}
}

func TestTeamModel(t *testing.T) {
	team := &Team{
		ID:   1,
		Name: "Test Team",
	}

	if team.ID != 1 {
		t.Errorf("expected ID 1, got %d", team.ID)
	}

	if team.Name != "Test Team" {
		t.Errorf("expected name 'Test Team', got '%s'", team.Name)
	}
}

func TestUserModel(t *testing.T) {
	teamID := 1
	user := &User{
		ID:       1,
		Username: "testuser",
		Name:     "Test User",
		IsActive: true,
		TeamID:   &teamID,
	}

	if user.ID != 1 {
		t.Errorf("expected ID 1, got %d", user.ID)
	}

	if user.Username != "testuser" {
		t.Errorf("expected username 'testuser', got '%s'", user.Username)
	}

	if !user.IsActive {
		t.Error("expected user to be active")
	}

	if user.TeamID == nil || *user.TeamID != 1 {
		t.Error("expected team ID to be 1")
	}
}

func TestPullRequestModel(t *testing.T) {
	pr := &PullRequest{
		ID:       1,
		Title:    "Test PR",
		AuthorID: 1,
		Status:   PRStatusOpen,
	}

	if pr.ID != 1 {
		t.Errorf("expected ID 1, got %d", pr.ID)
	}

	if pr.Title != "Test PR" {
		t.Errorf("expected title 'Test PR', got '%s'", pr.Title)
	}

	if pr.Status != PRStatusOpen {
		t.Errorf("expected status OPEN, got %s", pr.Status)
	}
}

func TestCreateTeamRequest(t *testing.T) {
	req := &CreateTeamRequest{
		Name: "New Team",
	}

	if req.Name != "New Team" {
		t.Errorf("expected name 'New Team', got '%s'", req.Name)
	}
}

func TestCreateUserRequest(t *testing.T) {
	teamID := 1
	req := &CreateUserRequest{
		Username: "newuser",
		Name:     "New User",
		TeamID:   &teamID,
	}

	if req.Username != "newuser" {
		t.Errorf("expected username 'newuser', got '%s'", req.Username)
	}

	if req.TeamID == nil || *req.TeamID != 1 {
		t.Error("expected team ID to be 1")
	}
}

func TestUpdateUserRequest(t *testing.T) {
	name := "Updated Name"
	isActive := false
	req := &UpdateUserRequest{
		Name:     &name,
		IsActive: &isActive,
	}

	if req.Name == nil || *req.Name != "Updated Name" {
		t.Error("expected name to be 'Updated Name'")
	}

	if req.IsActive == nil || *req.IsActive != false {
		t.Error("expected isActive to be false")
	}
}

func TestStatistics(t *testing.T) {
	stats := &Statistics{
		TotalPRs:  100,
		OpenPRs:   20,
		MergedPRs: 70,
		ClosedPRs: 10,
	}

	if stats.TotalPRs != 100 {
		t.Errorf("expected 100 total PRs, got %d", stats.TotalPRs)
	}

	if stats.OpenPRs != 20 {
		t.Errorf("expected 20 open PRs, got %d", stats.OpenPRs)
	}

	if stats.MergedPRs != 70 {
		t.Errorf("expected 70 merged PRs, got %d", stats.MergedPRs)
	}

	if stats.ClosedPRs != 10 {
		t.Errorf("expected 10 closed PRs, got %d", stats.ClosedPRs)
	}
}

func TestUserStatistic(t *testing.T) {
	stat := &UserStatistic{
		UserID:          1,
		UserName:        "Test User",
		AssignmentCount: 5,
	}

	if stat.UserID != 1 {
		t.Errorf("expected user ID 1, got %d", stat.UserID)
	}

	if stat.AssignmentCount != 5 {
		t.Errorf("expected 5 assignments, got %d", stat.AssignmentCount)
	}
}

func TestTeamStatistic(t *testing.T) {
	stat := &TeamStatistic{
		TeamID:   1,
		TeamName: "Test Team",
		PRCount:  10,
	}

	if stat.TeamID != 1 {
		t.Errorf("expected team ID 1, got %d", stat.TeamID)
	}

	if stat.PRCount != 10 {
		t.Errorf("expected 10 PRs, got %d", stat.PRCount)
	}
}

func TestHealthResponse(t *testing.T) {
	health := &HealthResponse{
		Status: "healthy",
	}

	if health.Status != "healthy" {
		t.Errorf("expected status 'healthy', got '%s'", health.Status)
	}
}

func TestBulkDeactivateRequest(t *testing.T) {
	req := &BulkDeactivateRequest{
		UserIDs: []int{1, 2, 3},
	}

	if len(req.UserIDs) != 3 {
		t.Errorf("expected 3 user IDs, got %d", len(req.UserIDs))
	}
}

func TestBulkDeactivateResponse(t *testing.T) {
	resp := &BulkDeactivateResponse{
		DeactivatedCount:  5,
		ReassignedPRCount: 3,
	}

	if resp.DeactivatedCount != 5 {
		t.Errorf("expected 5 deactivated users, got %d", resp.DeactivatedCount)
	}

	if resp.ReassignedPRCount != 3 {
		t.Errorf("expected 3 reassigned PRs, got %d", resp.ReassignedPRCount)
	}
}

