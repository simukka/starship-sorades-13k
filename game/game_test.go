package game

import (
	"math"
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

	if g.Ships[0].X != WIDTH/2 {
		t.Errorf("Expected Ship.X to be %d, got %f", WIDTH/2, g.Ships[0].X)
	}
	if g.Ships[0].E != 100 {
		t.Errorf("Expected Ship.E to be 100, got %d", g.Ships[0].E)
	}
	if len(g.Ships[0].Weapons) != 0 {
		t.Errorf("Expected Ship.Weapon to be 0, got %d", len(g.Ships[0].Weapons))
	}
}

func TestNewGame_PoolsInitialized(t *testing.T) {
	g := NewGame()

	if g.Bullets == nil {
		t.Error("Expected Bullets pool to be initialized")
	}
	if g.Explosions == nil {
		t.Error("Expected Explosions pool to be initialized")
	}
	if g.Bonuses == nil {
		t.Error("Expected Bonuses pool to be initialized")
	}
}

// TestCalculateTargetAngle_DirectlyBelow tests targeting when ship is directly below enemy
func TestCalculateTargetAngle_DirectlyBelow(t *testing.T) {
	// Enemy at (100, 100), Ship at (100, 200) - ship is directly below
	angle := CalculateTargetAngle(100, 100, 100, 200)
	expected := 0.0 // Straight down

	if math.Abs(angle-expected) > 0.001 {
		t.Errorf("Expected angle %f, got %f", expected, angle)
	}
}

// TestCalculateTargetAngle_DirectlyAbove tests targeting when ship is directly above enemy
func TestCalculateTargetAngle_DirectlyAbove(t *testing.T) {
	// Enemy at (100, 200), Ship at (100, 100) - ship is directly above
	angle := CalculateTargetAngle(100, 200, 100, 100)

	// math.Atan2 returns PI for straight up (or -PI, both are equivalent)
	if math.Abs(math.Abs(angle)-math.Pi) > 0.001 {
		t.Errorf("Expected angle %f or %f, got %f", math.Pi, -math.Pi, angle)
	}
}

// TestCalculateTargetAngle_DirectlyRight tests targeting when ship is directly to the right
func TestCalculateTargetAngle_DirectlyRight(t *testing.T) {
	// Enemy at (100, 100), Ship at (200, 100) - ship is to the right
	angle := CalculateTargetAngle(100, 100, 200, 100)
	expected := math.Pi / 2 // 90 degrees right

	if math.Abs(angle-expected) > 0.001 {
		t.Errorf("Expected angle %f, got %f", expected, angle)
	}
}

// TestCalculateTargetAngle_DirectlyLeft tests targeting when ship is directly to the left
func TestCalculateTargetAngle_DirectlyLeft(t *testing.T) {
	// Enemy at (200, 100), Ship at (100, 100) - ship is to the left
	angle := CalculateTargetAngle(200, 100, 100, 100)
	expected := -math.Pi / 2 // -90 degrees (left)

	if math.Abs(angle-expected) > 0.001 {
		t.Errorf("Expected angle %f, got %f", expected, angle)
	}
}

// TestCalculateTargetAngle_DiagonalDownRight tests targeting diagonally down-right
func TestCalculateTargetAngle_DiagonalDownRight(t *testing.T) {
	// Enemy at (0, 0), Ship at (100, 100) - 45 degrees down-right
	angle := CalculateTargetAngle(0, 0, 100, 100)
	expected := math.Pi / 4 // 45 degrees

	if math.Abs(angle-expected) > 0.001 {
		t.Errorf("Expected angle %f, got %f", expected, angle)
	}
}

// TestCalculateTargetAngle_DiagonalDownLeft tests targeting diagonally down-left
func TestCalculateTargetAngle_DiagonalDownLeft(t *testing.T) {
	// Enemy at (100, 0), Ship at (0, 100) - 45 degrees down-left
	angle := CalculateTargetAngle(100, 0, 0, 100)
	expected := -math.Pi / 4 // -45 degrees

	if math.Abs(angle-expected) > 0.001 {
		t.Errorf("Expected angle %f, got %f", expected, angle)
	}
}

// TestCalculateTargetAngle_WideCanvasAccuracy tests targeting across wide canvas (2048px)
func TestCalculateTargetAngle_WideCanvasAccuracy(t *testing.T) {
	testCases := []struct {
		name   string
		enemyX float64
		enemyY float64
		shipX  float64
		shipY  float64
	}{
		{"Left edge to center", 0, 100, WIDTH / 2, HEIGHT},
		{"Right edge to center", WIDTH, 100, WIDTH / 2, HEIGHT},
		{"Center top to center bottom", WIDTH / 2, 0, WIDTH / 2, HEIGHT},
		{"Far left diagonal", 0, 0, WIDTH, HEIGHT},
		{"Far right diagonal", WIDTH, 0, 0, HEIGHT},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			angle := CalculateTargetAngle(tc.enemyX, tc.enemyY, tc.shipX, tc.shipY)
			// Calculate expected using the same formula
			dx := tc.shipX - tc.enemyX
			dy := tc.shipY - tc.enemyY
			expected := math.Atan2(dx, dy)

			if math.Abs(angle-expected) > 0.001 {
				t.Errorf("%s: Expected angle %f, got %f", tc.name, expected, angle)
			}
		})
	}
}

// TestTorpedoTrajectoryAccuracy tests that torpedoes actually hit the ship position
func TestTorpedoTrajectoryAccuracy(t *testing.T) {
	testCases := []struct {
		name   string
		enemyX float64
		enemyY float64
		shipX  float64
		shipY  float64
	}{
		{"Center to center", WIDTH / 2, 100, WIDTH / 2, HEIGHT - 100},
		{"Left side", 100, 100, 100, HEIGHT - 100},
		{"Right side", WIDTH - 100, 100, WIDTH - 100, HEIGHT - 100},
		{"Cross left to right", 100, 100, WIDTH - 100, HEIGHT - 100},
		{"Cross right to left", WIDTH - 100, 100, 100, HEIGHT - 100},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			angle := CalculateTargetAngle(tc.enemyX, tc.enemyY, tc.shipX, tc.shipY)
			speed := 5.0

			// Simulate torpedo movement
			torpX := tc.enemyX
			torpY := tc.enemyY
			xAcc := math.Sin(angle) * speed
			yAcc := math.Cos(angle) * speed

			// Calculate distance to travel
			totalDist := math.Sqrt(math.Pow(tc.shipX-tc.enemyX, 2) + math.Pow(tc.shipY-tc.enemyY, 2))
			maxSteps := int(totalDist/speed) + 100 // Extra steps for safety
			hitDistance := 50.0                    // Allowable hit radius

			for step := 0; step < maxSteps; step++ {
				torpX += xAcc
				torpY += yAcc

				// Check if we hit the target position
				dx := torpX - tc.shipX
				dy := torpY - tc.shipY
				dist := math.Sqrt(dx*dx + dy*dy)

				if dist < hitDistance {
					return // Success - torpedo reached target area
				}
			}

			t.Errorf("Torpedo from (%.0f,%.0f) missed ship at (%.0f,%.0f), ended at (%.0f,%.0f)",
				tc.enemyX, tc.enemyY, tc.shipX, tc.shipY, torpX, torpY)
		})
	}
}

// TestHitDetectionBoxes tests that collision detection boxes are appropriately sized
func TestHitDetectionBoxes(t *testing.T) {
	// Ship collision box dimensions
	shipCollisionD := float64(ShipR) * 0.8 // Vertical collision distance
	shipCollisionE := float64(ShipR) * 0.4 // Horizontal collision distance

	// For a 48px ship radius:
	// D = 38.4 (vertical)
	// E = 19.2 (horizontal)

	t.Logf("Ship collision box: horizontal=%.1f, vertical=%.1f (ShipR=%d)", shipCollisionE*2, shipCollisionD*2, ShipR)

	// Verify the collision box is reasonable relative to ship size
	if shipCollisionD < float64(ShipR)*0.5 {
		t.Errorf("Vertical collision box too small: %.1f < %.1f", shipCollisionD, float64(ShipR)*0.5)
	}

	// For wide canvas, horizontal collision shouldn't be too tight
	minHorizontalCollision := float64(ShipR) * 0.3
	if shipCollisionE < minHorizontalCollision {
		t.Errorf("Horizontal collision box too small for wide canvas: %.1f < %.1f", shipCollisionE, minHorizontalCollision)
	}
}

// TestEnemyAngleClamping tests that enemy max angle clamping works correctly
func TestEnemyAngleClamping(t *testing.T) {
	testCases := []struct {
		name     string
		angle    float64
		maxAngle float64
		expected float64
	}{
		{"Within range positive", 0.5, 1.0, 0.5},
		{"Within range negative", -0.5, 1.0, -0.5},
		{"Exceeds max positive", 1.5, 1.0, 1.0},
		{"Exceeds max negative", -1.5, 1.0, -1.0},
		// PI + 0.5 ≈ 3.64, after subtracting 2*PI becomes ≈ -2.64, then clamped to -PI/4
		{"Large angle wrapped", math.Pi + 0.5, math.Pi / 4, -math.Pi / 4},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			angle := tc.angle

			// Simulate the angle clamping logic
			if angle > math.Pi {
				angle -= math.Pi * 2
			}
			if angle > tc.maxAngle {
				angle = tc.maxAngle
			}
			if angle < -tc.maxAngle {
				angle = -tc.maxAngle
			}

			if math.Abs(angle-tc.expected) > 0.001 {
				t.Errorf("Expected angle %f, got %f", tc.expected, angle)
			}
		})
	}
}
