/**
 * ============================================================================
 * STARSHIP SORADES 13K
 * ============================================================================
 * A retro-style space shooter game built for JS13K competition.
 * 
 * @author Thiemo Mättig
 * @license CC-BY-SA-3.0-DE
 * @see https://creativecommons.org/licenses/by-sa/3.0/de/
 * 
 * REFACTORED VERSION - Improvements:
 * - Object pooling for projectiles and effects (reduced GC pressure)
 * - TypedArrays for position data (better cache locality)
 * - requestAnimationFrame for smooth animation
 * - createPattern for efficient background rendering
 * - Swap-and-pop for O(1) array removals
 * - Web Audio API for low-latency sound
 * - Seeded PRNG for deterministic level generation
 * ============================================================================
 */

// =============================================================================
// SECTION: Seeded Pseudo-Random Number Generator
// =============================================================================

/**
 * Mulberry32 - A fast, high-quality 32-bit seeded PRNG.
 * Produces deterministic sequences for reproducible gameplay.
 * Used for enemy spawning, movement patterns, and shooting timers.
 * 
 * @class
 */
class SeededRNG {
    /**
     * Creates a new seeded random number generator.
     * 
     * @param {number} seed - Initial seed value (32-bit integer)
     */
    constructor(seed = Date.now()) {
        /** @type {number} Current state of the generator */
        this.state = seed >>> 0; // Ensure unsigned 32-bit
        /** @type {number} Original seed for reset */
        this.initialSeed = this.state;
    }

    /**
     * Sets a new seed and resets the generator state.
     * 
     * @param {number} seed - New seed value
     */
    setSeed(seed) {
        this.state = seed >>> 0;
        this.initialSeed = this.state;
    }

    /**
     * Resets the generator to its initial seed.
     * Useful for replaying the same sequence.
     */
    reset() {
        this.state = this.initialSeed;
    }

    /**
     * Generates the next random number in the sequence.
     * Uses Mulberry32 algorithm for high-quality distribution.
     * 
     * @returns {number} Random float between 0 (inclusive) and 1 (exclusive)
     */
    random() {
        // Mulberry32 algorithm
        let t = this.state += 0x6D2B79F5;
        t = Math.imul(t ^ (t >>> 15), t | 1);
        t ^= t + Math.imul(t ^ (t >>> 7), t | 61);
        return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
    }

    /**
     * Generates a random integer in the specified range.
     * 
     * @param {number} min - Minimum value (inclusive)
     * @param {number} max - Maximum value (exclusive)
     * @returns {number} Random integer in [min, max)
     */
    randomInt(min, max) {
        return Math.floor(this.random() * (max - min)) + min;
    }

    /**
     * Generates a random float in the specified range.
     * 
     * @param {number} min - Minimum value (inclusive)
     * @param {number} max - Maximum value (exclusive)
     * @returns {number} Random float in [min, max)
     */
    randomFloat(min, max) {
        return this.random() * (max - min) + min;
    }

    /**
     * Generates a seed for a specific level.
     * Combines a base game seed with the level number for unique but
     * reproducible level generation.
     * 
     * @param {number} baseSeed - Base game seed
     * @param {number} levelNumber - Current level/wave number
     * @returns {number} Deterministic seed for the level
     */
    static levelSeed(baseSeed, levelNumber) {
        // Mix the seeds using a simple hash
        let seed = baseSeed ^ (levelNumber * 2654435761);
        seed = Math.imul(seed ^ (seed >>> 16), 0x85ebca6b);
        seed = Math.imul(seed ^ (seed >>> 13), 0xc2b2ae35);
        return (seed ^ (seed >>> 16)) >>> 0;
    }
}

/**
 * Global seeded RNG instance for gameplay-affecting randomness.
 * This ensures deterministic enemy behavior based on level seed.
 * 
 * @type {SeededRNG}
 */
const gameRNG = new SeededRNG();

/**
 * Base seed for the current game session.
 * Used to generate level-specific seeds.
 * 
 * @type {number}
 */
let gameSeed = Date.now();

// =============================================================================
// SECTION: Object Pool Implementation
// =============================================================================

/**
 * Generic object pool for efficient reuse of frequently created/destroyed objects.
 * Eliminates garbage collection pauses by recycling objects instead of allocating new ones.
 * 
 * @class
 * @template T
 */
class ObjectPool {
    /**
     * Creates a new object pool with pre-allocated objects.
     * 
     * @param {number} maxSize - Maximum number of objects in the pool
     * @param {function(): T} factory - Factory function to create new objects
     */
    constructor(maxSize, factory) {
        /** @type {T[]} Pre-allocated object storage */
        this.pool = Array.from({ length: maxSize }, factory);
        /** @type {number} Number of currently active objects */
        this.activeCount = 0;
        /** @type {number} Maximum pool capacity */
        this.maxSize = maxSize;
    }

    /**
     * Acquires an object from the pool.
     * Returns null if pool is exhausted.
     * 
     * @returns {T|null} An available object or null if pool is full
     */
    acquire() {
        if (this.activeCount >= this.maxSize) return null;
        const obj = this.pool[this.activeCount];
        obj._poolIndex = this.activeCount;
        this.activeCount++;
        return obj;
    }

    /**
     * Releases an object back to the pool using swap-and-pop.
     * O(1) complexity by swapping with last active element.
     * 
     * @param {number} index - Index of object to release
     */
    release(index) {
        if (index >= this.activeCount || index < 0) return;
        
        // Swap with last active element (swap-and-pop pattern)
        const lastIndex = this.activeCount - 1;
        if (index !== lastIndex) {
            const temp = this.pool[index];
            this.pool[index] = this.pool[lastIndex];
            this.pool[lastIndex] = temp;
            // Update the swapped object's index
            this.pool[index]._poolIndex = index;
        }
        this.activeCount--;
    }

    /**
     * Iterates over all active objects.
     * Safe for releases during iteration when iterating backwards.
     * 
     * @param {function(T, number): void} callback - Function called for each active object
     */
    forEachReverse(callback) {
        for (let i = this.activeCount - 1; i >= 0; i--) {
            callback(this.pool[i], i);
        }
    }

    /**
     * Resets the pool, marking all objects as inactive.
     */
    clear() {
        this.activeCount = 0;
    }
}

// =============================================================================
// SECTION: Web Audio Manager
// =============================================================================

/**
 * Manages sound effects using the Web Audio API.
 * Provides low-latency audio playback with automatic buffer management.
 * 
 * @class
 */
class AudioManager {
    constructor() {
        /** @type {AudioContext|null} Web Audio context */
        this.ctx = null;
        /** @type {Map<number, AudioBuffer>} Decoded audio buffers by sound ID */
        this.buffers = new Map();
        /** @type {boolean} Whether audio is ready to play */
        this.ready = false;
        /** @type {GainNode|null} Master volume control */
        this.masterGain = null;
    }

    /**
     * Initializes the audio context (must be called after user interaction).
     * 
     * @returns {Promise<void>}
     */
    async init() {
        if (this.ctx) return;
        
        try {
            // Create audio context (handles both standard and webkit prefix)
            this.ctx = new (window.AudioContext || window.webkitAudioContext)();
            
            // Create master gain node for volume control
            this.masterGain = this.ctx.createGain();
            this.masterGain.connect(this.ctx.destination);
            this.masterGain.gain.value = 0.7;
            
            this.ready = true;
        } catch (e) {
            console.warn('Web Audio API not available:', e);
        }
    }

    /**
     * Loads and decodes a sound effect from a jsfxr data URL.
     * 
     * @param {number} id - Sound effect identifier
     * @param {string} dataUrl - Base64 WAV data URL from jsfxr
     * @returns {Promise<void>}
     */
    async loadSound(id, dataUrl) {
        if (!this.ctx) return;
        
        try {
            // Convert data URL to ArrayBuffer
            const response = await fetch(dataUrl);
            const arrayBuffer = await response.arrayBuffer();
            
            // Decode audio data
            const audioBuffer = await this.ctx.decodeAudioData(arrayBuffer);
            this.buffers.set(id, audioBuffer);
        } catch (e) {
            console.warn(`Failed to load sound ${id}:`, e);
        }
    }

    /**
     * Plays a sound effect by ID.
     * Creates a new buffer source for each play (allows overlapping sounds).
     * 
     * @param {number} id - Sound effect identifier
     */
    play(id) {
        if (!this.ready || !this.buffers.has(id)) return;
        
        // Resume context if suspended (browser autoplay policy)
        if (this.ctx.state === 'suspended') {
            this.ctx.resume();
        }
        
        // Create and configure buffer source
        const source = this.ctx.createBufferSource();
        source.buffer = this.buffers.get(id);
        source.connect(this.masterGain);
        source.start(0);
    }

    /**
     * Sets the master volume.
     * 
     * @param {number} volume - Volume level (0.0 to 1.0)
     */
    setVolume(volume) {
        if (this.masterGain) {
            this.masterGain.gain.value = Math.max(0, Math.min(1, volume));
        }
    }
}

// =============================================================================
// SECTION: Game Constants and Configuration
// =============================================================================

/**
 * Level/game state configuration object.
 * Contains display settings, scrolling parameters, and runtime state.
 * 
 * @namespace
 */
const l = {
    // --- Display Configuration ---
    /** @type {number} Canvas width in pixels */
    WIDTH: 1024,
    /** @type {number} Canvas height in pixels */
    HEIGHT: 768,
    
    // --- Gameplay Settings ---
    /** @type {number} Background scroll speed (pixels per frame) */
    SPEED: 2,
    /** @type {number} Duration of screen flash effect (frames) */
    MAX_BOMB: 5,
    
    // --- Runtime State ---
    /** @type {number} Current vertical scroll position */
    y: 0,
    /** @type {number} Current bomb flash countdown */
    bomb: 0,
    /** @type {number} Current score */
    p: 0,
    /** @type {number} Current wave/level number */
    level: 0,
    /** @type {boolean} Whether game is paused */
    paused: false,
    /** @type {number} Current level's seed for deterministic spawning */
    levelSeed: 0,
    
    // --- Text Display System ---
    text: {
        /** @type {number} Maximum text display duration (frames, 3 seconds at 30fps) */
        MAX_T: 3 * 30,
        /** @type {number} Current text display countdown */
        t: 0,
        /** @type {number} Text X position */
        x: 0,
        /** @type {number} Text Y position */
        y: 0,
        /** @type {number} Text vertical velocity */
        yAcc: 0,
        /** @type {HTMLCanvasElement|null} Rendered text image */
        image: null
    },
    
    // --- Score Display ---
    points: {
        /** @type {number} Width of each digit sprite */
        WIDTH: 32,
        /** @type {number} Height of each digit sprite */
        HEIGHT: 48,
        /** @type {number} Horizontal spacing between digits */
        STEP: 24,
        /** @type {HTMLCanvasElement[]} Pre-rendered digit images (0-9) */
        images: []
    },
    
    // --- Background ---
    /** @type {HTMLCanvasElement|null} Background tile image */
    background: null,
    /** @type {CanvasPattern|null} Repeating background pattern */
    backgroundPattern: null
};

/**
 * Player ship state and configuration.
 * Contains physics parameters, weapon state, and visual assets.
 * 
 * @namespace
 */
const ship = {
    // --- Size and Physics Constants ---
    /** @type {number} Ship collision/render radius */
    R: Math.max(l.WIDTH, l.HEIGHT) / 16 | 0,
    /** @type {number} Acceleration magnitude per input frame */
    ACC: 1.5,
    /** @type {number} Velocity damping factor (0-1, lower = more drag) */
    ACC_FACTOR: 0.9,
    /** @type {number} Visual rotation damping factor */
    ANGLE_FACTOR: 0.8,
    /** @type {number} Maximum visual rotation angle (degrees) */
    MAX_ANGLE: 10,
    /** @type {number} Duration to show energy bar after damage (frames) */
    MAX_OSD: 6 * 30,
    
    // --- Position State ---
    /** @type {number} Current X position (center) */
    x: l.WIDTH / 2,
    /** @type {number} Current Y position (center) */
    y: l.HEIGHT * 7 / 8 | 0,
    
    // --- Velocity State ---
    /** @type {number} Current X velocity */
    xAcc: 0,
    /** @type {number} Current Y velocity */
    yAcc: 0,
    
    // --- Visual State ---
    /** @type {number} Current visual rotation (-1 to 1, multiplied by MAX_ANGLE) */
    angle: 0,
    
    // --- Game State ---
    /** @type {number} Current energy/health (0-100) */
    e: 100,
    /** @type {number} Invulnerability timeout after hit */
    timeout: 0,
    /** @type {number} Current weapon level (0-4) */
    weapon: 0,
    /** @type {number} Cooldown frames until next shot */
    reload: 0,
    /** @type {number} Remaining frames to show energy bar */
    osd: 0,
    
    // --- Shield State ---
    shield: {
        /** @type {number} Maximum shield duration (frames, 5 seconds) */
        MAX_T: 5 * 30,
        /** @type {number} Remaining shield frames */
        t: 0,
        /** @type {HTMLCanvasElement|null} Shield visual effect image */
        image: null
    },
    
    // --- Visual Assets ---
    /** @type {HTMLCanvasElement|null} Ship sprite */
    image: null,
    /** @type {HTMLCanvasElement|null} Original ship image (for easter egg) */
    originalImage: null
};

// =============================================================================
// SECTION: Projectile Pool Configurations
// =============================================================================

/**
 * Bullet pool configuration and shared properties.
 * 
 * @namespace
 */
const bullets = {
    /** @type {number} Bullet collision radius */
    R: 8,
    /** @type {number} Bullet lifetime in frames */
    MAX_T: 35,
    /** @type {HTMLCanvasElement|null} Bullet sprite */
    image: null,
    /** @type {ObjectPool|null} Bullet object pool */
    pool: null
};

/**
 * Torpedo (enemy projectile) pool configuration.
 * 
 * @namespace
 */
const torpedos = {
    /** @type {number} Torpedo collision radius */
    R: 16,
    /** @type {number} Current animation frame index */
    frame: 0,
    /** @type {HTMLCanvasElement[]} Rotated torpedo sprites for animation */
    images: [],
    /** @type {ObjectPool|null} Torpedo object pool */
    pool: null
};

/**
 * Explosion effect pool configuration.
 * 
 * @namespace
 */
const explosions = {
    /** @type {HTMLCanvasElement|null} Base explosion sprite (scaled at runtime) */
    image: null,
    /** @type {ObjectPool|null} Explosion object pool */
    pool: null
};

/**
 * Bonus item pool configuration.
 * 
 * @namespace
 */
const bonus = {
    /** @type {number} Bonus item collision radius */
    R: 16,
    /** @type {Object<string, HTMLCanvasElement>} Lazily-rendered bonus sprites by type */
    images: {},
    /** @type {ObjectPool|null} Bonus object pool */
    pool: null
};

/**
 * Enemy management.
 * 
 * @type {Array}
 */
const enemies = [];

// =============================================================================
// SECTION: Animation Frame Management
// =============================================================================

/**
 * Animation frame request ID for cancellation.
 * @type {number}
 */
let animationFrameId = 0;

/**
 * Timestamp of last frame for fixed timestep.
 * @type {number}
 */
let lastFrameTime = 0;

/**
 * Target frame duration in milliseconds (~30 FPS).
 * @type {number}
 */
const FRAME_DURATION = 33.33;

// =============================================================================
// SECTION: Audio Manager Instance
// =============================================================================

/**
 * Global audio manager instance.
 * @type {AudioManager}
 */
const audioManager = new AudioManager();

// =============================================================================
// SECTION: Canvas Context References
// =============================================================================

/**
 * Main canvas element.
 * @type {HTMLCanvasElement}
 */
let c;

/**
 * 2D rendering context.
 * @type {CanvasRenderingContext2D}
 */
let a;

// =============================================================================
// SECTION: Keyboard Input State
// =============================================================================

/**
 * Currently pressed keys (using Set for O(1) lookup).
 * @type {Set<number>}
 */
const keys = new Set();

/**
 * Key code remapping for alternative control schemes.
 * Maps various keys to the canonical control codes (arrows + X).
 * 
 * @type {Object<number, number>}
 */
const keyMap = {
    27: 80,  // Esc => P (Pause)
    32: 88,  // Space => X (Fire)
    48: 88,  // 0 => X
    50: 40,  // 2 => Down (numpad)
    52: 37,  // 4 => Left (numpad)
    53: 40,  // 5 => Down (numpad)
    54: 39,  // 6 => Right (numpad)
    56: 38,  // 8 => Up (numpad)
    65: 37,  // A => Left (WASD)
    67: 88,  // C => X
    68: 39,  // D => Right (WASD)
    73: 38,  // I => Up (IJKL)
    74: 37,  // J => Left (IJKL)
    75: 40,  // K => Down (IJKL)
    76: 39,  // L => Right (IJKL)
    83: 40,  // S => Down (WASD)
    87: 38,  // W => Up (WASD)
    89: 88,  // Y => X (German keyboard)
    90: 88   // Z => X
};

// =============================================================================
// SECTION: Utility Functions
// =============================================================================

/**
 * Toggles fullscreen mode for the canvas element.
 * Uses standard Fullscreen API (no vendor prefixes needed in 2026).
 * 
 * @returns {boolean} Always returns false to prevent default behavior
 */
function toggleFullscreen() {
    const canvas = document.querySelector('canvas');
    
    if (document.fullscreenElement) {
        document.exitFullscreen().catch(() => {});
    } else {
        canvas.requestFullscreen().catch(() => {});
    }
    
    return false;
}

// Export for Closure Compiler compatibility
window['toggleFullscreen'] = toggleFullscreen;

/**
 * Plays a sound effect by ID through the audio manager.
 * 
 * @param {number} id - Sound effect identifier (0-21)
 */
function play(id) {
    audioManager.play(id);
}

/**
 * Helper function to render graphics to an off-screen canvas buffer.
 * Used for pre-rendering sprites and visual elements.
 * 
 * @param {function(HTMLCanvasElement, CanvasRenderingContext2D): void} renderFn - Drawing function
 * @param {number} width - Canvas width
 * @param {number} [height] - Canvas height (defaults to width for square)
 * @param {HTMLCanvasElement} [existingCanvas] - Reuse existing canvas (optional)
 * @returns {HTMLCanvasElement} The rendered canvas
 */
function render(renderFn, width, height, existingCanvas) {
    const canvas = existingCanvas || document.createElement('canvas');
    canvas.width = width | 0;
    canvas.height = (height || width) | 0;
    renderFn(canvas, canvas.getContext('2d'));
    return canvas;
}

/**
 * Renders the diamond-shaped "heart" indicator in enemy/ship centers.
 * Creates a glowing diamond effect using shadow blur.
 * 
 * @param {CanvasRenderingContext2D} ctx - Canvas context
 * @param {number} x - Center X coordinate
 * @param {number} y - Center Y coordinate
 */
function renderHeart(ctx, x, y) {
    const p = ship.R / 6 | 0;
    
    ctx.beginPath();
    ctx.moveTo(x - p, y);
    ctx.lineTo(x, y + p);
    ctx.lineTo(x + p, y);
    ctx.lineTo(x, y - p);
    ctx.closePath();
    
    ctx.globalCompositeOperation = 'lighter';
    ctx.shadowColor = '#FFF';
    ctx.stroke();
}

// =============================================================================
// SECTION: Game Object Spawning Functions
// =============================================================================

/**
 * Applies damage to the player ship.
 * Handles shield blocking, invulnerability frames, death, and weapon downgrade.
 * 
 * @param {number} damage - Amount of damage to apply
 */
function hurt(damage) {
    // Shield blocks all damage
    if (ship.shield.t) {
        play(2); // Shield hit sound
        return;
    }
    
    // Invulnerability window prevents damage stacking
    if (ship.timeout < 0) {
        ship.e -= damage;
        if (ship.e < 0) ship.e = 0;
        ship.timeout = 10; // Brief invulnerability
    }
    
    // Check for death
    if (!ship.e && !l.paused) {
        // Create dramatic death explosion
        explode(ship.x, ship.y, 512);
        explode(ship.x, ship.y, 1024);
        
        l.paused = true;
        spawnText("GAME OVER", -1);
        
        // Stop the game loop
        cancelAnimationFrame(animationFrameId);
        animationFrameId = 0;
        
        play(14); // Game over sound
    } else if (ship.e < 25) {
        play(17); // Low energy warning
    }
    
    play(1); // Hit sound
    
    // Downgrade weapon on hit
    if (ship.weapon > 2) {
        ship.weapon--;
    }
    
    // Show energy bar
    ship.osd = ship.MAX_OSD;
}

/**
 * Fires a bullet from the player's ship.
 * Bullet velocity is influenced by ship's current velocity.
 * 
 * @param {number} xVel - Base X velocity
 * @param {number} yVel - Base Y velocity
 */
function fire(xVel, yVel) {
    const bullet = bullets.pool.acquire();
    if (!bullet) return; // Pool exhausted
    
    // Add ship velocity influence
    const finalXVel = xVel + ship.xAcc / 2;
    const finalYVel = yVel + ship.yAcc / 2;
    
    // Initialize bullet properties
    bullet.t = bullets.MAX_T;
    bullet.x = ship.x + Math.random() * finalXVel;
    bullet.y = ship.y + Math.random() * finalYVel;
    bullet.xAcc = finalXVel;
    bullet.yAcc = finalYVel;
}

/**
 * Creates an explosion effect at the specified position.
 * 
 * @param {number} x - Center X coordinate
 * @param {number} y - Center Y coordinate
 * @param {number} [size] - Explosion size (random if not specified)
 */
function explode(x, y, size) {
    const exp = explosions.pool.acquire();
    if (!exp) return; // Pool exhausted
    
    exp.x = x;
    exp.y = y;
    exp.size = size || Math.random() * 64;
    exp.angle = Math.random();
    exp.d = Math.random() * 0.4 - 0.2; // Rotation speed
    exp.alpha = 1;
}

/**
 * Spawns a bonus item at the specified position.
 * Bonus types: '+' (weapon), 'E' (energy), 'S' (shield), 'B' (bomb), '10' (points)
 * Uses seeded RNG for deterministic bonus drops.
 * 
 * @param {Object} source - Source object with position and velocity
 * @param {string} [type] - Bonus type (random if not specified)
 */
function spawnBonus(source, type) {
    if (!type) {
        const r = gameRNG.random();
        // 10% chance each for special bonuses, rest is money
        type = r > 0.9 ? '+' : (r > 0.8 ? 'E' : (r > 0.7 ? 'S' : (r > 0.6 ? 'B' : '10')));
    }
    
    // Lazy-render bonus sprite if not cached
    if (!bonus.images[type]) {
        bonus.images[type] = render((canvas, ctx) => {
            ctx.shadowBlur = 6;
            ctx.fillStyle = type.length > 1 ? '#FEB' : '#EFF';
            ctx.shadowColor = ctx.fillStyle;
            ctx.arc(canvas.width / 2, canvas.height / 2, canvas.width / 2 - ctx.shadowBlur, 0, Math.PI * 2);
            ctx.fill();
            
            // Draw bonus type text
            ctx.fillStyle = 'rgba(0,0,0,.5)';
            ctx.font = `bold ${canvas.width / 1.8 - type.length * 7 + 7 | 0}px sans-serif`;
            ctx.textAlign = 'center';
            ctx.textBaseline = 'middle';
            ctx.fillText(type, canvas.width / 2, canvas.height / 2);
        }, bonus.R * 2);
    }
    
    const item = bonus.pool.acquire();
    if (!item) return; // Pool exhausted
    
    item.type = type;
    item.x = source.x || l.WIDTH / 2;
    item.y = source.y || -bonus.R;
    item.xAcc = source.xAcc ? source.xAcc / 2 : gameRNG.randomFloat(-l.SPEED / 2, l.SPEED / 2);
    item.yAcc = (source.yAcc ? source.yAcc / 2 : 0) + l.SPEED / 2;
}

/**
 * Spawns an enemy torpedo projectile.
 * 
 * @param {Object} source - Source enemy with position
 * @param {number} angle - Firing angle in radians
 * @param {number} [maxAngle] - Maximum allowed angle deviation
 * @returns {boolean} Whether torpedo was spawned (false if off-screen)
 */
function spawnTorpedo(source, angle, maxAngle) {
    const y = source.y + (source.yOffset || 0) | 0;
    
    // Don't spawn if enemy is too far above screen
    if (y < -l.HEIGHT / 4) return false;
    
    // Clamp angle if maxAngle specified
    if (maxAngle) {
        if (angle > Math.PI) angle -= Math.PI * 2;
        angle = Math.max(-maxAngle, Math.min(maxAngle, angle));
    }
    
    const torpedo = torpedos.pool.acquire();
    if (!torpedo) return false; // Pool exhausted
    
    // Speed increases with level
    const speed = 3 + l.level / 2;
    
    torpedo.x = source.x | 0;
    torpedo.y = y;
    torpedo.xAcc = angle ? Math.sin(angle) * speed : 0;
    torpedo.yAcc = angle ? Math.cos(angle) * speed : speed;
    torpedo.e = 0; // Health (can be destroyed)
    
    return true;
}

/**
 * Spawns enemies of a specific type.
 * 
 * @param {number} typeIndex - Index into enemies.TYPES array
 * @param {number} [yOffset] - Vertical spawn offset
 */
function spawnEnemy(typeIndex, yOffset) {
    // Clamp to valid type range
    if (typeIndex < 0 || typeIndex >= enemies.TYPES.length) {
        typeIndex = Math.random() * enemies.TYPES.length | 0;
    }
    
    const type = enemies.TYPES[typeIndex];
    
    // Initialize type properties if needed
    if (!type.R) {
        type.R = Math.max(l.WIDTH, l.HEIGHT) / 16 | 0;
    }
    if (!type.image) {
        type.image = render(type.render, type.R * 2);
    }
    
    type.spawn(yOffset);
}

/**
 * Displays centered text message (wave announcements, pause, game over).
 * 
 * @param {string} text - Text to display
 * @param {number} [duration] - Display duration in frames (-1 for permanent)
 */
function spawnText(text, duration) {
    // Render text to off-screen canvas
    l.text.image = render((canvas, ctx) => {
        ctx.shadowBlur = canvas.height / 10 | 0;
        ctx.font = `bold ${canvas.height * 0.9 - ctx.shadowBlur * 2 | 0}px Consolas,monospace`;
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        
        const maxWidth = canvas.width - ctx.shadowBlur * 2;
        const centerX = canvas.width / 2;
        const centerY = canvas.height / 2;
        
        // Outer glow
        ctx.fillStyle = '#62F';
        ctx.shadowColor = '#FFF';
        ctx.fillText(text, centerX, centerY, maxWidth);
        ctx.fillText(text, centerX, centerY, maxWidth);
        
        // Inner stroke
        ctx.fillStyle = '#FFF';
        ctx.shadowBlur /= 4;
        ctx.lineWidth = ctx.shadowBlur;
        ctx.lineJoin = 'round';
        ctx.strokeStyle = '#FFF';
        ctx.shadowColor = '#000';
        ctx.strokeText(text, centerX, centerY, maxWidth);
        
        // Scanline effect
        ctx.globalAlpha = 0.2;
        ctx.globalCompositeOperation = 'source-atop';
        ctx.fillStyle = '#62F';
        for (let i = 0; i < canvas.height; i += 3) {
            ctx.fillRect(0, i, canvas.width, 1);
        }
    }, l.WIDTH / 1.6, l.WIDTH / 8, l.text.image);
    
    l.text.x = (l.WIDTH - l.text.image.width) / 2 | 0;
    l.text.y = 16;
    l.text.yAcc = l.SPEED / 2;
    l.text.t = duration || l.text.MAX_T;
    
    // Permanent messages (pause, game over) draw immediately
    if (duration < 0) {
        l.text.t = 0;
        a.globalAlpha = 1;
        a.drawImage(l.text.image, l.text.x, l.text.y);
    }
}

// =============================================================================
// SECTION: Pre-rendered Graphics Assets
// =============================================================================

/**
 * Renders all static game graphics to off-screen canvases.
 * Called once during initialization.
 */
function initializeGraphics() {
    // --- Score Digit Sprites (0-9) ---
    for (let number = 0; number < 10; number++) {
        l.points.images[number] = render((canvas, ctx) => {
            ctx.shadowBlur = 6;
            ctx.font = `bold ${canvas.width * 1.3 | 0}px Consolas,monospace`;
            ctx.textAlign = 'center';
            ctx.textBaseline = 'middle';
            ctx.lineWidth = 2;
            ctx.lineJoin = 'round';
            ctx.shadowColor = '#9F0';
            ctx.strokeText(number, canvas.width / 2, canvas.height / 2);
            ctx.strokeStyle = '#9F0';
            ctx.strokeText(number, canvas.width / 2, canvas.height / 2);
        }, l.points.WIDTH, l.points.HEIGHT);
    }
    
    // --- Background Tile ---
    l.background = render((canvas, ctx) => {
        ctx.fillStyle = '#000';
        ctx.fillRect(0, 0, canvas.width, canvas.height);
        ctx.globalCompositeOperation = 'lighter';
        
        // Draw grid lines
        ctx.beginPath();
        for (let i = 6; i--; ) {
            ctx.moveTo(canvas.width * (i + 1) / 4, -canvas.height);
            ctx.lineTo(canvas.width * (i - 2) / 4, canvas.height * 2);
            ctx.moveTo(-canvas.width, canvas.height * (i - 2) / 4);
            ctx.lineTo(canvas.width * 2, canvas.height * (i + 1) / 4);
        }
        ctx.lineWidth = 3;
        ctx.shadowBlur = ctx.lineWidth * 2;
        ctx.strokeStyle = '#111';
        ctx.shadowColor = '#444';
        ctx.stroke();
        
        // Mirror and overlay
        ctx.shadowBlur = 0;
        ctx.globalAlpha = 0.25;
        ctx.translate(canvas.width, 0);
        ctx.scale(-1, 1);
        ctx.drawImage(canvas, 0, 0);
    }, 256);
    
    // --- Player Ship Sprite ---
    ship.image = render((canvas, ctx) => {
        ctx.beginPath();
        for (let i = 5; i--; ) {
            ctx.moveTo(canvas.width / 2, canvas.height * (1 + i) / 10);
            ctx.lineTo(canvas.width * (11 + i) / 16, canvas.height * (15 - i) / 16);
            ctx.lineTo(canvas.width * (5 - i) / 16, canvas.height * (15 - i) / 16);
            ctx.closePath();
        }
        ctx.lineWidth = canvas.width / 17 | 0;
        ctx.shadowBlur = ctx.lineWidth * 2;
        ctx.strokeStyle = '#9F0';
        ctx.shadowColor = ctx.strokeStyle;
        ctx.stroke();
        ctx.stroke();
        
        // Center diamond
        const p = canvas.width / 6 | 0;
        ctx.beginPath();
        ctx.moveTo(canvas.width / 2 - p, canvas.height / 2);
        ctx.lineTo(canvas.width / 2, canvas.height / 2 + p);
        ctx.lineTo(canvas.width / 2 + p, canvas.height / 2);
        ctx.lineTo(canvas.width / 2, canvas.height / 2 - p);
        ctx.closePath();
        ctx.strokeStyle = '#FFF';
        ctx.shadowColor = ctx.strokeStyle;
        ctx.stroke();
        ctx.stroke();
    }, ship.R, ship.R * 2);
    
    // --- Shield Effect Sprite ---
    ship.shield.image = render((canvas, ctx) => {
        const d = 8;
        ctx.lineWidth = 18;
        ctx.shadowBlur = ctx.lineWidth;
        ctx.strokeStyle = '#000';
        ctx.shadowColor = '#CF0';
        ctx.beginPath();
        ctx.arc(canvas.width / 2, canvas.height / 2, canvas.width / 2 + ctx.lineWidth / 2 - d, 0, Math.PI * 2);
        ctx.stroke();
        
        // Black arc to clip outer shadow
        ctx.lineWidth = 26 + d;
        ctx.shadowBlur = 0;
        ctx.beginPath();
        ctx.arc(canvas.width / 2, canvas.height / 2, canvas.width / 2 + ctx.lineWidth / 2 - d, 0, Math.PI * 2);
        ctx.stroke();
    }, ship.R * 2);
    
    // --- Bullet Sprite ---
    bullets.image = render((canvas, ctx) => {
        const p = 6;
        ctx.beginPath();
        ctx.moveTo(canvas.width / 2, p);
        ctx.lineTo(canvas.width - p, canvas.height / 2);
        ctx.lineTo(canvas.width / 2, canvas.height - p);
        ctx.lineTo(p, canvas.height / 2);
        ctx.closePath();
        ctx.lineWidth = 3;
        ctx.shadowBlur = ctx.lineWidth * 2;
        ctx.strokeStyle = '#CF0';
        ctx.shadowColor = ctx.strokeStyle;
        ctx.stroke();
        ctx.stroke();
    }, bullets.R * 2);
    
    // --- Explosion Sprite ---
    explosions.image = render((canvas, ctx) => {
        ctx.fillStyle = '#F63';
        ctx.shadowBlur = 6;
        ctx.shadowColor = ctx.fillStyle;
        const p = ctx.shadowBlur;
        
        // Overlapping blurred rectangles
        for (let i = 5; i--; ) {
            ctx.fillRect(p, p, canvas.width - p * 2, canvas.height - p * 2);
        }
        
        // Tiny cross in center
        ctx.lineWidth = 0.3;
        ctx.strokeStyle = '#FC6';
        const pp = p * 0.8;
        ctx.beginPath();
        ctx.moveTo(pp, pp);
        ctx.lineTo(canvas.width - pp, canvas.height - pp);
        ctx.moveTo(canvas.width - pp, pp);
        ctx.lineTo(pp, canvas.height - pp);
        ctx.stroke();
    }, 16);
    
    // --- Torpedo Animation Frames ---
    const frameCount = 8;
    for (let i = 0; i < frameCount; i++) {
        torpedos.images.push(render((canvas, ctx) => {
            ctx.translate(canvas.width / 2, canvas.height / 2);
            ctx.rotate(Math.PI / -2 * i / frameCount);
            ctx.translate(-canvas.width / 2, -canvas.height / 2);
            
            const p = 6;
            ctx.beginPath();
            ctx.lineWidth = 3;
            ctx.shadowBlur = ctx.lineWidth * 2;
            ctx.strokeStyle = '#62F';
            ctx.shadowColor = ctx.strokeStyle;
            ctx.moveTo(canvas.width / 2, p);
            ctx.lineTo(canvas.width - p, canvas.height / 2);
            ctx.lineTo(canvas.width / 2, canvas.height - p);
            ctx.lineTo(p, canvas.height / 2);
            ctx.closePath();
            ctx.stroke();
            ctx.stroke();
        }, torpedos.R * 2));
    }
}

// =============================================================================
// SECTION: Enemy Type Definitions
// =============================================================================

/**
 * Enemy type definitions with render, spawn, and behavior functions.
 * Each type has unique visuals and attack patterns.
 * 
 * @type {Array<{render: Function, spawn: Function, shoot: Function, R?: number, image?: HTMLCanvasElement}>}
 */
enemies.TYPES = [
    // --- Type 0: Small Fighter ---
    // Fast-spawning small enemies that fire single shots
    {
        render: function(canvas, ctx) {
            ctx.lineWidth = 3;
            ctx.shadowBlur = ctx.lineWidth * 2;
            ctx.strokeStyle = '#62F';
            ctx.shadowColor = ctx.strokeStyle;
            ctx.miterLimit = 128;
            ctx.beginPath();
            
            for (let i = 5; i--; ) {
                const x1 = canvas.width * (6 - i) / 11;
                const y1 = canvas.height * (6 - i) / 20;
                const x2 = canvas.width * (11 - i) / 26;
                const y2 = canvas.height * (1 + i) / 9;
                
                ctx.moveTo(canvas.width / 2, canvas.height * (12 - i) / 12 - ctx.shadowBlur);
                ctx.lineTo(canvas.width - x1, y1);
                ctx.lineTo(canvas.width - x2, y2);
                ctx.lineTo(x2, y2);
                ctx.lineTo(x1, y1);
                ctx.closePath();
            }
            
            ctx.stroke();
            ctx.stroke();
            renderHeart(ctx, canvas.width / 2, canvas.height / 2);
        },
        
        spawn: function(yOffset) {
            for (let i = 2 + l.level; i--; ) {
                const yStop = l.HEIGHT / 8 + gameRNG.random() * l.HEIGHT / 4 | 0;
                enemies.push({
                    image: this.image,
                    x: this.R + (l.WIDTH - this.R * 2) * gameRNG.random() | 0,
                    y: yOffset ? yStop + yOffset : -this.R,
                    yStop: yStop,
                    r: this.R,
                    angle: 0,
                    maxAngle: Math.PI / 32,
                    e: 12 + l.level,
                    t: gameRNG.randomInt(0, 120),
                    shoot: this.shoot
                });
            }
        },
        
        shoot: function(angle) {
            this.t = Math.max(5, 600 / (l.level + 4) | 0);
            if (spawnTorpedo(this, angle, this.maxAngle)) {
                play(19);
            }
        }
    },
    
    // --- Type 1: Medium Fighter ---
    // Fires two torpedoes in a spread pattern
    {
        render: function(canvas, ctx) {
            ctx.lineWidth = 3;
            ctx.shadowBlur = ctx.lineWidth * 2;
            ctx.strokeStyle = '#62F';
            ctx.shadowColor = ctx.strokeStyle;
            
            for (let i = 5; i--; ) {
                const x1 = canvas.width * (5 - i) / 14 + ctx.shadowBlur;
                const y1 = canvas.height * (16 - i) / 17;
                const x2 = canvas.width * (8 - i) / 22;
                const y2 = canvas.height * (1 + i) / 11;
                
                ctx.moveTo(canvas.width / 2, canvas.height * (6 - i) / 12);
                ctx.lineTo(canvas.width - x1, y1);
                ctx.lineTo(canvas.width - x2, y2);
                ctx.lineTo(x2, y2);
                ctx.lineTo(x1, y1);
                ctx.closePath();
            }
            
            ctx.stroke();
            ctx.stroke();
            renderHeart(ctx, canvas.width / 2, canvas.height / 2);
        },
        
        spawn: function(yOffset) {
            for (let i = 1 + l.level; i--; ) {
                const yStop = l.HEIGHT / 8 + gameRNG.random() * l.HEIGHT / 4 | 0;
                enemies.push({
                    image: this.image,
                    x: this.R + (l.WIDTH - this.R * 2) * gameRNG.random() | 0,
                    y: yOffset ? yStop + yOffset : -this.R,
                    yStop: yStop,
                    r: this.R,
                    angle: 0,
                    maxAngle: Math.PI / 16,
                    e: 20 + l.level * 2,
                    t: gameRNG.randomInt(0, 120),
                    shoot: this.shoot
                });
            }
        },
        
        shoot: function(angle) {
            this.t = Math.max(5, 600 / (l.level + 4) | 0);
            
            // Clamp angle
            if (angle > Math.PI) angle -= Math.PI * 2;
            angle = Math.max(-this.maxAngle, Math.min(this.maxAngle, angle));
            
            // Fire spread pattern
            if (spawnTorpedo(this, angle + 0.1) || spawnTorpedo(this, angle - 0.1)) {
                play(20);
            }
        }
    },
    
    // --- Type 2: Turret ---
    // Fires in rotating pattern, doesn't aim at player
    {
        render: function(canvas, ctx) {
            ctx.lineWidth = 3;
            ctx.shadowBlur = ctx.lineWidth * 2;
            ctx.strokeStyle = '#62F';
            ctx.shadowColor = ctx.strokeStyle;
            ctx.miterLimit = 32;
            ctx.beginPath();
            
            // Outer spiky ring
            for (let i = 0; i < Math.PI * 2; i += Math.PI / 4) {
                let d = Math.PI / 12;
                const r = canvas.width / 2 - ctx.shadowBlur;
                let x = canvas.width / 2 + Math.sin(i + d) * r;
                let y = canvas.width / 2 + Math.cos(i + d) * r;
                
                if (!i) ctx.moveTo(x, y);
                else ctx.lineTo(x, y);
                
                d -= Math.PI / 1.45;
                ctx.lineTo(canvas.width / 2 + Math.sin(i + d) * r, canvas.width / 2 + Math.cos(i + d) * r);
            }
            ctx.closePath();
            
            // Inner octagon
            for (let i = 0; i < Math.PI * 2; i += Math.PI / 4) {
                const r = canvas.width * 0.4;
                const x = canvas.width / 2 + Math.sin(i) * r;
                const y = canvas.width / 2 + Math.cos(i) * r;
                
                if (!i) ctx.moveTo(x, y);
                else ctx.lineTo(x, y);
            }
            ctx.closePath();
            
            ctx.stroke();
            ctx.stroke();
            renderHeart(ctx, canvas.width / 2, canvas.height / 2);
        },
        
        spawn: function(yOffset) {
            const yStep = l.HEIGHT * -1.5 / l.level | 0;
            
            for (let i = l.level; i--; ) {
                enemies.push({
                    image: this.image,
                    x: l.WIDTH / 4 + l.WIDTH / 2 * gameRNG.random() | 0,
                    y: i * yStep + (yOffset || -this.R),
                    yStop: l.HEIGHT / 2 | 0,
                    r: this.R,
                    angle: 0,
                    maxAngle: Math.PI * 32,
                    e: 28 + l.level * 3,
                    t: gameRNG.randomInt(0, 30),
                    fireDirection: gameRNG.random() * Math.PI,
                    tActive: 0,
                    shoot: this.shoot
                });
            }
        },
        
        shoot: function() {
            // Burst fire with pauses
            if (!this.tActive) {
                this.tActive = 5;
                this.t = Math.max(5, 540 / (l.level + 8) | 0);
            }
            this.tActive--;
            
            this.fireDirection += 0.2;
            let result = false;
            
            // Fire in two opposite directions
            for (let i = Math.PI / 8; i < Math.PI * 2; i += Math.PI) {
                result = spawnTorpedo(this, this.fireDirection + i) || result;
            }
            
            if (result) play(18);
        }
    },
    
    // --- Type 3: Boss ---
    // Large enemy that spawns in formation with smaller copies
    {
        R: Math.max(l.WIDTH, l.HEIGHT) / 8 | 0,
        
        render: function(canvas, ctx) {
            ctx.lineWidth = 3;
            ctx.shadowBlur = ctx.lineWidth * 2;
            ctx.strokeStyle = '#62F';
            ctx.shadowColor = ctx.strokeStyle;
            ctx.miterLimit = 32;
            ctx.beginPath();
            
            for (let i = 7; i--; ) {
                ctx.moveTo(canvas.width / 2, canvas.height * i / 12 + ctx.shadowBlur);
                const x1 = canvas.width * (11 + i) / 18 - ctx.shadowBlur;
                const y1 = canvas.height * (25 - i) / 28;
                ctx.lineTo(x1, y1);
                const x2 = canvas.width * (16 - i) / 16 - ctx.shadowBlur;
                const y2 = canvas.height * (i + 4) / 28;
                ctx.lineTo(x2, y2);
                ctx.lineTo(canvas.width / 2, canvas.height * (i + 30) / 36 - ctx.shadowBlur);
                ctx.lineTo(canvas.width - x2, y2);
                ctx.lineTo(canvas.width - x1, y1);
                ctx.closePath();
            }
            
            ctx.stroke();
            ctx.stroke();
            renderHeart(ctx, canvas.width / 2, canvas.height * 0.8);
        },
        
        spawn: function(yOffset) {
            const yStart = yOffset || -this.R * 3;
            const e = 36 + l.level * 4;
            const tStart = 2 * 30;
            
            // Main boss
            enemies.push({
                image: this.image,
                x: l.WIDTH / 2,
                y: yStart,
                yOffset: this.R * 0.6 | 0,
                yStop: this.R + 8,
                r: this.R,
                angle: 0,
                maxAngle: Math.PI / 8,
                e: e,
                t: tStart + gameRNG.randomInt(0, 30),
                shoot: this.shoot
            });
            
            // Flanking sub-bosses
            let size = this.R * 1.3;
            let d = size * 0.4;
            let x = d;
            
            for (let i = 1; i < l.level; i++) {
                let y = this.R + 16 - size + Math.sqrt(i) * 64;
                if (i % 2) y += this.R * (0.7 / i + 0.3);
                
                // Left and right flankers
                for (const xMult of [-1, 1]) {
                    enemies.push({
                        image: this.image,
                        x: l.WIDTH / 2 + x * xMult | 0,
                        y: yStart,
                        yOffset: size / 2 * 0.6 | 0,
                        yStop: y | 0,
                        r: size / 2,
                        angle: 0,
                        maxAngle: Math.PI / 8,
                        e: e / 2 | 0,
                        t: tStart + gameRNG.randomInt(0, 30),
                        shoot: this.shoot
                    });
                }
                
                x += d;
                d *= 0.84;
                size *= 0.9;
            }
        },
        
        shoot: function(angle) {
            // Deterministic timing based on level seed, aimed at player
            this.t = gameRNG.randomInt(0, 6 * 30);
            if (spawnTorpedo(this, angle)) {
                play(12);
            }
        }
    }
];

// =============================================================================
// SECTION: Object Pool Initialization
// =============================================================================

/**
 * Initializes all object pools with appropriate sizes.
 */
function initializePools() {
    // Bullet pool - player can fire rapidly
    bullets.pool = new ObjectPool(200, () => ({
        t: 0,
        x: 0,
        y: 0,
        xAcc: 0,
        yAcc: 0,
        _poolIndex: 0
    }));
    
    // Torpedo pool - enemies fire these
    torpedos.pool = new ObjectPool(150, () => ({
        x: 0,
        y: 0,
        xAcc: 0,
        yAcc: 0,
        e: 0,
        _poolIndex: 0
    }));
    
    // Explosion pool - visual effects
    explosions.pool = new ObjectPool(50, () => ({
        x: 0,
        y: 0,
        size: 0,
        angle: 0,
        d: 0,
        alpha: 0,
        _poolIndex: 0
    }));
    
    // Bonus item pool
    bonus.pool = new ObjectPool(30, () => ({
        type: '',
        x: 0,
        y: 0,
        xAcc: 0,
        yAcc: 0,
        _poolIndex: 0
    }));
}

// =============================================================================
// SECTION: Input Handlers
// =============================================================================

/**
 * Keyboard down event handler.
 * Maps alternative control keys and handles special actions.
 * 
 * @param {KeyboardEvent} e - Keyboard event
 */
document.onkeydown = function(e) {
    const code = (e || event).keyCode;
    const mappedCode = keyMap[code] || code;
    keys.add(mappedCode);
    
    // Initialize audio on first interaction (browser autoplay policy)
    if (!audioManager.ctx) {
        audioManager.init();
    }
    
    // F key - Toggle fullscreen
    if (keys.has(70)) {
        toggleFullscreen();
    }
    // M key - Easter egg (custom ship image)
    else if (keys.has(77)) {
        if (ship.originalImage) {
            ship.image = ship.originalImage;
            ship.originalImage = null;
        } else {
            const image = new Image();
            image.onload = function() {
                ship.originalImage = ship.image;
                ship.image = image;
            };
            image.src = 'starship-sorades.jpg';
        }
    }
    // P key - Pause (only when alive)
    else if (keys.has(80) && ship.e) {
        l.paused = !l.paused;
        if (l.paused && !l.text.t) {
            spawnText('PAUSE', -1);
        }
    }
    // X key - Unpause
    else if (l.paused && keys.has(88)) {
        l.paused = false;
    }
};

/**
 * Keyboard up event handler.
 * 
 * @param {KeyboardEvent} e - Keyboard event
 */
document.onkeyup = function(e) {
    const code = (e || event).keyCode;
    const mappedCode = keyMap[code] || code;
    keys.delete(mappedCode);
};

// =============================================================================
// SECTION: Main Game Loop
// =============================================================================

/**
 * Main game loop using requestAnimationFrame.
 * Uses fixed timestep for consistent gameplay regardless of display refresh rate.
 * 
 * @param {DOMHighResTimeStamp} currentTime - Current timestamp from rAF
 */
function gameLoopRAF(currentTime) {
    // Schedule next frame
    animationFrameId = requestAnimationFrame(gameLoopRAF);
    
    // Fixed timestep - only update if enough time has passed
    if (currentTime - lastFrameTime < FRAME_DURATION) return;
    lastFrameTime = currentTime;
    
    // Skip update if paused
    if (l.paused) return;
    
    gameloop();
}

/**
 * Core game logic - updates all game objects and renders frame.
 */
function gameloop() {
    // =========================================================================
    // Player Input Processing
    // =========================================================================
    
    // Fire weapon
    if (--ship.reload <= 0 && keys.has(88)) {
        ship.reload = ship.weapon ? 4 : 6;
        fire(0, -16);
        
        if (ship.weapon > 1) {
            fire(-8, -8);
            fire(8, -8);
            
            if (ship.weapon > 2) {
                fire(0, 16);
                
                if (ship.weapon > 3) {
                    fire(-16, 0);
                    fire(16, 0);
                }
            }
        }
        
        play(ship.weapon > 2 ? 16 : ship.weapon > 0 ? 0 : 15);
    }
    
    // Movement input with visual tilt
    ship.angle *= ship.ANGLE_FACTOR;
    
    // Left arrow
    if (keys.has(37)) {
        if (ship.x >= l.WIDTH && ship.xAcc > 0) ship.xAcc = 0;
        ship.xAcc -= ship.ACC;
        ship.angle = (ship.angle + 1) * ship.ANGLE_FACTOR - 1;
    }
    // Up arrow
    if (keys.has(38)) {
        if (ship.y >= l.HEIGHT && ship.yAcc > 0) ship.yAcc = 0;
        ship.yAcc -= ship.ACC;
    }
    // Right arrow
    if (keys.has(39)) {
        if (ship.x < 0 && ship.xAcc < 0) ship.xAcc = 0;
        ship.xAcc += ship.ACC;
        ship.angle = (ship.angle - 1) * ship.ANGLE_FACTOR + 1;
    }
    // Down arrow
    if (keys.has(40)) {
        if (ship.y < 0 && ship.yAcc < 0) ship.yAcc = 0;
        ship.yAcc += ship.ACC;
    }
    
    // Screen boundary collision
    if (ship.x < 0 && ship.xAcc < 0) ship.x = 0;
    else if (ship.x >= l.WIDTH && ship.xAcc > 0) ship.x = l.WIDTH - 1;
    if (ship.y < 0 && ship.yAcc < 0) ship.y = 0;
    else if (ship.y >= l.HEIGHT && ship.yAcc > 0) ship.y = l.HEIGHT - 1;
    
    // Apply velocity and damping
    ship.x += ship.xAcc;
    ship.y += ship.yAcc;
    ship.xAcc *= ship.ACC_FACTOR;
    ship.yAcc *= ship.ACC_FACTOR;
    
    // =========================================================================
    // Background Rendering (using createPattern for efficiency)
    // =========================================================================
    
    a.save();
    a.translate(0, l.y % l.background.width);
    a.fillStyle = l.backgroundPattern;
    a.fillRect(0, -l.background.width, l.WIDTH, l.HEIGHT + l.background.width);
    a.restore();
    l.y += l.SPEED;
    
    // =========================================================================
    // Score Display
    // =========================================================================
    
    let points = l.p;
    let scoreX = l.WIDTH - l.points.WIDTH - 8;
    while (points) {
        a.drawImage(l.points.images[points % 10], scoreX, 4);
        points = points / 10 | 0;
        scoreX -= l.points.STEP;
    }
    
    // =========================================================================
    // Text Display (wave announcements, etc.)
    // =========================================================================
    
    if (l.text.t) {
        a.globalAlpha = l.text.t < l.text.MAX_T ? l.text.t / l.text.MAX_T : 1;
        a.drawImage(l.text.image, l.text.x, l.text.y);
        a.globalAlpha = 1;
        l.text.t--;
        l.text.y += l.text.yAcc;
    }
    
    // =========================================================================
    // Bullet Update and Rendering
    // =========================================================================
    
    const bulletTorpedoCollisionDist = 12;
    
    bullets.pool.forEachReverse((bullet, bulletIdx) => {
        // Render bullet (avoid sub-pixel rendering)
        a.drawImage(bullets.image, bullet.x - bullets.R | 0, bullet.y - bullets.R | 0);
        
        // Update position
        bullet.x += bullet.xAcc;
        bullet.y += bullet.yAcc;
        
        // Check collision with torpedoes
        torpedos.pool.forEachReverse((torpedo, torpedoIdx) => {
            if (bullet.y < torpedo.y + bulletTorpedoCollisionDist &&
                bullet.y > torpedo.y - bulletTorpedoCollisionDist &&
                bullet.x < torpedo.x + bulletTorpedoCollisionDist &&
                bullet.x > torpedo.x - bulletTorpedoCollisionDist) {
                
                if (--torpedo.e < 0) {
                    // Torpedo destroyed
                    l.p += 5;
                    // Use seeded RNG for deterministic bonus drops
                    if (gameRNG.random() > 0.75) spawnBonus(torpedo);
                    explode(torpedo.x, torpedo.y);
                    torpedos.pool.release(torpedoIdx);
                    play(11);
                }
                bullet.t = 0;
            }
        });
        
        // Remove expired or off-screen bullets
        if (--bullet.t < 0 || bullet.x < -bullets.R ||
            bullet.x >= l.WIDTH + bullets.R || bullet.y >= l.HEIGHT + bullets.R) {
            bullets.pool.release(bulletIdx);
        }
    });
    
    // =========================================================================
    // Screen Flash Effect (bomb)
    // =========================================================================
    
    if (l.bomb) {
        a.fillStyle = `rgba(255,255,255,${l.bomb-- / l.MAX_BOMB / 2})`;
        a.fillRect(0, 0, l.WIDTH, l.HEIGHT);
    }
    
    // =========================================================================
    // Enable Additive Blending for Effects
    // =========================================================================
    
    a.globalCompositeOperation = 'lighter';
    
    // Collision distances
    const shipCollisionD = ship.R * 0.8 | 0;
    const shipCollisionE = ship.R * 0.4 | 0;
    
    // =========================================================================
    // Bonus Item Update and Rendering
    // =========================================================================
    
    bonus.pool.forEachReverse((item, idx) => {
        // Check collision with player
        if (ship.y < item.y + shipCollisionD && ship.y > item.y - shipCollisionD &&
            ship.x < item.x + shipCollisionE && ship.x > item.x - shipCollisionE) {
            
            l.p += 10;
            
            switch (item.type) {
                case '+': // Weapon upgrade
                    if (ship.weapon < 4) {
                        ship.weapon++;
                        play(5);
                    } else {
                        play(6);
                    }
                    break;
                    
                case 'E': // Energy
                    if (ship.e < 100) {
                        ship.osd = ship.MAX_OSD;
                        play(5);
                    } else {
                        play(6);
                    }
                    ship.e = Math.min(100, ship.e + 5);
                    break;
                    
                case 'S': // Shield
                    ship.shield.t += ship.shield.MAX_T * ship.shield.MAX_T *
                        2 / (ship.shield.t + ship.shield.MAX_T * 2) | 0;
                    play(3);
                    break;
                    
                case 'B': // Bomb
                    // Damage all enemies
                    for (let j = enemies.length; j--; ) {
                        enemies[j].e--;
                    }
                    // Explode some torpedoes
                    for (let j = Math.min(torpedos.pool.activeCount, 5); j--; ) {
                        explode(torpedos.pool.pool[j].x, torpedos.pool.pool[j].y);
                    }
                    torpedos.pool.clear();
                    l.bomb = l.MAX_BOMB;
                    play(13);
                    break;
                    
                default: // Money
                    play(7);
            }
            
            bonus.pool.release(idx);
            return;
        }
        
        // Render bonus
        a.drawImage(bonus.images[item.type], item.x - bonus.R | 0, item.y - bonus.R | 0);
        
        // Update position
        item.x += item.xAcc;
        item.y += item.yAcc;
        
        // Remove off-screen items
        if (item.y >= l.HEIGHT + bonus.R * 2 || item.x < -bonus.R ||
            item.x >= l.WIDTH + bonus.R || item.y < -bonus.R) {
            bonus.pool.release(idx);
        }
    });
    
    // =========================================================================
    // Torpedo Update and Rendering
    // =========================================================================
    
    torpedos.pool.forEachReverse((torpedo, idx) => {
        // Check collision with player
        if (ship.y < torpedo.y + shipCollisionD && ship.y > torpedo.y - shipCollisionD &&
            ship.x < torpedo.x + shipCollisionE && ship.x > torpedo.x - shipCollisionE) {
            hurt(10);
            explode(torpedo.x, torpedo.y);
            torpedos.pool.release(idx);
            return;
        }
        
        // Render torpedo with animation
        a.drawImage(torpedos.images[torpedos.frame],
            torpedo.x - torpedos.R | 0, torpedo.y - torpedos.R | 0);
        
        // Update position
        torpedo.x += torpedo.xAcc;
        torpedo.y += torpedo.yAcc;
        
        // Remove off-screen torpedoes
        if (torpedo.y >= l.HEIGHT + torpedos.R || torpedo.x < -torpedos.R ||
            torpedo.x >= l.WIDTH + torpedos.R || torpedo.y < -l.HEIGHT) {
            torpedos.pool.release(idx);
        }
    });
    
    // Advance torpedo animation
    torpedos.frame = (torpedos.frame + 1) % torpedos.images.length;
    ship.timeout--;
    
    // =========================================================================
    // Explosion Update and Rendering
    // =========================================================================
    
    explosions.pool.forEachReverse((exp, idx) => {
        a.save();
        a.globalAlpha = exp.alpha;
        a.translate(exp.x, exp.y);
        a.rotate(exp.angle);
        a.drawImage(explosions.image, -exp.size / 2, -exp.size / 2, exp.size, exp.size);
        a.restore();
        
        // Animate explosion
        exp.size += 16;
        exp.angle += exp.d;
        exp.alpha -= 0.1;
        
        // Remove faded explosions
        if (exp.alpha < 0.1) {
            explosions.pool.release(idx);
        }
    });
    
    // =========================================================================
    // Player Ship Rendering
    // =========================================================================
    
    a.save();
    a.translate(ship.x, ship.y);
    a.rotate(ship.angle * ship.MAX_ANGLE / 180 * Math.PI);
    a.drawImage(ship.image, -ship.R / 2 | 0, -ship.R);
    a.restore();
    
    // Shield effect
    if (ship.shield.t) {
        if (ship.shield.t > 30 || Math.random() > 0.5) {
            a.drawImage(ship.shield.image, ship.x - ship.R + 0.5 | 0, ship.y - ship.R + 0.5 | 0);
        }
        if (!--ship.shield.t) {
            play(4); // Shield depleted sound
        }
    }
    
    // =========================================================================
    // Wave Spawning
    // =========================================================================
    
    if (!enemies.length && !l.text.t) {
        l.p += (l.level || 0) * 1000;
        l.level = (l.level || 0) + 1;
        
        // Generate deterministic seed for this level
        l.levelSeed = SeededRNG.levelSeed(gameSeed, l.level);
        gameRNG.setSeed(l.levelSeed);
        
        // Spawn all enemy types with vertical offsets
        // Enemy spawning now uses gameRNG for deterministic placement
        spawnEnemy(0, -0.75 * l.HEIGHT);
        spawnEnemy(1, -1.5 * l.HEIGHT);
        spawnEnemy(2, -1 * l.HEIGHT);
        spawnEnemy(3, -2.25 * l.HEIGHT);
        
        spawnText('WAVE ' + l.level);
        l.bomb = l.MAX_BOMB;
        play(8);
    }
    
    // =========================================================================
    // Enemy Update and Rendering
    // =========================================================================
    
    enemyLoop:
    for (let i = enemies.length - 1; i >= 0; i--) {
        const enemy = enemies[i];
        const enemyY = enemy.y + (enemy.yOffset || 0);
        
        // Calculate angle to player for aiming
        let angle = Math.atan((enemy.x - ship.x) / (enemyY - ship.y));
        if (ship.y <= enemyY) angle += Math.PI;
        
        // Calculate visual rotation (smoothed, clamped)
        let bossAngle = (angle + Math.PI) % (Math.PI * 2) - Math.PI;
        const maxAngle = enemy.maxAngle || 0;
        bossAngle = Math.max(-maxAngle, Math.min(maxAngle, bossAngle));
        enemy.angle = (enemy.angle * 29 - bossAngle) / 30;
        
        // Render enemy
        a.save();
        a.translate(enemy.x, enemyY);
        a.rotate(enemy.angle);
        a.drawImage(enemy.image, -enemy.r, enemy.y - enemyY - enemy.r,
            enemy.r * 2, enemy.r * 2);
        a.restore();
        
        // Check bullet collisions
        const hitboxD = enemy.r * 0.6;
        
        bullets.pool.forEachReverse((bullet, bulletIdx) => {
            if (bullet.y < enemy.y + hitboxD && bullet.y > enemy.y - hitboxD &&
                bullet.x > enemy.x - hitboxD && bullet.x < enemy.x + hitboxD) {
                
                l.p += 1;
                explode(bullet.x, bullet.y);
                bullets.pool.release(bulletIdx);
                
                // Damage enemy
                if (--enemy.e <= 0) {
                    // Enemy destroyed
                    l.p += 100;
                    spawnBonus(enemy);
                    explode(enemy.x, enemy.y, enemy.r * 2);
                    explode(enemy.x, enemy.y, enemy.r * 3);
                    
                    // Swap-and-pop for enemy array
                    enemies[i] = enemies[enemies.length - 1];
                    enemies.pop();
                    
                    play(10);
                } else {
                    play(9);
                }
            }
        });
        
        // Skip if enemy was destroyed
        if (i >= enemies.length) continue;
        
        // Move enemy toward stop position
        if (enemy.y < enemy.yStop) {
            enemy.y += enemy.yAcc || 1;
        }
        
        // Fire at player
        if (--enemy.t < 0) {
            enemy.shoot(angle);
        }
    }
    
    // =========================================================================
    // Disable Additive Blending
    // =========================================================================
    
    a.globalCompositeOperation = 'source-over';
    
    // =========================================================================
    // Energy Bar HUD
    // =========================================================================
    
    if (ship.osd) {
        const barX = ship.x - 31.5 | 0;
        const barY = ship.y + 62.5 | 0;
        const colorValue = ship.e * 512 / 100 | 0;
        
        a.globalAlpha = ship.osd / ship.MAX_OSD;
        a.fillStyle = '#000';
        a.fillRect(barX, barY, 64, 4);
        
        // Color gradient from red to green based on health
        const r = colorValue > 255 ? 512 - colorValue : 255;
        const g = colorValue > 255 ? 255 : colorValue;
        a.fillStyle = `rgb(${r},${g},0)`;
        a.fillRect(barX, barY, ship.e * 64 / 100 | 0, 4);
        
        a.lineWidth = 0.5;
        a.strokeStyle = '#FFF';
        a.strokeRect(barX, barY, 64, 4);
        a.globalAlpha = 1;
        
        if (ship.e >= 25) ship.osd--;
    }
}

// =============================================================================
// SECTION: Sound Effect Definitions
// =============================================================================

/**
 * Sound effect parameter strings for jsfxr.
 * Format: "poolSize|sfxrParams"
 * 
 * @type {string[]}
 */
const sfxData = [
    // 0 = Player shoots (basic)
    '0,,.167,.1637,.1361,.7212,.0399,-.363,,,,,,.1314,.0517,,.0154,-.1633,1,,,.0515,,.2',
    // 1 = Player is hurt
    '3,.0704,.0462,.3388,.4099,.1599,,.0109,-.3247,.0006,,-.1592,.4477,.1028,.1787,,-.0157,-.3372,.1896,.1628,,.0016,-.0003,.5',
    // 2 = Shield hit
    '3,.1,.3899,.1901,.2847,.0399,,.0007,.1492,,,-.9636,,,-.3893,.1636,-.0047,.7799,.1099,-.1103,.5924,.484,.1547,1',
    // 3 = Shield activated
    '1,,.0398,,.4198,.3891,,.4383,,,,,,,,.616,,,1,,,,,.5',
    // 4 = Shield deactivated
    '1,.1299,.27,.1299,.4199,.1599,,.4383,,,,-.6399,,,-.4799,.7099,,,1,,,,,.5',
    // 5 = Weapon upgrade collected
    '0,.43,.1099,.67,.4499,.6999,,-.2199,-.2,.5299,.5299,-.0399,.3,,.0799,.1899,-.1194,.2327,.8815,-.2364,.43,.2099,-.5799,.5',
    // 6 = Non-applicable bonus collected
    '0,.2,.1099,.0733,.0854,.14,,-.1891,.36,,,.9826,,,.4642,,-.1194,.2327,.8815,-.2364,.0992,.0076,.2,.5',
    // 7 = Money collected
    '0,.09,.1099,.0733,.0854,.1099,,-.1891,.827,,,.9826,,,.4642,,-.1194,.2327,.8815,-.2364,.0992,.0076,.8314,.5',
    // 8 = New wave alarm
    '1,.1,1,.1901,.2847,.3199,,.0007,.1492,,,-.9636,,,-.3893,.1636,-.0047,.6646,.9653,-.1103,.5924,.484,.1547,.6',
    // 9 = Enemy hit
    '3,.1,.3899,.1901,.2847,.0399,,.0007,.1492,,,-.9636,,,-.3893,.1636,-.0047,.6646,.9653,-.1103,.5924,.484,.1547,.4',
    // 10 = Enemy destroyed
    '3,.2,.1899,.4799,.91,.0599,,-.2199,-.2,.5299,.5299,-.0399,.3,,.0799,.1899,-.1194,.2327,.8815,-.2364,.43,.2099,-.5799,.5',
    // 11 = Torpedo destroyed
    '3,,.3626,.5543,.191,.0731,,-.3749,,,,,,,,,,,1,,,,,.4',
    // 12 = Boss shoots
    '1,.071,.3474,.0506,.1485,.5799,.2,-.2184,-.1405,.1681,,-.1426,,.9603,-.0961,,.2791,-.8322,.2832,.0009,,.0088,-.0082,.3',
    // 13 = Bomb explosion
    '3,.05,.3365,.4591,.4922,.1051,,.015,,,,-.6646,.7394,,,,,,1,,,,,.7',
    // 14 = Game over
    '1,1,.09,.5,.4111,.506,.0942,.1499,.0199,.8799,.1099,-.68,.0268,.1652,.62,.6999,-.0399,.4799,.5199,-.0429,.0599,.8199,-.4199,.7',
    // 15 = Player shoots (weak weapon)
    '2,,.1199,.15,.1361,.5,.0399,-.363,-.4799,,,,,.1314,.0517,,.0154,-.1633,1,,,.0515,,.2',
    // 16 = Player shoots (strong weapon)
    '2,,.98,.4699,.07,.7799,.0399,-.28,-.4799,.2399,.1,,.36,.1314,.0517,,.0154,-.1633,1,,.37,.0399,.54,.1',
    // 17 = Low energy warning
    '0,.9705,.0514,.5364,.5273,.4816,.0849,.1422,.205,.7714,.1581,-.7685,.0822,.2147,.6062,.7448,-.0917,.4009,.6251,.1116,.0573,.9005,-.3763,.3',
    // 18 = Turret shoots
    '0,.0399,.1362,.0331,.2597,.85,.0137,-.3976,,,,,,.2099,-.72,,,,1,,,,,.3',
    // 19 = Small enemy shoots
    '0,,.2863,,.3048,.751,.2,-.316,,,,,,.4416,.1008,,,,1,,,.2962,,.3',
    // 20 = Medium enemy shoots
    '0,,.3138,,.0117,.7877,.1583,-.3391,-.04,,.0464,.0585,,.4085,-.4195,,-.024,-.0396,1,-.0437,.0124,.02,.0216,.3',
    // 21 = Intro sound
    '0,1,.8799,.3499,.17,.61,.1899,-.3,-.18,.3,.6399,-.0279,.0071,.8,-.1599,.5099,-.46,.5199,.25,.0218,.49,.4,-.2,.3'
];

// =============================================================================
// SECTION: Game Initialization
// =============================================================================

/**
 * Initializes and starts the game.
 * Sets up canvas, loads sounds, renders graphics, and starts game loop.
 */
async function initGame() {
    // Get canvas and context
    c = document.querySelector('canvas');
    c.width = l.WIDTH;
    c.height = l.HEIGHT;
    a = c.getContext('2d');
    
    // Show loading screen
    a.fillStyle = '#000';
    a.fillRect(0, 0, c.width, c.height);
    spawnText('LOADING', -1);
    
    // Initialize audio manager
    await audioManager.init();
    
    // Load all sound effects
    const loadPromises = sfxData.map((params, idx) => {
        const dataUrl = jsfxr(params);
        return audioManager.loadSound(idx, dataUrl);
    });
    
    // Wait for critical sounds to load (don't block on all)
    await Promise.race([
        Promise.all(loadPromises),
        new Promise(resolve => setTimeout(resolve, 2000)) // 2s timeout
    ]);
    
    // Initialize object pools
    initializePools();
    
    // Render all graphics
    initializeGraphics();
    
    // Create background pattern for efficient rendering
    l.backgroundPattern = a.createPattern(l.background, 'repeat');
    
    // Initialize game seed (can be set externally for multiplayer/replays)
    gameSeed = Date.now();
    gameRNG.setSeed(gameSeed);
    
    // Show title
    spawnText('SORADES 13K', 6 * 30);
    
    // Start game loop using requestAnimationFrame
    lastFrameTime = performance.now();
    animationFrameId = requestAnimationFrame(gameLoopRAF);
    
    // Play intro sound
    play(21);
}

/**
 * Sets the game seed for deterministic gameplay.
 * Call before starting a new game for reproducible levels.
 * 
 * @param {number} seed - Seed value for the game session
 */
function setGameSeed(seed) {
    gameSeed = seed >>> 0;
    gameRNG.setSeed(gameSeed);
    l.level = 0; // Reset level counter
}

/**
 * Gets the current game seed.
 * Useful for sharing seeds in multiplayer or saving replays.
 * 
 * @returns {number} Current game seed
 */
function getGameSeed() {
    return gameSeed;
}

/**
 * Gets the current level's seed.
 * 
 * @returns {number} Current level seed
 */
function getLevelSeed() {
    return l.levelSeed;
}

// Export seed functions for external access
window['setGameSeed'] = setGameSeed;
window['getGameSeed'] = getGameSeed;
window['getLevelSeed'] = getLevelSeed;

// Start initialization after a brief delay to ensure DOM is ready
// and to allow the "LOADING" text to render
setTimeout(initGame, 0);