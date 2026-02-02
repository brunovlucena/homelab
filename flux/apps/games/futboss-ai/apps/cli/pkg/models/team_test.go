// FutBoss AI - Team Model Tests (Go)
// Author: Bruno Lucena (bruno@lucena.cloud)

package models

import (
	"testing"
)

func TestTeamPoints(t *testing.T) {
	team := &Team{
		Wins:   10,
		Draws:  5,
		Losses: 3,
	}

	expected := 35 // 10*3 + 5*1
	if team.Points() != expected {
		t.Errorf("Expected %d points, got %d", expected, team.Points())
	}
}

func TestTeamMatchesPlayed(t *testing.T) {
	team := &Team{
		Wins:   10,
		Draws:  5,
		Losses: 3,
	}

	expected := 18
	if team.MatchesPlayed() != expected {
		t.Errorf("Expected %d matches, got %d", expected, team.MatchesPlayed())
	}
}

func TestTeamGoalDifference(t *testing.T) {
	tests := []struct {
		goalsFor     int
		goalsAgainst int
		expected     int
	}{
		{25, 15, 10},
		{15, 25, -10},
		{20, 20, 0},
	}

	for _, tt := range tests {
		team := &Team{
			GoalsFor:     tt.goalsFor,
			GoalsAgainst: tt.goalsAgainst,
		}
		if team.GoalDifference() != tt.expected {
			t.Errorf("GD for %d-%d: expected %d, got %d", tt.goalsFor, tt.goalsAgainst, tt.expected, team.GoalDifference())
		}
	}
}

func TestTeamWinRate(t *testing.T) {
	tests := []struct {
		wins     int
		draws    int
		losses   int
		expected float64
	}{
		{10, 0, 0, 100.0},
		{5, 0, 5, 50.0},
		{0, 10, 0, 0.0},
		{0, 0, 0, 0.0}, // No matches
	}

	for _, tt := range tests {
		team := &Team{
			Wins:   tt.wins,
			Draws:  tt.draws,
			Losses: tt.losses,
		}
		rate := team.WinRate()
		if rate != tt.expected {
			t.Errorf("WinRate for %dW-%dD-%dL: expected %.1f, got %.1f", tt.wins, tt.draws, tt.losses, tt.expected, rate)
		}
	}
}

