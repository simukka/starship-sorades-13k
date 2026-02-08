package game

import (
	"math"

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
	YStop         float64
	YOffset       float64
	Radius        float64
	Angle         float64
	MaxAngle      float64
	Health        int
	FireTimer     int
	FireDirection float64
	TActive       int
	TypeIndex     int

	Target *Ship
	Kind   EnemyKind
}

// EnemySpawnConfig holds configuration for spawning enemies.
type EnemySpawnConfig struct {
	CountBase      int
	CountPerLevel  int
	MaxAngle       float64
	HealthBase     int
	HealthPerLevel int
	TBase          int
	TRange         int
	TStart         int
	YStopBase      float64
	YStopRange     float64
	YOffsetMult    float64 // yoffset multiplyer
	UseYStep       bool    // For turret-style Y positioning
	HasFireDir     bool    // For turrets with rotating fire
}

// Standard enemy spawn configurations
var enemyConfigs = map[EnemyKind]EnemySpawnConfig{
	// Type 0: Small fighters
	SmallFighter: {
		CountBase: 10, CountPerLevel: 1,
		MaxAngle: math.Pi / 32, HealthBase: 12, HealthPerLevel: 1,
		TBase: 0, TRange: 120, YStopBase: HEIGHT / 8, YStopRange: HEIGHT / 4,
	},
	// Type 1: Medium fighters
	MediumFighter: {
		CountBase: 1, CountPerLevel: 1,
		MaxAngle: math.Pi / 16, HealthBase: 20, HealthPerLevel: 2,
		TBase: 0, TRange: 120, YStopBase: HEIGHT / 8, YStopRange: HEIGHT / 4,
	},
	// Type 2: Turrets
	TurretFighter: {
		CountBase: 0, CountPerLevel: 1,
		MaxAngle: math.Pi * 32, HealthBase: 28, HealthPerLevel: 3,
		TBase: 0, TRange: 30, YStopBase: HEIGHT / 2, YStopRange: 0,
		UseYStep: true, HasFireDir: true,
	},
	Boss: {
		CountBase: 0, CountPerLevel: 0,
		YOffsetMult: 0.6,
		MaxAngle:    math.Pi / 8, HealthBase: 36, HealthPerLevel: 4,
		TBase: 0, TRange: 20, TStart: 60,
	},
}

// AudioPan returns a pan value (-1.0 to 1.0) based on the ship's X position.
// Left edge = -1.0, center = 0.0, right edge = 1.0
func (s *Enemy) AudioPan() float64 {
	return (s.X/WIDTH)*2 - 1
}

// GetPosition implements Collidable interface.
func (e *Enemy) GetPosition() (x, y float64) {
	return e.X, e.Y + e.YOffset
}

// GetRadius implements Collidable interface.
func (e *Enemy) GetRadius() float64 {
	return e.Radius
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

	// Calculate angle to player using Atan2 for accurate targeting across all quadrants
	// TODO: Pick a target
	r := 0
	e.Target = g.Ships[r]

	// Normalize angle to [-π, π] range for consistent shortest-path rotation
	angle := math.Mod(e.TargetAngle()+math.Pi, math.Pi*2) - math.Pi

	// Clamp angle to enemy's maximum rotation limit
	if angle > e.MaxAngle {
		angle = e.MaxAngle
	}
	if angle < -e.MaxAngle {
		angle = -e.MaxAngle
	}

	// Apply exponential smoothing for gradual rotation toward target
	e.Angle = (e.Angle*EnemyAngleSmoothingFactor - angle) / (EnemyAngleSmoothingFactor + 1)

	if e.Health <= 0 {
		// we dead
		g.Level.P += 100
		g.SpawnBonus(e.X, e.Y, 0, 0, "")
		g.Explode(e.X, e.Y, e.Radius*2)
		g.Explode(e.X, e.Y, e.Radius*3)

		g.Audio.PlayWithPan(10,
			e.AudioPan(),
			e.DistanceVolume(g.Ship.X, g.Ship.Y, float64(HEIGHT)))
		return false
	}

	// Render enemy
	g.Ctx.Call("save")
	g.Ctx.Call("translate", e.X, enemyY)
	g.Ctx.Call("rotate", e.Angle)
	g.Ctx.Call("drawImage", e.Image,
		-e.Radius, e.Y-enemyY-e.Radius, e.Radius*2, e.Radius*2)
	g.Ctx.Call("restore")

	// Draw targeting indicator (red circle) when enemy has a target
	if e.Target != nil {
		g.Ctx.Call("save")
		g.Ctx.Set("strokeStyle", "rgba(255, 60, 60, 0.7)")
		g.Ctx.Set("lineWidth", 2)
		g.Ctx.Call("beginPath")
		g.Ctx.Call("arc", e.X, enemyY, e.Radius*1.3, 0, math.Pi*2)
		g.Ctx.Call("stroke")
		g.Ctx.Call("restore")
	}

	// move enemy
	if e.Y < e.YStop {
		e.Y++
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

	// if e.Angle != 0 {
	// 	torpedo.XAcc = math.Sin(e.Angle) * speed
	// 	torpedo.YAcc = math.Cos(e.Angle) * speed
	// } else {
	// 	torpedo.XAcc = 0
	// 	torpedo.YAcc = speed
	// }
	torpedo.E = 0

	return true
}

// TODO use a pool
func (g *Game) SpawnEnemy(kind EnemyKind, yOffset float64) bool {
	r := float64(ShipR)
	cfg, exists := enemyConfigs[kind]

	if !exists {
		panic("how?")
	}

	if g.EnemyTypes[kind].R > 0 {
		r = g.EnemyTypes[kind].R
	}

	count := cfg.CountBase + g.Level.LevelNum*cfg.CountPerLevel

	var yStep float64
	if cfg.UseYStep {
		yStep = float64(HEIGHT) * -1.5 / float64(g.Level.LevelNum)
	}

	for i := count; i > 0; i-- {
		var y, yStop float64

		if cfg.UseYStep {
			yStop = cfg.YStopBase
			if yOffset != 0 {
				y = float64(i-1)*yStep + yOffset
			} else {
				y = float64(i-1)*yStep - r
			}
		} else {
			yStop = cfg.YStopBase + g.GameRNG.Random()*cfg.YStopRange
			if yOffset != 0 {
				y = yStop + yOffset
			} else {
				y = -r
			}
		}

		// xPos
		x := r + (float64(WIDTH)-r*2)*g.GameRNG.Random()
		yOff := 0.0
		if kind == Boss {
			x = float64(WIDTH / 2)
			yStop = r + 8
			yOff = r * cfg.YOffsetMult
		}

		enemy := &Enemy{
			Image:     g.EnemyTypes[kind].Image,
			X:         x,
			Y:         y,
			YStop:     yStop,
			YOffset:   yOff,
			Radius:    r,
			MaxAngle:  cfg.MaxAngle,
			Health:    cfg.HealthBase + g.Level.LevelNum*cfg.HealthPerLevel,
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
