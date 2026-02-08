package game

import (
	"math"
	"strconv"

	"github.com/gopherjs/gopherjs/js"
)

// Ship holds player ship state.
type Ship struct {
	X, Y          float64
	VelX, VelY    float64 // Velocity for infinite world movement
	XAcc, YAcc    float64
	Angle         float64 // Now used as actual rotation angle in radians
	E             int     // Energy/health
	Points        int     // Score/points for this ship (determines enemy strength)
	Timeout       int
	Weapon        int
	Reload        int
	OSD           int
	Shield        Shield
	Image         *js.Object
	OriginalImage *js.Object
	Weapons       []*Weapon
	local         bool
	Paused        bool //Ship is paused
	InBase        bool // Ship is inside a base shield
	RepairTimer   int  // Frames until next repair tick while in base

	// Targeting system
	Target      *Enemy // Currently locked target
	LockingOn   *Enemy // Enemy being locked onto (not yet locked)
	LockTimer   int    // Frames remaining until lock completes
	LockMaxTime int    // Total frames required for this lock
}

type Weapon struct {
	X, Y    float64
	AudioID int
}

// AudioPan returns a pan value (-1.0 to 1.0) based on the ship's screen position.
// In infinite world mode, the local player ship is always at the camera center (pan = 0).
// For multiplayer, this would calculate position relative to the local camera.
func (s *Ship) AudioPan() float64 {
	// In single-player, the camera follows the ship, so it's always centered
	if s.local {
		return 0.0
	}
	// For non-local ships (future multiplayer), would need camera reference
	return 0.0
}

// GetPosition implements Entity interface - returns the ship's world coordinates.
func (s *Ship) GetPosition() (x, y float64) {
	return s.X, s.Y
}

// GetAngle implements Entity interface - returns the ship's facing angle in radians.
func (s *Ship) GetAngle() float64 {
	return s.Angle
}

// GetHealth implements Entity interface - returns the ship's current energy/health.
func (s *Ship) GetHealth() int {
	return s.E
}

// SetHealth implements Entity interface - sets the ship's energy/health.
func (s *Ship) SetHealth(health int) {
	s.E = health
}

// GetRadius implements Entity interface - returns the ship's collision radius.
func (s *Ship) GetRadius() float64 {
	return float64(ShipR)
}

// IsAlive implements Entity interface - returns true if the ship has health remaining.
func (s *Ship) IsAlive() bool {
	return s.E > 0
}

// Update implements Entity interface - updates the ship state and renders it.
// Returns false if the ship should be removed (death).
func (s *Ship) Update(g *Game) bool {
	if !s.IsAlive() {
		return false
	}

	// Render the ship
	s.Render(g)
	return true
}

func (s *Ship) Pickup(g *Game, item *Bonus) bool {
	if s.Y < (item.Y+ShipCollisionD) && s.Y > (item.Y-ShipCollisionD) &&
		s.X < (item.X+ShipCollisionE) && s.X > (item.X-ShipCollisionE) {
		switch item.Type {
		case "+":
			if len(s.Weapons) < MaxWeapons {
				s.AddWeapon()
				// todo: make audio level reflective of weapon count
				g.Audio.PlayLocal(5, 1.0)
			} else {
				g.Audio.PlayLocal(6, 1.0)
			}
		case "E":
			if s.E < 100 {
				s.OSD = ShipMaxOSD
				// todo: make audio level reflective of energy level
				g.Audio.PlayLocal(5, 1.0)
			} else {
				g.Audio.PlayLocal(6, 1.0)
			}
			s.E += 5
			if s.E > 100 {
				s.E = 100
			}
		case "S":
			s.Shield.T += s.Shield.MaxT * s.Shield.MaxT * 2 /
				(s.Shield.T + s.Shield.MaxT*2)
			g.Audio.PlayLocal(3, 1.0)
		case "B":
			// for j := len(g.Enemies) - 1; j >= 0; j-- {
			// 	g.Enemies[j].Health--
			// }
			// v := 0.0
			// // TODO: only target the nearest torpedos within a distance from the ship.
			// // As the ship has more weapons, the distance increases.
			// g.Bullets.ForEachKindReverse(TorpedoBullet, func(b *Bullet, i int) {
			// 	v += 0.01
			// 	if v > 0.7 {
			// 		v = 0.7
			// 	}
			// 	g.Explode(b.X, b.Y, 0)
			// 	g.Audio.PlayWithPan(13, b.AudioPan(), v)
			// 	g.Bullets.Release(i)
			// })
			// for j := 0; j < min(g.Torpedos.ActiveCount, 5); j++ {
			// 	g.Explode(g.Torpedos.Pool[j].X, g.Torpedos.Pool[j].Y, 0)
			// }
			// g.Torpedos.Clear()
			// g.Level.Bomb = MaxBomb
			// TODO make audio level reflective of the number of torpedos cleared
			// g.Audio.PlayWithPan(13, s.AudioPan(), 1.0)
		default:
			g.Audio.PlayLocal(7, 1.0)
		}

		return true
	}

	return false
}

func (s *Ship) Move(g *Game, keys map[int]bool) {
	// Rotation input (Left/Right arrows rotate the ship)
	// Left arrow - rotate counter-clockwise
	if keys[37] {
		s.Angle -= ShipRotationSpeed
	}
	// Right arrow - rotate clockwise
	if keys[39] {
		s.Angle += ShipRotationSpeed
	}

	// Track if thrusting this frame
	thrusting := false

	// Thrust input (Up/Down arrows control forward/backward)
	// Up arrow - thrust forward (in direction ship is facing)
	if keys[38] {
		s.VelX += math.Sin(s.Angle) * ShipThrustAcc
		s.VelY -= math.Cos(s.Angle) * ShipThrustAcc
		thrusting = true
	}
	// Down arrow - thrust backward (reverse)
	if keys[40] {
		s.VelX -= math.Sin(s.Angle) * ShipThrustAcc * 0.5
		s.VelY += math.Cos(s.Angle) * ShipThrustAcc * 0.5
		thrusting = true
	}

	// Clamp velocity to max speed
	speed := math.Sqrt(s.VelX*s.VelX + s.VelY*s.VelY)
	if speed > ShipMaxSpeed {
		scale := ShipMaxSpeed / speed
		s.VelX *= scale
		s.VelY *= scale
	}

	// Play thrust sound with volume relative to velocity
	if thrusting && speed > 1.0 {
		// Volume scales from 0.1 at low speed to 0.4 at max speed
		volume := 0.1 + (speed/ShipMaxSpeed)*0.3
		g.Audio.PlayLocal(23, volume)
	}

	// Apply velocity to position (infinite world - no boundaries)
	s.X += s.VelX
	s.Y += s.VelY

	// Apply drag/damping
	s.VelX *= ShipACCFactor
	s.VelY *= ShipACCFactor

	// Update camera to follow ship
	g.Camera.X = s.X
	g.Camera.Y = s.Y
}

// WeaponAngleStep defines the angular separation between weapon upgrades in radians.
// 15 degrees = π/12 radians
const WeaponAngleStep = math.Pi / 12

// WeaponSpeed defines the base projectile speed for weapons.
const WeaponSpeed = 50.0

// MaxWeapons defines the maximum number of weapons (full 360° coverage at 15° intervals = 24 weapons,
// but we stop at 180° coverage = 13 weapons: 0°, ±15°, ±30°, ±45°, ±60°, ±75°, ±90°)
const MaxWeapons = 13

// AddWeapon adds a new weapon to the ship's arsenal.
// Weapons are added in a pattern starting from forward (0°) and alternating left/right:
//   - 1st weapon: 0° (straight ahead)
//   - 2nd weapon: -15° (left)
//   - 3rd weapon: +15° (right)
//   - 4th weapon: -30° (left)
//   - 5th weapon: +30° (right)
//   - ... continues until ship is surrounded
//
// The weapon's X and Y values represent bullet velocity direction.
func (s *Ship) AddWeapon() {
	cur := len(s.Weapons)
	if cur >= MaxWeapons {
		return // Already at max weapons
	}

	// Calculate angle for this weapon
	// Pattern: 0, -15, +15, -30, +30, -45, +45, ...
	var angle float64
	if cur == 0 {
		angle = 0
	} else {
		// For cur=1: step=1, angle = -15°
		// For cur=2: step=1, angle = +15°
		// For cur=3: step=2, angle = -30°
		// For cur=4: step=2, angle = +30°
		step := (cur + 1) / 2
		if cur%2 == 1 {
			angle = -float64(step) * WeaponAngleStep // Left side (negative)
		} else {
			angle = float64(step) * WeaponAngleStep // Right side (positive)
		}
	}

	// Calculate velocity components
	// Forward is negative Y (up on screen), so:
	// X = sin(angle) * speed (positive angle = right)
	// Y = -cos(angle) * speed (negative = up/forward)
	weapon := &Weapon{
		X:       math.Sin(angle) * WeaponSpeed,
		Y:       -math.Cos(angle) * WeaponSpeed,
		AudioID: 0,
	}

	s.Weapons = append(s.Weapons, weapon)
}

// Collision checks if an enemy torpedo has hit the ship and processes the collision.
// It uses an axis-aligned bounding box (AABB) collision detection where:
//   - ShipCollisionD defines the vertical half-extent (Y-axis tolerance)
//   - ShipCollisionE defines the horizontal half-extent (X-axis tolerance)
//
// When a collision is detected:
//   - The ship takes 10 damage via Hurt()
//   - An explosion effect is spawned at the torpedo's position
//   - Returns true to signal the torpedo should be removed
//
// Returns false if no collision occurred.
// Note: When debug UI (F9) is active, ship is "invisible" to torpedos.
func (s *Ship) Collision(g *Game, bullet *Bullet) bool {
	if bullet.Kind != TorpedoBullet {
		return false
	}
	if g.IsShipProtectedByBase(s) {
		return false
	}
	if s.Y < (bullet.Y+ShipCollisionD) && s.Y > (bullet.Y-ShipCollisionD) &&
		s.X < bullet.X+ShipCollisionE && s.X > (bullet.X-ShipCollisionE) {
		s.Hurt(g, 10)
		g.Explode(bullet.X, bullet.Y, 0)
		return true
	}

	return false
}

func (s *Ship) Render(g *Game) {
	s.Timeout--

	// Update InBase status and handle repair
	s.InBase = g.IsShipProtectedByBase(s)
	if s.InBase && s.E < 100 {
		// Repair while in base - 1 health every 55 frames (~1.8 seconds)
		// Full repair from 1% takes about 3 minutes
		s.RepairTimer--
		if s.RepairTimer <= 0 {
			s.E++
			s.RepairTimer = 55 // Reset timer
			s.OSD = ShipMaxOSD // Show energy bar when repairing
		}
	}

	// Convert world position to screen position
	screenX, screenY := g.Camera.WorldToScreen(s.X, s.Y)

	g.Ctx.Call("save")
	g.Ctx.Call("translate", screenX, screenY)
	g.Ctx.Call("rotate", s.Angle) // Use actual rotation angle (radians)
	g.Ctx.Call("drawImage", s.Image, -ShipR/2, -ShipR)
	g.Ctx.Call("restore")

	// Shield effect
	if s.Shield.T > 0 {
		mathRandom := js.Global.Get("Math").Call("random").Float()

		if s.Shield.T > 30 || mathRandom > 0.5 {
			g.Ctx.Call("drawImage", s.Shield.Image,
				int(screenX)-ShipR, int(screenY)-ShipR)
		}
		s.Shield.T--
		if s.Shield.T == 0 {
			g.Audio.PlayLocal(4, 1.0)
		}
	}

	if s.OSD > 0 {
		s.RenderEnergyBar(g)
	}
}

// Fire creates bullets from all of the ship's equipped weapons.
// Each weapon in the Weapons slice fires one bullet per call.
// If a target is locked, all weapons aim at the target.
// Otherwise, bullet velocity is based on the weapon's angle offset rotated by ship's facing angle.
// Bullets spawn at the ship's position and travel in the combined direction.
// Cannot fire while inside a base shield (safe zone).
func (s *Ship) Fire(g *Game) {
	// Cannot fire while inside base shield
	if s.InBase {
		return
	}

	s.Reload--

	weapons := len(s.Weapons)
	if weapons > 0 {
		s.Reload = 4
	} else {
		s.Reload = 6
	}

	// Fire from each equipped weapon
	for _, w := range s.Weapons {
		// if no target, not able to fire
		if s.Target == nil {
			continue
		}

		// Try to acquire a bullet from the pool
		bullet := g.Bullets.AcquireKind(StandardBullet)
		if bullet == nil {
			continue // Pool exhausted, skip this weapon
		}

		var totalAngle float64

		// If we have a locked target, aim all weapons at it with prediction
		if s.Target != nil && s.Target.IsAlive() {
			// Calculate predicted intercept position
			totalAngle = s.PredictTargetAngle(s.Target, WeaponSpeed)
		} else {
			// No target
			continue
		}

		// Calculate bullet velocity in the combined direction
		bulletSpeed := WeaponSpeed
		finalXVel := math.Sin(totalAngle) * bulletSpeed
		finalYVel := -math.Cos(totalAngle) * bulletSpeed

		// Add ship velocity to projectile (projectile speed is relative to ship motion)
		finalXVel += s.VelX
		finalYVel += s.VelY

		// Initialize bullet properties
		bullet.T = BulletMaxT // Set lifetime

		// Spawn bullet at weapon position (offset from ship center based on weapon direction)
		// Scale weapon vector to get spawn offset (w.X/Y are at WeaponSpeed magnitude)
		spawnOffsetScale := 0.4
		bullet.X = s.X + w.X*spawnOffsetScale
		bullet.Y = s.Y + w.Y*spawnOffsetScale

		bullet.XAcc = finalXVel
		bullet.YAcc = finalYVel

		// Play weapon fire sound (local to player, not filtered by shield)
		g.Audio.PlayLocal(w.AudioID, 0.5)
	}
}

// TODO rename to Hit
// Hurt applies damage to the ship and handles damage effects.
// If the ship has an active shield, damage is blocked and a shield hit sound plays.
// Otherwise, damage is applied after the invincibility timeout expires.
// When health reaches 0, the ship explodes and the game ends.
// Taking damage also removes one weapon upgrade.
func (s *Ship) Hurt(g *Game, damage int) {
	if s.Paused {
		return
	}

	if g.IsShipProtectedByBase(s) {
		return
	}

	// Shield absorbs all damage while active
	if s.Shield.T > 0 {
		g.Audio.PlayLocal(2, 1.0)
		return
	}

	// Apply damage only if invincibility has expired (Timeout < 0)
	if s.Timeout < 0 {
		s.E -= damage
		if s.E < 0 {
			s.E = 0
		}
		// Grant brief invincibility after taking damage
		s.Timeout = 10
	}

	// Check for death - create large explosions
	if s.E == 0 {
		g.Explode(s.X, s.Y, 512)
		g.Explode(s.X, s.Y, 1024)
		g.Audio.PlayLocal(14, 1.0) // Death sound
	} else if s.E < 25 {
		g.Audio.PlayLocal(17, 1.0) // Low health warning
	}

	// Play hit sound
	g.Audio.PlayLocal(1, 1.0)

	// Lose one weapon upgrade on damage
	weapons := len(s.Weapons)
	if weapons > 1 {
		s.Weapons = s.Weapons[:weapons-1]
	}

	// Show on-screen damage indicator
	s.OSD = ShipMaxOSD
}

func (s *Ship) RenderEnergyBar(g *Game) {
	// Convert world position to screen position
	screenX, screenY := g.Camera.WorldToScreen(s.X, s.Y)

	barX := int(screenX) - 32
	barY := int(screenY) + 63
	colorValue := s.E * 512 / 100

	g.Ctx.Set("globalAlpha", float64(s.OSD)/float64(ShipMaxOSD))
	g.Ctx.Set("fillStyle", Theme.EnergyBarBackground)
	g.Ctx.Call("fillRect", barX, barY, 64, 4)

	var r, gr int
	if colorValue > 255 {
		r = 512 - colorValue
		gr = 255
	} else {
		r = 255
		gr = colorValue
	}
	g.Ctx.Set("fillStyle", "rgb("+strconv.Itoa(r)+","+strconv.Itoa(gr)+",0)")
	g.Ctx.Call("fillRect", barX, barY, s.E*64/100, 4)

	g.Ctx.Set("lineWidth", Theme.EnergyBarLineWidth)
	g.Ctx.Set("strokeStyle", Theme.EnergyBarBorder)
	g.Ctx.Call("strokeRect", barX, barY, 64, 4)
	g.Ctx.Set("globalAlpha", 1)

	if s.E >= 25 {
		s.OSD--
	}
}

// UpdateTargeting handles the targeting lock-on system.
// Call this each frame to update lock progress and validate targets.
func (s *Ship) UpdateTargeting(g *Game) {
	// If we have a locked target, validate it's still alive and on screen
	if s.Target != nil {
		if !s.Target.IsAlive() || !g.Camera.IsOnScreen(s.Target.X, s.Target.Y+s.Target.YOffset, s.Target.Radius) {
			s.Target = nil
		}
	}

	// If we're locking onto something, update the lock timer
	if s.LockingOn != nil {
		// Validate the locking target is still valid
		if !s.LockingOn.IsAlive() || !g.Camera.IsOnScreen(s.LockingOn.X, s.LockingOn.Y+s.LockingOn.YOffset, s.LockingOn.Radius) {
			s.LockingOn = nil
			s.LockTimer = 0
			s.LockMaxTime = 0
			return
		}

		s.LockTimer--
		if s.LockTimer <= 0 {
			// Lock complete!
			s.Target = s.LockingOn
			s.LockingOn = nil
			s.LockMaxTime = 0
			// Play lock-on sound
			g.Audio.PlayLocal(6, 1.0)
		}
	}
}

// InitiateTargetLock starts locking onto the nearest on-screen enemy.
// If already locked, this will switch to a new target.
// Lock time is random, up to 1 second (30 frames at 30 FPS).
func (s *Ship) InitiateTargetLock(g *Game) {
	// Find nearest enemy that is on screen
	var nearestEnemy *Enemy
	nearestDistSq := math.MaxFloat64

	for _, enemy := range g.Enemies {
		if !enemy.IsAlive() {
			continue
		}

		// Check if enemy is on screen
		if !g.Camera.IsOnScreen(enemy.X, enemy.Y+enemy.YOffset, enemy.Radius) {
			continue
		}

		// Calculate distance to ship
		dx := enemy.X - s.X
		dy := (enemy.Y + enemy.YOffset) - s.Y
		distSq := dx*dx + dy*dy

		if distSq < nearestDistSq {
			nearestDistSq = distSq
			nearestEnemy = enemy
		}
	}

	if nearestEnemy != nil {
		// Start locking onto this enemy
		s.LockingOn = nearestEnemy
		// Random lock time: 5-30 frames (0.17s to 1s at 30 FPS)
		s.LockTimer = 5 + int(js.Global.Get("Math").Call("random").Float()*25)
		s.LockMaxTime = s.LockTimer
		// Clear any existing lock
		s.Target = nil
		// Play targeting locking sound
		g.Audio.PlayLocal(22, 0.6)
	}
}

// ClearTarget removes the current target lock.
func (s *Ship) ClearTarget() {
	s.Target = nil
	s.LockingOn = nil
	s.LockTimer = 0
	s.LockMaxTime = 0
}

// PredictTargetAngle calculates the angle to shoot at to hit a moving target.
// Uses quadratic intercept calculation to predict where the target will be
// when the projectile reaches it, accounting for:
// - Distance between ship and target
// - Target's current velocity
// - Projectile speed (plus ship velocity contribution)
//
// Returns the angle in radians to aim at the predicted intercept point.
// Falls back to direct aim if no intercept solution exists.
func (s *Ship) PredictTargetAngle(target *Enemy, projectileSpeed float64) float64 {
	// Target position (accounting for YOffset)
	targetX := target.X
	targetY := target.Y + target.YOffset

	// Relative position
	dx := targetX - s.X
	dy := targetY - s.Y

	// Target velocity
	tvx := target.VelX
	tvy := target.VelY

	// Account for ship velocity in the effective projectile speed
	// The projectile will inherit ship velocity, so we consider relative velocity
	// For simplicity, we use the raw projectile speed here since ship velocity
	// is added to the final bullet velocity anyway

	// Quadratic coefficients for time-to-intercept:
	// a*t² + b*t + c = 0
	// where t is the time for the projectile to reach the target
	a := tvx*tvx + tvy*tvy - projectileSpeed*projectileSpeed
	b := 2 * (dx*tvx + dy*tvy)
	c := dx*dx + dy*dy

	var t float64

	// Solve quadratic equation
	if math.Abs(a) < 0.0001 {
		// Linear case (target and projectile have similar speeds)
		if math.Abs(b) > 0.0001 {
			t = -c / b
		} else {
			// No solution, aim directly
			t = 0
		}
	} else {
		discriminant := b*b - 4*a*c
		if discriminant < 0 {
			// No real solution - target is too fast to intercept
			// Fall back to direct aim
			return math.Atan2(dx, -dy)
		}

		sqrtDisc := math.Sqrt(discriminant)
		t1 := (-b - sqrtDisc) / (2 * a)
		t2 := (-b + sqrtDisc) / (2 * a)

		// Choose the smallest positive time
		if t1 > 0 && t2 > 0 {
			t = math.Min(t1, t2)
		} else if t1 > 0 {
			t = t1
		} else if t2 > 0 {
			t = t2
		} else {
			// Both negative - target is moving away too fast
			return math.Atan2(dx, -dy)
		}
	}

	// Predict target position at time t
	predictedX := targetX + tvx*t
	predictedY := targetY + tvy*t

	// Calculate angle to predicted position
	predDx := predictedX - s.X
	predDy := predictedY - s.Y

	return math.Atan2(predDx, -predDy)
}
