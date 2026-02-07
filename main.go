//go:build js
// +build js

package main

import (
	"github.com/gopherjs/gopherjs/js"
	"github.com/simukka/starship-sorades-13k/game"
)

func main() {
	// Get the canvas element
	doc := js.Global.Get("document")
	canvas := doc.Call("getElementById", "c")
	if canvas == nil || canvas == js.Undefined {
		panic("canvas element not found")
	}
	// Set canvas dimensions
	canvas.Set("width", game.WIDTH)
	canvas.Set("height", game.HEIGHT)

	// Get 2D context
	ctx := canvas.Call("getContext", "2d")

	// Create the game instance
	g := game.NewGame()
	g.Canvas = canvas
	g.Ctx = ctx

	// Initialize audio
	g.Audio.Init()

	// Load all sound effects using pure Go jsfxr implementation
	for i, sfxParams := range game.SfxData {
		dataURL := game.GenerateWavDataURL(sfxParams)
		g.Audio.LoadSound(i, dataURL)
	}

	// Initialize graphics (background, sprites, etc.)
	g.InitializeGraphics()

	// Setup input handlers
	g.SetupInputHandlers()

	// Render title screen
	g.RenderTitleScreen()

	// Keep program running
	select {}
}
