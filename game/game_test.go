package game

import (
	"testing"
)

func TestNewGame_LevelIsPaused(t *testing.T) {
	g := NewGame()

	if !g.Level.Paused {
		t.Errorf("Expected Level.Paused to be true, got false")
	}
}

func TestNewGame_ShipDefaults(t *testing.T) {
	g := NewGame()

	if g.Ship.X != WIDTH/2 {
		t.Errorf("Expected Ship.X to be %d, got %f", WIDTH/2, g.Ship.X)
	}
	if g.Ship.E != 100 {
		t.Errorf("Expected Ship.E to be 100, got %d", g.Ship.E)
	}
	if g.Ship.Weapon != 0 {
		t.Errorf("Expected Ship.Weapon to be 0, got %d", g.Ship.Weapon)
	}
}

func TestNewGame_PoolsInitialized(t *testing.T) {
	g := NewGame()

	if g.Bullets == nil {
		t.Error("Expected Bullets pool to be initialized")
	}
	if g.Torpedos == nil {
		t.Error("Expected Torpedos pool to be initialized")
	}
	if g.Explosions == nil {
		t.Error("Expected Explosions pool to be initialized")
	}
	if g.Bonuses == nil {
		t.Error("Expected Bonuses pool to be initialized")
	}
}
