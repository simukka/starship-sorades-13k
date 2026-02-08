package game

import (
	"math"
	"strconv"

	"github.com/gopherjs/gopherjs/js"
)

// ShipHUD displays ship velocity, angle, and position
type ShipHUD struct {
	Visible    bool
	PanelX     int
	PanelY     int
	LineHeight int
}

// NewShipHUD creates a new ship HUD instance
func NewShipHUD() *ShipHUD {
	return &ShipHUD{
		Visible:    true,
		PanelX:     10,
		PanelY:     10,
		LineHeight: 16,
	}
}

// Toggle toggles the ship HUD visibility
func (h *ShipHUD) Toggle() {
	h.Visible = !h.Visible
}

// Render draws the ship HUD overlay
func (h *ShipHUD) Render(ctx *js.Object, ship *Ship) {
	if !h.Visible || ship == nil {
		return
	}

	y := h.PanelY + 16

	// Calculate total velocity
	velocity := math.Sqrt(ship.VelX*ship.VelX + ship.VelY*ship.VelY)

	// Convert angle to degrees (0-360)
	angleDeg := math.Mod(ship.Angle*180/math.Pi+360, 360)

	// Set text style
	ctx.Set("font", "bold 12px monospace")
	ctx.Set("textAlign", "left")
	ctx.Set("shadowBlur", 4)
	ctx.Set("shadowColor", "#000000")

	// Position
	ctx.Set("fillStyle", "#00ffff")
	ctx.Call("fillText", "POS: "+strconv.FormatFloat(ship.X, 'f', 1, 64)+", "+strconv.FormatFloat(ship.Y, 'f', 1, 64), h.PanelX, y)
	y += h.LineHeight

	// Velocity
	ctx.Set("fillStyle", "#ffff00")
	ctx.Call("fillText", "VEL: "+strconv.FormatFloat(velocity, 'f', 2, 64)+" ("+strconv.FormatFloat(ship.VelX, 'f', 1, 64)+", "+strconv.FormatFloat(ship.VelY, 'f', 1, 64)+")", h.PanelX, y)
	y += h.LineHeight

	// Angle
	ctx.Set("fillStyle", "#ff88ff")
	ctx.Call("fillText", "ANG: "+strconv.FormatFloat(angleDeg, 'f', 1, 64)+"°", h.PanelX, y)
	y += h.LineHeight

	// Targeting status
	if ship.Target != nil {
		ctx.Set("fillStyle", "#ff0000")
		ctx.Call("fillText", "TGT: LOCKED", h.PanelX, y)
	} else if ship.LockingOn != nil {
		// Show lock progress
		progress := 100 - (ship.LockTimer * 100 / ship.LockMaxTime)
		ctx.Set("fillStyle", "#ff8800")
		ctx.Call("fillText", "TGT: LOCKING "+strconv.Itoa(progress)+"%", h.PanelX, y)
	} else {
		ctx.Set("fillStyle", "#888888")
		ctx.Call("fillText", "TGT: NONE [T]", h.PanelX, y)
	}

	// Reset shadow
	ctx.Set("shadowBlur", 0)

	// Render weapon indicator UI
	h.RenderWeaponIndicator(ctx, ship)
}

// RenderWeaponIndicator draws a circular UI showing weapon placements and aim angles
func (h *ShipHUD) RenderWeaponIndicator(ctx *js.Object, ship *Ship) {
	if len(ship.Weapons) == 0 {
		return
	}

	// Position the weapon indicator in bottom-left corner
	centerX := float64(h.PanelX + 60)
	centerY := float64(HEIGHT - 80)
	radius := 50.0
	weaponDotRadius := 6.0

	ctx.Call("save")

	// Draw outer ring (ship representation)
	ctx.Set("strokeStyle", "#444444")
	ctx.Set("lineWidth", 2)
	ctx.Call("beginPath")
	ctx.Call("arc", centerX, centerY, radius, 0, math.Pi*2)
	ctx.Call("stroke")

	// Draw ship direction indicator (forward arrow)
	ctx.Set("strokeStyle", "#666666")
	ctx.Set("lineWidth", 1)
	ctx.Call("beginPath")
	ctx.Call("moveTo", centerX, centerY-radius-5)
	ctx.Call("lineTo", centerX-5, centerY-radius+5)
	ctx.Call("moveTo", centerX, centerY-radius-5)
	ctx.Call("lineTo", centerX+5, centerY-radius+5)
	ctx.Call("stroke")

	// Calculate angle to target if we have one
	var targetAngle float64
	hasTarget := ship.Target != nil && ship.Target.IsAlive()
	if hasTarget {
		dx := ship.Target.X - ship.X
		dy := (ship.Target.Y + ship.Target.YOffset) - ship.Y
		targetAngle = math.Atan2(dx, -dy) - ship.Angle // Relative to ship facing
	}

	// Draw each weapon
	for _, w := range ship.Weapons {
		// Calculate weapon's base angle from its X,Y direction
		weaponAngle := math.Atan2(w.X, -w.Y)

		// Position on the ring
		wx := centerX + math.Sin(weaponAngle)*radius*0.8
		wy := centerY - math.Cos(weaponAngle)*radius*0.8

		// Draw weapon dot
		if hasTarget {
			ctx.Set("fillStyle", "#FF5B24") // Vipps orange when targeting
		} else {
			ctx.Set("fillStyle", "#888888") // Gray when no target
		}
		ctx.Call("beginPath")
		ctx.Call("arc", wx, wy, weaponDotRadius, 0, math.Pi*2)
		ctx.Call("fill")

		// Draw aim line toward target
		if hasTarget {
			// Line from weapon to edge of indicator showing aim direction
			aimLength := radius * 0.6
			aimX := wx + math.Sin(targetAngle)*aimLength
			aimY := wy - math.Cos(targetAngle)*aimLength

			ctx.Set("strokeStyle", "#FF5B24")
			ctx.Set("lineWidth", 1.5)
			ctx.Set("globalAlpha", 0.7)
			ctx.Call("beginPath")
			ctx.Call("moveTo", wx, wy)
			ctx.Call("lineTo", aimX, aimY)
			ctx.Call("stroke")
			ctx.Set("globalAlpha", 1)
		}
	}

	// Draw target indicator on the ring edge if targeting
	if hasTarget {
		targetX := centerX + math.Sin(targetAngle)*radius
		targetY := centerY - math.Cos(targetAngle)*radius

		ctx.Set("fillStyle", "#ff0000")
		ctx.Set("shadowBlur", 6)
		ctx.Set("shadowColor", "#ff0000")
		ctx.Call("beginPath")
		ctx.Call("arc", targetX, targetY, 5, 0, math.Pi*2)
		ctx.Call("fill")
		ctx.Set("shadowBlur", 0)

		// Draw crosshair at target position
		ctx.Set("strokeStyle", "#ff0000")
		ctx.Set("lineWidth", 1)
		ctx.Call("beginPath")
		ctx.Call("moveTo", targetX-8, targetY)
		ctx.Call("lineTo", targetX+8, targetY)
		ctx.Call("moveTo", targetX, targetY-8)
		ctx.Call("lineTo", targetX, targetY+8)
		ctx.Call("stroke")
	}

	// Draw center dot (ship position)
	ctx.Set("fillStyle", "#FF5B24")
	ctx.Call("beginPath")
	ctx.Call("arc", centerX, centerY, 4, 0, math.Pi*2)
	ctx.Call("fill")

	ctx.Call("restore")
}

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

	s.drawStatLine(ctx, "Game Seed", strconv.FormatUint(uint64(g.GameSeed), 10), "#aaaaaa", y)
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
