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

	g.InitializeAudio()
	g.SetupInputHandlers()
	g.RenderTitleScreen()

	select {}
}
