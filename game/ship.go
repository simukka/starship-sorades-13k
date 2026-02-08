package game

import (
	"math"
	"strconv"

	"github.com/gopherjs/gopherjs/js"
)

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
	Weapons       []*Weapon
	local         bool
}

type Weapon struct {
	X, Y    float64
	AudioID int
}

// AudioPan returns a pan value (-1.0 to 1.0) based on the ship's X position.
// Left edge = -1.0, center = 0.0, right edge = 1.0
func (s *Ship) AudioPan() float64 {
	return (s.X/WIDTH)*2 - 1
}

func (s *Ship) Pickup(g *Game, item *Bonus) bool {
	if s.Y < (item.Y+ShipCollisionD) && s.Y > (item.Y-ShipCollisionD) &&
		s.X < (item.X+ShipCollisionE) && s.X > (item.X-ShipCollisionE) {
		switch item.Type {
		case "+":
			if len(s.Weapons) < MaxWeapons {
				s.AddWeapon()
				// todo: make audio level reflective of weapon count
				g.Audio.PlayWithPan(5, s.AudioPan(), 1.0)
			} else {
				g.Audio.PlayWithPan(6, s.AudioPan(), 1.0)
			}
		case "E":
			if s.E < 100 {
				s.OSD = ShipMaxOSD
				// todo: make audio level reflective of energy level
				g.Audio.PlayWithPan(5, s.AudioPan(), 1.0)
			} else {
				g.Audio.PlayWithPan(6, s.AudioPan(), 1.0)
			}
			s.E += 5
			if s.E > 100 {
				s.E = 100
			}
		case "S":
			s.Shield.T += s.Shield.MaxT * s.Shield.MaxT * 2 /
				(s.Shield.T + s.Shield.MaxT*2)
			g.Audio.PlayWithPan(3, s.AudioPan(), 1.0)
		case "B":
			for j := len(g.Enemies) - 1; j >= 0; j-- {
				g.Enemies[j].Health--
			}
			v := 0.0
			g.Bullets.ForEachKindReverse(TorpedoBullet, func(b *Bullet, i int) {
				v += 0.05
				if v > 0.7 {
					v = 0.7
				}
				g.Explode(b.X, b.Y, 0)
				g.Audio.PlayWithPan(13, b.AudioPan(), v)
				g.Bullets.Release(i)
			})
			// for j := 0; j < min(g.Torpedos.ActiveCount, 5); j++ {
			// 	g.Explode(g.Torpedos.Pool[j].X, g.Torpedos.Pool[j].Y, 0)
			// }
			// g.Torpedos.Clear()
			g.Level.Bomb = MaxBomb
			// TODO make audio level reflective of the number of torpedos cleared
			// g.Audio.PlayWithPan(13, s.AudioPan(), 1.0)
		default:
			g.Audio.PlayWithPan(7, s.AudioPan(), 1.0)
		}

		return true
	}

	return false
}

func (s *Ship) Move(g *Game, keys map[int]bool) {
	// Movement input with visual tilt
	s.Angle *= ShipAngleFactor

	// Left arrow
	if keys[37] {
		if s.X >= WIDTH && s.XAcc > 0 {
			s.XAcc = 0
		}
		s.XAcc -= ShipACC
		s.Angle = (s.Angle+1)*ShipAngleFactor - 1
	}
	// Up arrow
	if keys[38] {
		if s.Y >= HEIGHT && s.YAcc > 0 {
			s.YAcc = 0
		}
		s.YAcc -= ShipACC
	}
	// Right arrow
	if keys[39] {
		if s.X < 0 && s.XAcc < 0 {
			s.XAcc = 0
		}
		s.XAcc += ShipACC
		s.Angle = (s.Angle-1)*ShipAngleFactor + 1
	}
	// Down arrow
	if keys[40] {
		if s.Y < 0 && s.YAcc < 0 {
			s.YAcc = 0
		}
		s.YAcc += ShipACC
	}

	// Screen boundary collision
	if s.X < 0 && s.XAcc < 0 {
		s.X = 0
	} else if s.X >= WIDTH && s.XAcc > 0 {
		s.X = WIDTH - 1
	}
	if s.Y < 0 && s.YAcc < 0 {
		s.Y = 0
	} else if s.Y >= HEIGHT && s.YAcc > 0 {
		s.Y = HEIGHT - 1
	}

	// Apply velocity and damping
	s.X += s.XAcc
	s.Y += s.YAcc
	s.XAcc *= ShipACCFactor
	s.YAcc *= ShipACCFactor
}

// WeaponAngleStep defines the angular separation between weapon upgrades in radians.
// 15 degrees = π/12 radians
const WeaponAngleStep = math.Pi / 12

// WeaponSpeed defines the base projectile speed for weapons.
const WeaponSpeed = 16.0

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

	g.Ctx.Call("save")
	g.Ctx.Call("translate", s.X, s.Y)
	g.Ctx.Call("rotate", s.Angle*ShipMaxAngle/180*math.Pi)
	g.Ctx.Call("drawImage", s.Image, -ShipR/2, -ShipR)
	g.Ctx.Call("restore")

	// Shield effect
	if s.Shield.T > 0 {
		mathRandom := js.Global.Get("Math").Call("random").Float()

		if s.Shield.T > 30 || mathRandom > 0.5 {
			g.Ctx.Call("drawImage", s.Shield.Image,
				int(s.X)-ShipR, int(s.Y)-ShipR)
		}
		s.Shield.T--
		if s.Shield.T == 0 {
			g.Audio.PlayWithPan(4, s.AudioPan(), 1.0)
		}
	}

	if s.OSD > 0 {
		s.RenderEnergyBar(g)
	}
}

// Fire creates bullets from all of the ship's equipped weapons.
// Each weapon in the Weapons slice fires one bullet per call.
// Bullet velocity is based on the weapon's offset (X, Y) plus half the ship's current momentum.
// Bullets spawn at the ship's position with slight random offset for visual variety.
func (s *Ship) Fire(g *Game) {
	Debug("FIRE")
	s.Reload--

	weapons := len(s.Weapons)
	if weapons > 0 {
		s.Reload = 4
	} else {
		s.Reload = 6
	}

	// Fire from each equipped weapon
	for _, w := range s.Weapons {
		// Try to acquire a bullet from the pool
		bullet := g.Bullets.AcquireKind(StandardBullet)
		if bullet == nil {
			continue // Pool exhausted, skip this weapon
		}

		// Calculate bullet velocity: weapon direction + ship momentum influence
		finalXVel := w.X + s.XAcc/2
		finalYVel := w.Y + s.YAcc/2

		// Initialize bullet properties
		bullet.T = BulletMaxT // Set lifetime
		// Add random offset based on velocity for spread effect
		bullet.X = s.X + math.Floor(js.Global.Get("Math").Call("random").Float()*finalXVel)
		bullet.Y = s.Y + math.Floor(js.Global.Get("Math").Call("random").Float()*finalYVel)
		bullet.XAcc = finalXVel
		bullet.YAcc = finalYVel

		// Play weapon fire sound with stereo panning based on ship position
		g.Audio.PlayWithPan(w.AudioID, s.AudioPan(), 1.0)
	}
}

// TODO rename to Hit
// Hurt applies damage to the ship and handles damage effects.
// If the ship has an active shield, damage is blocked and a shield hit sound plays.
// Otherwise, damage is applied after the invincibility timeout expires.
// When health reaches 0, the ship explodes and the game ends.
// Taking damage also removes one weapon upgrade.
func (s *Ship) Hurt(g *Game, damage int) {
	// Ship is invisible to torpedos when debug UI is active
	if g.DebugUI.Visible {
		return
	}
	// Shield absorbs all damage while active
	if s.Shield.T > 0 {
		g.Audio.PlayWithPan(2, s.AudioPan(), 1.0)
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
	if s.E == 0 && !g.Level.Paused {
		g.Explode(s.X, s.Y, 512)
		g.Explode(s.X, s.Y, 1024)
		g.Audio.PlayWithPan(14, s.AudioPan(), 1.0) // Death sound
	} else if s.E < 25 {
		g.Audio.PlayWithPan(17, s.AudioPan(), 1.0) // Low health warning
	}

	// Play hit sound
	g.Audio.PlayWithPan(1, s.AudioPan(), 1.0)

	// Lose one weapon upgrade on damage
	weapons := len(s.Weapons)
	if weapons > 0 {
		s.Weapons = s.Weapons[:weapons-1]
	}

	// Show on-screen damage indicator
	s.OSD = ShipMaxOSD
}

func (s *Ship) RenderEnergyBar(g *Game) {
	barX := int(s.X) - 32
	barY := int(s.Y) + 63
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
