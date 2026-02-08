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

	// Base shield colors
	BaseShieldColor     string
	BaseShieldGlowColor string
	BaseShieldBorder    string
	BaseColor           string

	// Bullet/projectile colors
	BulletColor string
	BulletGlow  string

	// Torpedo colors
	TorpedoColor string
	TorpedoGlow  string

	// Enemy colors - Apple silver/white theme (default)
	EnemyColor string
	EnemyGlow  string

	// Enemy type-specific colors (Apple brand history)
	// SmallFighter: 1984-1998 Rainbow Apple era
	EnemySmallColor string
	EnemySmallGlow  string
	// MediumFighter: 1998-2001 Bondi Blue iMac era
	EnemyMediumColor string
	EnemyMediumGlow  string
	// TurretFighter: 2001-2007 iPod era (white/chrome)
	EnemyTurretColor string
	EnemyTurretGlow  string
	// Boss: 2013+ Modern era (Space Gray/Product RED)
	EnemyBossColor  string
	EnemyBossGlow   string
	EnemyBossAccent string

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

	// Player ship colors - Vipps MobilePay orange theme
	ShipColor:       "#FF5B24",
	ShipGlow:        "#FF5B24",
	ShipCenterColor: "#FFF",

	// Shield colors - Vipps orange
	ShieldColor:     "#000",
	ShieldGlowColor: "#FF7A4D",

	// Base shield colors - MobilePay purple/violet theme
	BaseShieldColor:     "rgba(102, 34, 255, 0.15)",
	BaseShieldGlowColor: "#62F",
	BaseShieldBorder:    "rgba(102, 34, 255, 0.6)",
	BaseColor:           "#62F",

	// Bullet/projectile colors - Vipps orange
	BulletColor: "#FF5B24",
	BulletGlow:  "#FF7A4D",

	// Torpedo colors - Apple silver/gray
	TorpedoColor: "#A2AAAD",
	TorpedoGlow:  "#C0C0C0",

	// Enemy colors - Apple silver/white theme (default)
	EnemyColor: "#E0E0E0",
	EnemyGlow:  "#C0C0C0",

	// SmallFighter: 1984-1998 Rainbow Apple era (green from rainbow)
	EnemySmallColor: "#5AC94D",
	EnemySmallGlow:  "#6EDB5E",
	// MediumFighter: 1998-2001 Bondi Blue iMac era
	EnemyMediumColor: "#0095D9",
	EnemyMediumGlow:  "#00B4FF",
	// TurretFighter: 2001-2007 iPod era (white/chrome)
	EnemyTurretColor: "#F5F5F7",
	EnemyTurretGlow:  "#FFFFFF",
	// Boss: 2013+ Modern era (Space Gray with Product RED)
	EnemyBossColor:  "#86868B",
	EnemyBossGlow:   "#A1A1A6",
	EnemyBossAccent: "#FF3B30",

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
