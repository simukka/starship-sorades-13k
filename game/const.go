package game

import (
	"github.com/gopherjs/gopherjs/js"
)

// Constants for game configuration
const (
	WIDTH         = 2048
	HEIGHT        = 1080
	Speed         = 2
	MaxBomb       = 5
	FrameDuration = 33.33 // ~30 FPS
)

// Ship constants
const (
	ShipR           = 48 // Ship collision radius
	ShipACC         = 1.5
	ShipACCFactor   = 0.9
	ShipAngleFactor = 0.8
	ShipMaxAngle    = 10
	ShipMaxOSD      = 180 // 6 * 30
	ShipMaxShield   = 300 // 10 * 30 frames

	// ShipCollisionD is the vertical collision half-extent (depth along Y axis).
	// Derived as 80% of ShipR for a tighter vertical hitbox.
	ShipCollisionD = float64(ShipR) * 0.8

	// ShipCollisionE is the horizontal collision half-extent (extent along X axis).
	// Derived as 40% of ShipR for a narrower horizontal hitbox.
	ShipCollisionE = float64(ShipR) * 0.4
)

// Projectile constants
const (
	BulletR                    = 8
	BulletMaxT                 = 35
	TorpedoR                   = 16
	BonusR                     = 16
	BulletTorpedoCollisionDist = 12.0
)

// Enemy constants
const (
	// EnemyAngleSmoothingFactor controls how quickly enemies rotate toward their target.
	// Higher values = slower, smoother rotation. Lower values = faster, snappier rotation.
	// The formula is: newAngle = (currentAngle * factor - targetAngle) / (factor + 1)
	EnemyAngleSmoothingFactor = 1
)

// Shield holds shield state.
type Shield struct {
	MaxT  int
	T     int
	Image *js.Object
}

// TextDisplay holds text display state.
type TextDisplay struct {
	MaxT  int
	T     int
	X     int
	Y     int
	YAcc  float64
	Image *js.Object
}

// Points holds score display configuration.
type Points struct {
	Width  int
	Height int
	Step   int
	Images []*js.Object
}

// Level holds the game level/state.
type Level struct {
	Y                 float64
	Bomb              int
	P                 int // Score
	LevelNum          int
	Paused            bool
	LevelSeed         uint32
	Text              TextDisplay
	Points            Points
	Background        *js.Object
	BackgroundPattern *js.Object
}
