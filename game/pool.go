package game

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
