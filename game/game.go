package game

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/simukka/starship-sorades-13k/audio"
	"github.com/simukka/starship-sorades-13k/common"
)

// Game holds the complete game state.
type Game struct {
	// Core state
	Level    Level
	Ships    []*Ship
	Ship     *Ship
	Enemies  []*Enemy
	GameSeed uint32
	GameRNG  *common.SeededRNG

	// Object pools
	Bullets    *BulletPool
	Explosions *ExplosionPool
	Bonuses    *BonusPool

	// Audio
	Audio *audio.AudioManager

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
	EnemyTypes     map[EnemyKind]EnemyType

	// Debug UI
	DebugUI      *DebugUI
	StatsOverlay *StatsOverlay

	// Collision detection
	EnemyGrid  *SpatialGrid // Spatial hash for enemies
	BulletGrid *SpatialGrid // Spatial hash for bullets/torpedos
}

// NewGame creates a new game instance.
func NewGame() *Game {
	seed := common.NewSeededRNG(0)
	g := &Game{
		Enemies:      make([]*Enemy, 0, 64),
		GameRNG:      seed,
		Bullets:      NewBulletPool(350),
		Explosions:   NewExplosionPool(50),
		Bonuses:      NewBonusPool(30),
		Audio:        audio.NewAudioManager(seed, HEIGHT),
		Keys:         make(map[int]bool),
		BonusImages:  make(map[string]*js.Object),
		EnemyTypes:   make(map[EnemyKind]EnemyType, 4),
		DebugUI:      NewDebugUI(),
		StatsOverlay: NewStatsOverlay(),
		EnemyGrid:    NewSpatialGrid(WIDTH, HEIGHT, 64),
		BulletGrid:   NewSpatialGrid(WIDTH, HEIGHT, 64),
	}
	g.initLevelDefaults()
	g.initShipDefaults()
	return g
}

// initShipDefaults initializes ship state to default values.
func (g *Game) initShipDefaults() {
	g.Ships = append(g.Ships, &Ship{
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
		local: true,
	})

	g.Ship = g.Ships[0]
	g.Ship.AddWeapon()
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

// Hurt applies damage to a player's ship.
func (g *Game) Hurt(damage int, ship *Ship) {
	if ship.Shield.T > 0 {
		g.Audio.PlayWithPan(2, ship.AudioPan(), 1.0)
		return
	}

	if ship.Timeout < 0 {
		ship.E -= damage
		if ship.E < 0 {
			ship.E = 0
		}
		ship.Timeout = 10
	}

	if ship.E == 0 && !g.Level.Paused {
		g.Explode(ship.X, ship.Y, 512)
		g.Explode(ship.X, ship.Y, 1024)
		g.Level.Paused = true
		g.SpawnText("GAME OVER", -1)
		g.Audio.PlayWithPan(14, ship.AudioPan(), 1.0)
	} else if ship.E < 25 {
		g.Audio.PlayWithPan(17, ship.AudioPan(), 1.0)
	}

	g.Audio.PlayWithPan(1, ship.AudioPan(), 1.0)

	if ship.Weapon > 2 {
		ship.Weapon--
	}
	ship.OSD = ShipMaxOSD
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

	// Fade out title and start synth music (seeded for deterministic music)
	g.Audio.FadeOutTitle()
	g.Audio.StartSynthMusic(g.GameSeed, g.Level.LevelNum+1) // Use level 1 preset at start

	// Start game loop
	g.AnimationFrameID = js.Global.Call("requestAnimationFrame", g.GameLoopRAF).Int()
}

// GameOver handles game over state.
func (g *Game) GameOver() {
	g.Level.Paused = true

	// Stop music
	g.Audio.StopMusic()
	g.Audio.StopSynthMusic()

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
	g.initShipDefaults()

	// Reset level (preserve Background, Points assets)
	g.initLevelDefaults()

	// Clear all pools
	g.Bullets.Clear()
	g.Explosions.Clear()
	g.Bonuses.Clear()

	// Clear enemies
	g.Enemies = g.Enemies[:0]

	// Reset game RNG
	g.GameRNG.SetSeed(g.GameSeed)
}

// RemoveEnemy removes an enemy using swap-and-pop.
func (g *Game) RemoveEnemy(index int) {
	last := len(g.Enemies) - 1
	if index != last {
		g.Enemies[index] = g.Enemies[last]
	}
	g.Enemies = g.Enemies[:last]
}
