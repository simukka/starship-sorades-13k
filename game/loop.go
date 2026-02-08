package game

import (
	"strconv"

	"github.com/gopherjs/gopherjs/js"
)

// GameLoopRAF is the main game loop using requestAnimationFrame.
func (g *Game) GameLoopRAF(currentTime float64) {
	// Schedule next frame
	g.AnimationFrameID = js.Global.Call("requestAnimationFrame", g.GameLoopRAF).Int()

	// Update FPS counter
	g.StatsOverlay.UpdateFPS(currentTime)

	// Fixed timestep
	if currentTime-g.LastFrameTime < FrameDuration {
		return
	}

	g.LastFrameTime = currentTime

	g.GameLoop()
}

// GameLoop is the core game logic.
func (g *Game) GameLoop() {
	// Player Input Processing
	g.ProcessInput()

	// Update targeting system
	g.Ship.UpdateTargeting(g)

	// Update shield audio filter based on ship position
	g.UpdateShieldAudioFilter()

	// Background Rendering
	g.RenderBackground()

	// Score Display
	g.RenderScore()

	// Text Display
	g.RenderText()

	// Populate spatial grids for collision detection
	g.PopulateSpatialGrids()

	// Bullet Update and Rendering
	g.UpdateBullets()

	// Screen Flash Effect
	if g.Level.Bomb > 0 {
		alpha := float64(g.Level.Bomb) / float64(MaxBomb) / 2
		g.Ctx.Set("fillStyle", Theme.BombFlashColor+strconv.FormatFloat(alpha, 'f', 2, 64)+")")
		g.Ctx.Call("fillRect", 0, 0, WIDTH, HEIGHT)
		g.Level.Bomb--
	}

	// Enable additive blending
	g.Ctx.Set("globalCompositeOperation", "lighter")

	// Base Rendering (render before other entities)
	g.RenderBases()

	// Bonus Item Update
	g.UpdateBonuses()

	// Explosion Update
	g.UpdateExplosions()

	// Player Ship Rendering
	g.RenderShip()

	// Wave Spawning
	g.CheckWaveSpawn()

	// Enemy Update
	g.UpdateEnemies()

	// Disable additive blending
	g.Ctx.Set("globalCompositeOperation", "source-over")

	// Off-screen base indicators (render after blending disabled for visibility)
	g.RenderBaseIndicators()

	// Ship HUD overlay (velocity, angle, position)
	g.ShipHUD.Render(g.Ctx, g.Ship)

	// Stats overlay
	g.StatsOverlay.Render(g.Ctx, g)
}

// UpdateBullets updates and renders bullets.
func (g *Game) UpdateBullets() {
	g.Bullets.ForEachReverse(func(bullet *Bullet, bulletIdx int) {
		// Within the render the projectile checks for collisions
		// with other ships
		if !bullet.Render(g) {
			g.Bullets.Release(bulletIdx)
			return
		}

		// Check bullet-vs-torpedo collision using spatial grid
		if bullet.Kind == StandardBullet {
			g.CheckBulletTorpedoCollision(bullet)
		}
	})

	// Advance torpedo animation
	g.TorpedoFrame = (g.TorpedoFrame + 1) % len(g.TorpedoImages)
}

// PopulateSpatialGrids clears and repopulates the spatial hash grids.
// Called once per frame before collision detection.
// Uses camera-relative (screen) coordinates for infinite world support.
func (g *Game) PopulateSpatialGrids() {
	// Clear grids from previous frame
	g.EnemyGrid.Clear()
	g.BulletGrid.Clear()

	// Insert all enemies using screen coordinates
	for _, enemy := range g.Enemies {
		screenX, screenY := g.Camera.WorldToScreen(enemy.X, enemy.Y+enemy.YOffset)
		g.EnemyGrid.InsertAt(enemy, screenX, screenY)
	}

	// Insert all active bullets/torpedos using screen coordinates
	for i := 0; i < g.Bullets.ActiveCount; i++ {
		bullet := g.Bullets.Pool[i]
		screenX, screenY := g.Camera.WorldToScreen(bullet.X, bullet.Y)
		g.BulletGrid.InsertAt(bullet, screenX, screenY)
	}
}

// CheckBulletTorpedoCollision checks if a standard bullet hits any nearby torpedos.
func (g *Game) CheckBulletTorpedoCollision(bullet *Bullet) {
	// Convert bullet to screen coordinates for spatial grid lookup
	screenX, screenY := g.Camera.WorldToScreen(bullet.X, bullet.Y)

	// Get nearby objects from the spatial grid (using screen coordinates)
	nearby := g.BulletGrid.GetNearby(screenX, screenY)

	for _, obj := range nearby {
		torpedo, ok := obj.(*Bullet)
		if !ok || torpedo.Kind != TorpedoBullet {
			continue
		}

		// AABB collision check (using world coordinates for accuracy)
		if bullet.Y < torpedo.Y+BulletTorpedoCollisionDist &&
			bullet.Y > torpedo.Y-BulletTorpedoCollisionDist &&
			bullet.X < torpedo.X+BulletTorpedoCollisionDist &&
			bullet.X > torpedo.X-BulletTorpedoCollisionDist {

			torpedo.E--

			if torpedo.E < 0 {
				g.Explode(torpedo.X, torpedo.Y, 0)
				vol := torpedo.DistanceVolume(g.Ship.X, g.Ship.Y, float64(HEIGHT))
				g.Audio.PlayWithPan(11, torpedo.AudioPan(), vol)
				torpedo.T = -1 // Mark for removal
			}

			bullet.T = 0 // Consume the bullet
			return
		}
	}
}

// UpdateBonuses updates and renders bonus items.
func (g *Game) UpdateBonuses() {
	g.Bonuses.ForEachReverse(func(item *Bonus, idx int) {
		claimed := false

		for _, s := range g.Ships {
			if s.Pickup(g, item) && !claimed {
				claimed = true
				break
			}
		}

		if claimed {
			g.Bonuses.Release(idx)
			return
		}

		// Update position
		item.X += item.XAcc
		item.Y += item.YAcc

		// Convert to screen coordinates and render if visible
		screenX, screenY := g.Camera.WorldToScreen(item.X, item.Y)
		if g.Camera.IsOnScreen(item.X, item.Y, BonusR*2) {
			g.Ctx.Call("drawImage", g.BonusImages[item.Type],
				int(screenX)-BonusR, int(screenY)-BonusR)
		}

		// Remove items too far from camera (infinite world cleanup)
		dx := item.X - g.Camera.X
		dy := item.Y - g.Camera.Y
		if dx*dx+dy*dy > EnemyDespawnDistance*EnemyDespawnDistance {
			g.Bonuses.Release(idx)
		}
	})
}

// UpdateExplosions updates and renders explosions.
func (g *Game) UpdateExplosions() {
	g.Explosions.ForEachReverse(func(exp *Explosion, idx int) {
		// Convert to screen coordinates
		screenX, screenY := g.Camera.WorldToScreen(exp.X, exp.Y)

		// Only render if on screen
		if g.Camera.IsOnScreen(exp.X, exp.Y, exp.Size) {
			g.Ctx.Call("save")
			g.Ctx.Set("globalAlpha", exp.Alpha)
			g.Ctx.Call("translate", screenX, screenY)
			g.Ctx.Call("rotate", exp.Angle)
			g.Ctx.Call("drawImage", g.ExplosionImage,
				-exp.Size/2, -exp.Size/2, exp.Size, exp.Size)
			g.Ctx.Call("restore")
		}

		// Animate explosion
		exp.Size += 16
		exp.Angle += exp.D
		exp.Alpha -= 0.1

		if exp.Alpha < 0.1 {
			g.Explosions.Release(idx)
		}
	})
}

// RenderShip renders the player ship using the Entity interface.
func (g *Game) RenderShip() {
	for _, s := range g.Ships {
		s.Update(g)
	}
}

// CheckWaveSpawn spawns enemies continuously near each ship.
// Enemy count and strength scale with each ship's points.
func (g *Game) CheckWaveSpawn() {
	// Calculate how many enemies should exist based on all ships' points
	targetEnemies := 0
	for _, ship := range g.Ships {
		// Base enemies + scaling with points
		targetEnemies += 5 + ship.Points/500
	}

	// Spawn new enemies when count drops below target
	if len(g.Enemies) < targetEnemies {
		// Spawn enemies near each ship
		for _, ship := range g.Ships {
			// Small fighters - always spawn
			g.SpawnEnemyNearShip(SmallFighter, ship)

			// Medium fighters - spawn after 1000 points
			if ship.Points > 1000 {
				g.SpawnEnemyNearShip(MediumFighter, ship)
			}

			// Turrets - spawn after 3000 points
			if ship.Points > 3000 {
				g.SpawnEnemyNearShip(TurretFighter, ship)
			}

			// Boss - rare spawn after 5000 points
			if ship.Points > 5000 && g.GameRNG.Random() > 0.9 {
				g.SpawnEnemyNearShip(Boss, ship)
			}
		}
	}
}

// UpdateEnemies updates and renders enemies using the Entity interface.
func (g *Game) UpdateEnemies() {
	for i, e := range g.Enemies {
		if !e.Update(g) {
			// enemy no longer exists
			g.RemoveEnemy(i)
		}
	}
}

// RenderBases renders all bases with their shields.
func (g *Game) RenderBases() {
	for _, base := range g.Bases {
		base.Render(g)
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}
