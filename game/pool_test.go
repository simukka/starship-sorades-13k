package game

import (
	"testing"
)

// --- StandardBullet Collision Tests ---

// TestBulletRender_StandardBullet_HitsEnemy tests that standard bullet collides with enemy
func TestBulletRender_StandardBullet_HitsEnemy(t *testing.T) {
	g := NewGame()

	// Create an enemy at center of screen
	enemy := &Enemy{
		X:      WIDTH / 2,
		Y:      HEIGHT / 2,
		Radius: 32,
		Health: 3,
	}
	g.Enemies = append(g.Enemies, enemy)

	// Create a bullet at the same position
	bullet := g.Bullets.Acquire()
	bullet.Kind = StandardBullet
	bullet.X = enemy.X
	bullet.Y = enemy.Y
	bullet.XAcc = 0
	bullet.YAcc = -10
	bullet.T = BulletMaxT

	initialHealth := enemy.Health

	// Render should return false (bullet removed due to collision)
	result := bullet.Render(g)

	if result {
		t.Error("Expected Render to return false when bullet hits enemy")
	}
	if enemy.Health >= initialHealth {
		t.Errorf("Expected enemy health to decrease, was %d now %d", initialHealth, enemy.Health)
	}
}

// TestBulletRender_StandardBullet_MissesEnemy tests that standard bullet misses distant enemy
func TestBulletRender_StandardBullet_MissesEnemy(t *testing.T) {
	g := NewGame()

	// Create an enemy at center of screen
	enemy := &Enemy{
		X:      WIDTH / 2,
		Y:      HEIGHT / 2,
		Radius: 32,
		Health: 3,
	}
	g.Enemies = append(g.Enemies, enemy)

	// Create a bullet far from enemy
	bullet := g.Bullets.Acquire()
	bullet.Kind = StandardBullet
	bullet.X = 100 // Far from enemy
	bullet.Y = 100
	bullet.XAcc = 0
	bullet.YAcc = -10
	bullet.T = BulletMaxT

	initialHealth := enemy.Health

	// Render should return true (bullet still active)
	result := bullet.Render(g)

	if !result {
		t.Error("Expected Render to return true when bullet misses enemy")
	}
	if enemy.Health != initialHealth {
		t.Errorf("Expected enemy health unchanged, was %d now %d", initialHealth, enemy.Health)
	}
}

// TestBulletRender_StandardBullet_HitsMultipleEnemies tests bullet stops at first enemy hit
func TestBulletRender_StandardBullet_HitsMultipleEnemies(t *testing.T) {
	g := NewGame()

	// Create two enemies at same position
	enemy1 := &Enemy{X: WIDTH / 2, Y: HEIGHT / 2, Radius: 32, Health: 3}
	enemy2 := &Enemy{X: WIDTH / 2, Y: HEIGHT / 2, Radius: 32, Health: 3}
	g.Enemies = append(g.Enemies, enemy1, enemy2)

	// Create a bullet at the same position
	bullet := g.Bullets.Acquire()
	bullet.Kind = StandardBullet
	bullet.X = enemy1.X
	bullet.Y = enemy1.Y
	bullet.T = BulletMaxT

	// Render should return false (bullet hits first enemy and is removed)
	result := bullet.Render(g)

	if result {
		t.Error("Expected Render to return false when bullet hits enemy")
	}

	// Only first enemy should be hit (bullet consumed)
	totalDamage := (3 - enemy1.Health) + (3 - enemy2.Health)
	if totalDamage != 1 {
		t.Errorf("Expected only 1 total damage dealt, got %d", totalDamage)
	}
}

// --- TorpedoBullet Collision Tests ---

// TestBulletRender_TorpedoBullet_HitsShip tests that torpedo collides with player ship
func TestBulletRender_TorpedoBullet_HitsShip(t *testing.T) {
	g := NewGame()
	g.Ship.X = WIDTH / 2
	g.Ship.Y = HEIGHT / 2
	g.Ship.E = 100
	g.Ship.Shield.T = 0 // No shield
	g.Ship.Timeout = -1 // Not invincible

	// Create a torpedo at ship position
	bullet := g.Bullets.Acquire()
	bullet.Kind = TorpedoBullet
	bullet.X = g.Ship.X
	bullet.Y = g.Ship.Y
	bullet.XAcc = 0
	bullet.YAcc = 5

	initialHealth := g.Ship.E

	// Render should return false (torpedo removed due to collision)
	result := bullet.Render(g)

	if result {
		t.Error("Expected Render to return false when torpedo hits ship")
	}
	if g.Ship.E >= initialHealth {
		t.Errorf("Expected ship health to decrease, was %d now %d", initialHealth, g.Ship.E)
	}
}

// TestBulletRender_TorpedoBullet_MissesShip tests that torpedo misses distant ship
func TestBulletRender_TorpedoBullet_MissesShip(t *testing.T) {
	g := NewGame()
	g.Ship.X = WIDTH / 2
	g.Ship.Y = HEIGHT / 2
	g.Ship.E = 100

	// Create a torpedo far from ship
	bullet := g.Bullets.Acquire()
	bullet.Kind = TorpedoBullet
	bullet.X = 100 // Far from ship
	bullet.Y = 100
	bullet.XAcc = 0
	bullet.YAcc = 5

	initialHealth := g.Ship.E

	// Render should return true (torpedo still active)
	result := bullet.Render(g)

	if !result {
		t.Error("Expected Render to return true when torpedo misses ship")
	}
	if g.Ship.E != initialHealth {
		t.Errorf("Expected ship health unchanged, was %d now %d", initialHealth, g.Ship.E)
	}
}

// TestBulletRender_TorpedoBullet_ShieldBlocksDamage tests that shield blocks torpedo damage
func TestBulletRender_TorpedoBullet_ShieldBlocksDamage(t *testing.T) {
	g := NewGame()
	g.Ship.X = WIDTH / 2
	g.Ship.Y = HEIGHT / 2
	g.Ship.E = 100
	g.Ship.Shield.T = 50 // Active shield

	// Create a torpedo at ship position
	bullet := g.Bullets.Acquire()
	bullet.Kind = TorpedoBullet
	bullet.X = g.Ship.X
	bullet.Y = g.Ship.Y
	bullet.XAcc = 0
	bullet.YAcc = 5

	// Render should return false (collision still occurs, torpedo removed)
	result := bullet.Render(g)

	if result {
		t.Error("Expected Render to return false when torpedo hits shielded ship")
	}
	if g.Ship.E != 100 {
		t.Errorf("Shield should block damage, expected health 100, got %d", g.Ship.E)
	}
}

// TestBulletRender_TorpedoBullet_DoesNotHitEnemy tests that torpedo ignores enemies
func TestBulletRender_TorpedoBullet_DoesNotHitEnemy(t *testing.T) {
	g := NewGame()

	// Create an enemy
	enemy := &Enemy{X: WIDTH / 2, Y: HEIGHT / 2, Radius: 32, Health: 3}
	g.Enemies = append(g.Enemies, enemy)

	// Move ship away so torpedo doesn't hit it
	g.Ship.X = 100
	g.Ship.Y = 100

	// Create a torpedo at enemy position
	bullet := g.Bullets.Acquire()
	bullet.Kind = TorpedoBullet
	bullet.X = enemy.X
	bullet.Y = enemy.Y
	bullet.XAcc = 0
	bullet.YAcc = 5

	initialHealth := enemy.Health

	// Render should return true (torpedo doesn't collide with enemies)
	result := bullet.Render(g)

	if !result {
		t.Error("Expected Render to return true; torpedo should not hit enemies")
	}
	if enemy.Health != initialHealth {
		t.Errorf("Expected enemy health unchanged, was %d now %d", initialHealth, enemy.Health)
	}
}

// TestBulletRender_StandardBullet_DoesNotHitShip tests that standard bullet ignores player ship
func TestBulletRender_StandardBullet_DoesNotHitShip(t *testing.T) {
	g := NewGame()
	g.Ship.X = WIDTH / 2
	g.Ship.Y = HEIGHT / 2
	g.Ship.E = 100
	g.Ship.Shield.T = 0
	g.Ship.Timeout = -1

	// Create a bullet at ship position
	bullet := g.Bullets.Acquire()
	bullet.Kind = StandardBullet
	bullet.X = g.Ship.X
	bullet.Y = g.Ship.Y
	bullet.XAcc = 0
	bullet.YAcc = -10
	bullet.T = BulletMaxT

	initialHealth := g.Ship.E

	// Render should return true (bullet doesn't collide with own ship)
	result := bullet.Render(g)

	if !result {
		t.Error("Expected Render to return true; standard bullet should not hit player ship")
	}
	if g.Ship.E != initialHealth {
		t.Errorf("Expected ship health unchanged, was %d now %d", initialHealth, g.Ship.E)
	}
}

// --- Position Update Tests ---

// TestBulletRender_UpdatesPosition tests that Render updates bullet position
func TestBulletRender_UpdatesPosition(t *testing.T) {
	g := NewGame()

	bullet := g.Bullets.Acquire()
	bullet.Kind = StandardBullet
	bullet.X = 500
	bullet.Y = 500
	bullet.XAcc = 5
	bullet.YAcc = -10
	bullet.T = BulletMaxT

	initialX := bullet.X
	initialY := bullet.Y

	bullet.Render(g)

	expectedX := initialX + bullet.XAcc
	expectedY := initialY + bullet.YAcc

	if bullet.X != expectedX {
		t.Errorf("Expected X to be %f after update, got %f", expectedX, bullet.X)
	}
	if bullet.Y != expectedY {
		t.Errorf("Expected Y to be %f after update, got %f", expectedY, bullet.Y)
	}
}

// --- Off-Screen Tests ---

// TestBulletRender_StandardBullet_RemovedWhenOffScreenLeft tests removal at left edge
func TestBulletRender_StandardBullet_RemovedWhenOffScreenLeft(t *testing.T) {
	g := NewGame()

	bullet := g.Bullets.Acquire()
	bullet.Kind = StandardBullet
	bullet.X = -BulletR - 1 // Off left edge
	bullet.Y = HEIGHT / 2
	bullet.T = BulletMaxT

	result := bullet.Render(g)

	if result {
		t.Error("Expected bullet to be removed when off-screen left")
	}
}

// TestBulletRender_StandardBullet_RemovedWhenOffScreenRight tests removal at right edge
func TestBulletRender_StandardBullet_RemovedWhenOffScreenRight(t *testing.T) {
	g := NewGame()

	bullet := g.Bullets.Acquire()
	bullet.Kind = StandardBullet
	bullet.X = WIDTH + BulletR + 1 // Off right edge
	bullet.Y = HEIGHT / 2
	bullet.T = BulletMaxT

	result := bullet.Render(g)

	if result {
		t.Error("Expected bullet to be removed when off-screen right")
	}
}

// TestBulletRender_StandardBullet_RemovedWhenOffScreenBottom tests removal at bottom edge
func TestBulletRender_StandardBullet_RemovedWhenOffScreenBottom(t *testing.T) {
	g := NewGame()

	bullet := g.Bullets.Acquire()
	bullet.Kind = StandardBullet
	bullet.X = WIDTH / 2
	bullet.Y = HEIGHT + BulletR + 1 // Off bottom edge
	bullet.T = BulletMaxT

	result := bullet.Render(g)

	if result {
		t.Error("Expected bullet to be removed when off-screen bottom")
	}
}

// TestBulletRender_StandardBullet_RemovedWhenExpired tests removal when lifetime expires
func TestBulletRender_StandardBullet_RemovedWhenExpired(t *testing.T) {
	g := NewGame()

	bullet := g.Bullets.Acquire()
	bullet.Kind = StandardBullet
	bullet.X = WIDTH / 2
	bullet.Y = HEIGHT / 2
	bullet.T = 0 // About to expire

	result := bullet.Render(g)

	if result {
		t.Error("Expected bullet to be removed when lifetime expires")
	}
}

// TestBulletRender_StandardBullet_DecreasesLifetime tests that lifetime decrements
func TestBulletRender_StandardBullet_DecreasesLifetime(t *testing.T) {
	g := NewGame()

	bullet := g.Bullets.Acquire()
	bullet.Kind = StandardBullet
	bullet.X = WIDTH / 2
	bullet.Y = HEIGHT / 2
	bullet.T = BulletMaxT

	initialT := bullet.T

	bullet.Render(g)

	if bullet.T != initialT-1 {
		t.Errorf("Expected lifetime to decrease from %d to %d, got %d", initialT, initialT-1, bullet.T)
	}
}

// TestBulletRender_TorpedoBullet_RemovedWhenOffScreenBottom tests torpedo removal at bottom
func TestBulletRender_TorpedoBullet_RemovedWhenOffScreenBottom(t *testing.T) {
	g := NewGame()
	g.Ship.X = WIDTH / 2
	g.Ship.Y = HEIGHT / 2

	bullet := g.Bullets.Acquire()
	bullet.Kind = TorpedoBullet
	bullet.X = WIDTH / 2
	bullet.Y = HEIGHT + TorpedoR + 1 // Off bottom edge

	result := bullet.Render(g)

	if result {
		t.Error("Expected torpedo to be removed when off-screen bottom")
	}
}

// TestBulletRender_TorpedoBullet_RemovedWhenOffScreenTop tests torpedo removal at top
func TestBulletRender_TorpedoBullet_RemovedWhenOffScreenTop(t *testing.T) {
	g := NewGame()
	g.Ship.X = WIDTH / 2
	g.Ship.Y = HEIGHT / 2

	bullet := g.Bullets.Acquire()
	bullet.Kind = TorpedoBullet
	bullet.X = WIDTH / 2
	bullet.Y = -HEIGHT - 1 // Off top edge (extended boundary)

	result := bullet.Render(g)

	if result {
		t.Error("Expected torpedo to be removed when off-screen top")
	}
}

// --- Collision Boundary Tests ---

// TestBulletRender_StandardBullet_HitsEnemyAtBoundary tests collision at hitbox edge
func TestBulletRender_StandardBullet_HitsEnemyAtBoundary(t *testing.T) {
	g := NewGame()

	enemy := &Enemy{
		X:      WIDTH / 2,
		Y:      HEIGHT / 2,
		Radius: 32,
		Health: 3,
	}
	g.Enemies = append(g.Enemies, enemy)

	// Place bullet just inside enemy hitbox (hitboxD = Radius * 0.6)
	hitboxD := enemy.Radius * 0.6
	bullet := g.Bullets.Acquire()
	bullet.Kind = StandardBullet
	bullet.X = enemy.X + hitboxD - 1 // Just inside right edge
	bullet.Y = enemy.Y
	bullet.T = BulletMaxT

	initialHealth := enemy.Health

	result := bullet.Render(g)

	if result {
		t.Error("Expected collision when bullet is just inside enemy hitbox")
	}
	if enemy.Health >= initialHealth {
		t.Errorf("Expected enemy to take damage, was %d now %d", initialHealth, enemy.Health)
	}
}

// TestBulletRender_StandardBullet_MissesEnemyAtBoundary tests no collision just outside hitbox
func TestBulletRender_StandardBullet_MissesEnemyAtBoundary(t *testing.T) {
	g := NewGame()

	enemy := &Enemy{
		X:      WIDTH / 2,
		Y:      HEIGHT / 2,
		Radius: 32,
		Health: 3,
	}
	g.Enemies = append(g.Enemies, enemy)

	// Place bullet just outside enemy hitbox
	hitboxD := enemy.Radius * 0.6
	bullet := g.Bullets.Acquire()
	bullet.Kind = StandardBullet
	bullet.X = enemy.X + hitboxD + 1 // Just outside right edge
	bullet.Y = enemy.Y
	bullet.T = BulletMaxT

	initialHealth := enemy.Health

	result := bullet.Render(g)

	if !result {
		t.Error("Expected no collision when bullet is just outside enemy hitbox")
	}
	if enemy.Health != initialHealth {
		t.Errorf("Expected enemy health unchanged, was %d now %d", initialHealth, enemy.Health)
	}
}

// TestBulletRender_TorpedoBullet_HitsShipAtBoundary tests torpedo collision at ship hitbox edge
func TestBulletRender_TorpedoBullet_HitsShipAtBoundary(t *testing.T) {
	g := NewGame()
	g.Ship.X = WIDTH / 2
	g.Ship.Y = HEIGHT / 2
	g.Ship.E = 100
	g.Ship.Shield.T = 0
	g.Ship.Timeout = -1

	// Place torpedo just inside ship hitbox
	bullet := g.Bullets.Acquire()
	bullet.Kind = TorpedoBullet
	bullet.X = g.Ship.X + ShipCollisionE - 1 // Just inside right edge
	bullet.Y = g.Ship.Y
	bullet.XAcc = 0
	bullet.YAcc = 5

	initialHealth := g.Ship.E

	result := bullet.Render(g)

	if result {
		t.Error("Expected collision when torpedo is just inside ship hitbox")
	}
	if g.Ship.E >= initialHealth {
		t.Errorf("Expected ship to take damage, was %d now %d", initialHealth, g.Ship.E)
	}
}

// TestBulletRender_TorpedoBullet_MissesShipAtBoundary tests no collision just outside ship hitbox
func TestBulletRender_TorpedoBullet_MissesShipAtBoundary(t *testing.T) {
	g := NewGame()
	g.Ship.X = WIDTH / 2
	g.Ship.Y = HEIGHT / 2
	g.Ship.E = 100

	// Place torpedo just outside ship hitbox
	bullet := g.Bullets.Acquire()
	bullet.Kind = TorpedoBullet
	bullet.X = g.Ship.X + ShipCollisionE + 1 // Just outside right edge
	bullet.Y = g.Ship.Y
	bullet.XAcc = 0
	bullet.YAcc = 5

	initialHealth := g.Ship.E

	result := bullet.Render(g)

	if !result {
		t.Error("Expected no collision when torpedo is just outside ship hitbox")
	}
	if g.Ship.E != initialHealth {
		t.Errorf("Expected ship health unchanged, was %d now %d", initialHealth, g.Ship.E)
	}
}

// --- Explosion Tests ---

// TestBulletRender_StandardBullet_CreatesExplosionOnHit tests explosion spawns on enemy hit
func TestBulletRender_StandardBullet_CreatesExplosionOnHit(t *testing.T) {
	g := NewGame()

	enemy := &Enemy{X: WIDTH / 2, Y: HEIGHT / 2, Radius: 32, Health: 3}
	g.Enemies = append(g.Enemies, enemy)

	bullet := g.Bullets.Acquire()
	bullet.Kind = StandardBullet
	bullet.X = enemy.X
	bullet.Y = enemy.Y
	bullet.T = BulletMaxT

	initialExplosions := g.Explosions.ActiveCount

	bullet.Render(g)

	if g.Explosions.ActiveCount <= initialExplosions {
		t.Error("Expected explosion to be created when bullet hits enemy")
	}
}

// TestBulletRender_TorpedoBullet_CreatesExplosionOnHit tests explosion spawns on ship hit
func TestBulletRender_TorpedoBullet_CreatesExplosionOnHit(t *testing.T) {
	g := NewGame()
	g.Ship.X = WIDTH / 2
	g.Ship.Y = HEIGHT / 2
	g.Ship.E = 100
	g.Ship.Shield.T = 0
	g.Ship.Timeout = -1

	bullet := g.Bullets.Acquire()
	bullet.Kind = TorpedoBullet
	bullet.X = g.Ship.X
	bullet.Y = g.Ship.Y
	bullet.XAcc = 0
	bullet.YAcc = 5

	initialExplosions := g.Explosions.ActiveCount

	bullet.Render(g)

	if g.Explosions.ActiveCount <= initialExplosions {
		t.Error("Expected explosion to be created when torpedo hits ship")
	}
}
