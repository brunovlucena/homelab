// FutBoss AI - Ollama Agent Integration (Go)
// Author: Bruno Lucena (bruno@lucena.cloud)

package ollama

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/brunolucena/futboss-ai/cli/pkg/models"
)

type OllamaClient struct {
	BaseURL string
	Model   string
	Client  *http.Client
}

type GenerateRequest struct {
	Model   string            `json:"model"`
	Prompt  string            `json:"prompt"`
	Stream  bool              `json:"stream"`
	Options map[string]interface{} `json:"options,omitempty"`
}

type GenerateResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

type AgentDecision struct {
	Action string `json:"action"`
	Reason string `json:"reason"`
}

type GameState struct {
	Minute    int    `json:"minute"`
	HomeScore int    `json:"home_score"`
	AwayScore int    `json:"away_score"`
	HasBall   bool   `json:"has_ball"`
	BallZone  string `json:"ball_zone"` // defense, midfield, attack
}

func NewOllamaClient(baseURL, model string) *OllamaClient {
	return &OllamaClient{
		BaseURL: baseURL,
		Model:   model,
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *OllamaClient) BuildAgentPrompt(player *models.Player, state *GameState) string {
	tempDesc := map[models.Temperament]string{
		models.TemperamentCalm:       "You are calm and composed under pressure, preferring safe plays.",
		models.TemperamentExplosive:  "You are explosive and unpredictable, taking risks for big rewards.",
		models.TemperamentCalculated: "You are calculated and strategic, analyzing every option carefully.",
	}

	styleDesc := map[models.PlayStyle]string{
		models.PlayStyleOffensive: "You prefer attacking plays, always looking to score.",
		models.PlayStyleDefensive: "You focus on defense, prioritizing ball retention and safety.",
		models.PlayStyleBalanced:  "You balance attack and defense based on the situation.",
	}

	return fmt.Sprintf(`You are %s, a %s football player.

PERSONALITY:
%s
%s

YOUR ATTRIBUTES (1-100 scale):
- Speed: %d
- Strength: %d
- Finishing: %d
- Passing: %d
- Dribbling: %d
- Defense: %d
- Intelligence: %d
- Creativity: %d

CURRENT GAME STATE:
- Minute: %d
- Score: %d - %d
- Your team has ball: %v
- Ball position: %s

Based on your personality and attributes, what action do you take?
Choose one: PASS, DRIBBLE, SHOOT, TACKLE, HOLD, RUN_FORWARD, FALL_BACK

Respond with ONLY the action name and a brief reason (max 20 words).`,
		player.Name,
		player.Position,
		tempDesc[player.Personality.Temperament],
		styleDesc[player.Personality.PlayStyle],
		player.Attributes.Speed,
		player.Attributes.Strength,
		player.Attributes.Finishing,
		player.Attributes.Passing,
		player.Attributes.Dribbling,
		player.Attributes.Defense,
		player.Attributes.Intelligence,
		player.Attributes.Creativity,
		state.Minute,
		state.HomeScore,
		state.AwayScore,
		state.HasBall,
		state.BallZone,
	)
}

func (c *OllamaClient) GetAgentDecision(player *models.Player, state *GameState) (*AgentDecision, error) {
	prompt := c.BuildAgentPrompt(player, state)

	// Calculate temperature based on creativity
	temp := 0.7 + float64(player.Attributes.Creativity)/200

	reqBody := GenerateRequest{
		Model:  c.Model,
		Prompt: prompt,
		Stream: false,
		Options: map[string]interface{}{
			"temperature": temp,
			"num_predict": 50,
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return c.fallbackDecision(player, state), nil
	}

	resp, err := c.Client.Post(
		c.BaseURL+"/api/generate",
		"application/json",
		bytes.NewBuffer(jsonBody),
	)
	if err != nil {
		return c.fallbackDecision(player, state), nil
	}
	defer resp.Body.Close()

	var result GenerateResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return c.fallbackDecision(player, state), nil
	}

	// Parse response
	return c.parseDecision(result.Response), nil
}

func (c *OllamaClient) parseDecision(response string) *AgentDecision {
	validActions := map[string]bool{
		"PASS": true, "DRIBBLE": true, "SHOOT": true,
		"TACKLE": true, "HOLD": true, "RUN_FORWARD": true, "FALL_BACK": true,
	}

	// Simple parsing - first word should be action
	action := "HOLD"
	reason := response

	for a := range validActions {
		if len(response) >= len(a) && response[:len(a)] == a {
			action = a
			if len(response) > len(a)+3 {
				reason = response[len(a)+3:] // Skip " - "
			}
			break
		}
	}

	return &AgentDecision{
		Action: action,
		Reason: reason,
	}
}

func (c *OllamaClient) fallbackDecision(player *models.Player, state *GameState) *AgentDecision {
	attrs := player.Attributes

	if !state.HasBall {
		if attrs.Defense > 60 {
			return &AgentDecision{Action: "TACKLE", Reason: "Defensive instinct"}
		}
		return &AgentDecision{Action: "FALL_BACK", Reason: "Positioning"}
	}

	if state.BallZone == "attack" {
		if attrs.Finishing > 70 {
			return &AgentDecision{Action: "SHOOT", Reason: "In scoring position"}
		}
		if attrs.Passing > attrs.Dribbling {
			return &AgentDecision{Action: "PASS", Reason: "Better passing option"}
		}
		return &AgentDecision{Action: "DRIBBLE", Reason: "Create space"}
	}

	if attrs.Dribbling > 70 {
		return &AgentDecision{Action: "DRIBBLE", Reason: "Skill advantage"}
	}
	return &AgentDecision{Action: "PASS", Reason: "Move ball forward"}
}

func (c *OllamaClient) CheckHealth() bool {
	resp, err := c.Client.Get(c.BaseURL + "/api/tags")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == 200
}

