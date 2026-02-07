package game

// Theme holds all visual styling constants for easy customization.
var Theme = struct {
	// Background colors
	BackgroundColor     string
	BackgroundLineColor string
	BackgroundGlow      string

	// Player ship colors
	ShipColor       string
	ShipGlow        string
	ShipCenterColor string

	// Shield colors
	ShieldColor     string
	ShieldGlowColor string

	// Bullet/projectile colors
	BulletColor string
	BulletGlow  string

	// Torpedo colors
	TorpedoColor string
	TorpedoGlow  string

	// Enemy colors
	EnemyColor string
	EnemyGlow  string

	// Explosion colors
	ExplosionColor     string
	ExplosionGlow      string
	ExplosionLineColor string

	// UI/HUD colors
	ScoreColor          string
	ScoreGlow           string
	TextPrimaryColor    string
	TextSecondaryColor  string
	TextGlow            string
	TextScanlineColor   string
	EnergyBarBackground string
	EnergyBarBorder     string

	// Bonus item colors
	BonusColorPoints  string
	BonusColorPowerup string
	BonusTextColor    string

	// Screen flash color
	BombFlashColor string

	// Fonts
	ScoreFont    string
	TextFont     string
	BonusFont    string
	InstructFont string

	// Line widths
	ShipLineWidth      float64
	EnemyLineWidth     float64
	BulletLineWidth    float64
	TorpedoLineWidth   float64
	EnergyBarLineWidth float64

	// Shadow/glow blur values
	DefaultShadowBlur   float64
	ShipShadowBlur      float64
	ShieldShadowBlur    float64
	BulletShadowBlur    float64
	ExplosionShadowBlur float64
}{
	// Background colors - dark space theme
	BackgroundColor:     "#000",
	BackgroundLineColor: "#111",
	BackgroundGlow:      "#444",

	// Player ship colors - green/lime theme
	ShipColor:       "#9F0",
	ShipGlow:        "#9F0",
	ShipCenterColor: "#FFF",

	// Shield colors - bright green
	ShieldColor:     "#000",
	ShieldGlowColor: "#CF0",

	// Bullet/projectile colors - yellow-green
	BulletColor: "#CF0",
	BulletGlow:  "#CF0",

	// Torpedo colors - purple/violet
	TorpedoColor: "#62F",
	TorpedoGlow:  "#62F",

	// Enemy colors - purple/violet (same as torpedoes)
	EnemyColor: "#62F",
	EnemyGlow:  "#62F",

	// Explosion colors - orange/red
	ExplosionColor:     "#F63",
	ExplosionGlow:      "#F63",
	ExplosionLineColor: "#FC6",

	// UI/HUD colors
	ScoreColor:          "#9F0",
	ScoreGlow:           "#9F0",
	TextPrimaryColor:    "#62F",
	TextSecondaryColor:  "#FFF",
	TextGlow:            "#FFF",
	TextScanlineColor:   "#62F",
	EnergyBarBackground: "#000",
	EnergyBarBorder:     "#FFF",

	// Bonus item colors
	BonusColorPoints:  "#FEB",
	BonusColorPowerup: "#EFF",
	BonusTextColor:    "rgba(0,0,0,.5)",

	// Screen flash color
	BombFlashColor: "rgba(255,255,255,",

	// Fonts
	ScoreFont:    "Consolas,monospace",
	TextFont:     "Consolas,monospace",
	BonusFont:    "sans-serif",
	InstructFont: "16px sans-serif",

	// Line widths
	ShipLineWidth:      3.0,
	EnemyLineWidth:     3.0,
	BulletLineWidth:    3.0,
	TorpedoLineWidth:   3.0,
	EnergyBarLineWidth: 0.5,

	// Shadow/glow blur values
	DefaultShadowBlur:   6.0,
	ShipShadowBlur:      6.0,
	ShieldShadowBlur:    18.0,
	BulletShadowBlur:    6.0,
	ExplosionShadowBlur: 6.0,
}
