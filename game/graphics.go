package game

import (
	"math"
	"strconv"

	"github.com/gopherjs/gopherjs/js"
)

// maxInt returns the maximum of two integers.
func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// RenderToCanvas creates an off-screen canvas and renders to it.
func RenderToCanvas(width, height int, renderFn func(canvas, ctx *js.Object)) *js.Object {
	document := js.Global.Get("document")
	canvas := document.Call("createElement", "canvas")
	canvas.Set("width", width)
	canvas.Set("height", height)
	ctx := canvas.Call("getContext", "2d")
	renderFn(canvas, ctx)
	return canvas
}

// InitializeGraphics renders all static game graphics.
func (g *Game) InitializeGraphics() {
	// Score digit sprites (0-9)
	for i := 0; i < 10; i++ {
		num := i // capture for closure
		g.Level.Points.Images[i] = RenderToCanvas(g.Level.Points.Width, g.Level.Points.Height,
			func(canvas, ctx *js.Object) {
				ctx.Set("shadowBlur", Theme.DefaultShadowBlur)
				ctx.Set("font", "bold "+strconv.Itoa(canvas.Get("width").Int()*13/10)+"px "+Theme.ScoreFont)
				ctx.Set("textAlign", "center")
				ctx.Set("textBaseline", "middle")
				ctx.Set("lineWidth", 2)
				ctx.Set("lineJoin", "round")
				ctx.Set("shadowColor", Theme.ScoreGlow)
				ctx.Call("strokeText", strconv.Itoa(num), canvas.Get("width").Int()/2, canvas.Get("height").Int()/2)
				ctx.Set("strokeStyle", Theme.ScoreColor)
				ctx.Call("strokeText", strconv.Itoa(num), canvas.Get("width").Int()/2, canvas.Get("height").Int()/2)
			})
	}

	// Background tile
	g.Level.Background = RenderToCanvas(256, 256, func(canvas, ctx *js.Object) {
		ctx.Set("fillStyle", Theme.BackgroundColor)
		ctx.Call("fillRect", 0, 0, canvas.Get("width").Int(), canvas.Get("height").Int())
		ctx.Set("globalCompositeOperation", "lighter")

		ctx.Call("beginPath")
		w := canvas.Get("width").Float()
		h := canvas.Get("height").Float()
		for i := 5; i >= 0; i-- {
			fi := float64(i)
			ctx.Call("moveTo", w*(fi+1)/4, -h)
			ctx.Call("lineTo", w*(fi-2)/4, h*2)
			ctx.Call("moveTo", -w, h*(fi-2)/4)
			ctx.Call("lineTo", w*2, h*(fi+1)/4)
		}
		ctx.Set("lineWidth", 3)
		ctx.Set("shadowBlur", Theme.DefaultShadowBlur)
		ctx.Set("strokeStyle", Theme.BackgroundLineColor)
		ctx.Set("shadowColor", Theme.BackgroundGlow)
		ctx.Call("stroke")

		ctx.Set("shadowBlur", 0)
		ctx.Set("globalAlpha", 0.25)
		ctx.Call("translate", w, 0)
		ctx.Call("scale", -1, 1)
		ctx.Call("drawImage", canvas, 0, 0)
	})

	// Create repeating pattern from background tile
	g.Level.BackgroundPattern = g.Ctx.Call("createPattern", g.Level.Background, "repeat")

	// Player ship sprite
	g.Ship.Image = RenderToCanvas(ShipR, ShipR*2, func(canvas, ctx *js.Object) {
		w := canvas.Get("width").Float()
		h := canvas.Get("height").Float()

		ctx.Call("beginPath")
		for i := 4; i >= 0; i-- {
			fi := float64(i)
			ctx.Call("moveTo", w/2, h*(1+fi)/10)
			ctx.Call("lineTo", w*(11+fi)/16, h*(15-fi)/16)
			ctx.Call("lineTo", w*(5-fi)/16, h*(15-fi)/16)
			ctx.Call("closePath")
		}
		lineWidth := int(w / 17)
		ctx.Set("lineWidth", lineWidth)
		ctx.Set("shadowBlur", lineWidth*2)
		ctx.Set("strokeStyle", Theme.ShipColor)
		ctx.Set("shadowColor", Theme.ShipGlow)
		ctx.Call("stroke")
		ctx.Call("stroke")

		// Center diamond
		p := w / 6
		ctx.Call("beginPath")
		ctx.Call("moveTo", w/2-p, h/2)
		ctx.Call("lineTo", w/2, h/2+p)
		ctx.Call("lineTo", w/2+p, h/2)
		ctx.Call("lineTo", w/2, h/2-p)
		ctx.Call("closePath")
		ctx.Set("strokeStyle", Theme.ShipCenterColor)
		ctx.Set("shadowColor", Theme.ShipCenterColor)
		ctx.Call("stroke")
		ctx.Call("stroke")
	})

	// Shield effect sprite
	g.Ship.Shield.Image = RenderToCanvas(ShipR*2, ShipR*2, func(canvas, ctx *js.Object) {
		w := canvas.Get("width").Float()
		d := 8.0
		ctx.Set("lineWidth", 18)
		ctx.Set("shadowBlur", Theme.ShieldShadowBlur)
		ctx.Set("strokeStyle", Theme.ShieldColor)
		ctx.Set("shadowColor", Theme.ShieldGlowColor)
		ctx.Call("beginPath")
		ctx.Call("arc", w/2, w/2, w/2+9-d, 0, math.Pi*2)
		ctx.Call("stroke")

		ctx.Set("lineWidth", 26+d)
		ctx.Set("shadowBlur", 0)
		ctx.Call("beginPath")
		ctx.Call("arc", w/2, w/2, w/2+13+d/2-d, 0, math.Pi*2)
		ctx.Call("stroke")
	})

	// Bullet sprite
	g.BulletImage = RenderToCanvas(BulletR*2, BulletR*2, func(canvas, ctx *js.Object) {
		w := canvas.Get("width").Float()
		h := canvas.Get("height").Float()
		p := 6.0
		ctx.Call("beginPath")
		ctx.Call("moveTo", w/2, p)
		ctx.Call("lineTo", w-p, h/2)
		ctx.Call("lineTo", w/2, h-p)
		ctx.Call("lineTo", p, h/2)
		ctx.Call("closePath")
		ctx.Set("lineWidth", Theme.BulletLineWidth)
		ctx.Set("shadowBlur", Theme.BulletShadowBlur)
		ctx.Set("strokeStyle", Theme.BulletColor)
		ctx.Set("shadowColor", Theme.BulletGlow)
		ctx.Call("stroke")
		ctx.Call("stroke")
	})

	// Explosion sprite
	g.ExplosionImage = RenderToCanvas(16, 16, func(canvas, ctx *js.Object) {
		w := canvas.Get("width").Float()
		h := canvas.Get("height").Float()
		ctx.Set("fillStyle", Theme.ExplosionColor)
		ctx.Set("shadowBlur", Theme.ExplosionShadowBlur)
		ctx.Set("shadowColor", Theme.ExplosionGlow)
		p := 6.0

		for i := 0; i < 5; i++ {
			ctx.Call("fillRect", p, p, w-p*2, h-p*2)
		}

		ctx.Set("lineWidth", 0.3)
		ctx.Set("strokeStyle", Theme.ExplosionLineColor)
		pp := p * 0.8
		ctx.Call("beginPath")
		ctx.Call("moveTo", pp, pp)
		ctx.Call("lineTo", w-pp, h-pp)
		ctx.Call("moveTo", w-pp, pp)
		ctx.Call("lineTo", pp, h-pp)
		ctx.Call("stroke")
	})

	// Torpedo animation frames
	frameCount := 8
	g.TorpedoImages = make([]*js.Object, frameCount)
	for i := 0; i < frameCount; i++ {
		idx := i // capture
		g.TorpedoImages[i] = RenderToCanvas(TorpedoR*2, TorpedoR*2, func(canvas, ctx *js.Object) {
			w := canvas.Get("width").Float()
			h := canvas.Get("height").Float()

			ctx.Call("translate", w/2, h/2)
			ctx.Call("rotate", math.Pi/-2*float64(idx)/float64(frameCount))
			ctx.Call("translate", -w/2, -h/2)

			p := 6.0
			ctx.Call("beginPath")
			ctx.Set("lineWidth", Theme.TorpedoLineWidth)
			ctx.Set("shadowBlur", Theme.DefaultShadowBlur)
			ctx.Set("strokeStyle", Theme.TorpedoColor)
			ctx.Set("shadowColor", Theme.TorpedoGlow)
			ctx.Call("moveTo", w/2, p)
			ctx.Call("lineTo", w-p, h/2)
			ctx.Call("lineTo", w/2, h-p)
			ctx.Call("lineTo", p, h/2)
			ctx.Call("closePath")
			ctx.Call("stroke")
			ctx.Call("stroke")
		})
	}

	// Enemy sprites
	g.InitializeEnemyGraphics()

	Debug("Graphics ready...")
}

// InitializeEnemyGraphics renders enemy type sprites.
func (g *Game) InitializeEnemyGraphics() {
	r := float64(ShipR)

	// Enemy type 0 - small fighter
	g.EnemyTypes[0].R = r
	g.EnemyTypes[0].Image = RenderToCanvas(int(r*2), int(r*2), func(canvas, ctx *js.Object) {
		w := canvas.Get("width").Float()
		h := canvas.Get("height").Float()

		ctx.Set("lineWidth", Theme.EnemyLineWidth)
		ctx.Set("shadowBlur", Theme.DefaultShadowBlur)
		ctx.Set("strokeStyle", Theme.EnemyColor)
		ctx.Set("shadowColor", Theme.EnemyGlow)
		ctx.Set("miterLimit", 128)
		ctx.Call("beginPath")

		for i := 4; i >= 0; i-- {
			fi := float64(i)
			x1 := w * (6 - fi) / 11
			y1 := h * (6 - fi) / 20
			x2 := w * (11 - fi) / 26
			y2 := h * (1 + fi) / 9

			ctx.Call("moveTo", w/2, h*(12-fi)/12-6)
			ctx.Call("lineTo", w-x1, y1)
			ctx.Call("lineTo", w-x2, y2)
			ctx.Call("lineTo", x2, y2)
			ctx.Call("lineTo", x1, y1)
			ctx.Call("closePath")
		}

		ctx.Call("stroke")
		ctx.Call("stroke")
		renderHeart(ctx, w/2, h/2, r)
	})

	// Enemy type 1 - medium fighter
	g.EnemyTypes[1].R = r
	g.EnemyTypes[1].Image = RenderToCanvas(int(r*2), int(r*2), func(canvas, ctx *js.Object) {
		w := canvas.Get("width").Float()
		h := canvas.Get("height").Float()

		ctx.Set("lineWidth", Theme.EnemyLineWidth)
		ctx.Set("shadowBlur", Theme.DefaultShadowBlur)
		ctx.Set("strokeStyle", Theme.EnemyColor)
		ctx.Set("shadowColor", Theme.EnemyGlow)

		for i := 4; i >= 0; i-- {
			fi := float64(i)
			x1 := w*(5-fi)/14 + 6
			y1 := h * (16 - fi) / 17
			x2 := w * (8 - fi) / 22
			y2 := h * (1 + fi) / 11

			ctx.Call("moveTo", w/2, h*(6-fi)/12)
			ctx.Call("lineTo", w-x1, y1)
			ctx.Call("lineTo", w-x2, y2)
			ctx.Call("lineTo", x2, y2)
			ctx.Call("lineTo", x1, y1)
			ctx.Call("closePath")
		}

		ctx.Call("stroke")
		ctx.Call("stroke")
		renderHeart(ctx, w/2, h/2, r)
	})

	// Enemy type 2 - turret
	g.EnemyTypes[2].R = r
	g.EnemyTypes[2].Image = RenderToCanvas(int(r*2), int(r*2), func(canvas, ctx *js.Object) {
		w := canvas.Get("width").Float()

		ctx.Set("lineWidth", Theme.EnemyLineWidth)
		ctx.Set("shadowBlur", Theme.DefaultShadowBlur)
		ctx.Set("strokeStyle", Theme.EnemyColor)
		ctx.Set("shadowColor", Theme.EnemyGlow)
		ctx.Set("miterLimit", 32)
		ctx.Call("beginPath")

		// Outer spiky ring
		for i := 0.0; i < math.Pi*2; i += math.Pi / 4 {
			d := math.Pi / 12
			rr := w/2 - 6
			x := w/2 + math.Sin(i+d)*rr
			y := w/2 + math.Cos(i+d)*rr

			if i == 0 {
				ctx.Call("moveTo", x, y)
			} else {
				ctx.Call("lineTo", x, y)
			}

			d -= math.Pi / 1.45
			ctx.Call("lineTo", w/2+math.Sin(i+d)*rr, w/2+math.Cos(i+d)*rr)
		}
		ctx.Call("closePath")

		// Inner octagon
		for i := 0.0; i < math.Pi*2; i += math.Pi / 4 {
			rr := w * 0.4
			x := w/2 + math.Sin(i)*rr
			y := w/2 + math.Cos(i)*rr

			if i == 0 {
				ctx.Call("moveTo", x, y)
			} else {
				ctx.Call("lineTo", x, y)
			}
		}
		ctx.Call("closePath")

		ctx.Call("stroke")
		ctx.Call("stroke")
		renderHeart(ctx, w/2, w/2, r)
	})

	// Enemy type 3 - boss
	bossR := float64(maxInt(WIDTH, HEIGHT) / 8)
	g.EnemyTypes[3].R = bossR
	g.EnemyTypes[3].Image = RenderToCanvas(int(bossR*2), int(bossR*2), func(canvas, ctx *js.Object) {
		w := canvas.Get("width").Float()
		h := canvas.Get("height").Float()

		ctx.Set("lineWidth", Theme.EnemyLineWidth)
		ctx.Set("shadowBlur", Theme.DefaultShadowBlur)
		ctx.Set("strokeStyle", Theme.EnemyColor)
		ctx.Set("shadowColor", Theme.EnemyGlow)
		ctx.Set("miterLimit", 32)
		ctx.Call("beginPath")

		for i := 6; i >= 0; i-- {
			fi := float64(i)
			ctx.Call("moveTo", w/2, h*fi/12+6)
			x1 := w*(11+fi)/18 - 6
			y1 := h * (25 - fi) / 28
			ctx.Call("lineTo", x1, y1)
			x2 := w*(16-fi)/16 - 6
			y2 := h * (fi + 4) / 28
			ctx.Call("lineTo", x2, y2)
			ctx.Call("lineTo", w/2, h*(fi+30)/36-6)
			ctx.Call("lineTo", w-x2, y2)
			ctx.Call("lineTo", w-x1, y1)
			ctx.Call("closePath")
		}

		ctx.Call("stroke")
		ctx.Call("stroke")
		renderHeart(ctx, w/2, h*0.8, bossR)
	})
}

// RenderBonusImage renders a bonus item sprite.
func (g *Game) RenderBonusImage(bonusType string) {
	g.BonusImages[bonusType] = RenderToCanvas(BonusR*2, BonusR*2, func(canvas, ctx *js.Object) {
		w := canvas.Get("width").Float()
		h := canvas.Get("height").Float()

		ctx.Set("shadowBlur", Theme.DefaultShadowBlur)
		if len(bonusType) > 1 {
			ctx.Set("fillStyle", Theme.BonusColorPoints)
		} else {
			ctx.Set("fillStyle", Theme.BonusColorPowerup)
		}
		ctx.Set("shadowColor", ctx.Get("fillStyle"))
		ctx.Call("arc", w/2, h/2, w/2-6, 0, math.Pi*2)
		ctx.Call("fill")

		ctx.Set("fillStyle", Theme.BonusTextColor)
		fontSize := int(w/1.8) - len(bonusType)*7 + 7
		ctx.Set("font", "bold "+strconv.Itoa(fontSize)+"px "+Theme.BonusFont)
		ctx.Set("textAlign", "center")
		ctx.Set("textBaseline", "middle")
		ctx.Call("fillText", bonusType, w/2, h/2)
	})
}

// RenderTextImage renders a text message image.
func (g *Game) RenderTextImage(text string) {
	width := WIDTH * 10 / 16
	height := WIDTH / 8
	g.Level.Text.Image = RenderToCanvas(width, height, func(canvas, ctx *js.Object) {
		w := canvas.Get("width").Float()
		h := canvas.Get("height").Float()

		shadowBlur := h / 10
		ctx.Set("shadowBlur", shadowBlur)
		fontSize := int(h*0.9 - shadowBlur*2)
		ctx.Set("font", "bold "+strconv.Itoa(fontSize)+"px "+Theme.TextFont)
		ctx.Set("textAlign", "center")
		ctx.Set("textBaseline", "middle")

		maxWidth := w - shadowBlur*2
		centerX := w / 2
		centerY := h / 2

		// Outer glow
		ctx.Set("fillStyle", Theme.TextPrimaryColor)
		ctx.Set("shadowColor", Theme.TextGlow)
		ctx.Call("fillText", text, centerX, centerY, maxWidth)
		ctx.Call("fillText", text, centerX, centerY, maxWidth)

		// Inner stroke
		ctx.Set("fillStyle", Theme.TextSecondaryColor)
		ctx.Set("shadowBlur", shadowBlur/4)
		ctx.Set("lineWidth", shadowBlur/4)
		ctx.Set("lineJoin", "round")
		ctx.Set("strokeStyle", Theme.TextSecondaryColor)
		ctx.Set("shadowColor", Theme.BackgroundColor)
		ctx.Call("strokeText", text, centerX, centerY, maxWidth)

		// Scanline effect
		ctx.Set("globalAlpha", 0.2)
		ctx.Set("globalCompositeOperation", "source-atop")
		ctx.Set("fillStyle", Theme.TextScanlineColor)
		for i := 0.0; i < h; i += 3 {
			ctx.Call("fillRect", 0, i, w, 1)
		}
	})
}

// renderHeart renders the diamond-shaped heart indicator.
func renderHeart(ctx *js.Object, x, y, r float64) {
	p := r / 6

	ctx.Call("beginPath")
	ctx.Call("moveTo", x-p, y)
	ctx.Call("lineTo", x, y+p)
	ctx.Call("lineTo", x+p, y)
	ctx.Call("lineTo", x, y-p)
	ctx.Call("closePath")

	ctx.Set("globalCompositeOperation", "lighter")
	ctx.Set("shadowColor", Theme.ShipCenterColor)
	ctx.Call("stroke")
}
