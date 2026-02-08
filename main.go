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
	g := game.NewGame(canvas, ctx)

	// Expose multiplayer API to JavaScript
	js.Global.Set("StarshipMultiplayer", map[string]interface{}{
		"join": func(roomID string) {
			g.JoinMultiplayer(roomID)
		},
		"leave": func() {
			g.LeaveMultiplayer()
		},
		"isConnected": func() bool {
			return g.Network != nil && g.Network.IsConnected()
		},
		"isHost": func() bool {
			return g.Network != nil && g.Network.IsHost()
		},
		"getPlayerCount": func() int {
			if g.Network != nil {
				return g.Network.GetPlayerCount()
			}
			return 1
		},
	})

	// Clean up multiplayer connection when browser is closed
	js.Global.Call("addEventListener", "beforeunload", func() {
		g.LeaveMultiplayer()
	})

	select {}
}
