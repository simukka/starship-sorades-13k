SORADES 13K
===========

A scrolling shooter in the vein of "Raptor: Call of the Shadows" and
"Warning Forever". Are you able to survive 13 waves? Press X to fire,
F for fullscreen, P for pause, M to switch to the original ship model from
[our full-featured "Starship Sorades" game](http://www.gramambo.de/index.php?option=com_content&view=article&id=17%3Astarship-sorades&catid=2&Itemid=5)
(yes, I had 3 KB left and wasted them for an easter egg).
Multiple control schemes provided, e.g. cursor keys, WASD and numpad.

I use a lot of pre-rendering and other tricks to make this run as fast as
possible on a wide variety of browsers and machines. Developed with Opera.
Tested with Firefox, Chrome and even Internet Explorer 9 (no sound though).
Fullscreen works in Opera 12.50+, Firefox and WebKit based browsers.

## GopherJS Port

This version has been refactored to use [GopherJS](https://github.com/nicois/gopherjs) (Go compiled to JavaScript).

### Features

- **Seeded PRNG**: Deterministic level generation using Mulberry32 algorithm
- **Object Pooling**: Pre-allocated pools with swap-and-pop for efficient memory management
- **Web Audio API**: Low-latency sound effects using AudioContext
- **Canvas 2D Rendering**: Efficient sprite rendering with createPattern for background
- **requestAnimationFrame**: Smooth 30 FPS game loop

### Project Structure

```
.
├── main.go              # Entry point
├── game/
│   ├── audio.go         # Web Audio API sound manager
│   ├── enemies.go       # Enemy spawning and behavior
│   ├── graphics.go      # Canvas rendering and sprite generation
│   ├── input.go         # Keyboard input handling
│   ├── loop.go          # Main game loop
│   ├── pool.go          # Object pools (bullets, torpedos, etc.)
│   ├── rng.go           # Seeded random number generator
│   └── state.go         # Game state and core logic
├── jsfxr.js             # Sound synthesis library
├── index.html           # HTML entry point
├── Makefile             # Build commands
└── go.mod               # Go module definition
```

### Building

#### Prerequisites

- Go 1.21 or later
- GopherJS: `go install github.com/nicois/gopherjs@latest`

#### Build Commands

```bash
# Build the game
make build

# Build minified for production
make build-min

# Start local development server
make serve

# Clean build artifacts
make clean
```

#### Manual Build

```bash
gopherjs build -o game.js .
```

## Original Credits

SORADES 13K is a [Gramambo game](http://www.gramambo.de/index.php?option=com_content&view=article&id=4&Itemid=5#Spiele)
by [Thiemo M&auml;ttig](http://maettig.com/)
([@maettig](https://twitter.com/maettig) at Twitter).

Sound effects by [Sven Gramatke](http://www.sven-gramatke.de/).

Licensed under a Creative [Commons Attribution-ShareAlike 3.0 Germany License](http://creativecommons.org/licenses/by-sa/3.0/de/)
(CC-BY-SA-3.0-DE).

Contains a modified version of [jsfxr](https://github.com/mneubrand/jsfxr)
by [Markus Neubrand](https://twitter.com/markusneubrand),
based on [as3sfxr](http://code.google.com/p/as3sfxr/)
(licensed under [Apache License 2.0](http://www.apache.org/licenses/LICENSE-2.0))
by [Thomas Vian](http://tomvian.com/),
based on [sfxr](http://www.drpetter.se/project_sfxr.html)
(licensed under [MIT License](http://www.opensource.org/licenses/mit-license.php))
by [Tomas Pettersson](http://www.drpetter.se/).