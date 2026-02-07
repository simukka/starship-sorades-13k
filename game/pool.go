package game

// Bullet represents a player projectile.
type Bullet struct {
	T         int     // Lifetime remaining
	X, Y      float64 // Position
	XAcc      float64 // X velocity
	YAcc      float64 // Y velocity
	PoolIndex int     // Index in pool for swap-and-pop
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

// Torpedo represents an enemy projectile.
type Torpedo struct {
	X, Y      float64
	XAcc      float64
	YAcc      float64
	E         int // Health
	PoolIndex int
}

// TorpedoPool manages reusable torpedo objects.
type TorpedoPool struct {
	Pool        []*Torpedo
	ActiveCount int
	MaxSize     int
}

// NewTorpedoPool creates a new torpedo pool.
func NewTorpedoPool(maxSize int) *TorpedoPool {
	pool := &TorpedoPool{
		Pool:    make([]*Torpedo, maxSize),
		MaxSize: maxSize,
	}
	for i := 0; i < maxSize; i++ {
		pool.Pool[i] = &Torpedo{PoolIndex: i}
	}
	return pool
}

// Acquire gets an available torpedo from the pool.
func (p *TorpedoPool) Acquire() *Torpedo {
	if p.ActiveCount >= p.MaxSize {
		return nil
	}
	t := p.Pool[p.ActiveCount]
	t.PoolIndex = p.ActiveCount
	p.ActiveCount++
	return t
}

// Release returns a torpedo to the pool.
func (p *TorpedoPool) Release(index int) {
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
func (p *TorpedoPool) Clear() {
	p.ActiveCount = 0
}

// ForEachReverse iterates over active objects in reverse order.
func (p *TorpedoPool) ForEachReverse(fn func(*Torpedo, int)) {
	for i := p.ActiveCount - 1; i >= 0; i-- {
		fn(p.Pool[i], i)
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
