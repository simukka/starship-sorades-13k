package game

import (
	"math"
	"strconv"

	"github.com/gopherjs/gopherjs/js"
)

// Base represents a stationary base with a protective shield.
// Ships can enter the shield area, but enemies and torpedos cannot.
// Bases are permanent and never despawn from the infinite world.
type Base struct {
	X, Y         float64    // Fixed world coordinates
	Radius       float64    // Visual radius of the base structure
	ShieldRadius float64    // Radius of the protective shield
	ShieldPhase  float64    // Animation phase for shield effect
	ImpactTimer  int        // Frames remaining for impact vibration
	ImpactAngle  float64    // Angle of last impact for directional vibration
	Image        *js.Object // Base sprite (optional)
}

// NewBase creates a new base at the specified world coordinates.
func NewBase(x, y float64) *Base {
	return &Base{
		X:            x,
		Y:            y,
		Radius:       BaseRadius,
		ShieldRadius: BaseShieldRadius,
		ShieldPhase:  0,
	}
}

// GetPosition implements Collidable interface.
func (b *Base) GetPosition() (x, y float64) {
	return b.X, b.Y
}

// GetRadius returns the shield radius for collision purposes.
func (b *Base) GetRadius() float64 {
	return b.ShieldRadius
}

// ContainsPoint checks if a world coordinate is inside the shield.
func (b *Base) ContainsPoint(x, y float64) bool {
	dx := x - b.X
	dy := y - b.Y
	return dx*dx+dy*dy <= b.ShieldRadius*b.ShieldRadius
}

// IsShipInside checks if a ship is inside the shield.
func (b *Base) IsShipInside(s *Ship) bool {
	return b.ContainsPoint(s.X, s.Y)
}

// Render draws the base and its shield.
func (b *Base) Render(g *Game) {
	// Update shield animation phase
	b.ShieldPhase += 0.02

	// Convert world position to screen position
	screenX, screenY := g.Camera.WorldToScreen(b.X, b.Y)

	// Only render if on screen (with shield radius padding)
	if !g.Camera.IsOnScreen(b.X, b.Y, b.ShieldRadius+50) {
		return
	}

	// Apply vibration offset if shield was recently hit
	var vibrateX, vibrateY float64
	if b.ImpactTimer > 0 {
		// Vibration intensity decreases over time
		intensity := float64(b.ImpactTimer) * 0.8
		// High frequency oscillation for vibration effect
		vibrateX = math.Sin(float64(b.ImpactTimer)*2.5) * intensity * math.Cos(b.ImpactAngle)
		vibrateY = math.Sin(float64(b.ImpactTimer)*2.5) * intensity * math.Sin(b.ImpactAngle)
		b.ImpactTimer--
	}

	g.Ctx.Call("save")

	// Draw shield circle with glow effect
	g.Ctx.Set("shadowBlur", Theme.ShieldShadowBlur)
	g.Ctx.Set("shadowColor", Theme.BaseShieldGlowColor)

	// Shield fill (semi-transparent) - with vibration offset
	g.Ctx.Set("fillStyle", Theme.BaseShieldColor)
	g.Ctx.Call("beginPath")
	g.Ctx.Call("arc", screenX+vibrateX, screenY+vibrateY, b.ShieldRadius, 0, math.Pi*2)
	g.Ctx.Call("fill")

	// Shield border with pulsing effect - amplify pulse during impact
	pulseAlpha := 0.4 + 0.3*math.Sin(b.ShieldPhase)
	if b.ImpactTimer > 0 {
		pulseAlpha = 0.7 + 0.3*math.Sin(float64(b.ImpactTimer)*0.5)
	}
	g.Ctx.Set("globalAlpha", pulseAlpha)
	g.Ctx.Set("strokeStyle", Theme.BaseShieldGlowColor)
	g.Ctx.Set("lineWidth", 2.0)
	g.Ctx.Call("stroke")

	// Inner shield ring - with vibration
	g.Ctx.Set("globalAlpha", pulseAlpha*0.5)
	g.Ctx.Call("beginPath")
	g.Ctx.Call("arc", screenX+vibrateX*0.5, screenY+vibrateY*0.5, b.ShieldRadius*0.95, 0, math.Pi*2)
	g.Ctx.Call("stroke")

	g.Ctx.Set("globalAlpha", 1)
	g.Ctx.Set("shadowBlur", 0)

	// Draw base structure (simple hexagon shape)
	// g.Ctx.Set("fillStyle", Theme.BaseColor)
	// g.Ctx.Set("strokeStyle", Theme.BaseShieldGlowColor)
	// g.Ctx.Set("lineWidth", 2.0)

	// g.Ctx.Call("beginPath")
	// for i := 0; i < 6; i++ {
	// 	angle := float64(i)*math.Pi/3 - math.Pi/6
	// 	px := screenX + math.Cos(angle)*b.Radius
	// 	py := screenY + math.Sin(angle)*b.Radius
	// 	if i == 0 {
	// 		g.Ctx.Call("moveTo", px, py)
	// 	} else {
	// 		g.Ctx.Call("lineTo", px, py)
	// 	}
	// }
	// g.Ctx.Call("closePath")
	// g.Ctx.Call("fill")
	// g.Ctx.Call("stroke")

	// Draw center glow
	g.Ctx.Set("shadowBlur", 12)
	g.Ctx.Set("shadowColor", Theme.BaseShieldGlowColor)
	g.Ctx.Set("fillStyle", Theme.BaseShieldGlowColor)
	g.Ctx.Call("beginPath")
	g.Ctx.Call("arc", screenX, screenY, b.Radius*0.3, 0, math.Pi*2)
	g.Ctx.Call("fill")

	g.Ctx.Call("restore")
}

// BlocksTorpedo checks if a torpedo should be blocked by this base's shield.
// Returns true if the torpedo is trying to enter the shield from outside.
func (b *Base) BlocksBulletEntrance(torpedo *Bullet) bool {
	// Check if torpedo is near the shield boundary
	dx := torpedo.X - b.X
	dy := torpedo.Y - b.Y
	distSq := dx*dx + dy*dy

	// If inside shield, don't block (already inside)
	if distSq < b.ShieldRadius*b.ShieldRadius {
		return false
	}

	// Check if torpedo is heading toward the shield
	nextX := torpedo.X + torpedo.XAcc
	nextY := torpedo.Y + torpedo.YAcc
	nextDx := nextX - b.X
	nextDy := nextY - b.Y
	nextDistSq := nextDx*nextDx + nextDy*nextDy

	// Block if would enter the shield
	if nextDistSq < b.ShieldRadius*b.ShieldRadius {
		// Trigger impact vibration effect
		b.TriggerImpact(torpedo.X, torpedo.Y)
		return true
	}
	return false
}

// TriggerImpact starts the shield vibration effect from an impact at the given position.
func (b *Base) TriggerImpact(impactX, impactY float64) {
	// Calculate angle from base center to impact point
	b.ImpactAngle = math.Atan2(impactY-b.Y, impactX-b.X)
	// Set vibration duration (frames)
	b.ImpactTimer = 15
}

// BlocksBulletExit checks if a bullet fired from inside should be blocked.
// Returns true if the bullet is trying to leave the shield from inside.
func (b *Base) BlocksBulletExit(bullet *Bullet) bool {
	dx := bullet.X - b.X
	dy := bullet.Y - b.Y
	distSq := dx*dx + dy*dy

	// Block bullets trying to exit the shield
	return distSq > b.ShieldRadius*b.ShieldRadius
}

// IsPointInAnyBaseShield checks if a world position is inside any base's shield.
func (g *Game) IsPointInAnyBaseShield(x, y float64) bool {
	for _, base := range g.Bases {
		if base.ContainsPoint(x, y) {
			return true
		}
	}
	return false
}

// IsShipProtectedByBase checks if a ship is inside any base's shield.
func (g *Game) IsShipProtectedByBase(s *Ship) bool {
	for _, base := range g.Bases {
		if base.IsShipInside(s) {
			return true
		}
	}
	return false
}

// UpdateShieldAudioFilter updates the audio filter based on whether the ship is inside a base shield.
// When inside, external sounds are muffled through a low-pass filter.
func (g *Game) UpdateShieldAudioFilter() {
	inside := g.IsShipProtectedByBase(g.Ship)
	g.Ship.InBase = inside
	g.Audio.SetShieldMode(inside)
}

// GetBaseAtPoint returns the base whose shield contains the point, or nil.
func (g *Game) GetBaseAtPoint(x, y float64) *Base {
	for _, base := range g.Bases {
		if base.ContainsPoint(x, y) {
			return base
		}
	}
	return nil
}

// FindNearestBase finds the nearest base to the camera position.
func (g *Game) FindNearestBase() *Base {
	if len(g.Bases) == 0 {
		return nil
	}

	var nearest *Base
	nearestDistSq := math.MaxFloat64

	for _, base := range g.Bases {
		dx := base.X - g.Camera.X
		dy := base.Y - g.Camera.Y
		distSq := dx*dx + dy*dy
		if distSq < nearestDistSq {
			nearestDistSq = distSq
			nearest = base
		}
	}

	return nearest
}

// RenderBaseIndicators renders arrow indicators pointing to off-screen bases.
func (g *Game) RenderBaseIndicators() {
	base := g.FindNearestBase()
	if base == nil {
		return
	}

	// Convert base position to screen coordinates
	screenX, screenY := g.Camera.WorldToScreen(base.X, base.Y)

	// Check if base is already on screen (with some margin)
	margin := base.ShieldRadius
	if screenX >= -margin && screenX <= WIDTH+margin &&
		screenY >= -margin && screenY <= HEIGHT+margin {
		return // Base is visible, no indicator needed
	}

	// Calculate direction from screen center to base
	centerX := float64(WIDTH) / 2
	centerY := float64(HEIGHT) / 2
	dx := screenX - centerX
	dy := screenY - centerY
	angle := math.Atan2(dy, dx)

	// Calculate position on screen edge
	// Use parametric line intersection with screen bounds
	edgePadding := 40.0
	var indicatorX, indicatorY float64

	// Find intersection with screen edges
	// Check right edge
	if dx > 0 {
		t := (float64(WIDTH) - edgePadding - centerX) / dx
		y := centerY + t*dy
		if y >= edgePadding && y <= float64(HEIGHT)-edgePadding {
			indicatorX = float64(WIDTH) - edgePadding
			indicatorY = y
		}
	}
	// Check left edge
	if dx < 0 {
		t := (edgePadding - centerX) / dx
		y := centerY + t*dy
		if y >= edgePadding && y <= float64(HEIGHT)-edgePadding {
			indicatorX = edgePadding
			indicatorY = y
		}
	}
	// Check bottom edge
	if dy > 0 && indicatorX == 0 {
		t := (float64(HEIGHT) - edgePadding - centerY) / dy
		x := centerX + t*dx
		if x >= edgePadding && x <= float64(WIDTH)-edgePadding {
			indicatorX = x
			indicatorY = float64(HEIGHT) - edgePadding
		}
	}
	// Check top edge
	if dy < 0 && indicatorX == 0 {
		t := (edgePadding - centerY) / dy
		x := centerX + t*dx
		if x >= edgePadding && x <= float64(WIDTH)-edgePadding {
			indicatorX = x
			indicatorY = edgePadding
		}
	}

	// Fallback for corner cases
	if indicatorX == 0 && indicatorY == 0 {
		indicatorX = centerX + math.Cos(angle)*(centerX-edgePadding)
		indicatorY = centerY + math.Sin(angle)*(centerY-edgePadding)
		// Clamp to screen
		if indicatorX < edgePadding {
			indicatorX = edgePadding
		}
		if indicatorX > float64(WIDTH)-edgePadding {
			indicatorX = float64(WIDTH) - edgePadding
		}
		if indicatorY < edgePadding {
			indicatorY = edgePadding
		}
		if indicatorY > float64(HEIGHT)-edgePadding {
			indicatorY = float64(HEIGHT) - edgePadding
		}
	}

	// Calculate distance for display
	worldDx := base.X - g.Camera.X
	worldDy := base.Y - g.Camera.Y
	distance := math.Sqrt(worldDx*worldDx + worldDy*worldDy)

	// Draw the indicator arrow
	g.Ctx.Call("save")

	// Pulsing effect
	pulse := 0.7 + 0.3*math.Sin(base.ShieldPhase*2)
	g.Ctx.Set("globalAlpha", pulse)

	// Draw arrow pointing toward base
	g.Ctx.Call("translate", indicatorX, indicatorY)
	g.Ctx.Call("rotate", angle)

	// Arrow shape
	arrowSize := 16.0
	g.Ctx.Set("fillStyle", Theme.BaseShieldGlowColor)
	g.Ctx.Set("shadowBlur", 8)
	g.Ctx.Set("shadowColor", Theme.BaseShieldGlowColor)

	g.Ctx.Call("beginPath")
	g.Ctx.Call("moveTo", arrowSize, 0)               // Tip
	g.Ctx.Call("lineTo", -arrowSize/2, -arrowSize/2) // Top back
	g.Ctx.Call("lineTo", -arrowSize/4, 0)            // Notch
	g.Ctx.Call("lineTo", -arrowSize/2, arrowSize/2)  // Bottom back
	g.Ctx.Call("closePath")
	g.Ctx.Call("fill")

	g.Ctx.Call("restore")

	// Draw distance text
	g.Ctx.Call("save")
	g.Ctx.Set("globalAlpha", pulse)
	g.Ctx.Set("fillStyle", Theme.BaseShieldGlowColor)
	g.Ctx.Set("font", "12px "+Theme.ScoreFont)
	g.Ctx.Set("textAlign", "center")

	// Position text slightly offset from arrow
	textOffsetX := -math.Cos(angle) * 25
	textOffsetY := -math.Sin(angle) * 25
	distText := formatDistance(distance)
	g.Ctx.Call("fillText", distText, indicatorX+textOffsetX, indicatorY+textOffsetY+4)

	g.Ctx.Call("restore")
}

// formatDistance formats a distance value for display.
func formatDistance(dist float64) string {
	if dist >= 1000 {
		return strconv.FormatFloat(dist/1000, 'f', 1, 64) + "k"
	}
	return strconv.Itoa(int(dist))
}
