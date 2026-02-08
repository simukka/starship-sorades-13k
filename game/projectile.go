package game

import "github.com/gopherjs/gopherjs/js"

type BulletKind int

const (
	StandardBullet BulletKind = iota
	TorpedoBullet
)

// Bullet represents a player projectile.
type Bullet struct {
	T         int     // Lifetime remaining
	X, Y      float64 // Position
	XAcc      float64 // X velocity
	YAcc      float64 // Y velocity
	E         int     //Health
	PoolIndex int     // Index in pool for swap-and-pop
	Kind      BulletKind
}

// GetPosition implements Collidable interface.
func (b *Bullet) GetPosition() (x, y float64) {
	return b.X, b.Y
}

// GetRadius implements Collidable interface.
func (b *Bullet) GetRadius() float64 {
	if b.Kind == TorpedoBullet {
		return TorpedoR
	}
	return BulletR
}

// AudioPan returns a pan value (-1.0 to 1.0) based on the X position.
// Left edge = -1.0, center = 0.0, right edge = 1.0
func (b *Bullet) AudioPan() float64 {
	return (b.X/WIDTH)*2 - 1
}

// DistanceVolume calculates volume based on distance to a target position.
// Closer = louder (up to 1.0), farther = quieter (minimum 0.2).
func (b *Bullet) DistanceVolume(targetX, targetY, maxDist float64) float64 {
	dx := b.X - targetX
	dy := b.Y - targetY
	dist := dx*dx + dy*dy // Skip sqrt for performance
	maxDistSq := maxDist * maxDist

	if dist >= maxDistSq {
		return 0.2 // Minimum volume for distant objects
	}

	// Linear falloff: 1.0 at dist=0, 0.2 at dist=maxDist
	return 1.0 - (dist/maxDistSq)*0.8
}

// Render draws the bullet and updates its position each frame.
//
// The method handles two bullet kinds with different behaviors:
//
// StandardBullet (player projectile):
//   - Uses the standard bullet image and BulletR radius
//   - Has a limited lifetime (T frames) and is removed when expired
//   - Checks collision with enemies; on hit, the bullet is removed
//   - Removed when off-screen (left, right, or bottom edges)
//
// TorpedoBullet (enemy projectile):
//   - Uses animated torpedo images and TorpedoR radius
//   - Has no lifetime limit; persists until off-screen or collision
//   - Checks collision with player ships; on hit, the bullet is removed
//   - Removed when off-screen (any edge, with extended top boundary for spawning)
//
// Returns true if the bullet should remain active, false if it should be released.
func (b *Bullet) Render(g *Game) bool {
	var image *js.Object
	var projectileR float64

	if b.Kind == StandardBullet {
		image = g.BulletImage
		projectileR = BulletR
	}
	if b.Kind == TorpedoBullet {
		image = g.TorpedoImages[g.TorpedoFrame]
		projectileR = TorpedoR
	}

	// Update position
	b.X += b.XAcc
	b.Y += b.YAcc

	// Convert world position to screen position
	screenX, screenY := g.Camera.WorldToScreen(b.X, b.Y)

	// Only render if on screen
	if g.Camera.IsOnScreen(b.X, b.Y, projectileR) {
		g.Ctx.Call("drawImage", image, screenX-projectileR, screenY-projectileR)
	}

	// Check if bullet is too far from camera (despawn in infinite world)
	dx := b.X - g.Camera.X
	dy := b.Y - g.Camera.Y
	distSq := dx*dx + dy*dy
	maxDist := EnemyDespawnDistance * 1.5 // Bullets can travel a bit further
	if distSq > maxDist*maxDist {
		return false
	}

	// TODO: refactor the Ship and Enemy objects into a single
	// interface
	if b.Kind == TorpedoBullet {
		for _, base := range g.Bases {
			if base.BlocksBulletEntrance(b) {
				g.Explode(b.X, b.Y, 0)
				vol := b.DistanceVolume(g.Ship.X, g.Ship.Y, float64(HEIGHT))
				g.Audio.PlayWithPan(11, b.AudioPan(), vol)
				return false
			}
		}

		for _, s := range g.Ships {
			if s.Collision(g, b) {
				return false
			}
		}
	}

	if b.Kind == StandardBullet {
		// remove expired bullets
		b.T--
		if b.T < 0 {
			return false
		}

		for _, e := range g.Enemies {
			if e.Collision(g, b) {
				return false
			}
		}
	}

	return true
}

// Has this projection collided with another?
func (b *Bullet) Collision(g *Game, torpedo *Bullet) bool {

	if b.Y < (torpedo.Y+BulletTorpedoCollisionDist) &&
		b.Y > (torpedo.Y-BulletTorpedoCollisionDist) &&
		b.X < (torpedo.X+BulletTorpedoCollisionDist) &&
		b.X > (torpedo.X-BulletTorpedoCollisionDist) {
		torpedo.E--
		if g.GameRNG.Random() > 0.75 {
			g.SpawnBonus(torpedo.X, torpedo.Y, 0, 0, "")
		}
		g.Explode(torpedo.X, torpedo.Y, 0)

	}
	return false
}
