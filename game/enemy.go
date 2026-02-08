package game

import (
	"math"
	"strconv"

	"github.com/gopherjs/gopherjs/js"
)

type EnemyKind int

const (
	SmallFighter EnemyKind = iota
	MediumFighter
	TurretFighter
	Boss
)

// EnemyKindNames maps EnemyKind to display names for the UI
var EnemyKindNames = map[EnemyKind]string{
	SmallFighter:  "Small Fighter",
	MediumFighter: "Medium Fighter",
	TurretFighter: "Turret",
	Boss:          "Boss",
}

// Enemy represents an enemy entity.
type Enemy struct {
	Image         *js.Object
	X, Y          float64
	VelX, VelY    float64 // Velocity for predictive targeting
	YStop         float64
	YOffset       float64
	Radius        float64
	Angle         float64
	MaxAngle      float64
	Health        int
	MaxHealth     int // Maximum health for health bar calculation
	OSD           int // On-screen display timer for health bar
	FireTimer     int
	FireDirection float64
	TActive       int
	TypeIndex     int
	NetworkID     int // Unique ID for multiplayer synchronization

	Target *Ship
	Kind   EnemyKind
}

// EnemySpawnConfig holds configuration for spawning enemies.
type EnemySpawnConfig struct {
	CountBase       int
	CountPerPoints  int // Points threshold for +1 enemy count (e.g., 1000 = +1 per 1000 points)
	MaxAngle        float64
	HealthBase      int
	HealthPerPoints int // Points threshold for +1 health (e.g., 500 = +1 per 500 points)
	TBase           int
	TRange          int
	TStart          int
	YStopBase       float64
	YStopRange      float64
	YOffsetMult     float64 // yoffset multiplyer
	UseYStep        bool    // For turret-style Y positioning
	HasFireDir      bool    // For turrets with rotating fire
}

// Standard enemy spawn configurations
var enemyConfigs = map[EnemyKind]EnemySpawnConfig{
	// Type 0: Small fighters - spawn frequently, scale with points
	SmallFighter: {
		CountBase: 3, CountPerPoints: 2000,
		MaxAngle: math.Pi / 32, HealthBase: 8, HealthPerPoints: 1000,
		TBase: 0, TRange: 120, YStopBase: HEIGHT / 8, YStopRange: HEIGHT / 4,
	},
	// Type 1: Medium fighters - fewer but tougher
	MediumFighter: {
		CountBase: 1, CountPerPoints: 3000,
		MaxAngle: math.Pi / 16, HealthBase: 15, HealthPerPoints: 800,
		TBase: 0, TRange: 120, YStopBase: HEIGHT / 8, YStopRange: HEIGHT / 4,
	},
	// Type 2: Turrets - rare early, more common later
	TurretFighter: {
		CountBase: 0, CountPerPoints: 5000,
		MaxAngle: math.Pi * 32, HealthBase: 20, HealthPerPoints: 600,
		TBase: 0, TRange: 30, YStopBase: HEIGHT / 2, YStopRange: 0,
		UseYStep: true, HasFireDir: true,
	},
	Boss: {
		CountBase: 0, CountPerPoints: 10000,
		YOffsetMult: 0.6,
		MaxAngle:    math.Pi / 8, HealthBase: 30, HealthPerPoints: 500,
		TBase: 0, TRange: 20, TStart: 60,
	},
}

// AudioPan returns a pan value (-1.0 to 1.0) based on the enemy's screen position.
// In infinite world mode, this is relative to the camera (player ship).
// Left of camera = negative, center = 0.0, right of camera = positive.
// Clamped to [-1, 1] range.
func (e *Enemy) AudioPan() float64 { // If no target, return center pan
	if e.Target == nil {
		return 0.0
	}
	// Calculate screen X position relative to camera center
	// Camera center is at WIDTH/2, so we calculate offset from center
	screenX := e.X - e.Target.X  // Offset from player ship
	pan := screenX / (WIDTH / 2) // Normalize to [-1, 1] range (roughly)

	// Clamp to valid pan range
	if pan > 1.0 {
		return 1.0
	}
	if pan < -1.0 {
		return -1.0
	}
	return pan
}

// GetPosition implements Entity interface - returns the enemy's world coordinates.
func (e *Enemy) GetPosition() (x, y float64) {
	return e.X, e.Y + e.YOffset
}

// GetAngle implements Entity interface - returns the enemy's facing angle in radians.
func (e *Enemy) GetAngle() float64 {
	return e.Angle
}

// GetHealth implements Entity interface - returns the enemy's current health.
func (e *Enemy) GetHealth() int {
	return e.Health
}

// SetHealth implements Entity interface - sets the enemy's health.
func (e *Enemy) SetHealth(health int) {
	e.Health = health
}

// GetRadius implements Entity interface - returns the enemy's collision radius.
func (e *Enemy) GetRadius() float64 {
	return e.Radius
}

// IsAlive implements Entity interface - returns true if the enemy has health remaining.
func (e *Enemy) IsAlive() bool {
	return e.Health > 0
}

// Update implements Entity interface - updates enemy state, targeting, and movement.
// Returns false if the enemy should be removed (death or despawn).
func (e *Enemy) Update(g *Game) bool {
	return e.Render(g)
}

// DistanceVolume calculates volume based on distance to a target position.
// Closer = louder (up to 1.0), farther = quieter (minimum 0.2).
// maxDist is the distance at which volume reaches minimum.
func (s *Enemy) DistanceVolume(targetX, targetY, maxDist float64) float64 {
	dx := s.X - targetX
	dy := (s.Y + s.YOffset) - targetY
	dist := dx*dx + dy*dy // Skip sqrt for performance
	maxDistSq := maxDist * maxDist

	if dist >= maxDistSq {
		return 0.2 // Minimum volume for distant enemies
	}

	// Linear falloff: 1.0 at dist=0, 0.2 at dist=maxDist
	return 1.0 - (dist/maxDistSq)*0.8
}

// Render updates the enemy's orientation to face its target and draws the enemy sprite.
//
// Targeting behavior:
//   - Calculates the angle from enemy position to target (player ship) using CalculateTargetAngle
//   - Normalizes the angle to the range [-π, π] for consistent rotation direction
//   - Clamps the angle to MaxAngle to limit how far the enemy can rotate
//   - Applies exponential smoothing using EnemyAngleSmoothingFactor for gradual rotation:
//     newAngle = (currentAngle * EnemyAngleSmoothingFactor - clampedTargetAngle) / (EnemyAngleSmoothingFactor + 1)
//
// Rendering:
//   - Saves canvas state, translates to enemy position, rotates by current angle
//   - Draws the enemy image centered on its position
//   - Restores canvas state to avoid affecting other draw calls
func (e *Enemy) TargetAngle() float64 {
	dx := e.Target.X - e.X
	dy := e.Target.Y - e.Y

	// Atan2 gives angle from positive X axis, we want angle from positive Y axis (down)
	// Rotate by -PI/2 to convert from X-axis reference to Y-axis reference
	return math.Atan2(dx, dy)
}

func (e *Enemy) Render(g *Game) bool {
	enemyY := e.Y + e.YOffset

	// Find nearest ship to target
	var nearestShip *Ship
	nearestDist := math.MaxFloat64
	for _, ship := range g.Ships {
		dx := ship.X - e.X
		dy := ship.Y - enemyY
		dist := dx*dx + dy*dy
		if dist < nearestDist {
			nearestDist = dist
			nearestShip = ship
		}
	}
	e.Target = nearestShip

	// Check if enemy is too far from any ship - despawn
	if nearestDist > EnemyDespawnDistance*EnemyDespawnDistance {
		return false
	}

	// Normalize angle to [-π, π] range for consistent shortest-path rotation
	angle := math.Mod(e.TargetAngle()+math.Pi, math.Pi*2) - math.Pi

	// Apply exponential smoothing for gradual rotation toward target
	e.Angle = (e.Angle*EnemyAngleSmoothingFactor - angle) / (EnemyAngleSmoothingFactor + 1)

	if e.Health <= 0 {
		// Enemy destroyed - award points to nearest ship (the target)
		pointsAward := 100
		if e.Target != nil {
			e.Target.Points += pointsAward
		}

		g.SpawnBonus(e.X, e.Y, 0, 0, "")
		g.Explode(e.X, e.Y, e.Radius*2)
		g.Explode(e.X, e.Y, e.Radius*3)

		g.Audio.PlayWithPan(10,
			e.AudioPan(),
			e.DistanceVolume(g.Ship.X, g.Ship.Y, float64(HEIGHT)))
		return false
	}

	// Move enemy toward target (follow behavior)
	if e.Target != nil {
		followSpeed := 1.5 // Adjust for enemy speed
		dx := e.Target.X - e.X
		dy := e.Target.Y - enemyY
		dist := math.Sqrt(dx*dx + dy*dy)
		if dist > e.Radius*2 { // Don't get too close
			// Calculate velocity
			e.VelX = (dx / dist) * followSpeed
			e.VelY = (dy / dist) * followSpeed

			newX := e.X + e.VelX
			newY := e.Y + e.VelY

			// Check if new position would be inside a base shield
			// Check multiple points around the enemy to prevent slipping through
			blocked := false
			newEnemyY := newY + e.YOffset
			checkRadius := e.Radius * 0.8
			for _, base := range g.Bases {
				// Check center and cardinal directions
				if base.ContainsPoint(newX, newEnemyY) ||
					base.ContainsPoint(newX, newEnemyY-checkRadius) || // top
					base.ContainsPoint(newX, newEnemyY+checkRadius) || // bottom
					base.ContainsPoint(newX-checkRadius, newEnemyY) || // left
					base.ContainsPoint(newX+checkRadius, newEnemyY) { // right
					blocked = true
					break
				}
			}

			// Only move if not blocked by a shield
			if !blocked {
				e.X = newX
				e.Y = newY
			} else {
				// Blocked - zero velocity
				e.VelX = 0
				e.VelY = 0
			}
		} else {
			// Too close - stop moving
			e.VelX = 0
			e.VelY = 0
		}
	} else {
		e.VelX = 0
		e.VelY = 0
	}

	// Convert world position to screen position for rendering
	screenX, screenY := g.Camera.WorldToScreen(e.X, enemyY)

	// Only render if on screen
	if !g.Camera.IsOnScreen(e.X, enemyY, e.Radius*2) {
		// Still alive, just off-screen
		e.Fire(g)
		return true
	}

	// Render enemy
	g.Ctx.Call("save")
	g.Ctx.Call("translate", screenX, screenY)
	g.Ctx.Call("rotate", e.Angle)
	g.Ctx.Call("drawImage", e.Image,
		-e.Radius, -e.Radius, e.Radius*2, e.Radius*2)
	g.Ctx.Call("restore")

	// Render targeting reticle if this enemy is targeted or being locked
	if g.Ship.Target == e {
		// Locked target - red reticle
		g.Ctx.Call("save")
		g.Ctx.Set("strokeStyle", "#ff0000")
		g.Ctx.Set("lineWidth", 2)
		g.Ctx.Set("shadowBlur", 8)
		g.Ctx.Set("shadowColor", "#ff0000")
		g.Ctx.Call("beginPath")
		g.Ctx.Call("arc", screenX, screenY, e.Radius*1.5, 0, math.Pi*2)
		g.Ctx.Call("stroke")
		// Corner brackets
		bracketSize := e.Radius * 0.6
		bracketOffset := e.Radius * 1.2
		g.Ctx.Call("beginPath")
		// Top-left
		g.Ctx.Call("moveTo", screenX-bracketOffset, screenY-bracketOffset+bracketSize)
		g.Ctx.Call("lineTo", screenX-bracketOffset, screenY-bracketOffset)
		g.Ctx.Call("lineTo", screenX-bracketOffset+bracketSize, screenY-bracketOffset)
		// Top-right
		g.Ctx.Call("moveTo", screenX+bracketOffset-bracketSize, screenY-bracketOffset)
		g.Ctx.Call("lineTo", screenX+bracketOffset, screenY-bracketOffset)
		g.Ctx.Call("lineTo", screenX+bracketOffset, screenY-bracketOffset+bracketSize)
		// Bottom-right
		g.Ctx.Call("moveTo", screenX+bracketOffset, screenY+bracketOffset-bracketSize)
		g.Ctx.Call("lineTo", screenX+bracketOffset, screenY+bracketOffset)
		g.Ctx.Call("lineTo", screenX+bracketOffset-bracketSize, screenY+bracketOffset)
		// Bottom-left
		g.Ctx.Call("moveTo", screenX-bracketOffset+bracketSize, screenY+bracketOffset)
		g.Ctx.Call("lineTo", screenX-bracketOffset, screenY+bracketOffset)
		g.Ctx.Call("lineTo", screenX-bracketOffset, screenY+bracketOffset-bracketSize)
		g.Ctx.Call("stroke")
		g.Ctx.Call("restore")
	} else if g.Ship.LockingOn == e {
		// Locking - orange pulsing reticle
		progress := float64(g.Ship.LockMaxTime-g.Ship.LockTimer) / float64(g.Ship.LockMaxTime)
		g.Ctx.Call("save")
		g.Ctx.Set("strokeStyle", "#ff8800")
		g.Ctx.Set("lineWidth", 2)
		g.Ctx.Set("shadowBlur", 6)
		g.Ctx.Set("shadowColor", "#ff8800")
		g.Ctx.Call("beginPath")
		// Draw arc showing lock progress
		g.Ctx.Call("arc", screenX, screenY, e.Radius*1.5, -math.Pi/2, -math.Pi/2+progress*math.Pi*2)
		g.Ctx.Call("stroke")
		g.Ctx.Call("restore")
	}

	// Render health bar if recently hit
	if e.OSD > 0 {
		e.RenderHealthBar(g, screenX, screenY)
	}

	// and fire at target
	e.Fire(g)
	return true
}

func (e *Enemy) Collision(g *Game, b *Bullet) bool {
	hitboxD := e.Radius * 0.6
	if b.Kind != StandardBullet {
		return false
	}
	if b.Y < (e.Y+hitboxD) && b.Y > (e.Y-hitboxD) &&
		b.X > (e.X-hitboxD) && b.X < (e.X+hitboxD) {
		g.Level.P++
		g.Explode(b.X, b.Y, 0)
		e.Health--
		e.OSD = ShipMaxOSD // Show health bar when hit
		g.Audio.PlayWithPan(9,
			e.AudioPan(),
			e.DistanceVolume(g.Ship.X, g.Ship.Y, float64(HEIGHT)))
		return true
	}
	return false
}

func (e *Enemy) Fire(g *Game) bool {
	e.FireTimer--
	if e.FireTimer > 0 {
		// need to wait
		return false
	}

	if e.TypeIndex == 2 {

	}

	e.FireTimer = max(5, 600/(g.Level.LevelNum+4))

	torpedo := g.Bullets.AcquireKind(TorpedoBullet)
	if torpedo == nil {
		return false
	}

	// 19?
	// what is 20?
	g.Audio.PlayWithPan(20,
		torpedo.AudioPan(),
		torpedo.DistanceVolume(e.Target.X, e.Target.Y, float64(HEIGHT)))

	// speed := 3.0 + float64(g.Level.LevelNum)/2

	torpedo.X = math.Floor(e.X)
	torpedo.Y = e.Y + e.YOffset

	torpedo.XAcc = math.Sin(e.TargetAngle()) * (e.Radius / 2)
	torpedo.YAcc = math.Cos(e.TargetAngle()) * (e.Radius / 2)

	torpedo.E = 0

	return true
}

// SpawnEnemyNearShip spawns enemies of a given kind near a specific ship.
// Enemy count and strength scale with the ship's points.
func (g *Game) SpawnEnemyNearShip(kind EnemyKind, ship *Ship) bool {
	r := float64(ShipR)
	cfg, exists := enemyConfigs[kind]

	if !exists {
		return false
	}

	if g.EnemyTypes[kind].R > 0 {
		r = g.EnemyTypes[kind].R
	}

	// Calculate count based on ship's points
	count := cfg.CountBase
	if cfg.CountPerPoints > 0 {
		count += ship.Points / cfg.CountPerPoints
	}

	// Calculate health based on ship's points
	health := cfg.HealthBase
	if cfg.HealthPerPoints > 0 {
		health += ship.Points / cfg.HealthPerPoints
	}

	for i := count; i > 0; i-- {
		// Spawn at random angle around the ship
		spawnAngle := g.GameRNG.Random() * math.Pi * 2

		// Calculate spawn position at distance from ship
		spawnX := ship.X + math.Cos(spawnAngle)*EnemySpawnDistance
		spawnY := ship.Y + math.Sin(spawnAngle)*EnemySpawnDistance

		yOff := 0.0
		if kind == Boss {
			yOff = r * cfg.YOffsetMult
		}

		enemy := &Enemy{
			Image:     g.EnemyTypes[kind].Image,
			X:         spawnX,
			Y:         spawnY,
			YStop:     0, // Not used in infinite world
			YOffset:   yOff,
			Radius:    r,
			MaxAngle:  cfg.MaxAngle,
			Health:    health,
			MaxHealth: health, // Store max health for health bar
			FireTimer: cfg.TStart + g.GameRNG.RandomInt(cfg.TBase, cfg.TBase+cfg.TRange),
			Kind:      kind,
		}

		if cfg.HasFireDir {
			enemy.FireDirection = g.GameRNG.Random() * math.Pi
		}

		g.Enemies = append(g.Enemies, enemy)
	}

	return true
}

// EnemyType defines an enemy type's behavior and appearance.
type EnemyType struct {
	R     float64
	Image *js.Object
}

// CalculateTargetAngle calculates the angle from source to target position.
// Returns angle in radians where 0 = down, PI/2 = right, PI = up, -PI/2 = left.
// Uses Atan2 for correct handling of all quadrants.
func CalculateTargetAngle(srcX, srcY, targetX, targetY float64) float64 {
	dx := targetX - srcX
	dy := targetY - srcY
	// Atan2 gives angle from positive X axis, we want angle from positive Y axis (down)
	// Rotate by -PI/2 to convert from X-axis reference to Y-axis reference
	return math.Atan2(dx, dy)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// RenderHealthBar renders the enemy's health bar below the enemy sprite.
// Similar to Ship.RenderEnergyBar, it fades out over time.
func (e *Enemy) RenderHealthBar(g *Game, screenX, screenY float64) {
	barWidth := e.Radius * 2
	barX := int(screenX) - int(e.Radius)
	barY := int(screenY) + int(e.Radius) + 4

	// Calculate health percentage and color
	healthPercent := float64(e.Health) / float64(e.MaxHealth)
	if healthPercent < 0 {
		healthPercent = 0
	}
	colorValue := int(healthPercent * 512)

	g.Ctx.Set("globalAlpha", float64(e.OSD)/float64(ShipMaxOSD))
	g.Ctx.Set("fillStyle", Theme.EnergyBarBackground)
	g.Ctx.Call("fillRect", barX, barY, int(barWidth), 3)

	var r, gr int
	if colorValue > 255 {
		r = 512 - colorValue
		gr = 255
	} else {
		r = 255
		gr = colorValue
	}
	g.Ctx.Set("fillStyle", "rgb("+strconv.Itoa(r)+","+strconv.Itoa(gr)+",0)")
	g.Ctx.Call("fillRect", barX, barY, int(barWidth*healthPercent), 3)

	g.Ctx.Set("lineWidth", Theme.EnergyBarLineWidth)
	g.Ctx.Set("strokeStyle", Theme.EnergyBarBorder)
	g.Ctx.Call("strokeRect", barX, barY, int(barWidth), 3)
	g.Ctx.Set("globalAlpha", 1)

	e.OSD--
}
