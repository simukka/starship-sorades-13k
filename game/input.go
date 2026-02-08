package game

import (
	"github.com/gopherjs/gopherjs/js"
)

// KeyMap maps alternative keys to canonical control codes.
var KeyMap = map[int]int{
	27: 80, // Esc => P
	32: 88, // Space => X
	48: 88, // 0 => X
	50: 40, // 2 => Down
	52: 37, // 4 => Left
	53: 40, // 5 => Down
	54: 39, // 6 => Right
	56: 38, // 8 => Up
	65: 37, // A => Left
	67: 88, // C => X
	68: 39, // D => Right
	73: 38, // I => Up
	74: 37, // J => Left
	75: 40, // K => Down
	76: 39, // L => Right
	83: 40, // S => Down
	87: 38, // W => Up
	89: 88, // Y => X
	90: 88, // Z => X
}

// TranslateKeyCode converts alternative key codes to canonical control codes.
func TranslateKeyCode(keyCode int) int {
	if mapped, ok := KeyMap[keyCode]; ok {
		return mapped
	}
	return keyCode
}

// SetupInputHandlers initializes keyboard event handlers.
func (g *Game) SetupInputHandlers() {
	// Keydown handler
	js.Global.Get("document").Call("addEventListener", "keydown",
		func(event *js.Object) {
			rawKeyCode := event.Get("keyCode").Int()
			keyCode := TranslateKeyCode(rawKeyCode)
			g.Keys[keyCode] = true

			// Debug UI toggle (F9 = 120)
			if rawKeyCode == 120 {
				g.DebugUI.Toggle()
				event.Call("preventDefault")
				return
			}

			// Stats overlay toggle (F10 = 121)
			if rawKeyCode == 121 {
				g.StatsOverlay.Toggle()
				event.Call("preventDefault")
				return
			}

			// Pause toggle (P = 80, also mapped from Esc = 27)
			if keyCode == 80 {
				g.Level.Paused = !g.Level.Paused
				event.Call("preventDefault")
				return
			}

			// Debug UI controls when visible
			if g.DebugUI.Visible {
				switch rawKeyCode {
				case 81: // Q - Previous enemy type
					g.DebugUI.PrevEnemy()
				case 69: // E - Next enemy type
					g.DebugUI.NextEnemy()
				case 87: // W - Previous field
					g.DebugUI.PrevField()
				case 83: // S - Next field
					g.DebugUI.NextField()
				case 65: // A - Decrease value
					g.DebugUI.AdjustValue(-1)
				case 68: // D - Increase value
					g.DebugUI.AdjustValue(1)
				}
				event.Call("preventDefault")
				return
			}

			// Prevent default for game keys
			if keyCode >= 37 && keyCode <= 40 || keyCode == 88 {
				event.Call("preventDefault")
			}

			// Fullscreen toggle on 'F' key (70)
			if keyCode == 70 {
				canvas := js.Global.Get("document").Call("getElementById", "c")
				if canvas.Get("requestFullscreen") != nil && canvas.Get("requestFullscreen") != js.Undefined {
					canvas.Call("requestFullscreen")
				} else if canvas.Get("webkitRequestFullscreen") != nil && canvas.Get("webkitRequestFullscreen") != js.Undefined {
					canvas.Call("webkitRequestFullscreen")
				} else if canvas.Get("mozRequestFullScreen") != nil && canvas.Get("mozRequestFullScreen") != js.Undefined {
					canvas.Call("mozRequestFullScreen")
				}
			}
		})

	// Keyup handler
	js.Global.Get("document").Call("addEventListener", "keyup",
		func(event *js.Object) {
			rawKeyCode := event.Get("keyCode").Int()
			keyCode := TranslateKeyCode(rawKeyCode)
			g.Keys[keyCode] = false
		})

	// Click handler for starting/pausing
	js.Global.Get("document").Call("addEventListener", "click",
		func(event *js.Object) {
			if g.Audio.AudioCtx != nil && g.Audio.AudioCtx.Get("state").String() == "suspended" {
				g.Audio.AudioCtx.Call("resume")
			}

			if g.Level.Paused {
				g.Start()
			} else {
				// TODO: pause the game
			}
		})
}
