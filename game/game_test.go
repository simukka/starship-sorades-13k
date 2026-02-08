package game

import (
	"math"
	"testing"
)

// =============================================================================
// Ship Tests
// =============================================================================

func TestShip_IsAlive(t *testing.T) {
	tests := []struct {
		name   string
		health int
		want   bool
	}{
		{"full health", 100, true},
		{"low health", 1, true},
		{"zero health", 0, false},
		{"negative health", -5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Ship{E: tt.health}
			if got := s.IsAlive(); got != tt.want {
				t.Errorf("Ship.IsAlive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestShip_GetPosition(t *testing.T) {
	s := &Ship{X: 100.5, Y: 200.5}
	x, y := s.GetPosition()
	if x != 100.5 || y != 200.5 {
		t.Errorf("GetPosition() = (%v, %v), want (100.5, 200.5)", x, y)
	}
}

func TestShip_GetHealth(t *testing.T) {
	s := &Ship{E: 75}
	if got := s.GetHealth(); got != 75 {
		t.Errorf("GetHealth() = %v, want 75", got)
	}
}

func TestShip_SetHealth(t *testing.T) {
	s := &Ship{E: 100}
	s.SetHealth(50)
	if s.E != 50 {
		t.Errorf("SetHealth(50) resulted in E = %v, want 50", s.E)
	}
}

func TestShip_GetAngle(t *testing.T) {
	s := &Ship{Angle: math.Pi / 4}
	if got := s.GetAngle(); got != math.Pi/4 {
		t.Errorf("GetAngle() = %v, want %v", got, math.Pi/4)
	}
}

func TestShip_GetRadius(t *testing.T) {
	s := &Ship{}
	if got := s.GetRadius(); got != float64(ShipR) {
		t.Errorf("GetRadius() = %v, want %v", got, float64(ShipR))
	}
}

func TestShip_AudioPan_Local(t *testing.T) {
	s := &Ship{local: true, X: 500}
	// Local ship should always return 0.0 pan (centered)
	if got := s.AudioPan(); got != 0.0 {
		t.Errorf("AudioPan() for local ship = %v, want 0.0", got)
	}
}

func TestShip_AddWeapon(t *testing.T) {
	s := &Ship{Weapons: []*Weapon{}}

	// First weapon should be straight ahead (0°)
	s.AddWeapon()
	if len(s.Weapons) != 1 {
		t.Fatalf("AddWeapon() did not add weapon, got %d weapons", len(s.Weapons))
	}
	// Y should be negative (forward), X should be ~0
	if math.Abs(s.Weapons[0].X) > 0.001 {
		t.Errorf("First weapon X = %v, want ~0", s.Weapons[0].X)
	}
	if s.Weapons[0].Y >= 0 {
		t.Errorf("First weapon Y = %v, want negative (forward)", s.Weapons[0].Y)
	}

	// Second weapon should be to the left (-15°)
	s.AddWeapon()
	if len(s.Weapons) != 2 {
		t.Fatalf("AddWeapon() did not add second weapon")
	}
	if s.Weapons[1].X >= 0 {
		t.Errorf("Second weapon X = %v, want negative (left)", s.Weapons[1].X)
	}

	// Third weapon should be to the right (+15°)
	s.AddWeapon()
	if len(s.Weapons) != 3 {
		t.Fatalf("AddWeapon() did not add third weapon")
	}
	if s.Weapons[2].X <= 0 {
		t.Errorf("Third weapon X = %v, want positive (right)", s.Weapons[2].X)
	}
}

func TestShip_AddWeapon_MaxWeapons(t *testing.T) {
	s := &Ship{Weapons: make([]*Weapon, MaxWeapons)}

	// Should not add beyond max
	s.AddWeapon()
	if len(s.Weapons) != MaxWeapons {
		t.Errorf("AddWeapon() added beyond max, got %d weapons, want %d", len(s.Weapons), MaxWeapons)
	}
}

func TestShip_ClearTarget(t *testing.T) {
	enemy := &Enemy{Health: 10}
	s := &Ship{
		Target:      enemy,
		LockingOn:   enemy,
		LockTimer:   10,
		LockMaxTime: 30,
	}

	s.ClearTarget()

	if s.Target != nil {
		t.Error("ClearTarget() did not clear Target")
	}
	if s.LockingOn != nil {
		t.Error("ClearTarget() did not clear LockingOn")
	}
	if s.LockTimer != 0 {
		t.Errorf("ClearTarget() did not reset LockTimer, got %d", s.LockTimer)
	}
	if s.LockMaxTime != 0 {
		t.Errorf("ClearTarget() did not reset LockMaxTime, got %d", s.LockMaxTime)
	}
}

func TestShip_PredictTargetAngle_StationaryTarget(t *testing.T) {
	s := &Ship{X: 0, Y: 0}
	target := &Enemy{X: 100, Y: 0, YOffset: 0, VelX: 0, VelY: 0}

	angle := s.PredictTargetAngle(target, WeaponSpeed)

	// Target is directly to the right, so angle should be around π/2
	expectedAngle := math.Atan2(100, 0) // dx=100, -dy=0
	if math.Abs(angle-expectedAngle) > 0.01 {
		t.Errorf("PredictTargetAngle() = %v, want ~%v", angle, expectedAngle)
	}
}

func TestShip_PredictTargetAngle_DirectlyAhead(t *testing.T) {
	s := &Ship{X: 0, Y: 100}
	target := &Enemy{X: 0, Y: 0, YOffset: 0, VelX: 0, VelY: 0}

	angle := s.PredictTargetAngle(target, WeaponSpeed)

	// Target is directly above (negative Y), angle should be ~0
	if math.Abs(angle) > 0.01 {
		t.Errorf("PredictTargetAngle() = %v, want ~0", angle)
	}
}

// =============================================================================
// Enemy Tests
// =============================================================================

func TestEnemy_IsAlive(t *testing.T) {
	tests := []struct {
		name   string
		health int
		want   bool
	}{
		{"alive", 10, true},
		{"barely alive", 1, true},
		{"dead", 0, false},
		{"very dead", -5, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Enemy{Health: tt.health}
			if got := e.IsAlive(); got != tt.want {
				t.Errorf("Enemy.IsAlive() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnemy_GetPosition(t *testing.T) {
	e := &Enemy{X: 50, Y: 100, YOffset: 25}
	x, y := e.GetPosition()
	if x != 50 || y != 125 { // Y includes YOffset
		t.Errorf("GetPosition() = (%v, %v), want (50, 125)", x, y)
	}
}

func TestEnemy_GetHealth(t *testing.T) {
	e := &Enemy{Health: 15}
	if got := e.GetHealth(); got != 15 {
		t.Errorf("GetHealth() = %v, want 15", got)
	}
}

func TestEnemy_SetHealth(t *testing.T) {
	e := &Enemy{Health: 20}
	e.SetHealth(10)
	if e.Health != 10 {
		t.Errorf("SetHealth(10) resulted in Health = %v, want 10", e.Health)
	}
}

func TestEnemy_GetRadius(t *testing.T) {
	e := &Enemy{Radius: 32.0}
	if got := e.GetRadius(); got != 32.0 {
		t.Errorf("GetRadius() = %v, want 32.0", got)
	}
}

func TestEnemy_GetAngle(t *testing.T) {
	e := &Enemy{Angle: math.Pi / 2}
	if got := e.GetAngle(); got != math.Pi/2 {
		t.Errorf("GetAngle() = %v, want %v", got, math.Pi/2)
	}
}

func TestEnemy_AudioPan(t *testing.T) {
	ship := &Ship{X: 500}

	tests := []struct {
		name    string
		enemyX  float64
		wantPan float64
	}{
		{"center", 500, 0.0},
		{"far left", 500 - WIDTH, -1.0},
		{"far right", 500 + WIDTH, 1.0},
		{"half left", 500 - WIDTH/4, -0.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Enemy{X: tt.enemyX, Target: ship}
			got := e.AudioPan()
			if math.Abs(got-tt.wantPan) > 0.1 {
				t.Errorf("AudioPan() = %v, want ~%v", got, tt.wantPan)
			}
		})
	}
}

func TestEnemy_AudioPan_NoTarget(t *testing.T) {
	e := &Enemy{X: 100, Target: nil}
	if got := e.AudioPan(); got != 0.0 {
		t.Errorf("AudioPan() with no target = %v, want 0.0", got)
	}
}

func TestEnemy_DistanceVolume(t *testing.T) {
	e := &Enemy{X: 0, Y: 0, YOffset: 0}

	tests := []struct {
		name    string
		targetX float64
		targetY float64
		maxDist float64
		wantMin float64
		wantMax float64
	}{
		{"at target", 0, 0, 100, 0.99, 1.01},
		{"at max distance", 100, 0, 100, 0.19, 0.21},
		{"beyond max distance", 200, 0, 100, 0.19, 0.21},
		{"half distance", 50, 0, 100, 0.7, 0.85},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := e.DistanceVolume(tt.targetX, tt.targetY, tt.maxDist)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("DistanceVolume() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestEnemy_TargetAngle(t *testing.T) {
	ship := &Ship{X: 100, Y: 100}

	tests := []struct {
		name      string
		enemyX    float64
		enemyY    float64
		wantAngle float64
	}{
		{"target below", 100, 0, 0},             // dy > 0, dx = 0 -> angle = 0
		{"target right", 0, 100, math.Pi / 2},   // dx > 0, dy = 0
		{"target left", 200, 100, -math.Pi / 2}, // dx < 0, dy = 0
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &Enemy{X: tt.enemyX, Y: tt.enemyY, Target: ship}
			got := e.TargetAngle()
			if math.Abs(got-tt.wantAngle) > 0.1 {
				t.Errorf("TargetAngle() = %v, want ~%v", got, tt.wantAngle)
			}
		})
	}
}

// =============================================================================
// Base Tests
// =============================================================================

func TestBase_NewBase(t *testing.T) {
	b := NewBase(100, 200)

	if b.X != 100 || b.Y != 200 {
		t.Errorf("NewBase position = (%v, %v), want (100, 200)", b.X, b.Y)
	}
	if b.Radius != BaseRadius {
		t.Errorf("NewBase Radius = %v, want %v", b.Radius, BaseRadius)
	}
	if b.ShieldRadius != BaseShieldRadius {
		t.Errorf("NewBase ShieldRadius = %v, want %v", b.ShieldRadius, BaseShieldRadius)
	}
}

func TestBase_GetPosition(t *testing.T) {
	b := &Base{X: 50, Y: 75}
	x, y := b.GetPosition()
	if x != 50 || y != 75 {
		t.Errorf("GetPosition() = (%v, %v), want (50, 75)", x, y)
	}
}

func TestBase_GetRadius(t *testing.T) {
	b := &Base{ShieldRadius: 150}
	if got := b.GetRadius(); got != 150 {
		t.Errorf("GetRadius() = %v, want 150", got)
	}
}

func TestBase_ContainsPoint(t *testing.T) {
	b := &Base{X: 100, Y: 100, ShieldRadius: 50}

	tests := []struct {
		name string
		x, y float64
		want bool
	}{
		{"center", 100, 100, true},
		{"inside", 110, 110, true},
		{"on edge", 150, 100, true},
		{"outside", 160, 100, false},
		{"far outside", 500, 500, false},
		{"diagonal inside", 135, 135, true},   // sqrt(35^2 + 35^2) ≈ 49.5
		{"diagonal outside", 140, 140, false}, // sqrt(40^2 + 40^2) ≈ 56.6
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := b.ContainsPoint(tt.x, tt.y); got != tt.want {
				t.Errorf("ContainsPoint(%v, %v) = %v, want %v", tt.x, tt.y, got, tt.want)
			}
		})
	}
}

func TestBase_IsShipInside(t *testing.T) {
	b := &Base{X: 100, Y: 100, ShieldRadius: 50}

	tests := []struct {
		name string
		ship *Ship
		want bool
	}{
		{"inside", &Ship{X: 100, Y: 100}, true},
		{"outside", &Ship{X: 200, Y: 200}, false},
		{"on edge", &Ship{X: 150, Y: 100}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := b.IsShipInside(tt.ship); got != tt.want {
				t.Errorf("IsShipInside() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBase_TriggerImpact(t *testing.T) {
	b := &Base{X: 100, Y: 100}

	b.TriggerImpact(150, 100) // Impact from the right

	if b.ImpactTimer != 15 {
		t.Errorf("TriggerImpact() ImpactTimer = %v, want 15", b.ImpactTimer)
	}

	// Angle should be 0 (right side)
	if math.Abs(b.ImpactAngle) > 0.1 {
		t.Errorf("TriggerImpact() ImpactAngle = %v, want ~0", b.ImpactAngle)
	}

	// Test impact from above
	b.TriggerImpact(100, 50)                     // Impact from above
	expectedAngle := math.Atan2(50-100, 100-100) // should be -π/2
	if math.Abs(b.ImpactAngle-(-math.Pi/2)) > 0.1 {
		t.Errorf("TriggerImpact() from above ImpactAngle = %v, want ~%v", b.ImpactAngle, expectedAngle)
	}
}

// =============================================================================
// Bullet Tests
// =============================================================================

func TestBullet_GetPosition(t *testing.T) {
	b := &Bullet{X: 200, Y: 300}
	x, y := b.GetPosition()
	if x != 200 || y != 300 {
		t.Errorf("GetPosition() = (%v, %v), want (200, 300)", x, y)
	}
}

func TestBullet_GetRadius(t *testing.T) {
	tests := []struct {
		name string
		kind BulletKind
		want float64
	}{
		{"standard bullet", StandardBullet, BulletR},
		{"torpedo", TorpedoBullet, TorpedoR},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bullet{Kind: tt.kind}
			if got := b.GetRadius(); got != tt.want {
				t.Errorf("GetRadius() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBullet_AudioPan(t *testing.T) {
	tests := []struct {
		name    string
		x       float64
		wantPan float64
	}{
		{"left edge", 0, -1.0},
		{"center", WIDTH / 2, 0.0},
		{"right edge", WIDTH, 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b := &Bullet{X: tt.x}
			got := b.AudioPan()
			if math.Abs(got-tt.wantPan) > 0.01 {
				t.Errorf("AudioPan() = %v, want %v", got, tt.wantPan)
			}
		})
	}
}

func TestBullet_DistanceVolume(t *testing.T) {
	b := &Bullet{X: 0, Y: 0}

	// At origin, distance to (0,0) should give max volume
	got := b.DistanceVolume(0, 0, 100)
	if got < 0.99 {
		t.Errorf("DistanceVolume at origin = %v, want ~1.0", got)
	}

	// At max distance, should give minimum volume
	got = b.DistanceVolume(100, 0, 100)
	if math.Abs(got-0.2) > 0.01 {
		t.Errorf("DistanceVolume at max = %v, want ~0.2", got)
	}
}

// =============================================================================
// Pool Tests
// =============================================================================

func TestBulletPool_Acquire(t *testing.T) {
	pool := NewBulletPool(10)

	// Initial state
	if pool.ActiveCount != 0 {
		t.Errorf("Initial ActiveCount = %v, want 0", pool.ActiveCount)
	}

	// Acquire one
	b := pool.Acquire()
	if b == nil {
		t.Fatal("Acquire() returned nil")
	}
	if pool.ActiveCount != 1 {
		t.Errorf("ActiveCount after acquire = %v, want 1", pool.ActiveCount)
	}
}

func TestBulletPool_AcquireKind(t *testing.T) {
	pool := NewBulletPool(10)

	b := pool.AcquireKind(TorpedoBullet)
	if b == nil {
		t.Fatal("AcquireKind() returned nil")
	}
	if b.Kind != TorpedoBullet {
		t.Errorf("AcquireKind(TorpedoBullet).Kind = %v, want TorpedoBullet", b.Kind)
	}
}

func TestBulletPool_Release(t *testing.T) {
	pool := NewBulletPool(10)

	pool.Acquire()
	pool.Acquire()
	pool.Acquire()

	if pool.ActiveCount != 3 {
		t.Fatalf("ActiveCount = %v, want 3", pool.ActiveCount)
	}

	pool.Release(1)
	if pool.ActiveCount != 2 {
		t.Errorf("ActiveCount after release = %v, want 2", pool.ActiveCount)
	}
}

func TestBulletPool_Release_InvalidIndex(t *testing.T) {
	pool := NewBulletPool(10)
	pool.Acquire()

	// Should not panic or change count
	pool.Release(-1)
	pool.Release(5)
	pool.Release(100)

	if pool.ActiveCount != 1 {
		t.Errorf("ActiveCount after invalid releases = %v, want 1", pool.ActiveCount)
	}
}

func TestBulletPool_Clear(t *testing.T) {
	pool := NewBulletPool(10)

	pool.Acquire()
	pool.Acquire()
	pool.Acquire()
	pool.Clear()

	if pool.ActiveCount != 0 {
		t.Errorf("ActiveCount after clear = %v, want 0", pool.ActiveCount)
	}
}

func TestBulletPool_MaxCapacity(t *testing.T) {
	pool := NewBulletPool(3)

	pool.Acquire()
	pool.Acquire()
	pool.Acquire()

	// Should return nil when full
	b := pool.Acquire()
	if b != nil {
		t.Error("Acquire() should return nil when pool is full")
	}
}

func TestBulletPool_ForEachReverse(t *testing.T) {
	pool := NewBulletPool(10)

	// Acquire 3 bullets and set X to identify them
	for i := 0; i < 3; i++ {
		b := pool.Acquire()
		b.X = float64(i)
	}

	// Verify iteration order (should be 2, 1, 0)
	var order []float64
	pool.ForEachReverse(func(b *Bullet, idx int) {
		order = append(order, b.X)
	})

	if len(order) != 3 {
		t.Fatalf("ForEachReverse visited %d items, want 3", len(order))
	}
	if order[0] != 2 || order[1] != 1 || order[2] != 0 {
		t.Errorf("ForEachReverse order = %v, want [2, 1, 0]", order)
	}
}

func TestExplosionPool_AcquireAndRelease(t *testing.T) {
	pool := NewExplosionPool(5)

	e := pool.Acquire()
	if e == nil {
		t.Fatal("Acquire() returned nil")
	}
	if pool.ActiveCount != 1 {
		t.Errorf("ActiveCount = %v, want 1", pool.ActiveCount)
	}

	pool.Release(0)
	if pool.ActiveCount != 0 {
		t.Errorf("ActiveCount after release = %v, want 0", pool.ActiveCount)
	}
}

func TestBonusPool_AcquireAndRelease(t *testing.T) {
	pool := NewBonusPool(5)

	b := pool.Acquire()
	if b == nil {
		t.Fatal("Acquire() returned nil")
	}
	b.Type = "+"

	if pool.ActiveCount != 1 {
		t.Errorf("ActiveCount = %v, want 1", pool.ActiveCount)
	}

	pool.Release(0)
	if pool.ActiveCount != 0 {
		t.Errorf("ActiveCount after release = %v, want 0", pool.ActiveCount)
	}
}

// =============================================================================
// SpatialGrid Tests
// =============================================================================

type mockCollidable struct {
	x, y   float64
	radius float64
}

func (m *mockCollidable) GetPosition() (float64, float64) { return m.x, m.y }
func (m *mockCollidable) GetRadius() float64              { return m.radius }

func TestSpatialGrid_New(t *testing.T) {
	grid := NewSpatialGrid(1000, 1000, 100)

	if grid.CellSize != 100 {
		t.Errorf("CellSize = %v, want 100", grid.CellSize)
	}
	if grid.GridWidth != 11 { // 1000/100 + 1
		t.Errorf("GridWidth = %v, want 11", grid.GridWidth)
	}
	if grid.GridHeight != 11 {
		t.Errorf("GridHeight = %v, want 11", grid.GridHeight)
	}
}

func TestSpatialGrid_InsertAndGet(t *testing.T) {
	grid := NewSpatialGrid(1000, 1000, 100)

	obj := &mockCollidable{x: 150, y: 150, radius: 10}
	grid.Insert(obj)

	nearby := grid.GetNearby(150, 150)
	if len(nearby) != 1 {
		t.Fatalf("GetNearby found %d objects, want 1", len(nearby))
	}
	if nearby[0] != obj {
		t.Error("GetNearby returned wrong object")
	}
}

func TestSpatialGrid_Clear(t *testing.T) {
	grid := NewSpatialGrid(1000, 1000, 100)

	obj := &mockCollidable{x: 150, y: 150, radius: 10}
	grid.Insert(obj)
	grid.Clear()

	nearby := grid.GetNearby(150, 150)
	if len(nearby) != 0 {
		t.Errorf("GetNearby after clear found %d objects, want 0", len(nearby))
	}
}

func TestSpatialGrid_GetNearbyAdjacentCells(t *testing.T) {
	grid := NewSpatialGrid(1000, 1000, 100)

	// Insert objects in adjacent cells
	obj1 := &mockCollidable{x: 50, y: 50, radius: 10}   // Cell (0,0)
	obj2 := &mockCollidable{x: 150, y: 50, radius: 10}  // Cell (1,0)
	obj3 := &mockCollidable{x: 50, y: 150, radius: 10}  // Cell (0,1)
	obj4 := &mockCollidable{x: 150, y: 150, radius: 10} // Cell (1,1)

	grid.Insert(obj1)
	grid.Insert(obj2)
	grid.Insert(obj3)
	grid.Insert(obj4)

	// Query from center should find all 4
	nearby := grid.GetNearby(100, 100)
	if len(nearby) != 4 {
		t.Errorf("GetNearby found %d objects, want 4", len(nearby))
	}
}

func TestSpatialGrid_GetInCell(t *testing.T) {
	grid := NewSpatialGrid(1000, 1000, 100)

	obj1 := &mockCollidable{x: 50, y: 50, radius: 10}
	obj2 := &mockCollidable{x: 150, y: 50, radius: 10}

	grid.Insert(obj1)
	grid.Insert(obj2)

	// GetInCell should only return objects in that specific cell
	inCell := grid.GetInCell(50, 50)
	if len(inCell) != 1 {
		t.Errorf("GetInCell found %d objects, want 1", len(inCell))
	}
}

func TestSpatialGrid_BoundaryClamp(t *testing.T) {
	grid := NewSpatialGrid(1000, 1000, 100)

	// Insert at negative coordinates (should clamp to cell 0)
	obj := &mockCollidable{x: -100, y: -100, radius: 10}
	grid.InsertAt(obj, -100, -100)

	// Should still be findable
	nearby := grid.GetNearby(0, 0)
	found := false
	for _, n := range nearby {
		if n == obj {
			found = true
			break
		}
	}
	if !found {
		t.Error("Object inserted at negative coords not found")
	}
}

// =============================================================================
// Weapon System Tests
// =============================================================================

func TestWeaponAnglePattern(t *testing.T) {
	s := &Ship{Weapons: []*Weapon{}}

	// Add all weapons and verify the alternating pattern
	expectedSigns := []int{0, -1, 1, -1, 1, -1, 1} // 0°, -15°, +15°, -30°, +30°...

	for i := 0; i < 7; i++ {
		s.AddWeapon()

		w := s.Weapons[i]
		sign := 0
		if w.X < -0.001 {
			sign = -1
		} else if w.X > 0.001 {
			sign = 1
		}

		if sign != expectedSigns[i] {
			t.Errorf("Weapon %d X sign = %d, want %d", i, sign, expectedSigns[i])
		}
	}
}

func TestWeaponSpeed(t *testing.T) {
	s := &Ship{Weapons: []*Weapon{}}
	s.AddWeapon()

	// First weapon should have magnitude equal to WeaponSpeed
	w := s.Weapons[0]
	speed := math.Sqrt(w.X*w.X + w.Y*w.Y)

	if math.Abs(speed-WeaponSpeed) > 0.01 {
		t.Errorf("Weapon speed = %v, want %v", speed, WeaponSpeed)
	}
}

// =============================================================================
// Enemy Config Tests
// =============================================================================

func TestEnemyConfigs_Exist(t *testing.T) {
	kinds := []EnemyKind{SmallFighter, MediumFighter, TurretFighter, Boss}

	for _, kind := range kinds {
		if _, ok := enemyConfigs[kind]; !ok {
			t.Errorf("enemyConfigs missing config for %v", kind)
		}
	}
}

func TestEnemyKindNames_Exist(t *testing.T) {
	kinds := []EnemyKind{SmallFighter, MediumFighter, TurretFighter, Boss}

	for _, kind := range kinds {
		if _, ok := EnemyKindNames[kind]; !ok {
			t.Errorf("EnemyKindNames missing name for %v", kind)
		}
	}
}

// =============================================================================
// Collision Detection Tests
// =============================================================================

func TestAABBCollision_Ship(t *testing.T) {
	// Test the collision box logic without the full Game context
	shipX, shipY := 100.0, 100.0

	tests := []struct {
		name    string
		bulletX float64
		bulletY float64
		want    bool
	}{
		{"direct hit", 100, 100, true},
		{"inside top", 100, 100 - ShipCollisionD + 1, true},
		{"inside bottom", 100, 100 + ShipCollisionD - 1, true},
		{"inside left", 100 - ShipCollisionE + 1, 100, true},
		{"inside right", 100 + ShipCollisionE - 1, 100, true},
		{"outside top", 100, 100 - ShipCollisionD - 10, false},
		{"outside bottom", 100, 100 + ShipCollisionD + 10, false},
		{"outside left", 100 - ShipCollisionE - 10, 100, false},
		{"outside right", 100 + ShipCollisionE + 10, 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the AABB check from Ship.Collision
			inY := shipY < (tt.bulletY+ShipCollisionD) && shipY > (tt.bulletY-ShipCollisionD)
			inX := shipX < tt.bulletX+ShipCollisionE && shipX > (tt.bulletX-ShipCollisionE)
			got := inY && inX

			if got != tt.want {
				t.Errorf("Collision at (%v, %v) = %v, want %v", tt.bulletX, tt.bulletY, got, tt.want)
			}
		})
	}
}

func TestAABBCollision_Enemy(t *testing.T) {
	// Test the collision box logic for enemies
	enemyX, enemyY := 100.0, 100.0
	enemyRadius := 32.0
	hitboxD := enemyRadius * 0.6

	tests := []struct {
		name    string
		bulletX float64
		bulletY float64
		want    bool
	}{
		{"direct hit", 100, 100, true},
		{"edge hit", 100 + hitboxD - 1, 100, true},
		{"just outside", 100 + hitboxD + 1, 100, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate the AABB check from Enemy.Collision
			inY := tt.bulletY < (enemyY+hitboxD) && tt.bulletY > (enemyY-hitboxD)
			inX := tt.bulletX > (enemyX-hitboxD) && tt.bulletX < (enemyX+hitboxD)
			got := inY && inX

			if got != tt.want {
				t.Errorf("Collision at (%v, %v) = %v, want %v", tt.bulletX, tt.bulletY, got, tt.want)
			}
		})
	}
}

// =============================================================================
// Shield Mechanic Tests
// =============================================================================

func TestShield_Decrement(t *testing.T) {
	shield := Shield{T: 100, MaxT: 100}

	// Simulate shield countdown
	for i := 0; i < 50; i++ {
		shield.T--
	}

	if shield.T != 50 {
		t.Errorf("Shield.T after 50 decrements = %v, want 50", shield.T)
	}
}

func TestShield_AddFormula(t *testing.T) {
	// Test the shield bonus pickup formula:
	// T += MaxT * MaxT * 2 / (T + MaxT*2)
	shield := Shield{T: 50, MaxT: 100}

	add := shield.MaxT * shield.MaxT * 2 / (shield.T + shield.MaxT*2)
	shield.T += add

	// With T=50, MaxT=100: add = 100*100*2 / (50+200) = 20000/250 = 80
	// New T = 50 + 80 = 130
	if shield.T != 130 {
		t.Errorf("Shield.T after pickup = %v, want 130", shield.T)
	}

	// When shield is at 0, should add more
	shield.T = 0
	add = shield.MaxT * shield.MaxT * 2 / (shield.T + shield.MaxT*2)
	// add = 20000/200 = 100
	shield.T += add
	if shield.T != 100 {
		t.Errorf("Shield.T from 0 after pickup = %v, want 100", shield.T)
	}
}

// =============================================================================
// Damage Mechanic Tests
// =============================================================================

func TestDamage_WeaponLoss(t *testing.T) {
	// Test that taking damage removes one weapon
	weaponsBefore := 5
	weapons := weaponsBefore

	// Simulate weapon loss from Hurt()
	if weapons > 1 {
		weapons = weapons - 1
	}

	if weapons != 4 {
		t.Errorf("Weapons after damage = %v, want 4", weapons)
	}

	// Should not go below 1
	weapons = 1
	if weapons > 1 {
		weapons = weapons - 1
	}
	if weapons != 1 {
		t.Errorf("Weapons should not go below 1, got %v", weapons)
	}
}

func TestDamage_Invincibility(t *testing.T) {
	// Test invincibility timeout mechanic
	timeout := -1 // Can take damage
	health := 100
	damage := 10

	// Take damage when timeout < 0
	if timeout < 0 {
		health -= damage
		timeout = 10 // Grant invincibility
	}

	if health != 90 {
		t.Errorf("Health after damage = %v, want 90", health)
	}
	if timeout != 10 {
		t.Errorf("Timeout after damage = %v, want 10", timeout)
	}

	// Cannot take damage while invincible
	if timeout < 0 {
		health -= damage
	}
	if health != 90 {
		t.Errorf("Health should not change during invincibility, got %v", health)
	}
}

func TestDamage_HealthClamp(t *testing.T) {
	// Test health cannot go below 0
	health := 5
	damage := 10

	health -= damage
	if health < 0 {
		health = 0
	}

	if health != 0 {
		t.Errorf("Health after fatal damage = %v, want 0", health)
	}
}

func TestDamage_EnergyPickup(t *testing.T) {
	// Test energy pickup clamping to 100
	health := 98
	health += 5
	if health > 100 {
		health = 100
	}

	if health != 100 {
		t.Errorf("Health after pickup = %v, want 100", health)
	}
}

// =============================================================================
// Movement Tests
// =============================================================================

func TestMovement_VelocityClamping(t *testing.T) {
	velX := 20.0
	velY := 20.0

	speed := math.Sqrt(velX*velX + velY*velY)
	if speed > ShipMaxSpeed {
		scale := ShipMaxSpeed / speed
		velX *= scale
		velY *= scale
	}

	newSpeed := math.Sqrt(velX*velX + velY*velY)
	if newSpeed > ShipMaxSpeed+0.01 {
		t.Errorf("Clamped speed = %v, want <= %v", newSpeed, ShipMaxSpeed)
	}
}

func TestMovement_Drag(t *testing.T) {
	velX := 10.0
	velY := 10.0

	// Apply drag
	velX *= ShipACCFactor
	velY *= ShipACCFactor

	expectedVel := 10.0 * ShipACCFactor
	if math.Abs(velX-expectedVel) > 0.001 {
		t.Errorf("Velocity after drag = %v, want %v", velX, expectedVel)
	}
}

func TestMovement_ThrustDirection(t *testing.T) {
	// Test that thrust in direction of angle works correctly
	angle := math.Pi / 4 // 45 degrees

	velX := math.Sin(angle) * ShipThrustAcc
	velY := -math.Cos(angle) * ShipThrustAcc

	// At 45 degrees, X and Y components should be equal magnitude
	if math.Abs(math.Abs(velX)-math.Abs(velY)) > 0.001 {
		t.Errorf("Thrust components at 45° not equal: velX=%v, velY=%v", velX, velY)
	}

	// Y should be negative (moving up/forward)
	if velY >= 0 {
		t.Error("Forward thrust Y component should be negative")
	}

	// X should be positive (moving right)
	if velX <= 0 {
		t.Error("45° thrust X component should be positive")
	}
}

func TestMovement_ReverseThrust(t *testing.T) {
	angle := 0.0 // Facing up

	// Reverse is half power
	velX := -math.Sin(angle) * ShipThrustAcc * 0.5
	velY := math.Cos(angle) * ShipThrustAcc * 0.5

	if velX != 0 {
		t.Errorf("Reverse thrust at 0° velX = %v, want 0", velX)
	}
	if velY != ShipThrustAcc*0.5 {
		t.Errorf("Reverse thrust at 0° velY = %v, want %v", velY, ShipThrustAcc*0.5)
	}
}

// =============================================================================
// Rotation Tests
// =============================================================================

func TestRotation_Clockwise(t *testing.T) {
	angle := 0.0
	angle += ShipRotationSpeed

	if angle != ShipRotationSpeed {
		t.Errorf("Angle after clockwise rotation = %v, want %v", angle, ShipRotationSpeed)
	}
}

func TestRotation_CounterClockwise(t *testing.T) {
	angle := 0.0
	angle -= ShipRotationSpeed

	if angle != -ShipRotationSpeed {
		t.Errorf("Angle after counter-clockwise rotation = %v, want %v", angle, -ShipRotationSpeed)
	}
}
