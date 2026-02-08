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

// World constants for infinite world mode
const (
	// EnemySpawnDistance is how far from the ship enemies spawn
	EnemySpawnDistance = 2400.0
	// EnemyDespawnDistance is how far enemies can be before being removed
	EnemyDespawnDistance = EnemySpawnDistance * 5
	// ShipRotationSpeed is how fast the ship rotates (radians per frame)
	ShipRotationSpeed = 0.1
	// ShipThrustAcc is the acceleration when thrusting forward
	ShipThrustAcc = 5.0
	// ShipMaxSpeed is the maximum velocity magnitude
	ShipMaxSpeed = 16.0
)

// Camera represents the viewport into the infinite world.
// The camera is centered on the player ship.
type Camera struct {
	X, Y float64 // World position of camera center
}

// WorldToScreen converts world coordinates to screen coordinates.
func (c *Camera) WorldToScreen(worldX, worldY float64) (screenX, screenY float64) {
	screenX = worldX - c.X + WIDTH/2
	screenY = worldY - c.Y + HEIGHT/2
	return
}

// ScreenToWorld converts screen coordinates to world coordinates.
func (c *Camera) ScreenToWorld(screenX, screenY float64) (worldX, worldY float64) {
	worldX = screenX + c.X - WIDTH/2
	worldY = screenY + c.Y - HEIGHT/2
	return
}

// IsOnScreen checks if a world position is visible on screen (with padding).
func (c *Camera) IsOnScreen(worldX, worldY, padding float64) bool {
	screenX, screenY := c.WorldToScreen(worldX, worldY)
	return screenX >= -padding && screenX <= WIDTH+padding &&
		screenY >= -padding && screenY <= HEIGHT+padding
}

// --- Entity Interface ---

// Entity is the common interface for all game entities (ships, enemies).
// This allows the game loop to treat player ships and enemies uniformly.
type Entity interface {
	// GetPosition returns the entity's world coordinates.
	GetPosition() (x, y float64)

	// GetAngle returns the entity's facing angle in radians.
	GetAngle() float64

	// GetHealth returns the entity's current health/energy.
	GetHealth() int

	// SetHealth sets the entity's health/energy.
	SetHealth(health int)

	// GetRadius returns the entity's collision radius.
	GetRadius() float64

	// IsAlive returns true if the entity is still active.
	IsAlive() bool

	// Update updates the entity's state for one frame, including rendering.
	// Returns false if the entity should be removed.
	Update(g *Game) bool

	// AudioPan returns stereo pan value based on position (-1 to 1).
	AudioPan() float64
}

// Compile-time interface checks
var (
	_ Entity = (*Ship)(nil)
	_ Entity = (*Enemy)(nil)
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
	EnemyAngleSmoothingFactor = 20
)

// Base constants
const (
	// BaseShieldRadius is the radius of the base's protective shield.
	// Ships can enter, but enemies and torpedos cannot.
	BaseShieldRadius = float64(ShipR) * 6

	// BaseRadius is the visual radius of the base structure itself.
	BaseRadius = float64(ShipR) * 1.5
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
