// FutBoss AI - Match Model (Go)
// Author: Bruno Lucena (bruno@lucena.cloud)

package models

import "time"

type MatchStatus string

const (
	MatchStatusPending    MatchStatus = "pending"
	MatchStatusInProgress MatchStatus = "in_progress"
	MatchStatusFinished   MatchStatus = "finished"
	MatchStatusCancelled  MatchStatus = "cancelled"
)

type EventType string

const (
	EventTypeGoal        EventType = "goal"
	EventTypeAssist      EventType = "assist"
	EventTypeYellowCard  EventType = "yellow_card"
	EventTypeRedCard     EventType = "red_card"
	EventTypeSubstitution EventType = "substitution"
	EventTypeInjury      EventType = "injury"
	EventTypePenalty     EventType = "penalty"
	EventTypeSave        EventType = "save"
	EventTypeFoul        EventType = "foul"
)

type MatchEvent struct {
	Minute      int       `json:"minute"`
	EventType   EventType `json:"event_type"`
	PlayerID    string    `json:"player_id"`
	TeamID      string    `json:"team_id"`
	Description string    `json:"description"`
	AINarration string    `json:"ai_narration,omitempty"`
}

type MatchState struct {
	Minute         int  `json:"minute"`
	HomeScore      int  `json:"home_score"`
	AwayScore      int  `json:"away_score"`
	PossessionHome int  `json:"possession_home"`
	PossessionAway int  `json:"possession_away"`
	ShotsHome      int  `json:"shots_home"`
	ShotsAway      int  `json:"shots_away"`
	IsPaused       bool `json:"is_paused"`
}

type Match struct {
	ID          string       `json:"id"`
	HomeTeamID  string       `json:"home_team_id"`
	AwayTeamID  string       `json:"away_team_id"`
	HomeScore   int          `json:"home_score"`
	AwayScore   int          `json:"away_score"`
	Status      MatchStatus  `json:"status"`
	Events      []MatchEvent `json:"events"`
	State       MatchState   `json:"state"`
	StartedAt   *time.Time   `json:"started_at,omitempty"`
	FinishedAt  *time.Time   `json:"finished_at,omitempty"`
}

func (m *Match) IsHomeWinner() *bool {
	if m.Status != MatchStatusFinished {
		return nil
	}
	result := m.HomeScore > m.AwayScore
	if m.HomeScore == m.AwayScore {
		return nil // Draw
	}
	return &result
}

func (m *Match) IsDraw() bool {
	return m.Status == MatchStatusFinished && m.HomeScore == m.AwayScore
}

func (m *Match) TotalGoals() int {
	return m.HomeScore + m.AwayScore
}

