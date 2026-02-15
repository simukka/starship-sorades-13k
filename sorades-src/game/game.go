package game

import (
	"math"

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
	Bases    []*Base
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
	// DebugUI      *DebugUI
	StatsOverlay *StatsOverlay
	ShipHUD      *ShipHUD

	// Collision detection
	EnemyGrid  *SpatialGrid // Spatial hash for enemies
	BulletGrid *SpatialGrid // Spatial hash for bullets/torpedos

	// Camera for infinite world
	Camera *Camera

	// Multiplayer
	Network *NetworkManager
}

// NewGame creates a new game instance.
func NewGame(canvas, ctx *js.Object) *Game {
	seed := common.NewSeededRNG(0)
	g := &Game{
		Enemies:     make([]*Enemy, 0, 64),
		Bases:       make([]*Base, 0, 8),
		GameRNG:     seed,
		Bullets:     NewBulletPool(350),
		Explosions:  NewExplosionPool(50),
		Bonuses:     NewBonusPool(30),
		Audio:       audio.NewAudioManager(seed, HEIGHT),
		Keys:        make(map[int]bool),
		BonusImages: make(map[string]*js.Object),
		EnemyTypes:  make(map[EnemyKind]EnemyType, 4),
		// DebugUI:      NewDebugUI(),
		StatsOverlay: NewStatsOverlay(),
		ShipHUD:      NewShipHUD(),
		EnemyGrid:    NewSpatialGrid(WIDTH, HEIGHT, 64),
		BulletGrid:   NewSpatialGrid(WIDTH, HEIGHT, 64),
		Camera:       &Camera{X: 0, Y: 0},
		Canvas:       canvas,
		Ctx:          ctx,
	}
	// Initialize audio
	sounds := g.Audio.Init()

	// Load all sound effects using pure Go jsfxr implementation
	for i, dataURL := range sounds {
		g.Audio.LoadSound(i, dataURL)
	}

	// Initialize audio control panel (right-click to open)
	g.Audio.InitControlPanel(g.Canvas)

	g.initLevelDefaults()
	g.initShipDefaults()

	// Initialize graphics (background, sprites, etc.)
	// Must be after initShipDefaults() so ship sprites can be rendered
	g.InitializeGraphics()

	g.SetupInputHandlers()

	g.initBases()
	g.Start()
	return g
}

// initShipDefaults initializes ship state to default values.
func (g *Game) initShipDefaults() {
	g.Ships = append(g.Ships, &Ship{
		X:       0, // Start at world origin
		Y:       0,
		VelX:    0,
		VelY:    0,
		E:       100,
		XAcc:    0,
		YAcc:    0,
		Angle:   0, // Facing "up" (negative Y in world space)
		Weapon:  0,
		Reload:  0,
		Timeout: 0,
		OSD:     0,
		Shield: Shield{
			MaxT: ShipMaxShield,
		},
		local:  true,
		Paused: false,
	})

	g.Ship = g.Ships[0]
	g.Ship.AddWeapon()
}

// initBases initializes the starting base(s).
func (g *Game) initBases() {
	// Create a base at the world origin (where the ship starts)
	g.Bases = append(g.Bases, NewBase(0, 0))
}

// initLevelDefaults initializes level state to default values.
func (g *Game) initLevelDefaults() {
	g.Level = Level{
		LevelNum:  0,
		LevelSeed: 0,
		P:         0,
		Y:         0,
		Bomb:      0,
		Paused:    false,
		Text:      TextDisplay{MaxT: 90},
		Points: Points{
			Width:  32,
			Height: 48,
			Step:   24,
			Images: make([]*js.Object, 10),
		},
	}
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
	// Cancel existing animation frame
	if g.AnimationFrameID > 0 {
		js.Global.Call("cancelAnimationFrame", g.AnimationFrameID)
	}

	g.Audio.StartSynthMusic(g.GameSeed, 1) // Use level 1 preset at start

	// Start game loop
	g.AnimationFrameID = js.Global.Call("requestAnimationFrame", g.GameLoopRAF).Int()

	if g.Audio.AudioCtx != nil && g.Audio.AudioCtx.Get("state").String() == "suspended" {
		g.Audio.AudioCtx.Call("resume")
	}

	g.JoinMultiplayer("vipps")
}

// RemoveEnemy removes an enemy using swap-and-pop.
func (g *Game) RemoveEnemy(index int) {
	last := len(g.Enemies) - 1
	if index != last {
		g.Enemies[index] = g.Enemies[last]
	}
	g.Enemies = g.Enemies[:last]
}

// GetAllEntities returns all entities (ships and enemies) as an Entity interface slice.
// Useful for operations that need to iterate over all game entities uniformly.
func (g *Game) GetAllEntities() []Entity {
	entities := make([]Entity, 0, len(g.Ships)+len(g.Enemies))
	for _, s := range g.Ships {
		entities = append(entities, s)
	}
	for _, e := range g.Enemies {
		entities = append(entities, e)
	}
	return entities
}

// ForEachEntity calls the given function for each entity in the game.
// This provides a unified way to process all ships and enemies.
func (g *Game) ForEachEntity(fn func(Entity)) {
	for _, s := range g.Ships {
		fn(s)
	}
	for _, e := range g.Enemies {
		fn(e)
	}
}

// FindNearestEntity finds the nearest entity to a given world position.
// Returns nil if no entities exist.
func (g *Game) FindNearestEntity(x, y float64) Entity {
	var nearest Entity
	nearestDistSq := float64(1<<31 - 1) // Max float

	g.ForEachEntity(func(e Entity) {
		ex, ey := e.GetPosition()
		dx := x - ex
		dy := y - ey
		distSq := dx*dx + dy*dy
		if distSq < nearestDistSq {
			nearestDistSq = distSq
			nearest = e
		}
	})

	return nearest
}

// JoinMultiplayer joins a multiplayer room or creates one if it doesn't exist.
// The first player to join becomes the host.
func (g *Game) JoinMultiplayer(roomID string) {
	if g.Network != nil {
		g.Network.Disconnect()
	}
	g.Network = NewNetworkManager(g)
	g.Network.JoinRoom(roomID)
	g.Ship.NetworkID = g.Network.GetPlayerID()
}

// LeaveMultiplayer disconnects from the current multiplayer room.
func (g *Game) LeaveMultiplayer() {
	if g.Network != nil {
		g.Network.Disconnect()
		g.Network = nil
	}
}

// ProcessInput handles player input.
func (g *Game) ProcessInput() {
	// Determine if we're a network client (not host)
	isNetworkClient := g.Network != nil && g.Network.IsConnected() && !g.Network.IsHost()

	// Fire weapon (X key) - only on host or single player
	// Clients send input to host, host handles firing
	if g.Keys[88] && !isNetworkClient {
		g.Ship.Fire(g)
	}

	g.Ship.Move(g, g.Keys)
}

// RenderBackground renders the scrolling background based on camera position.
// Creates an infinite scrolling effect by tiling the background pattern.
func (g *Game) RenderBackground() {
	g.Ctx.Call("save")
	bgWidth := g.Level.Background.Get("width").Float()

	// Calculate offset based on camera position for parallax effect
	offsetX := math.Mod(-g.Camera.X*0.5, bgWidth)
	offsetY := math.Mod(-g.Camera.Y*0.5, bgWidth)

	g.Ctx.Call("translate", offsetX, offsetY)
	g.Ctx.Set("fillStyle", g.Level.BackgroundPattern)
	g.Ctx.Call("fillRect", -bgWidth, -bgWidth, WIDTH+bgWidth*2, HEIGHT+bgWidth*2)
	g.Ctx.Call("restore")
}
