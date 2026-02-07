package game

import (
	"github.com/gopherjs/gopherjs/js"
)

// Constants for game configuration
const (
	WIDTH         = 1024
	HEIGHT        = 768
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
)

// Projectile constants
const (
	BulletR    = 8
	BulletMaxT = 35
	TorpedoR   = 16
	BonusR     = 16
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

// Ship holds player ship state.
type Ship struct {
	X, Y          float64
	XAcc, YAcc    float64
	Angle         float64
	E             int // Energy/health
	Timeout       int
	Weapon        int
	Reload        int
	OSD           int
	Shield        Shield
	Image         *js.Object
	OriginalImage *js.Object
}

// Enemy represents an enemy entity.
type Enemy struct {
	Image         *js.Object
	X, Y          float64
	YStop         float64
	YOffset       float64
	R             float64
	Angle         float64
	MaxAngle      float64
	E             int
	T             int
	FireDirection float64
	TActive       int
	TypeIndex     int
}

// EnemyType defines an enemy type's behavior and appearance.
type EnemyType struct {
	R     float64
	Image *js.Object
}
