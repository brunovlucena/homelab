// FutBoss AI - Player Model Tests (Go)
// Author: Bruno Lucena (bruno@lucena.cloud)

package models

import (
	"testing"
)

func TestPlayerAttributesOverall(t *testing.T) {
	attrs := PlayerAttributes{
		Speed:       80,
		Strength:    70,
		Stamina:     75,
		Finishing:   85,
		Passing:     70,
		Dribbling:   78,
		Defense:     40,
		Intelligence: 72,
		Aggression:  60,
		Leadership:  55,
		Creativity:  80,
	}

	overall := attrs.Overall()

	// Average should be around 69-70
	if overall < 60 || overall > 80 {
		t.Errorf("Expected overall between 60-80, got %d", overall)
	}
}

func TestPlayerMarketValueYoung(t *testing.T) {
	player := &Player{
		Name:     "Young Star",
		Position: PositionST,
		Age:      20,
		Attributes: PlayerAttributes{
			Speed: 80, Strength: 70, Stamina: 75,
			Finishing: 85, Passing: 70, Dribbling: 78, Defense: 40,
			Intelligence: 72, Aggression: 60, Leadership: 55, Creativity: 80,
		},
	}

	value := player.GetMarketValue()
	baseValue := player.Attributes.Overall() * 10

	// Young players should have 1.3x modifier
	if value <= baseValue {
		t.Errorf("Young player value %d should be > base %d", value, baseValue)
	}
}

func TestPlayerMarketValueOld(t *testing.T) {
	player := &Player{
		Name:     "Veteran",
		Position: PositionCB,
		Age:      35,
		Attributes: PlayerAttributes{
			Speed: 50, Strength: 50, Stamina: 50,
			Finishing: 50, Passing: 50, Dribbling: 50, Defense: 50,
			Intelligence: 50, Aggression: 50, Leadership: 50, Creativity: 50,
		},
	}

	value := player.GetMarketValue()
	baseValue := player.Attributes.Overall() * 10

	// Old players should have 0.7x modifier
	if value >= baseValue {
		t.Errorf("Old player value %d should be < base %d", value, baseValue)
	}
}

func TestPlayerIsDefender(t *testing.T) {
	tests := []struct {
		position Position
		expected bool
	}{
		{PositionGK, false},
		{PositionCB, true},
		{PositionLB, true},
		{PositionRB, true},
		{PositionCM, false},
		{PositionST, false},
	}

	for _, tt := range tests {
		player := &Player{Position: tt.position}
		if player.IsDefender() != tt.expected {
			t.Errorf("Position %s: expected IsDefender=%v, got %v", tt.position, tt.expected, player.IsDefender())
		}
	}
}

func TestPlayerIsMidfielder(t *testing.T) {
	tests := []struct {
		position Position
		expected bool
	}{
		{PositionCDM, true},
		{PositionCM, true},
		{PositionCAM, true},
		{PositionCB, false},
		{PositionST, false},
	}

	for _, tt := range tests {
		player := &Player{Position: tt.position}
		if player.IsMidfielder() != tt.expected {
			t.Errorf("Position %s: expected IsMidfielder=%v, got %v", tt.position, tt.expected, player.IsMidfielder())
		}
	}
}

func TestPlayerIsAttacker(t *testing.T) {
	tests := []struct {
		position Position
		expected bool
	}{
		{PositionLW, true},
		{PositionRW, true},
		{PositionST, true},
		{PositionCB, false},
		{PositionCM, false},
	}

	for _, tt := range tests {
		player := &Player{Position: tt.position}
		if player.IsAttacker() != tt.expected {
			t.Errorf("Position %s: expected IsAttacker=%v, got %v", tt.position, tt.expected, player.IsAttacker())
		}
	}
}

