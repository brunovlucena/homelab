// FutBoss AI - Player Model (Go)
// Author: Bruno Lucena (bruno@lucena.cloud)

package models

type Position string

const (
	PositionGK  Position = "GK"
	PositionCB  Position = "CB"
	PositionLB  Position = "LB"
	PositionRB  Position = "RB"
	PositionCDM Position = "CDM"
	PositionCM  Position = "CM"
	PositionCAM Position = "CAM"
	PositionLW  Position = "LW"
	PositionRW  Position = "RW"
	PositionST  Position = "ST"
)

type Temperament string

const (
	TemperamentCalm       Temperament = "calm"
	TemperamentExplosive  Temperament = "explosive"
	TemperamentCalculated Temperament = "calculated"
)

type PlayStyle string

const (
	PlayStyleOffensive PlayStyle = "offensive"
	PlayStyleDefensive PlayStyle = "defensive"
	PlayStyleBalanced  PlayStyle = "balanced"
)

type PlayerAttributes struct {
	Speed       int `json:"speed"`
	Strength    int `json:"strength"`
	Stamina     int `json:"stamina"`
	Finishing   int `json:"finishing"`
	Passing     int `json:"passing"`
	Dribbling   int `json:"dribbling"`
	Defense     int `json:"defense"`
	Intelligence int `json:"intelligence"`
	Aggression  int `json:"aggression"`
	Leadership  int `json:"leadership"`
	Creativity  int `json:"creativity"`
}

func (a *PlayerAttributes) Overall() int {
	attrs := []int{
		a.Speed, a.Strength, a.Stamina,
		a.Finishing, a.Passing, a.Dribbling, a.Defense,
		a.Intelligence, a.Aggression, a.Leadership, a.Creativity,
	}
	sum := 0
	for _, v := range attrs {
		sum += v
	}
	return sum / len(attrs)
}

type PlayerPersonality struct {
	Temperament Temperament `json:"temperament"`
	PlayStyle   PlayStyle   `json:"play_style"`
}

type Player struct {
	ID          string            `json:"id"`
	Name        string            `json:"name"`
	Position    Position          `json:"position"`
	Nationality string            `json:"nationality"`
	Age         int               `json:"age"`
	Attributes  PlayerAttributes  `json:"attributes"`
	Personality PlayerPersonality `json:"personality"`
	TeamID      *string           `json:"team_id,omitempty"`
	Price       int               `json:"price"`
	IsListed    bool              `json:"is_listed"`
}

func (p *Player) GetMarketValue() int {
	baseValue := p.Attributes.Overall() * 10
	ageModifier := 1.0

	if p.Age < 23 {
		ageModifier = 1.3
	} else if p.Age > 32 {
		ageModifier = 0.7
	}

	return int(float64(baseValue) * ageModifier)
}

func (p *Player) IsDefender() bool {
	return p.Position == PositionCB || p.Position == PositionLB || p.Position == PositionRB
}

func (p *Player) IsMidfielder() bool {
	return p.Position == PositionCDM || p.Position == PositionCM || p.Position == PositionCAM
}

func (p *Player) IsAttacker() bool {
	return p.Position == PositionLW || p.Position == PositionRW || p.Position == PositionST
}

