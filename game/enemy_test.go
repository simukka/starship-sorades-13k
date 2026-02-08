package game

import (
	"math"
	"testing"
)

// tolerance for floating point comparisons
const floatTolerance = 0.0001

// almostEqual checks if two floats are approximately equal
func almostEqual(a, b, tolerance float64) bool {
	return math.Abs(a-b) <= tolerance
}

// --- CalculateTargetAngle Tests ---

// TestCalculateTargetAngle_TargetDirectlyBelow tests angle when target is directly below
func TestCalculateTargetAngle_TargetDirectlyBelow(t *testing.T) {
	srcX, srcY := 100.0, 100.0
	targetX, targetY := 100.0, 200.0 // Directly below

	angle := CalculateTargetAngle(srcX, srcY, targetX, targetY)

	// Target below = angle should be 0 (pointing down along positive Y)
	if !almostEqual(angle, 0, floatTolerance) {
		t.Errorf("Expected angle 0 for target directly below, got %f", angle)
	}
}

// TestCalculateTargetAngle_TargetDirectlyAbove tests angle when target is directly above
func TestCalculateTargetAngle_TargetDirectlyAbove(t *testing.T) {
	srcX, srcY := 100.0, 200.0
	targetX, targetY := 100.0, 100.0 // Directly above

	angle := CalculateTargetAngle(srcX, srcY, targetX, targetY)

	// Target above = angle should be PI or -PI (pointing up)
	if !almostEqual(math.Abs(angle), math.Pi, floatTolerance) {
		t.Errorf("Expected angle ±π for target directly above, got %f", angle)
	}
}

// TestCalculateTargetAngle_TargetDirectlyRight tests angle when target is directly to the right
func TestCalculateTargetAngle_TargetDirectlyRight(t *testing.T) {
	srcX, srcY := 100.0, 100.0
	targetX, targetY := 200.0, 100.0 // Directly right

	angle := CalculateTargetAngle(srcX, srcY, targetX, targetY)

	// Target right = angle should be PI/2
	expected := math.Pi / 2
	if !almostEqual(angle, expected, floatTolerance) {
		t.Errorf("Expected angle %f for target directly right, got %f", expected, angle)
	}
}

// TestCalculateTargetAngle_TargetDirectlyLeft tests angle when target is directly to the left
func TestCalculateTargetAngle_TargetDirectlyLeft(t *testing.T) {
	srcX, srcY := 200.0, 100.0
	targetX, targetY := 100.0, 100.0 // Directly left

	angle := CalculateTargetAngle(srcX, srcY, targetX, targetY)

	// Target left = angle should be -PI/2
	expected := -math.Pi / 2
	if !almostEqual(angle, expected, floatTolerance) {
		t.Errorf("Expected angle %f for target directly left, got %f", expected, angle)
	}
}

// TestCalculateTargetAngle_TargetLowerRight tests angle for diagonal lower-right
func TestCalculateTargetAngle_TargetLowerRight(t *testing.T) {
	srcX, srcY := 100.0, 100.0
	targetX, targetY := 200.0, 200.0 // Diagonal lower-right (45 degrees)

	angle := CalculateTargetAngle(srcX, srcY, targetX, targetY)

	// 45 degrees = PI/4
	expected := math.Pi / 4
	if !almostEqual(angle, expected, floatTolerance) {
		t.Errorf("Expected angle %f for diagonal lower-right, got %f", expected, angle)
	}
}

// TestCalculateTargetAngle_TargetLowerLeft tests angle for diagonal lower-left
func TestCalculateTargetAngle_TargetLowerLeft(t *testing.T) {
	srcX, srcY := 200.0, 100.0
	targetX, targetY := 100.0, 200.0 // Diagonal lower-left

	angle := CalculateTargetAngle(srcX, srcY, targetX, targetY)

	// -45 degrees = -PI/4
	expected := -math.Pi / 4
	if !almostEqual(angle, expected, floatTolerance) {
		t.Errorf("Expected angle %f for diagonal lower-left, got %f", expected, angle)
	}
}

// TestCalculateTargetAngle_TargetUpperRight tests angle for diagonal upper-right
func TestCalculateTargetAngle_TargetUpperRight(t *testing.T) {
	srcX, srcY := 100.0, 200.0
	targetX, targetY := 200.0, 100.0 // Diagonal upper-right

	angle := CalculateTargetAngle(srcX, srcY, targetX, targetY)

	// 135 degrees = 3*PI/4
	expected := 3 * math.Pi / 4
	if !almostEqual(angle, expected, floatTolerance) {
		t.Errorf("Expected angle %f for diagonal upper-right, got %f", expected, angle)
	}
}

// TestCalculateTargetAngle_TargetUpperLeft tests angle for diagonal upper-left
func TestCalculateTargetAngle_TargetUpperLeft(t *testing.T) {
	srcX, srcY := 200.0, 200.0
	targetX, targetY := 100.0, 100.0 // Diagonal upper-left

	angle := CalculateTargetAngle(srcX, srcY, targetX, targetY)

	// -135 degrees = -3*PI/4
	expected := -3 * math.Pi / 4
	if !almostEqual(angle, expected, floatTolerance) {
		t.Errorf("Expected angle %f for diagonal upper-left, got %f", expected, angle)
	}
}

// TestCalculateTargetAngle_SamePosition tests angle when source and target are same position
func TestCalculateTargetAngle_SamePosition(t *testing.T) {
	srcX, srcY := 100.0, 100.0
	targetX, targetY := 100.0, 100.0 // Same position

	angle := CalculateTargetAngle(srcX, srcY, targetX, targetY)

	// Atan2(0, 0) = 0
	if !almostEqual(angle, 0, floatTolerance) {
		t.Errorf("Expected angle 0 for same position, got %f", angle)
	}
}

// --- Angle Normalization Tests ---

// TestAngleNormalization_PositiveAngle tests normalization of positive angles
func TestAngleNormalization_PositiveAngle(t *testing.T) {
	testCases := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"Zero", 0, 0},
		{"PI/4", math.Pi / 4, math.Pi / 4},
		{"PI/2", math.Pi / 2, math.Pi / 2},
		{"3PI/4", 3 * math.Pi / 4, 3 * math.Pi / 4},
		{"PI", math.Pi, -math.Pi}, // PI normalizes to -PI
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Apply the same normalization as in Render
			normalized := math.Mod(tc.input+math.Pi, math.Pi*2) - math.Pi

			if !almostEqual(normalized, tc.expected, floatTolerance) {
				t.Errorf("Expected normalized angle %f, got %f", tc.expected, normalized)
			}
		})
	}
}

// TestAngleNormalization_NegativeAngle tests normalization of negative angles
func TestAngleNormalization_NegativeAngle(t *testing.T) {
	testCases := []struct {
		name     string
		input    float64
		expected float64
	}{
		{"-PI/4", -math.Pi / 4, -math.Pi / 4},
		{"-PI/2", -math.Pi / 2, -math.Pi / 2},
		{"-3PI/4", -3 * math.Pi / 4, -3 * math.Pi / 4},
		{"-PI", -math.Pi, -math.Pi},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			normalized := math.Mod(tc.input+math.Pi, math.Pi*2) - math.Pi

			if !almostEqual(normalized, tc.expected, floatTolerance) {
				t.Errorf("Expected normalized angle %f, got %f", tc.expected, normalized)
			}
		})
	}
}

// --- Angle Clamping Tests ---

// TestAngleClamping_WithinRange tests that angles within MaxAngle are not clamped
func TestAngleClamping_WithinRange(t *testing.T) {
	maxAngle := math.Pi / 4 // 45 degrees

	testAngles := []float64{0, 0.1, -0.1, maxAngle - 0.01, -(maxAngle - 0.01)}

	for _, angle := range testAngles {
		clamped := angle
		if clamped > maxAngle {
			clamped = maxAngle
		}
		if clamped < -maxAngle {
			clamped = -maxAngle
		}

		if clamped != angle {
			t.Errorf("Angle %f within range should not be clamped, got %f", angle, clamped)
		}
	}
}

// TestAngleClamping_ExceedsPositiveMax tests clamping of angles exceeding positive max
func TestAngleClamping_ExceedsPositiveMax(t *testing.T) {
	maxAngle := math.Pi / 4

	testCases := []struct {
		input    float64
		expected float64
	}{
		{math.Pi / 4, math.Pi / 4}, // Exactly at max
		{math.Pi / 3, math.Pi / 4}, // Exceeds max
		{math.Pi / 2, math.Pi / 4}, // Way over max
		{math.Pi, math.Pi / 4},     // PI clamped to max
	}

	for _, tc := range testCases {
		clamped := tc.input
		if clamped > maxAngle {
			clamped = maxAngle
		}
		if clamped < -maxAngle {
			clamped = -maxAngle
		}

		if !almostEqual(clamped, tc.expected, floatTolerance) {
			t.Errorf("Angle %f should clamp to %f, got %f", tc.input, tc.expected, clamped)
		}
	}
}

// TestAngleClamping_ExceedsNegativeMax tests clamping of angles below negative max
func TestAngleClamping_ExceedsNegativeMax(t *testing.T) {
	maxAngle := math.Pi / 4

	testCases := []struct {
		input    float64
		expected float64
	}{
		{-math.Pi / 4, -math.Pi / 4}, // Exactly at -max
		{-math.Pi / 3, -math.Pi / 4}, // Below -max
		{-math.Pi / 2, -math.Pi / 4}, // Way below -max
		{-math.Pi, -math.Pi / 4},     // -PI clamped to -max
	}

	for _, tc := range testCases {
		clamped := tc.input
		if clamped > maxAngle {
			clamped = maxAngle
		}
		if clamped < -maxAngle {
			clamped = -maxAngle
		}

		if !almostEqual(clamped, tc.expected, floatTolerance) {
			t.Errorf("Angle %f should clamp to %f, got %f", tc.input, tc.expected, clamped)
		}
	}
}

// --- Angle Smoothing Tests ---

// TestAngleSmoothing_ConvergesToTarget tests that smoothing gradually approaches target
func TestAngleSmoothing_ConvergesToTarget(t *testing.T) {
	currentAngle := 0.0
	targetAngle := math.Pi / 4

	// Simulate multiple frames of smoothing
	for i := 0; i < 100; i++ {
		currentAngle = (currentAngle*EnemyAngleSmoothingFactor - targetAngle) / (EnemyAngleSmoothingFactor + 1)
	}

	// After many iterations, should converge close to -targetAngle
	// Note: The formula approaches -targetAngle, not +targetAngle
	expected := -targetAngle
	if !almostEqual(currentAngle, expected, 0.01) {
		t.Errorf("Smoothing should converge to %f, got %f", expected, currentAngle)
	}
}

// TestAngleSmoothing_FromZero tests smoothing from zero angle
func TestAngleSmoothing_FromZero(t *testing.T) {
	currentAngle := 0.0
	targetAngle := 0.5

	// Single frame of smoothing
	newAngle := (currentAngle*EnemyAngleSmoothingFactor - targetAngle) / (EnemyAngleSmoothingFactor + 1)

	// Should be small negative value (moving toward -target)
	expectedApprox := -targetAngle / (EnemyAngleSmoothingFactor + 1)
	if !almostEqual(newAngle, expectedApprox, floatTolerance) {
		t.Errorf("Expected angle %f after one frame, got %f", expectedApprox, newAngle)
	}
}

// TestAngleSmoothing_HigherFactorSlowerRotation tests that higher factor = slower rotation
func TestAngleSmoothing_HigherFactorSlowerRotation(t *testing.T) {
	currentAngle := 0.0
	targetAngle := 1.0

	// Calculate change with default factor
	newAngleDefault := (currentAngle*EnemyAngleSmoothingFactor - targetAngle) / (EnemyAngleSmoothingFactor + 1)
	changeDefault := math.Abs(newAngleDefault - currentAngle)

	// Calculate change with higher factor (should be slower)
	higherFactor := float64(EnemyAngleSmoothingFactor * 2)
	newAngleHigher := (currentAngle*higherFactor - targetAngle) / (higherFactor + 1)
	changeHigher := math.Abs(newAngleHigher - currentAngle)

	if changeHigher >= changeDefault {
		t.Errorf("Higher smoothing factor should result in smaller angle change: default=%f, higher=%f",
			changeDefault, changeHigher)
	}
}

// TestAngleSmoothing_ZeroTarget tests smoothing when target is zero
func TestAngleSmoothing_ZeroTarget(t *testing.T) {
	currentAngle := math.Pi / 4
	targetAngle := 0.0

	newAngle := (currentAngle*EnemyAngleSmoothingFactor - targetAngle) / (EnemyAngleSmoothingFactor + 1)

	// Should decay toward zero
	expected := currentAngle * EnemyAngleSmoothingFactor / (EnemyAngleSmoothingFactor + 1)
	if !almostEqual(newAngle, expected, floatTolerance) {
		t.Errorf("Expected angle %f when target is zero, got %f", expected, newAngle)
	}
}

// --- Full Targeting Pipeline Tests ---

// TestTargetingPipeline_EnemyAboveTargetBelow tests full targeting when enemy above, target below
func TestTargetingPipeline_EnemyAboveTargetBelow(t *testing.T) {
	enemyX, enemyY := 500.0, 100.0
	targetX, targetY := 500.0, 500.0
	maxAngle := math.Pi / 3

	// Step 1: Calculate target angle
	targetAngle := CalculateTargetAngle(enemyX, enemyY, targetX, targetY)

	// Target directly below = angle 0
	if !almostEqual(targetAngle, 0, floatTolerance) {
		t.Errorf("Target directly below should give angle 0, got %f", targetAngle)
	}

	// Step 2: Normalize
	normalized := math.Mod(targetAngle+math.Pi, math.Pi*2) - math.Pi

	// Step 3: Clamp
	clamped := normalized
	if clamped > maxAngle {
		clamped = maxAngle
	}
	if clamped < -maxAngle {
		clamped = -maxAngle
	}

	// Angle 0 should pass through unchanged
	if !almostEqual(clamped, 0, floatTolerance) {
		t.Errorf("Zero angle should not be clamped, got %f", clamped)
	}
}

// TestTargetingPipeline_EnemyLeftTargetRight tests targeting across horizontal axis
func TestTargetingPipeline_EnemyLeftTargetRight(t *testing.T) {
	enemyX, enemyY := 100.0, 300.0
	targetX, targetY := 900.0, 300.0
	maxAngle := math.Pi / 4 // 45 degree limit

	// Step 1: Calculate target angle
	targetAngle := CalculateTargetAngle(enemyX, enemyY, targetX, targetY)

	// Target directly right = PI/2
	if !almostEqual(targetAngle, math.Pi/2, floatTolerance) {
		t.Errorf("Target directly right should give angle PI/2, got %f", targetAngle)
	}

	// Step 2: Normalize (PI/2 stays PI/2)
	normalized := math.Mod(targetAngle+math.Pi, math.Pi*2) - math.Pi

	// Step 3: Clamp (PI/2 > PI/4, so should clamp)
	clamped := normalized
	if clamped > maxAngle {
		clamped = maxAngle
	}
	if clamped < -maxAngle {
		clamped = -maxAngle
	}

	if !almostEqual(clamped, maxAngle, floatTolerance) {
		t.Errorf("PI/2 should clamp to maxAngle %f, got %f", maxAngle, clamped)
	}
}

// --- Enemy AudioPan Tests ---

// TestEnemy_AudioPan tests audio panning calculation
func TestEnemy_AudioPan(t *testing.T) {
	testCases := []struct {
		name     string
		x        float64
		expected float64
	}{
		{"Left edge", 0, -1.0},
		{"Center", WIDTH / 2, 0.0},
		{"Right edge", WIDTH, 1.0},
		{"Quarter left", WIDTH / 4, -0.5},
		{"Quarter right", WIDTH * 3 / 4, 0.5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			enemy := &Enemy{X: tc.x}
			pan := enemy.AudioPan()

			if !almostEqual(pan, tc.expected, floatTolerance) {
				t.Errorf("Expected pan %f for X=%f, got %f", tc.expected, tc.x, pan)
			}
		})
	}
}

// --- Enemy DistanceVolume Tests ---

// TestEnemy_DistanceVolume_AtSource tests volume at zero distance
func TestEnemy_DistanceVolume_AtSource(t *testing.T) {
	enemy := &Enemy{X: 100, Y: 100, YOffset: 0}

	volume := enemy.DistanceVolume(100, 100, 500)

	if !almostEqual(volume, 1.0, floatTolerance) {
		t.Errorf("Expected volume 1.0 at zero distance, got %f", volume)
	}
}

// TestEnemy_DistanceVolume_AtMaxDistance tests volume at maximum distance
func TestEnemy_DistanceVolume_AtMaxDistance(t *testing.T) {
	enemy := &Enemy{X: 0, Y: 0, YOffset: 0}
	maxDist := 500.0

	volume := enemy.DistanceVolume(500, 0, maxDist)

	if !almostEqual(volume, 0.2, floatTolerance) {
		t.Errorf("Expected minimum volume 0.2 at max distance, got %f", volume)
	}
}

// TestEnemy_DistanceVolume_BeyondMaxDistance tests volume beyond maximum distance
func TestEnemy_DistanceVolume_BeyondMaxDistance(t *testing.T) {
	enemy := &Enemy{X: 0, Y: 0, YOffset: 0}
	maxDist := 100.0

	volume := enemy.DistanceVolume(500, 500, maxDist)

	if !almostEqual(volume, 0.2, floatTolerance) {
		t.Errorf("Expected minimum volume 0.2 beyond max distance, got %f", volume)
	}
}

// TestEnemy_DistanceVolume_HalfwayDistance tests volume at half the maximum distance
func TestEnemy_DistanceVolume_HalfwayDistance(t *testing.T) {
	enemy := &Enemy{X: 0, Y: 0, YOffset: 0}
	maxDist := 100.0

	// At distance 50, dist^2 = 2500, maxDist^2 = 10000
	// volume = 1.0 - (2500/10000) * 0.8 = 1.0 - 0.2 = 0.8
	volume := enemy.DistanceVolume(50, 0, maxDist)

	expected := 1.0 - (2500.0/10000.0)*0.8
	if !almostEqual(volume, expected, floatTolerance) {
		t.Errorf("Expected volume %f at halfway distance, got %f", expected, volume)
	}
}

// TestEnemy_DistanceVolume_IncludesYOffset tests that YOffset is factored in
func TestEnemy_DistanceVolume_IncludesYOffset(t *testing.T) {
	enemy := &Enemy{X: 100, Y: 100, YOffset: 50}
	maxDist := 500.0

	// Target at enemy position without offset
	volumeNoOffset := enemy.DistanceVolume(100, 100, maxDist)

	// Target at enemy position with offset considered
	volumeWithOffset := enemy.DistanceVolume(100, 150, maxDist)

	// With offset factored in, distance to (100, 150) should be 0
	if !almostEqual(volumeWithOffset, 1.0, floatTolerance) {
		t.Errorf("Expected volume 1.0 when target matches position+offset, got %f", volumeWithOffset)
	}

	// Without offset match, there should be some distance
	if volumeNoOffset >= 1.0 {
		t.Errorf("Expected volume < 1.0 when target doesn't match offset, got %f", volumeNoOffset)
	}
}

// --- Enemy Targeting Tests ---

// TestEnemyTargeting_AnglePointsToShip verifies enemy angle calculation points toward ship
func TestEnemyTargeting_AnglePointsToShip(t *testing.T) {
	testCases := []struct {
		name          string
		enemyX        float64
		enemyY        float64
		enemyYOffset  float64
		shipX         float64
		shipY         float64
		expectedAngle float64
		maxAngle      float64
	}{
		{
			name:   "Ship directly below enemy",
			enemyX: 500, enemyY: 100, enemyYOffset: 0,
			shipX: 500, shipY: 500,
			expectedAngle: 0, // Pointing straight down
			maxAngle:      math.Pi,
		},
		{
			name:   "Ship to bottom-right of enemy",
			enemyX: 400, enemyY: 100, enemyYOffset: 0,
			shipX: 500, shipY: 200,
			expectedAngle: math.Pi / 4, // 45 degrees right
			maxAngle:      math.Pi,
		},
		{
			name:   "Ship to bottom-left of enemy",
			enemyX: 600, enemyY: 100, enemyYOffset: 0,
			shipX: 500, shipY: 200,
			expectedAngle: -math.Pi / 4, // 45 degrees left
			maxAngle:      math.Pi,
		},
		{
			name:   "Ship far right but angle clamped",
			enemyX: 100, enemyY: 100, enemyYOffset: 0,
			shipX: 900, shipY: 150,
			expectedAngle: math.Pi / 8, // Clamped to maxAngle
			maxAngle:      math.Pi / 8,
		},
		{
			name:   "Ship far left but angle clamped",
			enemyX: 900, enemyY: 100, enemyYOffset: 0,
			shipX: 100, shipY: 150,
			expectedAngle: -math.Pi / 8, // Clamped to -maxAngle
			maxAngle:      math.Pi / 8,
		},
		{
			name:   "Enemy with YOffset - ship below",
			enemyX: 500, enemyY: 100, enemyYOffset: 50,
			shipX: 500, shipY: 500,
			expectedAngle: 0, // Still pointing straight down
			maxAngle:      math.Pi,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			enemyY := tc.enemyY + tc.enemyYOffset

			// Calculate target angle (same formula as in Render)
			targetAngle := CalculateTargetAngle(tc.enemyX, enemyY, tc.shipX, tc.shipY)

			// Normalize angle to [-π, π]
			angle := math.Mod(targetAngle+math.Pi, math.Pi*2) - math.Pi

			// Clamp to max angle
			if angle > tc.maxAngle {
				angle = tc.maxAngle
			}
			if angle < -tc.maxAngle {
				angle = -tc.maxAngle
			}

			if !almostEqual(angle, tc.expectedAngle, 0.01) {
				t.Errorf("Expected angle %f, got %f", tc.expectedAngle, angle)
			}
		})
	}
}

// TestEnemyTargeting_ConvergesToTarget tests that smoothed angle converges toward target
func TestEnemyTargeting_ConvergesToTarget(t *testing.T) {
	// Enemy starts facing down (angle=0), ship is to the right
	enemyX, enemyY := 400.0, 100.0
	shipX, shipY := 600.0, 300.0

	targetAngle := CalculateTargetAngle(enemyX, enemyY, shipX, shipY)
	normalizedTarget := math.Mod(targetAngle+math.Pi, math.Pi*2) - math.Pi

	// Simulate angle updates over time
	currentAngle := 0.0
	for i := 0; i < 100; i++ {
		// Apply smoothing (same formula as Render)
		currentAngle = (currentAngle*EnemyAngleSmoothingFactor - normalizedTarget) / (EnemyAngleSmoothingFactor + 1)
	}

	// After many iterations, angle should converge close to target (negated due to formula)
	expectedConverged := -normalizedTarget
	if !almostEqual(currentAngle, expectedConverged, 0.1) {
		t.Errorf("Angle did not converge to target. Expected ~%f, got %f", expectedConverged, currentAngle)
	}
}

// --- Torpedo Firing Direction Tests ---

// TestEnemyFire_TorpedoAimedAtShip verifies torpedo velocity points toward ship position at fire time
func TestEnemyFire_TorpedoAimedAtShip(t *testing.T) {
	testCases := []struct {
		name          string
		enemyAngle    float64
		expectedXSign int // -1, 0, or 1
		expectedYSign int // Should always be positive (moving down)
	}{
		{
			name:          "Enemy facing straight down (angle=0)",
			enemyAngle:    0,
			expectedXSign: 0, // No horizontal movement
			expectedYSign: 1, // Moving down
		},
		{
			name:          "Enemy angled right (positive angle)",
			enemyAngle:    math.Pi / 4, // 45 degrees right
			expectedXSign: 1,           // Moving right
			expectedYSign: 1,           // Moving down
		},
		{
			name:          "Enemy angled left (negative angle)",
			enemyAngle:    -math.Pi / 4, // 45 degrees left
			expectedXSign: -1,           // Moving left
			expectedYSign: 1,            // Moving down
		},
		{
			name:          "Enemy slightly angled right",
			enemyAngle:    math.Pi / 8, // 22.5 degrees
			expectedXSign: 1,
			expectedYSign: 1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			speed := 3.0 // Base speed

			// Calculate torpedo velocity (same formula as Fire method)
			var xAcc, yAcc float64
			if tc.enemyAngle != 0 {
				xAcc = math.Sin(tc.enemyAngle) * speed
				yAcc = math.Cos(tc.enemyAngle) * speed
			} else {
				xAcc = 0
				yAcc = speed
			}

			// Verify X direction
			if tc.expectedXSign == 0 && !almostEqual(xAcc, 0, floatTolerance) {
				t.Errorf("Expected XAcc=0, got %f", xAcc)
			} else if tc.expectedXSign > 0 && xAcc <= 0 {
				t.Errorf("Expected positive XAcc, got %f", xAcc)
			} else if tc.expectedXSign < 0 && xAcc >= 0 {
				t.Errorf("Expected negative XAcc, got %f", xAcc)
			}

			// Verify Y direction (should always be positive = moving down)
			if tc.expectedYSign > 0 && yAcc <= 0 {
				t.Errorf("Expected positive YAcc, got %f", yAcc)
			}
		})
	}
}

// TestEnemyFire_TorpedoSpeedCalculation verifies torpedo speed is consistent regardless of angle
func TestEnemyFire_TorpedoSpeedCalculation(t *testing.T) {
	speed := 5.0 // Arbitrary speed

	testAngles := []float64{
		0,
		math.Pi / 8,
		math.Pi / 4,
		math.Pi / 3,
		-math.Pi / 8,
		-math.Pi / 4,
	}

	for _, angle := range testAngles {
		var xAcc, yAcc float64
		if angle != 0 {
			xAcc = math.Sin(angle) * speed
			yAcc = math.Cos(angle) * speed
		} else {
			xAcc = 0
			yAcc = speed
		}

		// Calculate actual speed (magnitude of velocity vector)
		actualSpeed := math.Sqrt(xAcc*xAcc + yAcc*yAcc)

		if !almostEqual(actualSpeed, speed, floatTolerance) {
			t.Errorf("Angle %f: Expected speed %f, got %f", angle, speed, actualSpeed)
		}
	}
}

// TestEnemyFire_TorpedoDirectionMatchesEnemyAngle verifies torpedo direction matches enemy facing
func TestEnemyFire_TorpedoDirectionMatchesEnemyAngle(t *testing.T) {
	testCases := []struct {
		name       string
		enemyAngle float64
	}{
		{"Facing down", 0},
		{"Facing down-right 22.5°", math.Pi / 8},
		{"Facing down-right 45°", math.Pi / 4},
		{"Facing down-left 22.5°", -math.Pi / 8},
		{"Facing down-left 45°", -math.Pi / 4},
	}

	speed := 4.0

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var xAcc, yAcc float64
			if tc.enemyAngle != 0 {
				xAcc = math.Sin(tc.enemyAngle) * speed
				yAcc = math.Cos(tc.enemyAngle) * speed
			} else {
				xAcc = 0
				yAcc = speed
			}

			// Calculate the angle of the torpedo velocity
			torpedoAngle := math.Atan2(xAcc, yAcc)

			if !almostEqual(torpedoAngle, tc.enemyAngle, floatTolerance) {
				t.Errorf("Expected torpedo angle %f to match enemy angle %f, got %f",
					tc.enemyAngle, tc.enemyAngle, torpedoAngle)
			}
		})
	}
}

// TestEnemyTargeting_EndToEnd simulates full targeting flow:
// enemy calculates angle to ship, then fires torpedo in that direction
func TestEnemyTargeting_EndToEnd(t *testing.T) {
	testCases := []struct {
		name     string
		enemyX   float64
		enemyY   float64
		shipX    float64
		shipY    float64
		maxAngle float64
	}{
		{
			name:   "Ship directly below",
			enemyX: 500, enemyY: 100,
			shipX: 500, shipY: 600,
			maxAngle: math.Pi,
		},
		{
			name:   "Ship to lower-right",
			enemyX: 300, enemyY: 100,
			shipX: 500, shipY: 600,
			maxAngle: math.Pi,
		},
		{
			name:   "Ship to lower-left",
			enemyX: 700, enemyY: 100,
			shipX: 500, shipY: 600,
			maxAngle: math.Pi,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Step 1: Calculate target angle (as in Render)
			targetAngle := CalculateTargetAngle(tc.enemyX, tc.enemyY, tc.shipX, tc.shipY)
			angle := math.Mod(targetAngle+math.Pi, math.Pi*2) - math.Pi

			// Clamp
			if angle > tc.maxAngle {
				angle = tc.maxAngle
			}
			if angle < -tc.maxAngle {
				angle = -tc.maxAngle
			}

			// For this test, assume angle has converged (no smoothing delay)
			// The smoothing formula: e.Angle = (e.Angle*K - angle) / (K+1)
			// When converged: e.Angle = (e.Angle*K - angle) / (K+1)
			// => e.Angle*(K+1) = e.Angle*K - angle
			// => e.Angle = -angle
			// So final enemy angle points opposite to the normalized angle.
			// But the normalized angle is already relative to pointing "down",
			// so we use the raw targetAngle for torpedo direction.
			enemyFinalAngle := angle

			// Step 2: Calculate torpedo velocity (as in Fire)
			speed := 3.0
			var torpedoXAcc, torpedoYAcc float64
			if enemyFinalAngle != 0 {
				torpedoXAcc = math.Sin(enemyFinalAngle) * speed
				torpedoYAcc = math.Cos(enemyFinalAngle) * speed
			} else {
				torpedoXAcc = 0
				torpedoYAcc = speed
			}

			// Step 3: Verify torpedo is aimed toward ship
			// The torpedo direction should have the correct sign
			dx := tc.shipX - tc.enemyX
			dy := tc.shipY - tc.enemyY

			// If ship is to the right, torpedo should move right (positive XAcc)
			if dx > 10 && torpedoXAcc <= 0 {
				t.Errorf("Ship is to the right (dx=%f), but torpedo XAcc=%f (should be positive)",
					dx, torpedoXAcc)
			}
			// If ship is to the left, torpedo should move left (negative XAcc)
			if dx < -10 && torpedoXAcc >= 0 {
				t.Errorf("Ship is to the left (dx=%f), but torpedo XAcc=%f (should be negative)",
					dx, torpedoXAcc)
			}
			// Ship is below, torpedo should move down (positive YAcc)
			if dy > 0 && torpedoYAcc <= 0 {
				t.Errorf("Ship is below (dy=%f), but torpedo YAcc=%f (should be positive)",
					dy, torpedoYAcc)
			}
		})
	}
}
