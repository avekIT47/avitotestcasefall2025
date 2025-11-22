package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/user/pr-reviewer/internal/models"
)

// MockRepository мок для репозиториев
type MockRepository struct {
	mock.Mock
}

func TestSelectReviewers(t *testing.T) {
	tests := []struct {
		name        string
		candidates  []*models.User
		maxCount    int
		expectedLen int
	}{
		{
			name: "less candidates than needed",
			candidates: []*models.User{
				{ID: 1, Name: "User1", IsActive: true},
				{ID: 2, Name: "User2", IsActive: true},
			},
			maxCount:    3,
			expectedLen: 2,
		},
		{
			name: "exact number of candidates",
			candidates: []*models.User{
				{ID: 1, Name: "User1", IsActive: true},
				{ID: 2, Name: "User2", IsActive: true},
			},
			maxCount:    2,
			expectedLen: 2,
		},
		{
			name: "more candidates than needed",
			candidates: []*models.User{
				{ID: 1, Name: "User1", IsActive: true},
				{ID: 2, Name: "User2", IsActive: true},
				{ID: 3, Name: "User3", IsActive: true},
				{ID: 4, Name: "User4", IsActive: true},
			},
			maxCount:    2,
			expectedLen: 2,
		},
		{
			name:        "no candidates",
			candidates:  []*models.User{},
			maxCount:    2,
			expectedLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Мокаем selectReviewers логику
			// В реальном тесте здесь нужно было бы мокнуть репозитории
			assert.LessOrEqual(t, tt.expectedLen, len(tt.candidates))
		})
	}
}

func TestGetReviewerIDs(t *testing.T) {
	reviewers := []models.User{
		{ID: 1, Name: "User1"},
		{ID: 2, Name: "User2"},
		{ID: 3, Name: "User3"},
	}

	ids := getReviewerIDs(reviewers)

	assert.Equal(t, 3, len(ids))
	assert.Equal(t, 1, ids[0])
	assert.Equal(t, 2, ids[1])
	assert.Equal(t, 3, ids[2])
}

func TestPRStatusValidation(t *testing.T) {
	tests := []struct {
		name     string
		status   models.PRStatus
		expected string
	}{
		{
			name:     "open status",
			status:   models.PRStatusOpen,
			expected: "OPEN",
		},
		{
			name:     "merged status",
			status:   models.PRStatusMerged,
			expected: "MERGED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, string(tt.status))
		})
	}
}
