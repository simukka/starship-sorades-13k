package game

import (
	"math"

	"github.com/gopherjs/gopherjs/js"
)

// Game holds the complete game state.
type Game struct {
	// Core state
	Level    Level
	Ship     Ship
	Enemies  []*Enemy
	GameSeed uint32
	GameRNG  *SeededRNG

	// Object pools
	Bullets    *BulletPool
	Torpedos   *TorpedoPool
	Explosions *ExplosionPool
	Bonuses    *BonusPool

	// Audio
	Audio *AudioManager

	// Rendering
	Canvas *js.Object
	Ctx    *js.Object

	// Input
	Keys map[int]bool

	// Animation
	AnimationFrameID int
	LastFrameTime    float64

	// Graphics assets
	BulletImage    *js.Object
	ExplosionImage *js.Object
	TorpedoImages  []*js.Object
	TorpedoFrame   int
	BonusImages    map[string]*js.Object
	EnemyTypes     []EnemyType
}

// NewGame creates a new game instance.
func NewGame() *Game {
	Debug("new game")
	g := &Game{
		Enemies:     make([]*Enemy, 0, 64),
		GameRNG:     NewSeededRNG(0),
		Bullets:     NewBulletPool(200),
		Torpedos:    NewTorpedoPool(150),
		Explosions:  NewExplosionPool(50),
		Bonuses:     NewBonusPool(30),
		Audio:       NewAudioManager(),
		Keys:        make(map[int]bool),
		BonusImages: make(map[string]*js.Object),
		EnemyTypes:  make([]EnemyType, 4),
	}
	g.initLevelDefaults()
	g.initShipDefaults()
	return g
}

// initShipDefaults initializes ship state to default values.
func (g *Game) initShipDefaults() {
	g.Ship = Ship{
		X:       WIDTH / 2,
		Y:       HEIGHT * 7 / 8,
		E:       100,
		XAcc:    0,
		YAcc:    0,
		Angle:   0,
		Weapon:  0,
		Reload:  0,
		Timeout: 0,
		OSD:     0,
		Shield: Shield{
			MaxT: ShipMaxShield,
		},
	}
}

// initLevelDefaults initializes level state to default values.
func (g *Game) initLevelDefaults() {
	g.Level = Level{
		LevelNum:  0,
		LevelSeed: 0,
		P:         0,
		Y:         0,
		Bomb:      0,
		Paused:    true,
		Text:      TextDisplay{MaxT: 90},
		Points: Points{
			Width:  32,
			Height: 48,
			Step:   24,
			Images: make([]*js.Object, 10),
		},
	}
}

// Hurt applies damage to the player ship.
func (g *Game) Hurt(damage int) {
	if g.Ship.Shield.T > 0 {
		g.Audio.Play(2)
		return
	}

	if g.Ship.Timeout < 0 {
		g.Ship.E -= damage
		if g.Ship.E < 0 {
			g.Ship.E = 0
		}
		g.Ship.Timeout = 10
	}

	if g.Ship.E == 0 && !g.Level.Paused {
		g.Explode(g.Ship.X, g.Ship.Y, 512)
		g.Explode(g.Ship.X, g.Ship.Y, 1024)
		g.Level.Paused = true
		g.SpawnText("GAME OVER", -1)
		g.Audio.Play(14)
	} else if g.Ship.E < 25 {
		g.Audio.Play(17)
	}

	g.Audio.Play(1)

	if g.Ship.Weapon > 2 {
		g.Ship.Weapon--
	}
	g.Ship.OSD = ShipMaxOSD
}

// Fire creates a bullet from the player's ship.
func (g *Game) Fire(xVel, yVel float64) {
	bullet := g.Bullets.Acquire()
	if bullet == nil {
		return
	}

	finalXVel := xVel + g.Ship.XAcc/2
	finalYVel := yVel + g.Ship.YAcc/2

	bullet.T = BulletMaxT
	bullet.X = g.Ship.X + math.Floor(js.Global.Get("Math").Call("random").Float()*finalXVel)
	bullet.Y = g.Ship.Y + math.Floor(js.Global.Get("Math").Call("random").Float()*finalYVel)
	bullet.XAcc = finalXVel
	bullet.YAcc = finalYVel
}

// Explode creates an explosion effect.
func (g *Game) Explode(x, y, size float64) {
	exp := g.Explosions.Acquire()
	if exp == nil {
		return
	}

	exp.X = x
	exp.Y = y
	if size == 0 {
		exp.Size = js.Global.Get("Math").Call("random").Float() * 64
	} else {
		exp.Size = size
	}
	exp.Angle = js.Global.Get("Math").Call("random").Float()
	exp.D = js.Global.Get("Math").Call("random").Float()*0.4 - 0.2
	exp.Alpha = 1
}

// SpawnBonus spawns a bonus item.
func (g *Game) SpawnBonus(x, y, xAcc, yAcc float64, bonusType string) {
	if bonusType == "" {
		r := g.GameRNG.Random()
		if r > 0.9 {
			bonusType = "+"
		} else if r > 0.8 {
			bonusType = "E"
		} else if r > 0.7 {
			bonusType = "S"
		} else if r > 0.6 {
			bonusType = "B"
		} else {
			bonusType = "10"
		}
	}

	// Lazy render bonus image if not cached
	if _, ok := g.BonusImages[bonusType]; !ok {
		g.RenderBonusImage(bonusType)
	}

	item := g.Bonuses.Acquire()
	if item == nil {
		return
	}

	item.Type = bonusType
	if x == 0 {
		item.X = WIDTH / 2
	} else {
		item.X = x
	}
	if y == 0 {
		item.Y = -BonusR
	} else {
		item.Y = y
	}
	if xAcc != 0 {
		item.XAcc = xAcc / 2
	} else {
		item.XAcc = g.GameRNG.RandomFloat(-Speed/2, Speed/2)
	}
	item.YAcc = yAcc/2 + Speed/2
}

// SpawnTorpedo spawns an enemy torpedo.
func (g *Game) SpawnTorpedo(x, y, yOffset, angle, maxAngle float64) bool {
	yPos := y + yOffset

	if yPos < -HEIGHT/4 {
		return false
	}

	if maxAngle > 0 {
		if angle > math.Pi {
			angle -= math.Pi * 2
		}
		if angle > maxAngle {
			angle = maxAngle
		}
		if angle < -maxAngle {
			angle = -maxAngle
		}
	}

	torpedo := g.Torpedos.Acquire()
	if torpedo == nil {
		return false
	}

	speed := 3.0 + float64(g.Level.LevelNum)/2

	torpedo.X = math.Floor(x)
	torpedo.Y = yPos
	if angle != 0 {
		torpedo.XAcc = math.Sin(angle) * speed
		torpedo.YAcc = math.Cos(angle) * speed
	} else {
		torpedo.XAcc = 0
		torpedo.YAcc = speed
	}
	torpedo.E = 0

	return true
}

// SpawnText displays a centered text message.
func (g *Game) SpawnText(text string, duration int) {
	g.RenderTextImage(text)
	g.spawnTextWithDuration(duration)

	if duration < 0 {
		g.Ctx.Set("globalAlpha", 1)
		g.Ctx.Call("drawImage", g.Level.Text.Image, g.Level.Text.X, g.Level.Text.Y)
	}
}

// spawnTextWithDuration sets up text display position and timing.
// Extracted for testability without browser APIs.
func (g *Game) spawnTextWithDuration(duration int) {
	g.Level.Text.X = (WIDTH - g.Level.Text.Image.Get("width").Int()) / 2
	g.Level.Text.Y = 16
	g.Level.Text.YAcc = Speed / 2
	if duration == 0 {
		g.Level.Text.T = g.Level.Text.MaxT
	} else if duration < 0 {
		g.Level.Text.T = 0
	} else {
		g.Level.Text.T = duration
	}
}

// SpawnEnemy spawns enemies of a specific type.
func (g *Game) SpawnEnemy(typeIndex int, yOffset float64) {
	if typeIndex < 0 || typeIndex >= len(g.EnemyTypes) {
		typeIndex = g.GameRNG.RandomInt(0, len(g.EnemyTypes))
	}

	// Call the appropriate spawn function based on type
	switch typeIndex {
	case 0:
		g.SpawnEnemyType0(yOffset)
	case 1:
		g.SpawnEnemyType1(yOffset)
	case 2:
		g.SpawnEnemyType2(yOffset)
	case 3:
		g.SpawnEnemyType3(yOffset)
	}
}

// SetGameSeed sets the game seed for deterministic gameplay.
func (g *Game) SetGameSeed(seed uint32) {
	g.GameSeed = seed
	g.GameRNG.SetSeed(seed)
	g.Level.LevelNum = 0
}

// GetGameSeed returns the current game seed.
func (g *Game) GetGameSeed() uint32 {
	return g.GameSeed
}

// GetLevelSeed returns the current level seed.
func (g *Game) GetLevelSeed() uint32 {
	return g.Level.LevelSeed
}

// Start begins or resumes the game.
func (g *Game) Start() {
	Debug("Start!")
	g.Level.Paused = false

	// Cancel existing animation frame
	if g.AnimationFrameID > 0 {
		js.Global.Call("cancelAnimationFrame", g.AnimationFrameID)
	}

	// Fade out title and play music
	g.Audio.FadeOutTitle()
	g.Audio.PlayMusic()

	// Start game loop
	g.AnimationFrameID = js.Global.Call("requestAnimationFrame", g.GameLoopRAF).Int()
}

// GameOver handles game over state.
func (g *Game) GameOver() {
	g.Level.Paused = true

	// Stop music
	g.Audio.StopMusic()

	// Play game over sound
	g.Audio.Play(14)

	// Display game over text
	g.SpawnText("GAME OVER", 0)

	// Render title screen after delay
	js.Global.Call("setTimeout", func() {
		g.RenderTitleScreen()
	}, 3000)
}

// RenderTitleScreen renders the title screen.
func (g *Game) RenderTitleScreen() {
	Debug("RenderTitleScreen")

	// Clear screen
	g.Ctx.Set("fillStyle", Theme.BackgroundColor)
	g.Ctx.Call("fillRect", 0, 0, WIDTH, HEIGHT)

	// Draw title text
	g.SpawnText("STARSHIP SORADES", 10)

	// Draw instructions
	g.Ctx.Set("fillStyle", Theme.TextSecondaryColor)
	g.Ctx.Set("font", Theme.InstructFont)
	g.Ctx.Set("textAlign", "center")
	g.Ctx.Call("fillText", "C", WIDTH/2, HEIGHT/2+50)
	g.Ctx.Call("fillText", "ARROWS TO MOVE, X TO FIRE, F FOR FULLSCREEN", WIDTH/2, HEIGHT/2+80)

	// Play title music
	g.Audio.PlayTitle()
}

// ResetGame resets all game state for a new game.
func (g *Game) ResetGame() {
	// Reset ship (preserve Image assets)
	shipImage := g.Ship.Image
	shieldImage := g.Ship.Shield.Image

	g.initShipDefaults()
	g.Ship.Image = shipImage
	g.Ship.Shield.Image = shieldImage

	// Reset level (preserve Background, Points assets)
	g.initLevelDefaults()

	// Clear all pools
	g.Bullets.Clear()
	g.Torpedos.Clear()
	g.Explosions.Clear()
	g.Bonuses.Clear()

	// Clear enemies
	g.Enemies = g.Enemies[:0]

	// Reset game RNG
	g.GameRNG.SetSeed(g.GameSeed)
}
