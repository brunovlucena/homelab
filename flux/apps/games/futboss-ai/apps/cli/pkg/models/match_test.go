// FutBoss AI - Match Model Tests (Go)
// Author: Bruno Lucena (bruno@lucena.cloud)

package models

import (
	"testing"
)

func TestMatchIsHomeWinnerYes(t *testing.T) {
	match := &Match{
		HomeScore: 3,
		AwayScore: 1,
		Status:    MatchStatusFinished,
	}

	result := match.IsHomeWinner()
	if result == nil || *result != true {
		t.Error("Expected home winner to be true")
	}
}

func TestMatchIsHomeWinnerNo(t *testing.T) {
	match := &Match{
		HomeScore: 1,
		AwayScore: 3,
		Status:    MatchStatusFinished,
	}

	result := match.IsHomeWinner()
	if result == nil || *result != false {
		t.Error("Expected home winner to be false")
	}
}

func TestMatchIsHomeWinnerDraw(t *testing.T) {
	match := &Match{
		HomeScore: 2,
		AwayScore: 2,
		Status:    MatchStatusFinished,
	}

	result := match.IsHomeWinner()
	if result != nil {
		t.Error("Expected nil for draw")
	}
}

func TestMatchIsHomeWinnerNotFinished(t *testing.T) {
	match := &Match{
		HomeScore: 2,
		AwayScore: 1,
		Status:    MatchStatusInProgress,
	}

	result := match.IsHomeWinner()
	if result != nil {
		t.Error("Expected nil for in-progress match")
	}
}

func TestMatchIsDraw(t *testing.T) {
	tests := []struct {
		homeScore int
		awayScore int
		status    MatchStatus
		expected  bool
	}{
		{2, 2, MatchStatusFinished, true},
		{0, 0, MatchStatusFinished, true},
		{3, 1, MatchStatusFinished, false},
		{2, 2, MatchStatusInProgress, false}, // Not finished
	}

	for _, tt := range tests {
		match := &Match{
			HomeScore: tt.homeScore,
			AwayScore: tt.awayScore,
			Status:    tt.status,
		}
		if match.IsDraw() != tt.expected {
			t.Errorf("IsDraw for %d-%d (%s): expected %v", tt.homeScore, tt.awayScore, tt.status, tt.expected)
		}
	}
}

func TestMatchTotalGoals(t *testing.T) {
	match := &Match{
		HomeScore: 3,
		AwayScore: 2,
	}

	if match.TotalGoals() != 5 {
		t.Errorf("Expected 5 total goals, got %d", match.TotalGoals())
	}
}

