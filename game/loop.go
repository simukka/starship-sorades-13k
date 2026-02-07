package game

import (
	"math"
	"strconv"

	"github.com/gopherjs/gopherjs/js"
)

// GameLoopRAF is the main game loop using requestAnimationFrame.
func (g *Game) GameLoopRAF(currentTime float64) {
	// Schedule next frame
	g.AnimationFrameID = js.Global.Call("requestAnimationFrame", g.GameLoopRAF).Int()

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
	Debug("=== Game Loop Start ===")
	Debug("Ship pos:", g.Ship.X, g.Ship.Y, "Energy:", g.Ship.E, "Weapon:", g.Ship.Weapon)
	Debug("Level:", g.Level.LevelNum, "Score:", g.Level.P, "Enemies:", len(g.Enemies))
	Debug("Bullets:", g.Bullets.ActiveCount, "Torpedos:", g.Torpedos.ActiveCount, "Explosions:", g.Explosions.ActiveCount)

	// Player Input Processing
	g.ProcessInput()

	// Background Rendering
	g.RenderBackground()

	// Score Display
	g.RenderScore()

	// Text Display
	g.RenderText()

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

	// Torpedo Update
	g.UpdateTorpedos()

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

	// Energy Bar HUD
	g.RenderEnergyBar()
}

// ProcessInput handles player input.
func (g *Game) ProcessInput() {
	// Fire weapon
	g.Ship.Reload--
	if g.Ship.Reload <= 0 && g.Keys[88] {
		if g.Ship.Weapon > 0 {
			g.Ship.Reload = 4
		} else {
			g.Ship.Reload = 6
		}
		g.Fire(0, -16)

		if g.Ship.Weapon > 1 {
			g.Fire(-8, -8)
			g.Fire(8, -8)

			if g.Ship.Weapon > 2 {
				g.Fire(0, 16)

				if g.Ship.Weapon > 3 {
					g.Fire(-16, 0)
					g.Fire(16, 0)
				}
			}
		}

		if g.Ship.Weapon > 2 {
			g.Audio.Play(16)
		} else if g.Ship.Weapon > 0 {
			g.Audio.Play(0)
		} else {
			g.Audio.Play(15)
		}
	}

	// Movement input with visual tilt
	g.Ship.Angle *= ShipAngleFactor

	// Left arrow
	if g.Keys[37] {
		if g.Ship.X >= WIDTH && g.Ship.XAcc > 0 {
			g.Ship.XAcc = 0
		}
		g.Ship.XAcc -= ShipACC
		g.Ship.Angle = (g.Ship.Angle+1)*ShipAngleFactor - 1
	}
	// Up arrow
	if g.Keys[38] {
		if g.Ship.Y >= HEIGHT && g.Ship.YAcc > 0 {
			g.Ship.YAcc = 0
		}
		g.Ship.YAcc -= ShipACC
	}
	// Right arrow
	if g.Keys[39] {
		if g.Ship.X < 0 && g.Ship.XAcc < 0 {
			g.Ship.XAcc = 0
		}
		g.Ship.XAcc += ShipACC
		g.Ship.Angle = (g.Ship.Angle-1)*ShipAngleFactor + 1
	}
	// Down arrow
	if g.Keys[40] {
		if g.Ship.Y < 0 && g.Ship.YAcc < 0 {
			g.Ship.YAcc = 0
		}
		g.Ship.YAcc += ShipACC
	}

	// Screen boundary collision
	if g.Ship.X < 0 && g.Ship.XAcc < 0 {
		g.Ship.X = 0
	} else if g.Ship.X >= WIDTH && g.Ship.XAcc > 0 {
		g.Ship.X = WIDTH - 1
	}
	if g.Ship.Y < 0 && g.Ship.YAcc < 0 {
		g.Ship.Y = 0
	} else if g.Ship.Y >= HEIGHT && g.Ship.YAcc > 0 {
		g.Ship.Y = HEIGHT - 1
	}

	// Apply velocity and damping
	g.Ship.X += g.Ship.XAcc
	g.Ship.Y += g.Ship.YAcc
	g.Ship.XAcc *= ShipACCFactor
	g.Ship.YAcc *= ShipACCFactor
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
	bulletTorpedoCollisionDist := 12.0

	g.Bullets.ForEachReverse(func(bullet *Bullet, bulletIdx int) {
		// Render bullet
		g.Ctx.Call("drawImage", g.BulletImage,
			int(bullet.X)-BulletR, int(bullet.Y)-BulletR)

		// Update position
		bullet.X += bullet.XAcc
		bullet.Y += bullet.YAcc

		// Check collision with torpedoes
		g.Torpedos.ForEachReverse(func(torpedo *Torpedo, torpedoIdx int) {
			if bullet.Y < torpedo.Y+bulletTorpedoCollisionDist &&
				bullet.Y > torpedo.Y-bulletTorpedoCollisionDist &&
				bullet.X < torpedo.X+bulletTorpedoCollisionDist &&
				bullet.X > torpedo.X-bulletTorpedoCollisionDist {

				torpedo.E--
				if torpedo.E < 0 {
					g.Level.P += 5
					if g.GameRNG.Random() > 0.75 {
						g.SpawnBonus(torpedo.X, torpedo.Y, 0, 0, "")
					}
					g.Explode(torpedo.X, torpedo.Y, 0)
					g.Torpedos.Release(torpedoIdx)
					g.Audio.Play(11)
				}
				bullet.T = 0
			}
		})

		// Remove expired or off-screen bullets
		bullet.T--
		if bullet.T < 0 || bullet.X < -BulletR ||
			bullet.X >= WIDTH+BulletR || bullet.Y >= HEIGHT+BulletR {
			g.Bullets.Release(bulletIdx)
		}
	})
}

// UpdateBonuses updates and renders bonus items.
func (g *Game) UpdateBonuses() {
	shipCollisionD := float64(ShipR) * 0.8
	shipCollisionE := float64(ShipR) * 0.4

	g.Bonuses.ForEachReverse(func(item *Bonus, idx int) {
		// Check collision with player
		if g.Ship.Y < item.Y+shipCollisionD && g.Ship.Y > item.Y-shipCollisionD &&
			g.Ship.X < item.X+shipCollisionE && g.Ship.X > item.X-shipCollisionE {

			g.Level.P += 10

			switch item.Type {
			case "+":
				if g.Ship.Weapon < 4 {
					g.Ship.Weapon++
					g.Audio.Play(5)
				} else {
					g.Audio.Play(6)
				}
			case "E":
				if g.Ship.E < 100 {
					g.Ship.OSD = ShipMaxOSD
					g.Audio.Play(5)
				} else {
					g.Audio.Play(6)
				}
				g.Ship.E += 5
				if g.Ship.E > 100 {
					g.Ship.E = 100
				}
			case "S":
				g.Ship.Shield.T += g.Ship.Shield.MaxT * g.Ship.Shield.MaxT *
					2 / (g.Ship.Shield.T + g.Ship.Shield.MaxT*2)
				g.Audio.Play(3)
			case "B":
				for j := len(g.Enemies) - 1; j >= 0; j-- {
					g.Enemies[j].E--
				}
				for j := 0; j < min(g.Torpedos.ActiveCount, 5); j++ {
					g.Explode(g.Torpedos.Pool[j].X, g.Torpedos.Pool[j].Y, 0)
				}
				g.Torpedos.Clear()
				g.Level.Bomb = MaxBomb
				g.Audio.Play(13)
			default:
				g.Audio.Play(7)
			}

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
func (g *Game) UpdateTorpedos() {
	shipCollisionD := float64(ShipR) * 0.8
	shipCollisionE := float64(ShipR) * 0.4

	g.Torpedos.ForEachReverse(func(torpedo *Torpedo, idx int) {
		// Check collision with player
		if g.Ship.Y < torpedo.Y+shipCollisionD && g.Ship.Y > torpedo.Y-shipCollisionD &&
			g.Ship.X < torpedo.X+shipCollisionE && g.Ship.X > torpedo.X-shipCollisionE {
			g.Hurt(10)
			g.Explode(torpedo.X, torpedo.Y, 0)
			g.Torpedos.Release(idx)
			return
		}

		// Render torpedo with animation
		g.Ctx.Call("drawImage", g.TorpedoImages[g.TorpedoFrame],
			int(torpedo.X)-TorpedoR, int(torpedo.Y)-TorpedoR)

		// Update position
		torpedo.X += torpedo.XAcc
		torpedo.Y += torpedo.YAcc

		// Remove off-screen torpedoes
		if torpedo.Y >= HEIGHT+TorpedoR || torpedo.X < -TorpedoR ||
			torpedo.X >= WIDTH+TorpedoR || torpedo.Y < -HEIGHT {
			g.Torpedos.Release(idx)
		}
	})

	// Advance torpedo animation
	g.TorpedoFrame = (g.TorpedoFrame + 1) % len(g.TorpedoImages)
	g.Ship.Timeout--
}

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
	g.Ctx.Call("save")
	g.Ctx.Call("translate", g.Ship.X, g.Ship.Y)
	g.Ctx.Call("rotate", g.Ship.Angle*ShipMaxAngle/180*math.Pi)
	g.Ctx.Call("drawImage", g.Ship.Image, -ShipR/2, -ShipR)
	g.Ctx.Call("restore")

	// Shield effect
	if g.Ship.Shield.T > 0 {
		mathRandom := js.Global.Get("Math").Call("random").Float()
		if g.Ship.Shield.T > 30 || mathRandom > 0.5 {
			g.Ctx.Call("drawImage", g.Ship.Shield.Image,
				int(g.Ship.X)-ShipR, int(g.Ship.Y)-ShipR)
		}
		g.Ship.Shield.T--
		if g.Ship.Shield.T == 0 {
			g.Audio.Play(4)
		}
	}
}

// CheckWaveSpawn checks if a new wave should spawn.
func (g *Game) CheckWaveSpawn() {
	if len(g.Enemies) == 0 && g.Level.Text.T == 0 {
		g.Level.P += g.Level.LevelNum * 1000
		g.Level.LevelNum++

		// Generate deterministic seed for this level
		g.Level.LevelSeed = LevelSeed(g.GameSeed, g.Level.LevelNum)
		g.GameRNG.SetSeed(g.Level.LevelSeed)

		// Spawn all enemy types
		g.SpawnEnemy(0, -0.75*HEIGHT)
		g.SpawnEnemy(1, -1.5*HEIGHT)
		g.SpawnEnemy(2, -1*HEIGHT)
		g.SpawnEnemy(3, -2.25*HEIGHT)

		g.SpawnText("WAVE "+strconv.Itoa(g.Level.LevelNum), 0)
		g.Level.Bomb = MaxBomb
		g.Audio.Play(8)
	}
}

// UpdateEnemies updates and renders enemies.
func (g *Game) UpdateEnemies() {
	for i := len(g.Enemies) - 1; i >= 0; i-- {
		enemy := g.Enemies[i]
		enemyY := enemy.Y + enemy.YOffset

		// Calculate angle to player
		angle := math.Atan((enemy.X - g.Ship.X) / (enemyY - g.Ship.Y))
		if g.Ship.Y <= enemyY {
			angle += math.Pi
		}

		// Calculate visual rotation
		bossAngle := math.Mod(angle+math.Pi, math.Pi*2) - math.Pi
		if bossAngle > enemy.MaxAngle {
			bossAngle = enemy.MaxAngle
		}
		if bossAngle < -enemy.MaxAngle {
			bossAngle = -enemy.MaxAngle
		}
		enemy.Angle = (enemy.Angle*29 - bossAngle) / 30

		// Render enemy
		g.Ctx.Call("save")
		g.Ctx.Call("translate", enemy.X, enemyY)
		g.Ctx.Call("rotate", enemy.Angle)
		g.Ctx.Call("drawImage", enemy.Image,
			-enemy.R, enemy.Y-enemyY-enemy.R, enemy.R*2, enemy.R*2)
		g.Ctx.Call("restore")

		// Check bullet collisions
		hitboxD := enemy.R * 0.6
		g.Bullets.ForEachReverse(func(bullet *Bullet, bulletIdx int) {
			if bullet.Y < enemy.Y+hitboxD && bullet.Y > enemy.Y-hitboxD &&
				bullet.X > enemy.X-hitboxD && bullet.X < enemy.X+hitboxD {

				g.Level.P++
				g.Explode(bullet.X, bullet.Y, 0)
				g.Bullets.Release(bulletIdx)

				enemy.E--
				if enemy.E <= 0 {
					g.Level.P += 100
					g.SpawnBonus(enemy.X, enemy.Y, 0, 0, "")
					g.Explode(enemy.X, enemy.Y, enemy.R*2)
					g.Explode(enemy.X, enemy.Y, enemy.R*3)
					g.RemoveEnemy(i)
					g.Audio.Play(10)
				} else {
					g.Audio.Play(9)
				}
			}
		})

		// Skip if enemy was destroyed
		if i >= len(g.Enemies) {
			continue
		}

		// Move enemy toward stop position
		if enemy.Y < enemy.YStop {
			enemy.Y++
		}

		// Fire at player
		enemy.T--
		if enemy.T < 0 {
			g.EnemyShoot(enemy, angle)
		}
	}
}

// RenderEnergyBar renders the energy bar HUD.
func (g *Game) RenderEnergyBar() {
	if g.Ship.OSD > 0 {
		barX := int(g.Ship.X) - 32
		barY := int(g.Ship.Y) + 63
		colorValue := g.Ship.E * 512 / 100

		g.Ctx.Set("globalAlpha", float64(g.Ship.OSD)/float64(ShipMaxOSD))
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
		g.Ctx.Call("fillRect", barX, barY, g.Ship.E*64/100, 4)

		g.Ctx.Set("lineWidth", Theme.EnergyBarLineWidth)
		g.Ctx.Set("strokeStyle", Theme.EnergyBarBorder)
		g.Ctx.Call("strokeRect", barX, barY, 64, 4)
		g.Ctx.Set("globalAlpha", 1)

		if g.Ship.E >= 25 {
			g.Ship.OSD--
		}
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
