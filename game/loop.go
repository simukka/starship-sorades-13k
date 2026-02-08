package game

import (
	"math"
	"strconv"

	"github.com/gopherjs/gopherjs/js"
	"github.com/simukka/starship-sorades-13k/common"
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

	Debug("Frame time:", currentTime, "Delta:", currentTime-g.LastFrameTime)
	g.LastFrameTime = currentTime

	if g.Level.Paused {
		return
	}

	g.GameLoop()
}

// GameLoop is the core game logic.
func (g *Game) GameLoop() {
	Debug("Level:", g.Level.LevelNum, "Score:", g.Level.P, "Enemies:", len(g.Enemies))
	Debug("Bullets:", g.Bullets.ActiveCount, "Explosions:", g.Explosions.ActiveCount)

	// Player Input Processing
	g.ProcessInput()

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

	// Debug UI (rendered last, on top of everything)
	g.DebugUI.Render(g.Ctx)

	// Stats overlay
	g.StatsOverlay.Render(g.Ctx, g)
}

// ProcessInput handles player input.
func (g *Game) ProcessInput() {
	// Fire weapon (X key)
	if g.Keys[88] {
		g.Ship.Fire(g)
	}

	// Ship movement disabled when debug UI is visible (F9)
	// This allows fine-tuning enemy configs without being hit
	if g.DebugUI.Visible {
		return
	}

	g.Ship.Move(g, g.Keys)
}

// ProcessDebugInput handles input for the debug UI
func (g *Game) ProcessDebugInput() {
	// Input is handled in keydown events in input.go
}

// RenderBackground renders the scrolling background.
func (g *Game) RenderBackground() {
	g.Ctx.Call("save")
	bgWidth := g.Level.Background.Get("width").Float()
	g.Ctx.Call("translate", 0, math.Mod(g.Level.Y, bgWidth))
	g.Ctx.Set("fillStyle", g.Level.BackgroundPattern)
	g.Ctx.Call("fillRect", 0, -bgWidth, WIDTH, HEIGHT+bgWidth)
	g.Ctx.Call("restore")
	g.Level.Y += Speed
}

// RenderScore renders the score display.
func (g *Game) RenderScore() {
	points := g.Level.P
	scoreX := WIDTH - g.Level.Points.Width - 8
	for points > 0 {
		g.Ctx.Call("drawImage", g.Level.Points.Images[points%10], scoreX, 4)
		points /= 10
		scoreX -= g.Level.Points.Step
	}
}

// RenderText renders the text display.
func (g *Game) RenderText() {
	if g.Level.Text.T > 0 {
		if g.Level.Text.T < g.Level.Text.MaxT {
			g.Ctx.Set("globalAlpha", float64(g.Level.Text.T)/float64(g.Level.Text.MaxT))
		} else {
			g.Ctx.Set("globalAlpha", 1)
		}
		g.Ctx.Call("drawImage", g.Level.Text.Image, g.Level.Text.X, g.Level.Text.Y)
		g.Ctx.Set("globalAlpha", 1)
		g.Level.Text.T--
		g.Level.Text.Y += int(g.Level.Text.YAcc)
	}
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
func (g *Game) PopulateSpatialGrids() {
	// Clear grids from previous frame
	g.EnemyGrid.Clear()
	g.BulletGrid.Clear()

	// Insert all enemies
	for _, enemy := range g.Enemies {
		g.EnemyGrid.Insert(enemy)
	}

	// Insert all active bullets/torpedos
	for i := 0; i < g.Bullets.ActiveCount; i++ {
		g.BulletGrid.Insert(g.Bullets.Pool[i])
	}
}

// CheckBulletTorpedoCollision checks if a standard bullet hits any nearby torpedos.
func (g *Game) CheckBulletTorpedoCollision(bullet *Bullet) {
	// Get nearby objects from the spatial grid
	nearby := g.BulletGrid.GetNearby(bullet.X, bullet.Y)

	for _, obj := range nearby {
		torpedo, ok := obj.(*Bullet)
		if !ok || torpedo.Kind != TorpedoBullet {
			continue
		}

		// AABB collision check
		if bullet.Y < torpedo.Y+BulletTorpedoCollisionDist &&
			bullet.Y > torpedo.Y-BulletTorpedoCollisionDist &&
			bullet.X < torpedo.X+BulletTorpedoCollisionDist &&
			bullet.X > torpedo.X-BulletTorpedoCollisionDist {

			torpedo.E--
			g.Level.P += 5

			if torpedo.E < 0 {
				// Torpedo destroyed
				if g.GameRNG.Random() > 0.75 {
					g.SpawnBonus(torpedo.X, torpedo.Y, 0, 0, "")
				}
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

		// Render bonus
		g.Ctx.Call("drawImage", g.BonusImages[item.Type],
			int(item.X)-BonusR, int(item.Y)-BonusR)

		// Update position
		item.X += item.XAcc
		item.Y += item.YAcc

		// Remove off-screen items
		if item.Y >= HEIGHT+BonusR*2 || item.X < -BonusR ||
			item.X >= WIDTH+BonusR || item.Y < -BonusR {
			g.Bonuses.Release(idx)
		}
	})
}

// UpdateTorpedos updates and renders torpedoes.
// func (g *Game) UpdateTorpedos() {
// 	g.Torpedos.ForEachReverse(func(torpedo *Torpedo, idx int) {
// 		// Check collision with player
// 		collision := false
// 		for _, s := range g.Ships {
// 			if s.Collision(g, torpedo) && !collision {
// 				collision = true
// 				break
// 			}
// 		}

// 		if collision {
// 			g.Torpedos.Release(idx)
// 			return
// 		}

// 		// Render torpedo with animation
// 		g.Ctx.Call("drawImage", g.TorpedoImages[g.TorpedoFrame],
// 			int(torpedo.X)-TorpedoR, int(torpedo.Y)-TorpedoR)

// 		// Update position
// 		torpedo.X += torpedo.XAcc
// 		torpedo.Y += torpedo.YAcc

// 		// Remove off-screen torpedoes
// 		if torpedo.Y >= HEIGHT+TorpedoR || torpedo.X < -TorpedoR ||
// 			torpedo.X >= WIDTH+TorpedoR || torpedo.Y < -HEIGHT {
// 			g.Torpedos.Release(idx)
// 		}
// 	})

// 	// Advance torpedo animation
// 	g.TorpedoFrame = (g.TorpedoFrame + 1) % len(g.TorpedoImages)
// }

// UpdateExplosions updates and renders explosions.
func (g *Game) UpdateExplosions() {
	g.Explosions.ForEachReverse(func(exp *Explosion, idx int) {
		g.Ctx.Call("save")
		g.Ctx.Set("globalAlpha", exp.Alpha)
		g.Ctx.Call("translate", exp.X, exp.Y)
		g.Ctx.Call("rotate", exp.Angle)
		g.Ctx.Call("drawImage", g.ExplosionImage,
			-exp.Size/2, -exp.Size/2, exp.Size, exp.Size)
		g.Ctx.Call("restore")

		// Animate explosion
		exp.Size += 16
		exp.Angle += exp.D
		exp.Alpha -= 0.1

		if exp.Alpha < 0.1 {
			g.Explosions.Release(idx)
		}
	})
}

// RenderShip renders the player ship.
func (g *Game) RenderShip() {
	for _, s := range g.Ships {
		s.Render(g)
	}
}

// CheckWaveSpawn checks if a new wave should spawn.
func (g *Game) CheckWaveSpawn() {
	if len(g.Enemies) == 0 && g.Level.Text.T == 0 {
		g.Level.P += g.Level.LevelNum * 1000
		g.Level.LevelNum++

		// Generate deterministic seed for this level
		g.Level.LevelSeed = common.LevelSeed(g.GameSeed, g.Level.LevelNum)
		g.GameRNG.SetSeed(g.Level.LevelSeed)

		// Transition music to new level's key/preset
		g.Audio.SetMusicPreset(g.Level.LevelNum)

		// Spawn all enemy types
		g.SpawnEnemy(SmallFighter, -0.75*HEIGHT)
		g.SpawnEnemy(MediumFighter, -1.5*HEIGHT)
		g.SpawnEnemy(TurretFighter, -1*HEIGHT)
		g.SpawnEnemy(Boss, -2.25*HEIGHT)

		g.SpawnText("WAVE "+strconv.Itoa(g.Level.LevelNum), 0)
		g.Level.Bomb = MaxBomb
		g.Audio.Play(8)
	}
}

// UpdateEnemies updates and renders enemies.
func (g *Game) UpdateEnemies() {
	for i, e := range g.Enemies {
		if !e.Render(g) {
			// enemy no longer exists
			g.RemoveEnemy(i)
		}
	}

	// for i := len(g.Enemies) - 1; i >= 0; i-- {

	// 	enemyDestroyed := false
	// 	g.Bullets.ForEachReverse(func(bullet *Bullet, bulletIdx int) {
	// 		if enemyDestroyed {
	// 			return
	// 		}
	// 		if bullet.Y < enemy.Y+hitboxD && bullet.Y > enemy.Y-hitboxD &&
	// 			bullet.X > enemy.X-hitboxD && bullet.X < enemy.X+hitboxD {

	// 			g.Level.P++
	// 			g.Explode(bullet.X, bullet.Y, 0)
	// 			g.Bullets.Release(bulletIdx)

	// 			enemy.Health--
	// 			if enemy.Health <= 0 {
	// 				g.Level.P += 100
	// 				g.SpawnBonus(enemy.X, enemy.Y, 0, 0, "")
	// 				g.Explode(enemy.X, enemy.Y, enemy.Radius*2)
	// 				g.Explode(enemy.X, enemy.Y, enemy.Radius*3)
	// 				vol := enemy.DistanceVolume(g.Ships.X, g.Ships.Y, float64(HEIGHT))
	// 				g.Audio.PlayWithPan(10, enemy.AudioPan(), vol)
	// 				g.RemoveEnemy(i)
	// 				enemyDestroyed = true
	// 			} else {
	// 				// hit sound
	// 				vol := enemy.DistanceVolume(g.Ships.X, g.Ships.Y, float64(HEIGHT))
	// 				g.Audio.PlayWithPan(9, enemy.AudioPan(), vol)
	// 			}
	// 		}
	// 	})
	// }

	// Update enemy-reactive synth layers
	g.updateEnemySynthLayers()
}

// updateEnemySynthLayers calculates enemy metrics and updates synth.
func (g *Game) updateEnemySynthLayers() {
	// enemyCount := len(g.Enemies)
	// if enemyCount == 0 {
	// 	g.Audio.UpdateEnemySynth(0, float64(HEIGHT), 9999)
	// 	return
	// }

	// // Calculate average enemy Y position and closest distance
	// var totalY float64
	// closestDist := 9999.0

	// for _, enemy := range g.Enemies {
	// 	enemyY := enemy.Y + enemy.YOffset
	// 	totalY += enemyY

	// 	// Distance from enemy to player
	// 	dx := enemy.X - g.Ships.X
	// 	dy := enemyY - g.Ships.Y
	// 	dist := math.Sqrt(dx*dx + dy*dy)
	// 	if dist < closestDist {
	// 		closestDist = dist
	// 	}
	// }

	// avgY := totalY / float64(enemyCount)
	// g.Audio.UpdateEnemySynth(enemyCount, avgY, closestDist)
}

// RenderEnergyBar renders the energy bar HUD.
// func (g *Game) RenderEnergyBar() {
// 	if g.Ships.OSD > 0 {
// 		barX := int(g.Ships.X) - 32
// 		barY := int(g.Ships.Y) + 63
// 		colorValue := g.Ships.E * 512 / 100

// 		g.Ctx.Set("globalAlpha", float64(g.Ships.OSD)/float64(ShipMaxOSD))
// 		g.Ctx.Set("fillStyle", Theme.EnergyBarBackground)
// 		g.Ctx.Call("fillRect", barX, barY, 64, 4)

// 		var r, gr int
// 		if colorValue > 255 {
// 			r = 512 - colorValue
// 			gr = 255
// 		} else {
// 			r = 255
// 			gr = colorValue
// 		}
// 		g.Ctx.Set("fillStyle", "rgb("+strconv.Itoa(r)+","+strconv.Itoa(gr)+",0)")
// 		g.Ctx.Call("fillRect", barX, barY, g.Ships.E*64/100, 4)

// 		g.Ctx.Set("lineWidth", Theme.EnergyBarLineWidth)
// 		g.Ctx.Set("strokeStyle", Theme.EnergyBarBorder)
// 		g.Ctx.Call("strokeRect", barX, barY, 64, 4)
// 		g.Ctx.Set("globalAlpha", 1)

// 		if g.Ships.E >= 25 {
// 			g.Ships.OSD--
// 		}
// 	}
// }

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
