// FutBoss AI - Team Model (Go)
// Author: Bruno Lucena (bruno@lucena.cloud)

package models

type Formation string

const (
	Formation442 Formation = "4-4-2"
	Formation433 Formation = "4-3-3"
	Formation352 Formation = "3-5-2"
	Formation451 Formation = "4-5-1"
	Formation343 Formation = "3-4-3"
	Formation532 Formation = "5-3-2"
)

type Team struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	OwnerID      string    `json:"owner_id"`
	Formation    Formation `json:"formation"`
	PlayerIDs    []string  `json:"player_ids"`
	Wins         int       `json:"wins"`
	Draws        int       `json:"draws"`
	Losses       int       `json:"losses"`
	GoalsFor     int       `json:"goals_for"`
	GoalsAgainst int       `json:"goals_against"`
}

func (t *Team) Points() int {
	return (t.Wins * 3) + t.Draws
}

func (t *Team) MatchesPlayed() int {
	return t.Wins + t.Draws + t.Losses
}

func (t *Team) GoalDifference() int {
	return t.GoalsFor - t.GoalsAgainst
}

func (t *Team) WinRate() float64 {
	if t.MatchesPlayed() == 0 {
		return 0
	}
	return float64(t.Wins) / float64(t.MatchesPlayed()) * 100
}

