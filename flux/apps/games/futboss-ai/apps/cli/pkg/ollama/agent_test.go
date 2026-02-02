// FutBoss AI - Ollama Agent Tests (Go)
// Author: Bruno Lucena (bruno@lucena.cloud)

package ollama

import (
	"testing"

	"github.com/brunolucena/futboss-ai/cli/pkg/models"
)

func TestBuildAgentPrompt(t *testing.T) {
	client := NewOllamaClient("http://localhost:11434", "llama3.2")

	player := &models.Player{
		Name:     "Test Striker",
		Position: models.PositionST,
		Attributes: models.PlayerAttributes{
			Speed: 85, Strength: 70, Stamina: 80,
			Finishing: 90, Passing: 65, Dribbling: 80, Defense: 30,
			Intelligence: 75, Aggression: 70, Leadership: 50, Creativity: 85,
		},
		Personality: models.PlayerPersonality{
			Temperament: models.TemperamentExplosive,
			PlayStyle:   models.PlayStyleOffensive,
		},
	}

	state := &GameState{
		Minute:    45,
		HomeScore: 1,
		AwayScore: 1,
		HasBall:   true,
		BallZone:  "attack",
	}

	prompt := client.BuildAgentPrompt(player, state)

	// Check prompt contains player info
	if len(prompt) < 100 {
		t.Error("Prompt too short")
	}

	// Check contains player name
	if !contains(prompt, "Test Striker") {
		t.Error("Prompt should contain player name")
	}

	// Check contains attributes
	if !contains(prompt, "Speed: 85") {
		t.Error("Prompt should contain speed attribute")
	}

	// Check contains game state
	if !contains(prompt, "Minute: 45") {
		t.Error("Prompt should contain minute")
	}
}

func TestFallbackDecisionWithBallAttack(t *testing.T) {
	client := NewOllamaClient("http://localhost:11434", "llama3.2")

	player := &models.Player{
		Position: models.PositionST,
		Attributes: models.PlayerAttributes{
			Finishing: 90,
			Dribbling: 70,
		},
	}

	state := &GameState{
		HasBall:  true,
		BallZone: "attack",
	}

	decision := client.fallbackDecision(player, state)

	if decision.Action != "SHOOT" {
		t.Errorf("Expected SHOOT for high finishing player in attack, got %s", decision.Action)
	}
}

func TestFallbackDecisionWithBallMidfield(t *testing.T) {
	client := NewOllamaClient("http://localhost:11434", "llama3.2")

	player := &models.Player{
		Position: models.PositionCM,
		Attributes: models.PlayerAttributes{
			Finishing: 50,
			Dribbling: 85,
			Passing:   70,
		},
	}

	state := &GameState{
		HasBall:  true,
		BallZone: "midfield",
	}

	decision := client.fallbackDecision(player, state)

	if decision.Action != "DRIBBLE" {
		t.Errorf("Expected DRIBBLE for high dribbling player, got %s", decision.Action)
	}
}

func TestFallbackDecisionWithoutBallDefender(t *testing.T) {
	client := NewOllamaClient("http://localhost:11434", "llama3.2")

	player := &models.Player{
		Position: models.PositionCB,
		Attributes: models.PlayerAttributes{
			Defense: 85,
		},
	}

	state := &GameState{
		HasBall:  false,
		BallZone: "defense",
	}

	decision := client.fallbackDecision(player, state)

	if decision.Action != "TACKLE" {
		t.Errorf("Expected TACKLE for defender without ball, got %s", decision.Action)
	}
}

func TestFallbackDecisionWithoutBallMidfielder(t *testing.T) {
	client := NewOllamaClient("http://localhost:11434", "llama3.2")

	player := &models.Player{
		Position: models.PositionCM,
		Attributes: models.PlayerAttributes{
			Defense: 50,
		},
	}

	state := &GameState{
		HasBall:  false,
		BallZone: "midfield",
	}

	decision := client.fallbackDecision(player, state)

	if decision.Action != "FALL_BACK" {
		t.Errorf("Expected FALL_BACK for low defense player, got %s", decision.Action)
	}
}

func TestParseDecision(t *testing.T) {
	client := NewOllamaClient("http://localhost:11434", "llama3.2")

	tests := []struct {
		response string
		expected string
	}{
		{"SHOOT - Great position to score", "SHOOT"},
		{"PASS - Better option available", "PASS"},
		{"DRIBBLE - Can beat defender", "DRIBBLE"},
		{"TACKLE - Win back possession", "TACKLE"},
		{"invalid response", "HOLD"}, // Default
	}

	for _, tt := range tests {
		decision := client.parseDecision(tt.response)
		if decision.Action != tt.expected {
			t.Errorf("For '%s': expected %s, got %s", tt.response, tt.expected, decision.Action)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && (s[:len(substr)] == substr || contains(s[1:], substr)))
}

