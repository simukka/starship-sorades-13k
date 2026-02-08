package common

// SeededRNG implements a Mulberry32 seeded pseudo-random number generator.
// Produces deterministic sequences for reproducible gameplay.
type SeededRNG struct {
	state       uint32
	initialSeed uint32
}

// NewSeededRNG creates a new seeded random number generator.
func NewSeededRNG(seed uint32) *SeededRNG {
	return &SeededRNG{
		state:       seed,
		initialSeed: seed,
	}
}

// SetSeed sets a new seed and resets the generator state.
func (r *SeededRNG) SetSeed(seed uint32) {
	r.state = seed
	r.initialSeed = seed
}

// Reset resets the generator to its initial seed.
func (r *SeededRNG) Reset() {
	r.state = r.initialSeed
}

// Random generates the next random number using Mulberry32 algorithm.
// Returns a float64 between 0 (inclusive) and 1 (exclusive).
func (r *SeededRNG) Random() float64 {
	r.state += 0x6D2B79F5
	t := r.state
	t = (t ^ (t >> 15)) * (t | 1)
	t ^= t + (t^(t>>7))*(t|61)
	return float64((t^(t>>14))>>0) / 4294967296.0
}

// RandomInt generates a random integer in the specified range [min, max).
func (r *SeededRNG) RandomInt(min, max int) int {
	return int(r.Random()*float64(max-min)) + min
}

// RandomFloat generates a random float in the specified range [min, max).
func (r *SeededRNG) RandomFloat(min, max float64) float64 {
	return r.Random()*(max-min) + min
}

// LevelSeed generates a deterministic seed for a specific level.
func LevelSeed(baseSeed uint32, levelNumber int) uint32 {
	seed := baseSeed ^ (uint32(levelNumber) * 2654435761)
	seed = (seed ^ (seed >> 16)) * 0x85ebca6b
	seed = (seed ^ (seed >> 13)) * 0xc2b2ae35
	return seed ^ (seed >> 16)
}
