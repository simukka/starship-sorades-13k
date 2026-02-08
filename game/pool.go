package game

import "github.com/gopherjs/gopherjs/js"

// --- Spatial Hash Grid for Collision Detection ---

// Collidable is an interface for objects that can participate in spatial hashing.
type Collidable interface {
	GetPosition() (x, y float64)
	GetRadius() float64
}

// SpatialGrid is a grid-based spatial hash for efficient collision detection.
// Objects are inserted into cells based on their position, allowing O(1) lookup
// of nearby objects for collision checks.
type SpatialGrid struct {
	CellSize   float64
	GridWidth  int
	GridHeight int
	Cells      [][]Collidable
}

// NewSpatialGrid creates a new spatial grid for the given world dimensions.
// cellSize determines the granularity - smaller cells = more precise but more overhead.
// Recommended: cellSize should be roughly 2x the largest object radius.
func NewSpatialGrid(worldWidth, worldHeight int, cellSize float64) *SpatialGrid {
	gridWidth := int(float64(worldWidth)/cellSize) + 1
	gridHeight := int(float64(worldHeight)/cellSize) + 1

	cells := make([][]Collidable, gridWidth*gridHeight)
	for i := range cells {
		cells[i] = make([]Collidable, 0, 4) // Pre-allocate small capacity
	}

	return &SpatialGrid{
		CellSize:   cellSize,
		GridWidth:  gridWidth,
		GridHeight: gridHeight,
		Cells:      cells,
	}
}

// cellIndex calculates the cell index for a given position.
func (sg *SpatialGrid) cellIndex(x, y float64) int {
	cx := int(x / sg.CellSize)
	cy := int(y / sg.CellSize)

	// Clamp to grid bounds
	if cx < 0 {
		cx = 0
	}
	if cx >= sg.GridWidth {
		cx = sg.GridWidth - 1
	}
	if cy < 0 {
		cy = 0
	}
	if cy >= sg.GridHeight {
		cy = sg.GridHeight - 1
	}

	return cy*sg.GridWidth + cx
}

// Clear removes all objects from the grid. Call at the start of each frame.
func (sg *SpatialGrid) Clear() {
	for i := range sg.Cells {
		sg.Cells[i] = sg.Cells[i][:0] // Keep capacity, reset length
	}
}

// Insert adds an object to the grid at its current position.
func (sg *SpatialGrid) Insert(obj Collidable) {
	x, y := obj.GetPosition()
	idx := sg.cellIndex(x, y)
	sg.Cells[idx] = append(sg.Cells[idx], obj)
}

// InsertAt adds an object to the grid at a specific position (for objects that span cells).
func (sg *SpatialGrid) InsertAt(obj Collidable, x, y float64) {
	idx := sg.cellIndex(x, y)
	sg.Cells[idx] = append(sg.Cells[idx], obj)
}

// GetNearby returns all objects in the same cell and adjacent cells.
// This is the main query method for collision detection.
func (sg *SpatialGrid) GetNearby(x, y float64) []Collidable {
	cx := int(x / sg.CellSize)
	cy := int(y / sg.CellSize)

	var nearby []Collidable

	// Check 3x3 grid of cells centered on (cx, cy)
	for dy := -1; dy <= 1; dy++ {
		for dx := -1; dx <= 1; dx++ {
			ncx := cx + dx
			ncy := cy + dy

			// Skip out-of-bounds cells
			if ncx < 0 || ncx >= sg.GridWidth || ncy < 0 || ncy >= sg.GridHeight {
				continue
			}

			idx := ncy*sg.GridWidth + ncx
			nearby = append(nearby, sg.Cells[idx]...)
		}
	}

	return nearby
}

// GetInCell returns all objects in a specific cell (no adjacent cells).
func (sg *SpatialGrid) GetInCell(x, y float64) []Collidable {
	idx := sg.cellIndex(x, y)
	return sg.Cells[idx]
}

// --- Bullet Types and Pool ---

type BulletKind int

const (
	StandardBullet BulletKind = iota
	TorpedoBullet
)

// Bullet represents a player projectile.
type Bullet struct {
	T         int     // Lifetime remaining
	X, Y      float64 // Position
	XAcc      float64 // X velocity
	YAcc      float64 // Y velocity
	E         int     //Health
	PoolIndex int     // Index in pool for swap-and-pop
	Kind      BulletKind
}

// GetPosition implements Collidable interface.
func (b *Bullet) GetPosition() (x, y float64) {
	return b.X, b.Y
}

// GetRadius implements Collidable interface.
func (b *Bullet) GetRadius() float64 {
	if b.Kind == TorpedoBullet {
		return TorpedoR
	}
	return BulletR
}

// BulletPool manages reusable bullet objects.
type BulletPool struct {
	Pool        []*Bullet
	ActiveCount int
	MaxSize     int
}

// NewBulletPool creates a new bullet pool with pre-allocated objects.
func NewBulletPool(maxSize int) *BulletPool {
	pool := &BulletPool{
		Pool:    make([]*Bullet, maxSize),
		MaxSize: maxSize,
	}
	for i := 0; i < maxSize; i++ {
		pool.Pool[i] = &Bullet{PoolIndex: i}
	}
	return pool
}

// Acquire gets an available bullet from the pool.
func (p *BulletPool) Acquire() *Bullet {
	if p.ActiveCount >= p.MaxSize {
		return nil
	}
	b := p.Pool[p.ActiveCount]
	b.PoolIndex = p.ActiveCount
	p.ActiveCount++
	return b
}

func (p *BulletPool) AcquireKind(kind BulletKind) *Bullet {
	b := p.Acquire()
	if b == nil {
		return nil
	}
	b.Kind = kind
	return b
}

// Release returns a bullet to the pool using swap-and-pop.
func (p *BulletPool) Release(index int) {
	if index >= p.ActiveCount || index < 0 {
		return
	}
	lastIndex := p.ActiveCount - 1
	if index != lastIndex {
		p.Pool[index], p.Pool[lastIndex] = p.Pool[lastIndex], p.Pool[index]
		p.Pool[index].PoolIndex = index
	}
	p.ActiveCount--
}

// Clear resets the pool, marking all objects as inactive.
func (p *BulletPool) Clear() {
	p.ActiveCount = 0
}

// ForEachReverse iterates over active objects in reverse order.
func (p *BulletPool) ForEachReverse(fn func(*Bullet, int)) {
	for i := p.ActiveCount - 1; i >= 0; i-- {
		fn(p.Pool[i], i)
	}
}

// ForEachReverse iterates over active objects in reverse order.
func (p *BulletPool) ForEachKindReverse(kind BulletKind, fn func(*Bullet, int)) {
	for i := p.ActiveCount - 1; i >= 0; i-- {
		if p.Pool[i].Kind == kind {
			fn(p.Pool[i], i)
		}
	}
}

// AudioPan returns a pan value (-1.0 to 1.0) based on the X position.
// Left edge = -1.0, center = 0.0, right edge = 1.0
func (b *Bullet) AudioPan() float64 {
	return (b.X/WIDTH)*2 - 1
}

// DistanceVolume calculates volume based on distance to a target position.
// Closer = louder (up to 1.0), farther = quieter (minimum 0.2).
func (b *Bullet) DistanceVolume(targetX, targetY, maxDist float64) float64 {
	dx := b.X - targetX
	dy := b.Y - targetY
	dist := dx*dx + dy*dy // Skip sqrt for performance
	maxDistSq := maxDist * maxDist

	if dist >= maxDistSq {
		return 0.2 // Minimum volume for distant objects
	}

	// Linear falloff: 1.0 at dist=0, 0.2 at dist=maxDist
	return 1.0 - (dist/maxDistSq)*0.8
}

// Render draws the bullet and updates its position each frame.
//
// The method handles two bullet kinds with different behaviors:
//
// StandardBullet (player projectile):
//   - Uses the standard bullet image and BulletR radius
//   - Has a limited lifetime (T frames) and is removed when expired
//   - Checks collision with enemies; on hit, the bullet is removed
//   - Removed when off-screen (left, right, or bottom edges)
//
// TorpedoBullet (enemy projectile):
//   - Uses animated torpedo images and TorpedoR radius
//   - Has no lifetime limit; persists until off-screen or collision
//   - Checks collision with player ships; on hit, the bullet is removed
//   - Removed when off-screen (any edge, with extended top boundary for spawning)
//
// Returns true if the bullet should remain active, false if it should be released.
func (b *Bullet) Render(g *Game) bool {
	var image *js.Object
	var projectileR float64

	if b.Kind == StandardBullet {
		image = g.BulletImage
		projectileR = BulletR
	}
	if b.Kind == TorpedoBullet {
		image = g.TorpedoImages[g.TorpedoFrame]
		projectileR = TorpedoR
	}

	// Render
	g.Ctx.Call("drawImage", image, b.X-projectileR, b.Y-projectileR)

	// Update position
	b.X += b.XAcc
	b.Y += b.YAcc

	// TODO: refactor the Ship and Enemy objects into a single
	// interface
	if b.Kind == TorpedoBullet {
		// remove off-screen torpedos
		if b.Y >= HEIGHT+projectileR || b.X < -float64(projectileR) ||
			b.X >= WIDTH+projectileR || b.Y < -HEIGHT {
			return false
		}

		for _, s := range g.Ships {
			if s.Collision(g, b) {
				return false
			}
		}
	}

	if b.Kind == StandardBullet {
		// remove expired or off-screen bullets
		b.T--
		if b.T < 0 || b.X < -projectileR ||
			b.X >= WIDTH+projectileR || b.Y >= HEIGHT+projectileR {
			return false
		}

		for _, e := range g.Enemies {
			if e.Collision(g, b) {
				return false
			}
		}
	}

	return true
}

// Has this projection collided with another?
func (b *Bullet) Collision(g *Game, torpedo *Bullet) bool {

	if b.Y < (torpedo.Y+BulletTorpedoCollisionDist) &&
		b.Y > (torpedo.Y-BulletTorpedoCollisionDist) &&
		b.X < (torpedo.X+BulletTorpedoCollisionDist) &&
		b.X > (torpedo.X-BulletTorpedoCollisionDist) {
		torpedo.E--
		g.Level.P += 5
		if g.GameRNG.Random() > 0.75 {
			g.SpawnBonus(torpedo.X, torpedo.Y, 0, 0, "")
		}
		g.Explode(torpedo.X, torpedo.Y, 0)

	}
	return false
}

// // Torpedo represents an enemy projectile.
// type Torpedo struct {
// 	X, Y      float64
// 	XAcc      float64
// 	YAcc      float64
// 	E         int // Health
// 	PoolIndex int
// }

// func (s *Torpedo) AudioPan() float64 {
// 	return (s.X/WIDTH)*2 - 1
// }

// // DistanceVolume calculates volume based on distance to a target position.
// // Closer = louder (up to 1.0), farther = quieter (minimum 0.2).
// func (s *Torpedo) DistanceVolume(targetX, targetY, maxDist float64) float64 {
// 	dx := s.X - targetX
// 	dy := s.Y - targetY
// 	dist := dx*dx + dy*dy // Skip sqrt for performance
// 	maxDistSq := maxDist * maxDist

// 	if dist >= maxDistSq {
// 		return 0.2 // Minimum volume for distant torpedos
// 	}

// 	// Linear falloff: 1.0 at dist=0, 0.2 at dist=maxDist
// 	return 1.0 - (dist/maxDistSq)*0.8
// }

// // TorpedoPool manages reusable torpedo objects.
// type TorpedoPool struct {
// 	Pool        []*Torpedo
// 	ActiveCount int
// 	MaxSize     int
// }

// // NewTorpedoPool creates a new torpedo pool.
// func NewTorpedoPool(maxSize int) *TorpedoPool {
// 	pool := &TorpedoPool{
// 		Pool:    make([]*Torpedo, maxSize),
// 		MaxSize: maxSize,
// 	}
// 	for i := 0; i < maxSize; i++ {
// 		pool.Pool[i] = &Torpedo{PoolIndex: i}
// 	}
// 	return pool
// }

// // Acquire gets an available torpedo from the pool.
// func (p *TorpedoPool) Acquire() *Torpedo {
// 	if p.ActiveCount >= p.MaxSize {
// 		return nil
// 	}
// 	t := p.Pool[p.ActiveCount]
// 	t.PoolIndex = p.ActiveCount
// 	p.ActiveCount++
// 	return t
// }

// // Release returns a torpedo to the pool.
// func (p *TorpedoPool) Release(index int) {
// 	if index >= p.ActiveCount || index < 0 {
// 		return
// 	}
// 	lastIndex := p.ActiveCount - 1
// 	if index != lastIndex {
// 		p.Pool[index], p.Pool[lastIndex] = p.Pool[lastIndex], p.Pool[index]
// 		p.Pool[index].PoolIndex = index
// 	}
// 	p.ActiveCount--
// }

// // Clear resets the pool.
// func (p *TorpedoPool) Clear() {
// 	p.ActiveCount = 0
// }

// // ForEachReverse iterates over active objects in reverse order.
// func (p *TorpedoPool) ForEachReverse(fn func(*Torpedo, int)) {
// 	for i := p.ActiveCount - 1; i >= 0; i-- {
// 		fn(p.Pool[i], i)
// 	}
// }

// Explosion represents a visual explosion effect.
type Explosion struct {
	X, Y      float64
	Size      float64
	Angle     float64
	D         float64 // Rotation speed
	Alpha     float64
	PoolIndex int
}

// ExplosionPool manages reusable explosion objects.
type ExplosionPool struct {
	Pool        []*Explosion
	ActiveCount int
	MaxSize     int
}

// NewExplosionPool creates a new explosion pool.
func NewExplosionPool(maxSize int) *ExplosionPool {
	pool := &ExplosionPool{
		Pool:    make([]*Explosion, maxSize),
		MaxSize: maxSize,
	}
	for i := 0; i < maxSize; i++ {
		pool.Pool[i] = &Explosion{PoolIndex: i}
	}
	return pool
}

// Acquire gets an available explosion from the pool.
func (p *ExplosionPool) Acquire() *Explosion {
	if p.ActiveCount >= p.MaxSize {
		return nil
	}
	e := p.Pool[p.ActiveCount]
	e.PoolIndex = p.ActiveCount
	p.ActiveCount++
	return e
}

// Release returns an explosion to the pool.
func (p *ExplosionPool) Release(index int) {
	if index >= p.ActiveCount || index < 0 {
		return
	}
	lastIndex := p.ActiveCount - 1
	if index != lastIndex {
		p.Pool[index], p.Pool[lastIndex] = p.Pool[lastIndex], p.Pool[index]
		p.Pool[index].PoolIndex = index
	}
	p.ActiveCount--
}

// Clear resets the pool.
func (p *ExplosionPool) Clear() {
	p.ActiveCount = 0
}

// ForEachReverse iterates over active objects in reverse order.
func (p *ExplosionPool) ForEachReverse(fn func(*Explosion, int)) {
	for i := p.ActiveCount - 1; i >= 0; i-- {
		fn(p.Pool[i], i)
	}
}

// Bonus represents a collectible bonus item.
type Bonus struct {
	Type      string
	X, Y      float64
	XAcc      float64
	YAcc      float64
	PoolIndex int
}

// BonusPool manages reusable bonus objects.
type BonusPool struct {
	Pool        []*Bonus
	ActiveCount int
	MaxSize     int
}

// NewBonusPool creates a new bonus pool.
func NewBonusPool(maxSize int) *BonusPool {
	pool := &BonusPool{
		Pool:    make([]*Bonus, maxSize),
		MaxSize: maxSize,
	}
	for i := 0; i < maxSize; i++ {
		pool.Pool[i] = &Bonus{PoolIndex: i}
	}
	return pool
}

// Acquire gets an available bonus from the pool.
func (p *BonusPool) Acquire() *Bonus {
	if p.ActiveCount >= p.MaxSize {
		return nil
	}
	b := p.Pool[p.ActiveCount]
	b.PoolIndex = p.ActiveCount
	p.ActiveCount++
	return b
}

// Release returns a bonus to the pool.
func (p *BonusPool) Release(index int) {
	if index >= p.ActiveCount || index < 0 {
		return
	}
	lastIndex := p.ActiveCount - 1
	if index != lastIndex {
		p.Pool[index], p.Pool[lastIndex] = p.Pool[lastIndex], p.Pool[index]
		p.Pool[index].PoolIndex = index
	}
	p.ActiveCount--
}

// Clear resets the pool.
func (p *BonusPool) Clear() {
	p.ActiveCount = 0
}

// ForEachReverse iterates over active objects in reverse order.
func (p *BonusPool) ForEachReverse(fn func(*Bonus, int)) {
	for i := p.ActiveCount - 1; i >= 0; i-- {
		fn(p.Pool[i], i)
	}
}
