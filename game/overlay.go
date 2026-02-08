package game

import (
	"math"
	"strconv"

	"github.com/gopherjs/gopherjs/js"
)

// StatsOverlay displays real-time game statistics
type StatsOverlay struct {
	Visible bool

	// FPS tracking
	FrameCount    int
	LastFPSUpdate float64
	CurrentFPS    float64

	// Position and styling
	PanelX      int
	PanelY      int
	LineHeight  int
	PanelWidth  int
	PanelHeight int
}

// NewStatsOverlay creates a new stats overlay instance
func NewStatsOverlay() *StatsOverlay {
	return &StatsOverlay{
		Visible:       false,
		PanelX:        WIDTH - 280,
		PanelY:        16,
		LineHeight:    18,
		PanelWidth:    264,
		PanelHeight:   280,
		LastFPSUpdate: 0,
		CurrentFPS:    0,
	}
}

// Toggle toggles the stats overlay visibility
func (s *StatsOverlay) Toggle() {
	s.Visible = !s.Visible
}

// UpdateFPS updates the FPS counter
func (s *StatsOverlay) UpdateFPS(currentTime float64) {
	s.FrameCount++

	// Update FPS every second
	elapsed := currentTime - s.LastFPSUpdate
	if elapsed >= 1000 {
		s.CurrentFPS = float64(s.FrameCount) / (elapsed / 1000)
		s.FrameCount = 0
		s.LastFPSUpdate = currentTime
	}
}

// Render draws the stats overlay
func (s *StatsOverlay) Render(ctx *js.Object, g *Game) {
	if !s.Visible {
		return
	}

	// Render enemy debug overlays first (in world space)
	s.RenderEnemyDebug(ctx, g)

	// Render player hitbox
	s.RenderPlayerDebug(ctx, g)

	// Draw stats panel background
	ctx.Set("fillStyle", "rgba(0, 0, 0, 0.75)")
	ctx.Call("fillRect", s.PanelX, s.PanelY, s.PanelWidth, s.PanelHeight)

	// Draw panel border
	ctx.Set("strokeStyle", "#00aaff")
	ctx.Set("lineWidth", 1)
	ctx.Call("strokeRect", s.PanelX, s.PanelY, s.PanelWidth, s.PanelHeight)

	// Title
	ctx.Set("fillStyle", "#00aaff")
	ctx.Set("font", "bold 14px monospace")
	ctx.Set("textAlign", "left")
	ctx.Call("fillText", "GAME STATS [F10]", s.PanelX+10, s.PanelY+20)

	// Separator
	ctx.Set("strokeStyle", "#444444")
	ctx.Call("beginPath")
	ctx.Call("moveTo", s.PanelX+10, s.PanelY+28)
	ctx.Call("lineTo", s.PanelX+s.PanelWidth-10, s.PanelY+28)
	ctx.Call("stroke")

	// Stats content
	ctx.Set("font", "12px monospace")
	y := s.PanelY + 48

	// Performance stats
	s.drawStatLine(ctx, "FPS", strconv.FormatFloat(s.CurrentFPS, 'f', 1, 64), "#00ff00", y)
	y += s.LineHeight

	// Separator - Performance
	y += 5
	ctx.Set("fillStyle", "#666666")
	ctx.Call("fillText", "── Game State ──", s.PanelX+10, y)
	y += s.LineHeight

	// Game state
	s.drawStatLine(ctx, "Level", strconv.Itoa(g.Level.LevelNum+1), "#ffffff", y)
	y += s.LineHeight
	s.drawStatLine(ctx, "Score", strconv.Itoa(g.Level.P), "#ffff00", y)
	y += s.LineHeight
	s.drawStatLine(ctx, "Game Seed", strconv.FormatUint(uint64(g.GameSeed), 10), "#aaaaaa", y)
	y += s.LineHeight
	s.drawStatLine(ctx, "Level Seed", strconv.FormatUint(uint64(g.Level.LevelSeed), 10), "#aaaaaa", y)
	y += s.LineHeight

	// Separator - Objects
	y += 5
	ctx.Set("fillStyle", "#666666")
	ctx.Call("fillText", "── Object Pools ──", s.PanelX+10, y)
	y += s.LineHeight

	// Object counts
	bulletCount := strconv.Itoa(g.Bullets.ActiveCount) + "/" + strconv.Itoa(g.Bullets.MaxSize)
	s.drawStatLine(ctx, "Bullets", bulletCount, "#ff8800", y)
	y += s.LineHeight

	explosionCount := strconv.Itoa(g.Explosions.ActiveCount) + "/" + strconv.Itoa(g.Explosions.MaxSize)
	s.drawStatLine(ctx, "Explosions", explosionCount, "#ff4400", y)
	y += s.LineHeight

	bonusCount := strconv.Itoa(g.Bonuses.ActiveCount) + "/" + strconv.Itoa(g.Bonuses.MaxSize)
	s.drawStatLine(ctx, "Bonuses", bonusCount, "#44ff44", y)
	y += s.LineHeight

	s.drawStatLine(ctx, "Enemies", strconv.Itoa(len(g.Enemies)), "#ff0066", y)
	y += s.LineHeight

	// Separator - Player
	y += 5
	ctx.Set("fillStyle", "#666666")
	ctx.Call("fillText", "── Player ──", s.PanelX+10, y)
	y += s.LineHeight

	// Player stats
	s.drawStatLine(ctx, "Health", strconv.Itoa(g.Ship.E)+"%", s.healthColor(g.Ship.E), y)
	y += s.LineHeight
	s.drawStatLine(ctx, "Weapons", strconv.Itoa(len(g.Ship.Weapons)), "#8888ff", y)
	y += s.LineHeight
	s.drawStatLine(ctx, "Shield", strconv.Itoa(g.Ship.Shield.T), "#00ffff", y)
	y += s.LineHeight
	s.drawStatLine(ctx, "Position", strconv.FormatFloat(g.Ship.X, 'f', 0, 64)+", "+strconv.FormatFloat(g.Ship.Y, 'f', 0, 64), "#aaaaaa", y)
}

// RenderEnemyDebug draws debug information for all enemies
func (s *StatsOverlay) RenderEnemyDebug(ctx *js.Object, g *Game) {
	for _, enemy := range g.Enemies {
		enemyY := enemy.Y + enemy.YOffset
		hitboxD := enemy.Radius * 0.6

		// Draw hitbox (square bounding box)
		ctx.Set("strokeStyle", "#ff0066")
		ctx.Set("lineWidth", 2)
		ctx.Call("strokeRect",
			enemy.X-hitboxD, enemyY-hitboxD,
			hitboxD*2, hitboxD*2)

		// Draw radius circle (full enemy radius for reference)
		ctx.Set("strokeStyle", "rgba(255, 0, 102, 0.3)")
		ctx.Set("lineWidth", 1)
		ctx.Call("beginPath")
		ctx.Call("arc", enemy.X, enemyY, enemy.Radius, 0, math.Pi*2)
		ctx.Call("stroke")

		// Draw angle indicator line
		ctx.Set("strokeStyle", "#ffff00")
		ctx.Set("lineWidth", 2)
		lineLength := enemy.Radius * 1.5
		endX := enemy.X + math.Sin(enemy.TargetAngle())*lineLength
		endY := enemyY + math.Cos(enemy.TargetAngle())*lineLength
		ctx.Call("beginPath")
		ctx.Call("moveTo", enemy.X, enemyY)
		ctx.Call("lineTo", endX, endY)
		ctx.Call("stroke")

		// Draw health bar above enemy
		barWidth := 40.0
		barHeight := 4.0
		barX := enemy.X - barWidth/2
		barY := enemyY - enemy.Radius - 20

		// Background
		ctx.Set("fillStyle", "rgba(0, 0, 0, 0.7)")
		ctx.Call("fillRect", barX-1, barY-1, barWidth+2, barHeight+2)

		// Health fill (estimate max health for bar - use config)
		config := enemyConfigs[enemy.Kind]
		maxHealth := config.HealthBase + config.HealthPerLevel*g.Level.LevelNum
		if maxHealth <= 0 {
			maxHealth = 1
		}
		healthPercent := float64(enemy.Health) / float64(maxHealth)
		if healthPercent > 1 {
			healthPercent = 1
		}

		ctx.Set("fillStyle", s.healthColor(int(healthPercent*100)))
		ctx.Call("fillRect", barX, barY, barWidth*healthPercent, barHeight)

		// Border
		ctx.Set("strokeStyle", "#ffffff")
		ctx.Set("lineWidth", 1)
		ctx.Call("strokeRect", barX, barY, barWidth, barHeight)

		// Draw text info
		ctx.Set("fillStyle", "#ffffff")
		ctx.Set("font", "10px monospace")
		ctx.Set("textAlign", "center")

		// Health value
		ctx.Call("fillText", "HP:"+strconv.Itoa(enemy.Health), enemy.X, barY-3)

		// Position and angle below enemy
		ctx.Set("fillStyle", "#aaaaaa")
		ctx.Set("font", "9px monospace")
		posText := "(" + strconv.FormatFloat(enemy.X, 'f', 0, 64) + "," + strconv.FormatFloat(enemyY, 'f', 0, 64) + ")"
		ctx.Call("fillText", posText, enemy.X, enemyY+enemy.Radius+12)

		// Angle in degrees
		angleDeg := enemy.Angle * 180 / math.Pi
		angleText := "∠" + strconv.FormatFloat(angleDeg, 'f', 1, 64) + "°"
		ctx.Call("fillText", angleText, enemy.X, enemyY+enemy.Radius+22)

		// Enemy kind/type
		ctx.Set("fillStyle", "#ff0066")
		kindName := EnemyKindNames[enemy.Kind]
		ctx.Call("fillText", kindName, enemy.X, enemyY+enemy.Radius+32)
	}

	// Reset text alignment
	ctx.Set("textAlign", "left")
}

// RenderPlayerDebug draws debug information for the player ship
func (s *StatsOverlay) RenderPlayerDebug(ctx *js.Object, g *Game) {
	ship := g.Ship

	// Draw player hitbox (AABB)
	ctx.Set("strokeStyle", "#00ff00")
	ctx.Set("lineWidth", 2)
	ctx.Call("strokeRect",
		ship.X-ShipCollisionE, ship.Y-ShipCollisionD,
		ShipCollisionE*2, ShipCollisionD*2)

	// Draw ship radius circle for reference
	ctx.Set("strokeStyle", "rgba(0, 255, 0, 0.3)")
	ctx.Set("lineWidth", 1)
	ctx.Call("beginPath")
	ctx.Call("arc", ship.X, ship.Y, ShipR, 0, math.Pi*2)
	ctx.Call("stroke")
}

// drawStatLine draws a single stat line with label and value
func (s *StatsOverlay) drawStatLine(ctx *js.Object, label, value, valueColor string, y int) {
	ctx.Set("fillStyle", "#cccccc")
	ctx.Call("fillText", label+":", s.PanelX+15, y)

	ctx.Set("fillStyle", valueColor)
	ctx.Set("textAlign", "right")
	ctx.Call("fillText", value, s.PanelX+s.PanelWidth-15, y)
	ctx.Set("textAlign", "left")
}

// healthColor returns a color based on health percentage
func (s *StatsOverlay) healthColor(health int) string {
	if health > 75 {
		return "#00ff00"
	} else if health > 50 {
		return "#88ff00"
	} else if health > 25 {
		return "#ffff00"
	} else {
		return "#ff0000"
	}
}
