package game

import (
	"math"
)

// SpawnEnemyType0 spawns small fighter enemies.
func (g *Game) SpawnEnemyType0(yOffset float64) {
	r := float64(ShipR)
	if g.EnemyTypes[0].R > 0 {
		r = g.EnemyTypes[0].R
	}

	for i := 2 + g.Level.LevelNum; i > 0; i-- {
		yStop := float64(HEIGHT/8) + g.GameRNG.Random()*float64(HEIGHT/4)
		var y float64
		if yOffset != 0 {
			y = yStop + yOffset
		} else {
			y = -r
		}

		enemy := &Enemy{
			Image:     g.EnemyTypes[0].Image,
			X:         r + (float64(WIDTH)-r*2)*g.GameRNG.Random(),
			Y:         y,
			YStop:     yStop,
			R:         r,
			MaxAngle:  math.Pi / 32,
			E:         12 + g.Level.LevelNum,
			T:         g.GameRNG.RandomInt(0, 120),
			TypeIndex: 0,
		}
		g.Enemies = append(g.Enemies, enemy)
	}
}

// SpawnEnemyType1 spawns medium fighter enemies.
func (g *Game) SpawnEnemyType1(yOffset float64) {
	r := float64(ShipR)
	if g.EnemyTypes[1].R > 0 {
		r = g.EnemyTypes[1].R
	}

	for i := 1 + g.Level.LevelNum; i > 0; i-- {
		yStop := float64(HEIGHT/8) + g.GameRNG.Random()*float64(HEIGHT/4)
		var y float64
		if yOffset != 0 {
			y = yStop + yOffset
		} else {
			y = -r
		}

		enemy := &Enemy{
			Image:     g.EnemyTypes[1].Image,
			X:         r + (float64(WIDTH)-r*2)*g.GameRNG.Random(),
			Y:         y,
			YStop:     yStop,
			R:         r,
			MaxAngle:  math.Pi / 16,
			E:         20 + g.Level.LevelNum*2,
			T:         g.GameRNG.RandomInt(0, 120),
			TypeIndex: 1,
		}
		g.Enemies = append(g.Enemies, enemy)
	}
}

// SpawnEnemyType2 spawns turret enemies.
func (g *Game) SpawnEnemyType2(yOffset float64) {
	r := float64(ShipR)
	if g.EnemyTypes[2].R > 0 {
		r = g.EnemyTypes[2].R
	}

	yStep := float64(HEIGHT) * -1.5 / float64(g.Level.LevelNum)

	for i := g.Level.LevelNum; i > 0; i-- {
		var y float64
		if yOffset != 0 {
			y = float64(i-1)*yStep + yOffset
		} else {
			y = float64(i-1)*yStep - r
		}

		enemy := &Enemy{
			Image:         g.EnemyTypes[2].Image,
			X:             float64(WIDTH/4) + float64(WIDTH/2)*g.GameRNG.Random(),
			Y:             y,
			YStop:         float64(HEIGHT / 2),
			R:             r,
			MaxAngle:      math.Pi * 32,
			E:             28 + g.Level.LevelNum*3,
			T:             g.GameRNG.RandomInt(0, 30),
			FireDirection: g.GameRNG.Random() * math.Pi,
			TypeIndex:     2,
		}
		g.Enemies = append(g.Enemies, enemy)
	}
}

// SpawnEnemyType3 spawns boss enemies.
func (g *Game) SpawnEnemyType3(yOffset float64) {
	var r float64
	if g.EnemyTypes[3].R > 0 {
		r = g.EnemyTypes[3].R
	} else {
		r = float64(WIDTH / 8)
		if HEIGHT > WIDTH {
			r = float64(HEIGHT / 8)
		}
	}

	var yStart float64
	if yOffset != 0 {
		yStart = yOffset
	} else {
		yStart = -r * 3
	}

	e := 36 + g.Level.LevelNum*4
	tStart := 60 // 2 * 30

	// Main boss
	enemy := &Enemy{
		Image:     g.EnemyTypes[3].Image,
		X:         float64(WIDTH / 2),
		Y:         yStart,
		YOffset:   r * 0.6,
		YStop:     r + 8,
		R:         r,
		MaxAngle:  math.Pi / 8,
		E:         e,
		T:         tStart + g.GameRNG.RandomInt(0, 30),
		TypeIndex: 3,
	}
	g.Enemies = append(g.Enemies, enemy)

	// Flanking sub-bosses
	size := r * 1.3
	d := size * 0.4
	x := d

	for i := 1; i < g.Level.LevelNum; i++ {
		y := r + 16 - size + math.Sqrt(float64(i))*64
		if i%2 == 1 {
			y += r * (0.7/float64(i) + 0.3)
		}

		// Left flanker
		leftEnemy := &Enemy{
			Image:     g.EnemyTypes[3].Image,
			X:         float64(WIDTH/2) - x,
			Y:         yStart,
			YOffset:   size / 2 * 0.6,
			YStop:     y,
			R:         size / 2,
			MaxAngle:  math.Pi / 8,
			E:         e / 2,
			T:         tStart + g.GameRNG.RandomInt(0, 30),
			TypeIndex: 3,
		}
		g.Enemies = append(g.Enemies, leftEnemy)

		// Right flanker
		rightEnemy := &Enemy{
			Image:     g.EnemyTypes[3].Image,
			X:         float64(WIDTH/2) + x,
			Y:         yStart,
			YOffset:   size / 2 * 0.6,
			YStop:     y,
			R:         size / 2,
			MaxAngle:  math.Pi / 8,
			E:         e / 2,
			T:         tStart + g.GameRNG.RandomInt(0, 30),
			TypeIndex: 3,
		}
		g.Enemies = append(g.Enemies, rightEnemy)

		x += d
		d *= 0.84
		size *= 0.9
	}
}

// EnemyShoot handles enemy shooting based on type.
func (g *Game) EnemyShoot(enemy *Enemy, angle float64) {
	switch enemy.TypeIndex {
	case 0:
		g.EnemyShootType0(enemy, angle)
	case 1:
		g.EnemyShootType1(enemy, angle)
	case 2:
		g.EnemyShootType2(enemy)
	case 3:
		g.EnemyShootType3(enemy, angle)
	}
}

// EnemyShootType0 - small fighter shooting behavior.
func (g *Game) EnemyShootType0(enemy *Enemy, angle float64) {
	enemy.T = max(5, 600/(g.Level.LevelNum+4))
	if g.SpawnTorpedo(enemy.X, enemy.Y, enemy.YOffset, angle, enemy.MaxAngle) {
		g.Audio.Play(19)
	}
}

// EnemyShootType1 - medium fighter shooting behavior.
func (g *Game) EnemyShootType1(enemy *Enemy, angle float64) {
	enemy.T = max(5, 600/(g.Level.LevelNum+4))

	if angle > math.Pi {
		angle -= math.Pi * 2
	}
	if angle > enemy.MaxAngle {
		angle = enemy.MaxAngle
	}
	if angle < -enemy.MaxAngle {
		angle = -enemy.MaxAngle
	}

	if g.SpawnTorpedo(enemy.X, enemy.Y, enemy.YOffset, angle+0.1, 0) ||
		g.SpawnTorpedo(enemy.X, enemy.Y, enemy.YOffset, angle-0.1, 0) {
		g.Audio.Play(20)
	}
}

// EnemyShootType2 - turret shooting behavior.
func (g *Game) EnemyShootType2(enemy *Enemy) {
	if enemy.TActive == 0 {
		enemy.TActive = 5
		enemy.T = max(5, 540/(g.Level.LevelNum+8))
	}
	enemy.TActive--

	enemy.FireDirection += 0.2
	result := false

	for i := math.Pi / 8; i < math.Pi*2; i += math.Pi {
		if g.SpawnTorpedo(enemy.X, enemy.Y, enemy.YOffset, enemy.FireDirection+i, 0) {
			result = true
		}
	}

	if result {
		g.Audio.Play(18)
	}
}

// EnemyShootType3 - boss shooting behavior.
func (g *Game) EnemyShootType3(enemy *Enemy, angle float64) {
	enemy.T = g.GameRNG.RandomInt(0, 180) // 6 * 30
	if g.SpawnTorpedo(enemy.X, enemy.Y, enemy.YOffset, angle, 0) {
		g.Audio.Play(12)
	}
}

// RemoveEnemy removes an enemy using swap-and-pop.
func (g *Game) RemoveEnemy(index int) {
	last := len(g.Enemies) - 1
	if index != last {
		g.Enemies[index] = g.Enemies[last]
	}
	g.Enemies = g.Enemies[:last]
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
