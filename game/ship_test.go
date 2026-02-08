package game

import (
	"testing"
)

// createTestShip creates a ship with default test values
func createTestShip() *Ship {
	return &Ship{
		X:       WIDTH / 2,
		Y:       HEIGHT / 2,
		E:       100,
		Timeout: -1, // Not invincible
		Weapons: []*Weapon{
			{X: 0, Y: -16, AudioID: 0},
			{X: -8, Y: -8, AudioID: 0},
			{X: 8, Y: -8, AudioID: 0},
		},
	}
}

// TestHurt_ShieldBlocksDamage tests that active shield prevents damage
func TestHurt_ShieldBlocksDamage(t *testing.T) {
	g := NewGame()
	ship := createTestShip()
	ship.Shield.T = 100 // Active shield

	initialHealth := ship.E
	initialWeapons := len(ship.Weapons)

	ship.Hurt(g, 25)

	if ship.E != initialHealth {
		t.Errorf("Shield should block damage: expected health %d, got %d", initialHealth, ship.E)
	}
	if len(ship.Weapons) != initialWeapons {
		t.Errorf("Shield should prevent weapon loss: expected %d weapons, got %d", initialWeapons, len(ship.Weapons))
	}
}

// TestHurt_DamageApplied tests that damage reduces health
func TestHurt_DamageApplied(t *testing.T) {
	g := NewGame()
	ship := createTestShip()
	ship.E = 100
	ship.Timeout = -1 // Ensure not invincible

	ship.Hurt(g, 25)

	if ship.E != 75 {
		t.Errorf("Expected health 75 after 25 damage, got %d", ship.E)
	}
}

// TestHurt_InvincibilityPreventsMultipleDamage tests timeout prevents damage stacking
func TestHurt_InvincibilityPreventsMultipleDamage(t *testing.T) {
	g := NewGame()
	ship := createTestShip()
	ship.E = 100
	ship.Timeout = -1

	ship.Hurt(g, 25) // Should apply damage and set Timeout to 10

	if ship.Timeout != 10 {
		t.Errorf("Expected Timeout to be 10 after damage, got %d", ship.Timeout)
	}

	healthAfterFirst := ship.E
	ship.Hurt(g, 25) // Should NOT apply damage (Timeout > 0)

	if ship.E != healthAfterFirst {
		t.Errorf("Invincibility should prevent damage: expected %d, got %d", healthAfterFirst, ship.E)
	}
}

// TestHurt_HealthDoesNotGoBelowZero tests health floors at 0
func TestHurt_HealthDoesNotGoBelowZero(t *testing.T) {
	g := NewGame()
	ship := createTestShip()
	ship.E = 10
	ship.Timeout = -1
	g.Level.Paused = true // Prevent game over logic

	ship.Hurt(g, 50) // Damage exceeds health

	if ship.E != 0 {
		t.Errorf("Health should not go below 0: expected 0, got %d", ship.E)
	}
}

// TestHurt_WeaponLostOnDamage tests that one weapon is removed on damage
func TestHurt_WeaponLostOnDamage(t *testing.T) {
	g := NewGame()
	ship := createTestShip()
	ship.Timeout = -1

	initialWeapons := len(ship.Weapons)
	ship.Hurt(g, 10)

	if len(ship.Weapons) != initialWeapons-1 {
		t.Errorf("Expected %d weapons after damage, got %d", initialWeapons-1, len(ship.Weapons))
	}
}

// TestHurt_NoWeaponsDoesNotPanic tests damage with no weapons doesn't crash
func TestHurt_NoWeaponsDoesNotPanic(t *testing.T) {
	g := NewGame()
	ship := createTestShip()
	ship.Weapons = []*Weapon{} // No weapons
	ship.Timeout = -1

	// Should not panic
	ship.Hurt(g, 10)

	if len(ship.Weapons) != 0 {
		t.Errorf("Expected 0 weapons, got %d", len(ship.Weapons))
	}
}

// TestHurt_OSDSetOnDamage tests that OSD is set to show damage indicator
func TestHurt_OSDSetOnDamage(t *testing.T) {
	g := NewGame()
	ship := createTestShip()
	ship.OSD = 0
	ship.Timeout = -1

	ship.Hurt(g, 10)

	if ship.OSD != ShipMaxOSD {
		t.Errorf("Expected OSD to be %d after damage, got %d", ShipMaxOSD, ship.OSD)
	}
}

// TestHurt_MultipleHitsRemoveMultipleWeapons tests weapon loss accumulates
func TestHurt_MultipleHitsRemoveMultipleWeapons(t *testing.T) {
	g := NewGame()
	ship := createTestShip()

	initialWeapons := len(ship.Weapons)

	// Take 3 hits (resetting timeout each time to allow damage)
	for i := 0; i < 3; i++ {
		ship.Timeout = -1
		ship.Hurt(g, 5)
	}

	expectedWeapons := initialWeapons - 3
	if expectedWeapons < 0 {
		expectedWeapons = 0
	}

	if len(ship.Weapons) != expectedWeapons {
		t.Errorf("Expected %d weapons after 3 hits, got %d", expectedWeapons, len(ship.Weapons))
	}
}

// TestHurt_ShieldDoesNotSetOSD tests that shield hit doesn't trigger OSD
func TestHurt_ShieldDoesNotSetOSD(t *testing.T) {
	g := NewGame()
	ship := createTestShip()
	ship.Shield.T = 100
	ship.OSD = 0

	ship.Hurt(g, 25)

	if ship.OSD != 0 {
		t.Errorf("Shield hit should not set OSD: expected 0, got %d", ship.OSD)
	}
}

// TestShip_AudioPan tests audio panning calculation
func TestShip_AudioPan(t *testing.T) {
	testCases := []struct {
		name     string
		x        float64
		expected float64
	}{
		{"Left edge", 0, -1.0},
		{"Center", WIDTH / 2, 0.0},
		{"Right edge", WIDTH, 1.0},
		{"Quarter left", WIDTH / 4, -0.5},
		{"Quarter right", WIDTH * 3 / 4, 0.5},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ship := &Ship{X: tc.x}
			pan := ship.AudioPan()

			if pan != tc.expected {
				t.Errorf("Expected pan %f for X=%f, got %f", tc.expected, tc.x, pan)
			}
		})
	}
}

// TestFire_CreatesBulletsForEachWeapon tests that Fire creates one bullet per weapon
func TestFire_CreatesBulletsForEachWeapon(t *testing.T) {
	g := NewGame()
	ship := createTestShip()

	initialBullets := g.Bullets.ActiveCount
	weaponCount := len(ship.Weapons)

	ship.Fire(g)

	expectedBullets := initialBullets + weaponCount
	if g.Bullets.ActiveCount != expectedBullets {
		t.Errorf("Expected %d bullets after firing %d weapons, got %d",
			expectedBullets, weaponCount, g.Bullets.ActiveCount)
	}
}

// TestFire_NoBulletsWithNoWeapons tests that Fire does nothing without weapons
func TestFire_NoBulletsWithNoWeapons(t *testing.T) {
	g := NewGame()
	ship := createTestShip()
	ship.Weapons = []*Weapon{} // No weapons

	initialBullets := g.Bullets.ActiveCount

	ship.Fire(g)

	if g.Bullets.ActiveCount != initialBullets {
		t.Errorf("Expected %d bullets with no weapons, got %d",
			initialBullets, g.Bullets.ActiveCount)
	}
}

// TestFire_SetsReloadWithWeapons tests reload time is 4 with weapons
func TestFire_SetsReloadWithWeapons(t *testing.T) {
	g := NewGame()
	ship := createTestShip()
	ship.Reload = 0

	ship.Fire(g)

	if ship.Reload != 4 {
		t.Errorf("Expected Reload to be 4 with weapons, got %d", ship.Reload)
	}
}

// TestFire_SetsReloadWithoutWeapons tests reload time is 6 without weapons
func TestFire_SetsReloadWithoutWeapons(t *testing.T) {
	g := NewGame()
	ship := createTestShip()
	ship.Weapons = []*Weapon{}
	ship.Reload = 0

	ship.Fire(g)

	if ship.Reload != 6 {
		t.Errorf("Expected Reload to be 6 without weapons, got %d", ship.Reload)
	}
}

// TestFire_BulletVelocityIncludesShipMomentum tests bullet velocity calculation
func TestFire_BulletVelocityIncludesShipMomentum(t *testing.T) {
	g := NewGame()
	ship := &Ship{
		X:       WIDTH / 2,
		Y:       HEIGHT / 2,
		XAcc:    10.0, // Moving right
		YAcc:    -5.0, // Moving up
		Weapons: []*Weapon{{X: 0, Y: -16, AudioID: 0}},
	}

	ship.Fire(g)

	if g.Bullets.ActiveCount < 1 {
		t.Fatal("Expected at least 1 bullet to be created")
	}

	bullet := g.Bullets.Pool[0]

	// Expected: weapon velocity + half ship momentum
	expectedXAcc := 0 + 10.0/2   // weapon.X + ship.XAcc/2
	expectedYAcc := -16 + -5.0/2 // weapon.Y + ship.YAcc/2

	if bullet.XAcc != expectedXAcc {
		t.Errorf("Expected bullet XAcc %f, got %f", expectedXAcc, bullet.XAcc)
	}
	if bullet.YAcc != expectedYAcc {
		t.Errorf("Expected bullet YAcc %f, got %f", expectedYAcc, bullet.YAcc)
	}
}

// TestFire_BulletLifetimeSet tests that bullet lifetime is initialized
func TestFire_BulletLifetimeSet(t *testing.T) {
	g := NewGame()
	ship := createTestShip()

	ship.Fire(g)

	if g.Bullets.ActiveCount < 1 {
		t.Fatal("Expected at least 1 bullet to be created")
	}

	bullet := g.Bullets.Pool[0]
	if bullet.T != BulletMaxT {
		t.Errorf("Expected bullet lifetime %d, got %d", BulletMaxT, bullet.T)
	}
}

// TestFire_HandlesPoolExhaustion tests Fire doesn't crash when pool is full
func TestFire_HandlesPoolExhaustion(t *testing.T) {
	g := NewGame()
	ship := createTestShip()

	// Exhaust the bullet pool
	for i := 0; i < g.Bullets.MaxSize; i++ {
		g.Bullets.Acquire()
	}

	// Should not panic when pool is exhausted
	ship.Fire(g)

	// Pool should still be at max capacity
	if g.Bullets.ActiveCount != g.Bullets.MaxSize {
		t.Errorf("Expected pool at max capacity %d, got %d",
			g.Bullets.MaxSize, g.Bullets.ActiveCount)
	}
}

// TestFire_MultipleFires tests firing multiple times creates correct bullet count
func TestFire_MultipleFires(t *testing.T) {
	g := NewGame()
	ship := createTestShip()
	weaponCount := len(ship.Weapons)

	// Fire 3 times
	for i := 0; i < 3; i++ {
		ship.Fire(g)
	}

	expectedBullets := weaponCount * 3
	if g.Bullets.ActiveCount != expectedBullets {
		t.Errorf("Expected %d bullets after 3 fires, got %d",
			expectedBullets, g.Bullets.ActiveCount)
	}
}

// TestMove_LeftArrowAccelerates tests left arrow key decreases XAcc
func TestMove_LeftArrowAccelerates(t *testing.T) {
	g := NewGame()
	ship := &Ship{X: WIDTH / 2, Y: HEIGHT / 2, XAcc: 0, YAcc: 0}
	keys := map[int]bool{37: true} // Left arrow

	ship.Move(g, keys)

	if ship.XAcc >= 0 {
		t.Errorf("Expected negative XAcc when pressing left, got %f", ship.XAcc)
	}
}

// TestMove_RightArrowAccelerates tests right arrow key increases XAcc
func TestMove_RightArrowAccelerates(t *testing.T) {
	g := NewGame()
	ship := &Ship{X: WIDTH / 2, Y: HEIGHT / 2, XAcc: 0, YAcc: 0}
	keys := map[int]bool{39: true} // Right arrow

	ship.Move(g, keys)

	if ship.XAcc <= 0 {
		t.Errorf("Expected positive XAcc when pressing right, got %f", ship.XAcc)
	}
}

// TestMove_UpArrowAccelerates tests up arrow key decreases YAcc
func TestMove_UpArrowAccelerates(t *testing.T) {
	g := NewGame()
	ship := &Ship{X: WIDTH / 2, Y: HEIGHT / 2, XAcc: 0, YAcc: 0}
	keys := map[int]bool{38: true} // Up arrow

	ship.Move(g, keys)

	if ship.YAcc >= 0 {
		t.Errorf("Expected negative YAcc when pressing up, got %f", ship.YAcc)
	}
}

// TestMove_DownArrowAccelerates tests down arrow key increases YAcc
func TestMove_DownArrowAccelerates(t *testing.T) {
	g := NewGame()
	ship := &Ship{X: WIDTH / 2, Y: HEIGHT / 2, XAcc: 0, YAcc: 0}
	keys := map[int]bool{40: true} // Down arrow

	ship.Move(g, keys)

	if ship.YAcc <= 0 {
		t.Errorf("Expected positive YAcc when pressing down, got %f", ship.YAcc)
	}
}

// TestMove_AppliesVelocity tests that velocity is applied to position
func TestMove_AppliesVelocity(t *testing.T) {
	g := NewGame()
	ship := &Ship{X: 100, Y: 100, XAcc: 10, YAcc: 5}
	keys := map[int]bool{}

	initialX := ship.X
	initialY := ship.Y

	ship.Move(g, keys)

	if ship.X != initialX+10 {
		t.Errorf("Expected X to be %f, got %f", initialX+10, ship.X)
	}
	if ship.Y != initialY+5 {
		t.Errorf("Expected Y to be %f, got %f", initialY+5, ship.Y)
	}
}

// TestMove_AppliesDamping tests that velocity is damped each frame
func TestMove_AppliesDamping(t *testing.T) {
	g := NewGame()
	ship := &Ship{X: WIDTH / 2, Y: HEIGHT / 2, XAcc: 10, YAcc: 10}
	keys := map[int]bool{}

	ship.Move(g, keys)

	expectedXAcc := 10 * ShipACCFactor
	expectedYAcc := 10 * ShipACCFactor

	if ship.XAcc != expectedXAcc {
		t.Errorf("Expected XAcc %f after damping, got %f", expectedXAcc, ship.XAcc)
	}
	if ship.YAcc != expectedYAcc {
		t.Errorf("Expected YAcc %f after damping, got %f", expectedYAcc, ship.YAcc)
	}
}

// TestMove_LeftBoundaryClamp tests ship doesn't go past left edge
func TestMove_LeftBoundaryClamp(t *testing.T) {
	g := NewGame()
	ship := &Ship{X: -10, Y: HEIGHT / 2, XAcc: -5, YAcc: 0}
	keys := map[int]bool{}

	ship.Move(g, keys)

	if ship.X != 0 {
		t.Errorf("Expected X clamped to 0 at left edge, got %f", ship.X)
	}
}

// TestMove_RightBoundaryClamp tests ship doesn't go past right edge
func TestMove_RightBoundaryClamp(t *testing.T) {
	g := NewGame()
	ship := &Ship{X: float64(WIDTH + 10), Y: HEIGHT / 2, XAcc: 5, YAcc: 0}
	keys := map[int]bool{}

	ship.Move(g, keys)

	if ship.X != WIDTH-1 {
		t.Errorf("Expected X clamped to %d at right edge, got %f", WIDTH-1, ship.X)
	}
}

// TestMove_TopBoundaryClamp tests ship doesn't go past top edge
func TestMove_TopBoundaryClamp(t *testing.T) {
	g := NewGame()
	ship := &Ship{X: WIDTH / 2, Y: -10, XAcc: 0, YAcc: -5}
	keys := map[int]bool{}

	ship.Move(g, keys)

	if ship.Y != 0 {
		t.Errorf("Expected Y clamped to 0 at top edge, got %f", ship.Y)
	}
}

// TestMove_BottomBoundaryClamp tests ship doesn't go past bottom edge
func TestMove_BottomBoundaryClamp(t *testing.T) {
	g := NewGame()
	ship := &Ship{X: WIDTH / 2, Y: float64(HEIGHT + 10), XAcc: 0, YAcc: 5}
	keys := map[int]bool{}

	ship.Move(g, keys)

	if ship.Y != HEIGHT-1 {
		t.Errorf("Expected Y clamped to %d at bottom edge, got %f", HEIGHT-1, ship.Y)
	}
}

// TestMove_LeftArrowTiltsShip tests left arrow tilts ship left
func TestMove_LeftArrowTiltsShip(t *testing.T) {
	g := NewGame()
	ship := &Ship{X: WIDTH / 2, Y: HEIGHT / 2, Angle: 0}
	keys := map[int]bool{37: true} // Left arrow

	ship.Move(g, keys)

	// After pressing left, angle should be negative (tilting left)
	if ship.Angle >= 0 {
		t.Errorf("Expected negative angle when pressing left, got %f", ship.Angle)
	}
}

// TestMove_RightArrowTiltsShip tests right arrow tilts ship right
func TestMove_RightArrowTiltsShip(t *testing.T) {
	g := NewGame()
	ship := &Ship{X: WIDTH / 2, Y: HEIGHT / 2, Angle: 0}
	keys := map[int]bool{39: true} // Right arrow

	ship.Move(g, keys)

	// After pressing right, angle should be positive (tilting right)
	if ship.Angle <= 0 {
		t.Errorf("Expected positive angle when pressing right, got %f", ship.Angle)
	}
}

// TestMove_AngleDampsOverTime tests angle returns to neutral over time
func TestMove_AngleDampsOverTime(t *testing.T) {
	g := NewGame()
	ship := &Ship{X: WIDTH / 2, Y: HEIGHT / 2, Angle: 5.0}
	keys := map[int]bool{} // No keys pressed

	ship.Move(g, keys)

	expectedAngle := 5.0 * ShipAngleFactor
	if ship.Angle != expectedAngle {
		t.Errorf("Expected angle %f after damping, got %f", expectedAngle, ship.Angle)
	}
}

// TestMove_NoKeysNoAcceleration tests no acceleration without input
func TestMove_NoKeysNoAcceleration(t *testing.T) {
	g := NewGame()
	ship := &Ship{X: WIDTH / 2, Y: HEIGHT / 2, XAcc: 0, YAcc: 0}
	keys := map[int]bool{}

	ship.Move(g, keys)

	if ship.XAcc != 0 || ship.YAcc != 0 {
		t.Errorf("Expected no acceleration without input, got XAcc=%f YAcc=%f", ship.XAcc, ship.YAcc)
	}
}

// TestMove_DiagonalMovement tests pressing two arrows simultaneously
func TestMove_DiagonalMovement(t *testing.T) {
	g := NewGame()
	ship := &Ship{X: WIDTH / 2, Y: HEIGHT / 2, XAcc: 0, YAcc: 0}
	keys := map[int]bool{38: true, 39: true} // Up + Right

	ship.Move(g, keys)

	if ship.XAcc <= 0 {
		t.Errorf("Expected positive XAcc for right movement, got %f", ship.XAcc)
	}
	if ship.YAcc >= 0 {
		t.Errorf("Expected negative YAcc for up movement, got %f", ship.YAcc)
	}
}

// --- Collision Detection Tests ---

// TestCollision_DirectHit tests that a torpedo directly on the ship triggers collision
func TestCollision_DirectHit(t *testing.T) {
	g := NewGame()
	ship := createTestShip()
	ship.Shield.T = 0 // No shield
	ship.Timeout = -1 // Not invincible
	initialHealth := ship.E

	torpedo := &Bullet{X: ship.X, Y: ship.Y, Kind: TorpedoBullet}

	result := ship.Collision(g, torpedo)

	if !result {
		t.Error("Expected collision to return true for direct hit")
	}
	if ship.E >= initialHealth {
		t.Errorf("Expected damage to be applied, health was %d now %d", initialHealth, ship.E)
	}
}

// TestCollision_TorpedoWithinVerticalRange tests collision within vertical tolerance
func TestCollision_TorpedoWithinVerticalRange(t *testing.T) {
	g := NewGame()
	ship := createTestShip()
	ship.Shield.T = 0
	ship.Timeout = -1

	// Torpedo just inside vertical collision range
	torpedo := &Bullet{Kind: TorpedoBullet, X: ship.X, Y: ship.Y + ShipCollisionD - 1}

	result := ship.Collision(g, torpedo)

	if !result {
		t.Error("Expected collision when torpedo is within vertical range")
	}
}

// TestCollision_TorpedoWithinHorizontalRange tests collision within horizontal tolerance
func TestCollision_TorpedoWithinHorizontalRange(t *testing.T) {
	g := NewGame()
	ship := createTestShip()
	ship.Shield.T = 0
	ship.Timeout = -1

	// Torpedo just inside horizontal collision range
	torpedo := &Bullet{Kind: TorpedoBullet, X: ship.X + ShipCollisionE - 1, Y: ship.Y}

	result := ship.Collision(g, torpedo)

	if !result {
		t.Error("Expected collision when torpedo is within horizontal range")
	}
}

// TestCollision_TorpedoOutsideVerticalRange tests no collision when torpedo too far vertically
func TestCollision_TorpedoOutsideVerticalRange(t *testing.T) {
	g := NewGame()
	ship := createTestShip()
	ship.Shield.T = 0
	ship.Timeout = -1
	initialHealth := ship.E

	// Torpedo outside vertical collision range
	torpedo := &Bullet{Kind: TorpedoBullet, X: ship.X, Y: ship.Y + ShipCollisionD + 10}

	result := ship.Collision(g, torpedo)

	if result {
		t.Error("Expected no collision when torpedo is outside vertical range")
	}
	if ship.E != initialHealth {
		t.Errorf("Expected no damage, health was %d now %d", initialHealth, ship.E)
	}
}

// TestCollision_TorpedoOutsideHorizontalRange tests no collision when torpedo too far horizontally
func TestCollision_TorpedoOutsideHorizontalRange(t *testing.T) {
	g := NewGame()
	ship := createTestShip()
	ship.Shield.T = 0
	ship.Timeout = -1
	initialHealth := ship.E

	// Torpedo outside horizontal collision range
	torpedo := &Bullet{Kind: TorpedoBullet, X: ship.X + ShipCollisionE + 10, Y: ship.Y}

	result := ship.Collision(g, torpedo)

	if result {
		t.Error("Expected no collision when torpedo is outside horizontal range")
	}
	if ship.E != initialHealth {
		t.Errorf("Expected no damage, health was %d now %d", initialHealth, ship.E)
	}
}

// TestCollision_TorpedoAboveShip tests collision when torpedo is above ship
func TestCollision_TorpedoAboveShip(t *testing.T) {
	g := NewGame()
	ship := createTestShip()
	ship.Shield.T = 0
	ship.Timeout = -1

	// Torpedo above but within range
	torpedo := &Bullet{Kind: TorpedoBullet, X: ship.X, Y: ship.Y - ShipCollisionD + 1}

	result := ship.Collision(g, torpedo)

	if !result {
		t.Error("Expected collision when torpedo is above ship within range")
	}
}

// TestCollision_TorpedoLeftOfShip tests collision when torpedo is left of ship
func TestCollision_TorpedoLeftOfShip(t *testing.T) {
	g := NewGame()
	ship := createTestShip()
	ship.Shield.T = 0
	ship.Timeout = -1

	// Torpedo to the left but within range
	torpedo := &Bullet{Kind: TorpedoBullet, X: ship.X - ShipCollisionE + 1, Y: ship.Y}

	result := ship.Collision(g, torpedo)

	if !result {
		t.Error("Expected collision when torpedo is left of ship within range")
	}
}

// TestCollision_AppliesTenDamage tests that collision applies exactly 10 damage
func TestCollision_AppliesTenDamage(t *testing.T) {
	g := NewGame()
	ship := createTestShip()
	ship.E = 100
	ship.Shield.T = 0
	ship.Timeout = -1

	torpedo := &Bullet{Kind: TorpedoBullet, X: ship.X, Y: ship.Y}

	ship.Collision(g, torpedo)

	if ship.E != 90 {
		t.Errorf("Expected health 90 after 10 damage, got %d", ship.E)
	}
}

// TestCollision_ShieldBlocksDamage tests that shield prevents collision damage
func TestCollision_ShieldBlocksDamage(t *testing.T) {
	g := NewGame()
	ship := createTestShip()
	ship.E = 100
	ship.Shield.T = 50 // Active shield

	torpedo := &Bullet{Kind: TorpedoBullet, X: ship.X, Y: ship.Y}

	result := ship.Collision(g, torpedo)

	// Collision still occurs (returns true) but damage is blocked by Hurt()
	if !result {
		t.Error("Expected collision to return true even with shield")
	}
	if ship.E != 100 {
		t.Errorf("Shield should block damage, expected health 100, got %d", ship.E)
	}
}

// TestCollision_CreatesExplosion tests that collision spawns an explosion
func TestCollision_CreatesExplosion(t *testing.T) {
	g := NewGame()
	ship := createTestShip()
	ship.Shield.T = 0
	ship.Timeout = -1

	initialExplosions := g.Explosions.ActiveCount
	torpedo := &Bullet{Kind: TorpedoBullet, X: ship.X, Y: ship.Y}

	ship.Collision(g, torpedo)

	if g.Explosions.ActiveCount <= initialExplosions {
		t.Error("Expected explosion to be created on collision")
	}
}

// TestCollision_BoundaryConditionExactEdge tests collision at exact boundary edge
func TestCollision_BoundaryConditionExactEdge(t *testing.T) {
	g := NewGame()
	ship := createTestShip()
	ship.Shield.T = 0
	ship.Timeout = -1

	// Torpedo exactly at the collision boundary (should NOT collide due to < comparison)
	torpedo := &Bullet{Kind: TorpedoBullet, X: ship.X, Y: ship.Y + ShipCollisionD}

	result := ship.Collision(g, torpedo)

	if result {
		t.Error("Expected no collision at exact boundary edge (due to < not <=)")
	}
}

// TestCollision_DiagonalPosition tests collision with torpedo at diagonal offset
func TestCollision_DiagonalPosition(t *testing.T) {
	g := NewGame()
	ship := createTestShip()
	ship.Shield.T = 0
	ship.Timeout = -1

	// Torpedo at diagonal but within both ranges
	torpedo := &Bullet{Kind: TorpedoBullet,
		X: ship.X + ShipCollisionE/2,
		Y: ship.Y + ShipCollisionD/2,
	}

	result := ship.Collision(g, torpedo)

	if !result {
		t.Error("Expected collision when torpedo is within both horizontal and vertical range")
	}
}

// TestCollision_DiagonalPositionOutOfRange tests no collision when diagonal offset exceeds range
func TestCollision_DiagonalPositionOutOfRange(t *testing.T) {
	g := NewGame()
	ship := createTestShip()
	ship.Shield.T = 0
	ship.Timeout = -1

	// Torpedo at diagonal but outside horizontal range
	torpedo := &Bullet{Kind: TorpedoBullet,
		X: ship.X + ShipCollisionE + 5,
		Y: ship.Y + ShipCollisionD/2,
	}

	result := ship.Collision(g, torpedo)

	if result {
		t.Error("Expected no collision when torpedo is outside horizontal range")
	}
}

// --- AddWeapon Tests ---

// TestAddWeapon_FirstWeaponFiresForward tests that the first weapon fires straight ahead
func TestAddWeapon_FirstWeaponFiresForward(t *testing.T) {
	ship := &Ship{Weapons: []*Weapon{}}

	ship.AddWeapon()

	if len(ship.Weapons) != 1 {
		t.Fatalf("Expected 1 weapon after AddWeapon, got %d", len(ship.Weapons))
	}

	weapon := ship.Weapons[0]

	// First weapon should fire straight forward (0°)
	// X should be 0 (no horizontal component)
	// Y should be negative (forward/up on screen)
	if weapon.X != 0 {
		t.Errorf("First weapon X velocity should be 0, got %f", weapon.X)
	}
	if weapon.Y >= 0 {
		t.Errorf("First weapon Y velocity should be negative (forward), got %f", weapon.Y)
	}
	if weapon.Y != -WeaponSpeed {
		t.Errorf("First weapon Y velocity should be %f, got %f", -WeaponSpeed, weapon.Y)
	}
}

// TestAddWeapon_SecondWeaponFiresLeft tests that the second weapon fires 15° left
func TestAddWeapon_SecondWeaponFiresLeft(t *testing.T) {
	ship := &Ship{Weapons: []*Weapon{}}

	ship.AddWeapon() // First weapon (0°)
	ship.AddWeapon() // Second weapon (-15°)

	if len(ship.Weapons) != 2 {
		t.Fatalf("Expected 2 weapons after two AddWeapon calls, got %d", len(ship.Weapons))
	}

	weapon := ship.Weapons[1]

	// Second weapon should fire 15° left (-15°)
	// X should be negative (left)
	// Y should be negative (mostly forward)
	if weapon.X >= 0 {
		t.Errorf("Second weapon X velocity should be negative (left), got %f", weapon.X)
	}
	if weapon.Y >= 0 {
		t.Errorf("Second weapon Y velocity should be negative (forward), got %f", weapon.Y)
	}
}

// TestAddWeapon_ThirdWeaponFiresRight tests that the third weapon fires 15° right
func TestAddWeapon_ThirdWeaponFiresRight(t *testing.T) {
	ship := &Ship{Weapons: []*Weapon{}}

	ship.AddWeapon() // First weapon (0°)
	ship.AddWeapon() // Second weapon (-15°)
	ship.AddWeapon() // Third weapon (+15°)

	if len(ship.Weapons) != 3 {
		t.Fatalf("Expected 3 weapons after three AddWeapon calls, got %d", len(ship.Weapons))
	}

	weapon := ship.Weapons[2]

	// Third weapon should fire 15° right (+15°)
	// X should be positive (right)
	// Y should be negative (mostly forward)
	if weapon.X <= 0 {
		t.Errorf("Third weapon X velocity should be positive (right), got %f", weapon.X)
	}
	if weapon.Y >= 0 {
		t.Errorf("Third weapon Y velocity should be negative (forward), got %f", weapon.Y)
	}
}

// TestAddWeapon_SymmetricLeftRight tests that left and right weapons are symmetric
func TestAddWeapon_SymmetricLeftRight(t *testing.T) {
	ship := &Ship{Weapons: []*Weapon{}}

	// Add weapons: 0°, -15°, +15°, -30°, +30°
	for i := 0; i < 5; i++ {
		ship.AddWeapon()
	}

	// Weapon 1 (-15°) and Weapon 2 (+15°) should be symmetric
	leftWeapon := ship.Weapons[1]
	rightWeapon := ship.Weapons[2]

	// X velocities should be opposite
	if leftWeapon.X+rightWeapon.X > 0.0001 || leftWeapon.X+rightWeapon.X < -0.0001 {
		t.Errorf("Left and right weapons should have opposite X: left=%f, right=%f", leftWeapon.X, rightWeapon.X)
	}

	// Y velocities should be equal
	if leftWeapon.Y != rightWeapon.Y {
		t.Errorf("Left and right weapons should have equal Y: left=%f, right=%f", leftWeapon.Y, rightWeapon.Y)
	}
}

// TestAddWeapon_AlternatingPattern tests the left-right alternating pattern
func TestAddWeapon_AlternatingPattern(t *testing.T) {
	ship := &Ship{Weapons: []*Weapon{}}

	// Add all weapons
	for i := 0; i < MaxWeapons; i++ {
		ship.AddWeapon()
	}

	// Verify pattern: 0°, left, right, left, right, ...
	// First weapon (index 0): X = 0
	if ship.Weapons[0].X != 0 {
		t.Errorf("Weapon 0 should fire straight (X=0), got %f", ship.Weapons[0].X)
	}

	// Odd indices (1, 3, 5, ...) should fire left (negative X)
	for i := 1; i < len(ship.Weapons); i += 2 {
		if ship.Weapons[i].X >= 0 {
			t.Errorf("Weapon %d should fire left (negative X), got %f", i, ship.Weapons[i].X)
		}
	}

	// Even indices > 0 (2, 4, 6, ...) should fire right (positive X)
	for i := 2; i < len(ship.Weapons); i += 2 {
		if ship.Weapons[i].X <= 0 {
			t.Errorf("Weapon %d should fire right (positive X), got %f", i, ship.Weapons[i].X)
		}
	}
}

// TestAddWeapon_MaxWeaponsLimit tests that weapons stop at MaxWeapons
func TestAddWeapon_MaxWeaponsLimit(t *testing.T) {
	ship := &Ship{Weapons: []*Weapon{}}

	// Try to add more than MaxWeapons
	for i := 0; i < MaxWeapons+5; i++ {
		ship.AddWeapon()
	}

	if len(ship.Weapons) != MaxWeapons {
		t.Errorf("Should have exactly %d weapons (max), got %d", MaxWeapons, len(ship.Weapons))
	}
}

// TestAddWeapon_WeaponSpeedConsistent tests all weapons have the same speed magnitude
func TestAddWeapon_WeaponSpeedConsistent(t *testing.T) {
	ship := &Ship{Weapons: []*Weapon{}}

	for i := 0; i < MaxWeapons; i++ {
		ship.AddWeapon()
	}

	tolerance := 0.0001
	for i, weapon := range ship.Weapons {
		// Speed magnitude = sqrt(X² + Y²)
		speedSq := weapon.X*weapon.X + weapon.Y*weapon.Y
		expectedSpeedSq := WeaponSpeed * WeaponSpeed

		if speedSq < expectedSpeedSq-tolerance || speedSq > expectedSpeedSq+tolerance {
			t.Errorf("Weapon %d speed² should be %f, got %f", i, expectedSpeedSq, speedSq)
		}
	}
}

// TestAddWeapon_IncreasingAngles tests that angles increase with each pair
func TestAddWeapon_IncreasingAngles(t *testing.T) {
	ship := &Ship{Weapons: []*Weapon{}}

	for i := 0; i < 7; i++ {
		ship.AddWeapon()
	}

	// Compare absolute X values - should increase (more sideways) for higher indices
	// Weapons 1,2 are at ±15°, weapons 3,4 are at ±30°, weapons 5,6 are at ±45°
	absX1 := ship.Weapons[1].X
	if absX1 < 0 {
		absX1 = -absX1
	}
	absX3 := ship.Weapons[3].X
	if absX3 < 0 {
		absX3 = -absX3
	}
	absX5 := ship.Weapons[5].X
	if absX5 < 0 {
		absX5 = -absX5
	}

	if absX1 >= absX3 {
		t.Errorf("30° weapon should have larger |X| than 15° weapon: |X1|=%f, |X3|=%f", absX1, absX3)
	}
	if absX3 >= absX5 {
		t.Errorf("45° weapon should have larger |X| than 30° weapon: |X3|=%f, |X5|=%f", absX3, absX5)
	}
}

// TestAddWeapon_EmptyShip tests adding weapon to ship with no weapons
func TestAddWeapon_EmptyShip(t *testing.T) {
	ship := &Ship{} // Weapons is nil

	ship.AddWeapon()

	if len(ship.Weapons) != 1 {
		t.Errorf("Expected 1 weapon after AddWeapon on empty ship, got %d", len(ship.Weapons))
	}
}
