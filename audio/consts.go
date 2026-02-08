package audio

var AudioConfig = Config{
	// Master settings
	MasterVolume: 0.7,
	ReverbMix:    0.3,
	ReverbTime:   0.8,
	ReverbDecay:  0.6,

	// Synth music settings
	SynthVolume:     0.15,
	SynthReverbSend: 0.4,

	// Drone pad settings
	DronePadVolume:  0.15,
	DroneFilterBase: 400,
	DroneFilterMod:  200,
	DroneLFORate:    0.05,
	DroneDetune:     10,
	SubBassVolume:   0.3,

	// Bass line settings
	BassVolume:     0.2,
	BassSubVolume:  0.25,
	BassDetune:     8,
	BassFilterBase: 300,
	BassFilterPeak: 600,
	BassFilterQ:    4,
	BassLFORate:    0.08,
	BassGlideTime:  0.08,
	BassHoldChance: 0.2,
	BassSkipChance: 0.5,

	// Arpeggio settings
	ArpAttackTime:    0.1,
	ArpPeakVolume:    0.12,
	ArpSustainVolume: 0.08,
	ArpReleaseTime:   1.5,
	ArpFilterCutoff:  800,
	ArpDetune:        0.01,
	ArpRandomChance:  0.3,
	ArpDelayMs:       2000,

	// Reactive layer master settings
	ReactiveVolume:     0.35,
	ReactiveReverbSend: 0.7,

	// Tension layer settings
	TensionFilterBase: 200,
	TensionFilterQ:    8,
	TensionPerEnemy:   0.012,
	TensionMaxVolume:  0.15,
	TensionFreqMod:    180,

	// Shimmer layer settings
	ShimmerFilterFreq:   2000,
	ShimmerFilterQ:      0.5,
	ShimmerPerEnemy:     0.005,
	ShimmerMaxVolume:    0.06,
	ShimmerVibratoRate:  0.15,
	ShimmerVibratoDepth: 2.0,

	// Danger siren settings
	SirenThreatThreshold: 0.7,
	SirenMaxVolume:       0.06,
	SirenFreqMod:         150,
	SirenVibratoRate:     2.5,
	SirenVibratoDepth:    20,
	SirenFilterCutoff:    1200,
	SirenFilterQ:         3,

	// Sub bass reactive settings
	SubBassReactiveMax: 0.2,
	SubBassFreqMod:     8,
	SubBassLFORate:     0.0625,
	SubBassLFODepth:    1.5,

	// Atmospheric pad settings
	PadFilterBase:  400,
	PadFilterQ:     2,
	PadPerEnemy:    0.008,
	PadMaxVolume:   0.1,
	PadDetuneRange: 6,

	// Pulse layer settings
	PulseFilterCutoff:    250,
	PulseLFODepth:        0.4,
	PulseMinVolume:       0.008,
	PulseMaxVolume:       0.1,
	PulseThreatThreshold: 0.25,
}

// LevelMusicPresets contains presets for each level style
var LevelMusicPresets = map[int]*LevelMusicPreset{
	// Level 1: A minor - mysterious, cinematic (Blade Runner)
	1: {
		Name:        "A Minor - Deep Space",
		DroneFreqs:  []float64{55.0, 82.5, 110.0, 165.0}, // A1, E2, A2, E3
		SubBassFreq: 27.5,                                // A0

		BassNotes: []float64{55, 55, 73.42, 73.42, 82.41, 82.41, 55, 41.2}, // Am-Dm-Em progression

		ArpNotes: []float64{110, 130.81, 146.83, 164.81, 196, 220, 261.63, 293.66}, // A minor pentatonic extended

		ShimmerFreqs:        []float64{880, 1320, 1760, 2200},       // A5 harmonics
		PadFreqs:            []float64{146.83, 174.61, 220, 261.63}, // Dm7: D, F, A, C
		TensionBaseFreq:     440,                                    // A4
		SirenBaseFreq:       660,                                    // E5
		SubBassReactiveFreq: 36.7,                                   // D1

		ArpTempo:  800,
		BassTempo: 1600,
	},

	// Level 2: D minor - darker, more aggressive
	2: {
		Name:        "D Minor - Hostile Territory",
		DroneFreqs:  []float64{73.42, 110.0, 146.83, 220.0}, // D2, A2, D3, A3
		SubBassFreq: 36.71,                                  // D1

		BassNotes: []float64{73.42, 73.42, 55, 55, 61.74, 61.74, 73.42, 49}, // Dm-Am-Bb-Dm progression

		ArpNotes: []float64{146.83, 174.61, 196, 220, 261.63, 293.66, 349.23, 392}, // D minor pentatonic extended

		ShimmerFreqs:        []float64{587.33, 880, 1174.66, 1760}, // D5 harmonics
		PadFreqs:            []float64{110, 130.81, 164.81, 196},   // Am7: A, C, E, G
		TensionBaseFreq:     587.33,                                // D5
		SirenBaseFreq:       880,                                   // A5
		SubBassReactiveFreq: 55,                                    // A1

		ArpTempo:  750,
		BassTempo: 1500,
	},

	// Level 3: E minor - epic, driving
	3: {
		Name:        "E Minor - Final Assault",
		DroneFreqs:  []float64{82.41, 123.47, 164.81, 246.94}, // E2, B2, E3, B3
		SubBassFreq: 41.2,                                     // E1

		BassNotes: []float64{82.41, 82.41, 73.42, 73.42, 55, 55, 82.41, 61.74}, // Em-D-Am-Em progression

		ArpNotes: []float64{164.81, 196, 220, 246.94, 293.66, 329.63, 392, 440}, // E minor pentatonic extended

		ShimmerFreqs:        []float64{659.26, 987.77, 1318.51, 1975.53}, // E5 harmonics
		PadFreqs:            []float64{123.47, 146.83, 185, 220},         // B7sus4: B, D, F#, A
		TensionBaseFreq:     659.26,                                      // E5
		SirenBaseFreq:       987.77,                                      // B5
		SubBassReactiveFreq: 73.42,                                       // D2

		ArpTempo:  700,
		BassTempo: 1400,
	},

	// Level 4: C# minor - Vangelis "Tears in Rain" style (melancholic, ethereal)
	4: {
		Name:        "C# Minor - Tears in Rain",
		DroneFreqs:  []float64{69.3, 103.83, 138.59, 207.65}, // C#2, G#2, C#3, G#3
		SubBassFreq: 34.65,                                   // C#1

		BassNotes: []float64{69.3, 69.3, 92.5, 92.5, 82.41, 82.41, 69.3, 55}, // C#m-F#m-E-C#m

		ArpNotes: []float64{138.59, 155.56, 185, 207.65, 277.18, 311.13, 370, 415.3}, // C# minor pentatonic

		ShimmerFreqs:        []float64{554.37, 831.61, 1108.73, 1661.22}, // C#5 harmonics
		PadFreqs:            []float64{92.5, 110, 138.59, 164.81},        // F#m7: F#, A, C#, E
		TensionBaseFreq:     554.37,                                      // C#5
		SirenBaseFreq:       831.61,                                      // G#5
		SubBassReactiveFreq: 46.25,                                       // F#1

		ArpTempo:  900, // Slower, more melancholic
		BassTempo: 1800,
	},

	// Level 5: F minor - "Blush Response" interrogation (tense, mechanical)
	5: {
		Name:        "F Minor - Voight-Kampff",
		DroneFreqs:  []float64{87.31, 130.81, 174.61, 261.63}, // F2, C3, F3, C4
		SubBassFreq: 43.65,                                    // F1

		BassNotes: []float64{87.31, 87.31, 77.78, 77.78, 65.41, 65.41, 87.31, 58.27}, // Fm-Ebm-Db-Fm

		ArpNotes: []float64{174.61, 207.65, 233.08, 261.63, 349.23, 415.3, 466.16, 523.25}, // F minor scale

		ShimmerFreqs:        []float64{698.46, 1046.5, 1396.91, 2093}, // F5 harmonics
		PadFreqs:            []float64{103.83, 123.47, 155.56, 185},   // Ab maj7: Ab, C, Eb, G
		TensionBaseFreq:     698.46,                                   // F5
		SirenBaseFreq:       1046.5,                                   // C6
		SubBassReactiveFreq: 51.91,                                    // Ab1

		ArpTempo:  650, // Faster, more mechanical
		BassTempo: 1300,
	},

	// Level 6: Bb minor - "Rachel's Song" romantic noir
	6: {
		Name:        "Bb Minor - Replicant Dreams",
		DroneFreqs:  []float64{58.27, 87.31, 116.54, 174.61}, // Bb1, F2, Bb2, F3
		SubBassFreq: 29.14,                                   // Bb0

		BassNotes: []float64{58.27, 58.27, 51.91, 51.91, 69.3, 69.3, 58.27, 46.25}, // Bbm-Abm-C#-Bbm

		ArpNotes: []float64{116.54, 138.59, 155.56, 174.61, 233.08, 277.18, 311.13, 349.23}, // Bb minor pentatonic

		ShimmerFreqs:        []float64{466.16, 698.46, 932.33, 1396.91}, // Bb4 harmonics
		PadFreqs:            []float64{138.59, 164.81, 207.65, 246.94},  // Db maj7: Db, F, Ab, C
		TensionBaseFreq:     466.16,                                     // Bb4
		SirenBaseFreq:       698.46,                                     // F5
		SubBassReactiveFreq: 41.2,                                       // E1

		ArpTempo:  950, // Very slow, romantic
		BassTempo: 1900,
	},

	// Level 7: G minor - "End Titles" epic synthwave
	7: {
		Name:        "G Minor - Neon Streets",
		DroneFreqs:  []float64{98, 146.83, 196, 293.66}, // G2, D3, G3, D4
		SubBassFreq: 49,                                 // G1

		BassNotes: []float64{98, 98, 87.31, 87.31, 73.42, 73.42, 98, 65.41}, // Gm-F-Dm-Gm

		ArpNotes: []float64{196, 233.08, 261.63, 293.66, 392, 466.16, 523.25, 587.33}, // G minor scale

		ShimmerFreqs:        []float64{784, 1174.66, 1568, 2349.32}, // G5 harmonics
		PadFreqs:            []float64{116.54, 146.83, 174.61, 220}, // Bb maj7: Bb, D, F, A
		TensionBaseFreq:     784,                                    // G5
		SirenBaseFreq:       1174.66,                                // D6
		SubBassReactiveFreq: 58.27,                                  // Bb1

		ArpTempo:  600, // Fast, energetic
		BassTempo: 1200,
	},

	// Level 8: F# minor - "Memory" introspective, crystalline
	8: {
		Name:        "F# Minor - Memory Implants",
		DroneFreqs:  []float64{92.5, 138.59, 185, 277.18}, // F#2, C#3, F#3, C#4
		SubBassFreq: 46.25,                                // F#1

		BassNotes: []float64{92.5, 92.5, 82.41, 82.41, 69.3, 69.3, 92.5, 55}, // F#m-E-C#m-F#m

		ArpNotes: []float64{185, 220, 246.94, 277.18, 370, 440, 493.88, 554.37}, // F# minor scale

		ShimmerFreqs:        []float64{739.99, 1108.73, 1479.98, 2217.46}, // F#5 harmonics
		PadFreqs:            []float64{110, 138.59, 164.81, 207.65},       // A maj7: A, C#, E, G#
		TensionBaseFreq:     739.99,                                       // F#5
		SirenBaseFreq:       1108.73,                                      // C#6
		SubBassReactiveFreq: 55,                                           // A1

		ArpTempo:  850, // Medium-slow, contemplative
		BassTempo: 1700,
	},

	// Level 9: Eb minor - "Los Angeles 2019" dark urban sprawl
	9: {
		Name:        "Eb Minor - Sprawl",
		DroneFreqs:  []float64{77.78, 116.54, 155.56, 233.08}, // Eb2, Bb2, Eb3, Bb3
		SubBassFreq: 38.89,                                    // Eb1

		BassNotes: []float64{77.78, 77.78, 69.3, 69.3, 58.27, 58.27, 77.78, 51.91}, // Ebm-C#m-Bbm-Ebm

		ArpNotes: []float64{155.56, 185, 207.65, 233.08, 311.13, 370, 415.3, 466.16}, // Eb minor pentatonic

		ShimmerFreqs:        []float64{622.25, 932.33, 1244.51, 1864.66}, // Eb5 harmonics
		PadFreqs:            []float64{103.83, 130.81, 155.56, 196},      // Gb maj7: Gb, Bb, Db, F
		TensionBaseFreq:     622.25,                                      // Eb5
		SirenBaseFreq:       932.33,                                      // Bb5
		SubBassReactiveFreq: 51.91,                                       // Ab1

		ArpTempo:  720,
		BassTempo: 1440,
	},
}

// SfxData contains sound effect parameter strings for jsfxr.
var SfxData = []string{
	// 0 = Player shoots (basic)
	"0,,.167,.1637,.1361,.7212,.0399,-.363,,,,,,.1314,.0517,,.0154,-.1633,1,,,.0515,,.2",
	// 1 = Player is hurt
	"3,.0704,.0462,.3388,.4099,.1599,,.0109,-.3247,.0006,,-.1592,.4477,.1028,.1787,,-.0157,-.3372,.1896,.1628,,.0016,-.0003,.5",
	// 2 = Shield hit
	"3,.1,.3899,.1901,.2847,.0399,,.0007,.1492,,,-.9636,,,-.3893,.1636,-.0047,.7799,.1099,-.1103,.5924,.484,.1547,1",
	// 3 = Shield activated
	"1,,.0398,,.4198,.3891,,.4383,,,,,,,,.616,,,1,,,,,.5",
	// 4 = Shield deactivated
	"1,.1299,.27,.1299,.4199,.1599,,.4383,,,,-.6399,,,-.4799,.7099,,,1,,,,,.5",
	// 5 = Weapon upgrade collected
	"0,.43,.1099,.67,.4499,.6999,,-.2199,-.2,.5299,.5299,-.0399,.3,,.0799,.1899,-.1194,.2327,.8815,-.2364,.43,.2099,-.5799,.5",
	// 6 = Non-applicable bonus collected
	"0,.2,.1099,.0733,.0854,.14,,-.1891,.36,,,.9826,,,.4642,,-.1194,.2327,.8815,-.2364,.0992,.0076,.2,.5",
	// 7 = Money collected
	"0,.09,.1099,.0733,.0854,.1099,,-.1891,.827,,,.9826,,,.4642,,-.1194,.2327,.8815,-.2364,.0992,.0076,.8314,.5",
	// 8 = New wave alarm
	"1,.1,1,.1901,.2847,.3199,,.0007,.1492,,,-.9636,,,-.3893,.1636,-.0047,.6646,.9653,-.1103,.5924,.484,.1547,.6",
	// 9 = Enemy hit
	"3,.1,.3899,.1901,.2847,.0399,,.0007,.1492,,,-.9636,,,-.3893,.1636,-.0047,.6646,.9653,-.1103,.5924,.484,.1547,.4",
	// 10 = Enemy destroyed
	"3,.2,.1899,.4799,.91,.0599,,-.2199,-.2,.5299,.5299,-.0399,.3,,.0799,.1899,-.1194,.2327,.8815,-.2364,.43,.2099,-.5799,.5",
	// 11 = Torpedo destroyed
	"3,,.3626,.5543,.191,.0731,,-.3749,,,,,,,,,,,1,,,,,.4",
	// 12 = Boss shoots
	"1,.071,.3474,.0506,.1485,.5799,.2,-.2184,-.1405,.1681,,-.1426,,.9603,-.0961,,.2791,-.8322,.2832,.0009,,.0088,-.0082,.3",
	// 13 = Bomb explosion
	"3,.05,.3365,.4591,.4922,.1051,,.015,,,,-.6646,.7394,,,,,,1,,,,,.7",
	// 14 = Game over
	"1,1,.09,.5,.4111,.506,.0942,.1499,.0199,.8799,.1099,-.68,.0268,.1652,.62,.6999,-.0399,.4799,.5199,-.0429,.0599,.8199,-.4199,.7",
	// 15 = Player shoots (weak weapon)
	"2,,.1199,.15,.1361,.5,.0399,-.363,-.4799,,,,,.1314,.0517,,.0154,-.1633,1,,,.0515,,.2",
	// 16 = Player shoots (strong weapon)
	"2,,.98,.4699,.07,.7799,.0399,-.28,-.4799,.2399,.1,,.36,.1314,.0517,,.0154,-.1633,1,,.37,.0399,.54,.1",
	// 17 = Low energy warning
	"0,.9705,.0514,.5364,.5273,.4816,.0849,.1422,.205,.7714,.1581,-.7685,.0822,.2147,.6062,.7448,-.0917,.4009,.6251,.1116,.0573,.9005,-.3763,.3",
	// 18 = Turret shoots
	"0,.0399,.1362,.0331,.2597,.85,.0137,-.3976,,,,,,.2099,-.72,,,,1,,,,,.3",
	// 19 = Small enemy shoots
	"0,,.2863,,.3048,.751,.2,-.316,,,,,,.4416,.1008,,,,1,,,.2962,,.3",
	// 20 = Medium enemy shoots
	"0,,.3138,,.0117,.7877,.1583,-.3391,-.04,,.0464,.0585,,.4085,-.4195,,-.024,-.0396,1,-.0437,.0124,.02,.0216,.3",
	// 21 = Intro sound
	"0,1,.8799,.3499,.17,.61,.1899,-.3,-.18,.3,.6399,-.0279,.0071,.8,-.1599,.5099,-.46,.5199,.25,.0218,.49,.4,-.2,.3",
	// 22 = Target locking (rising electronic tone)
	"1,.0099,.15,,.2299,.45,,.1799,.48,.5099,.4599,-.4399,.6299,,,,,.0099,.6599,.0099,,.1699,,.4",
	// 23 = Ship thrust (magnetic hum sound)
	"2,.01,.12,.03,.15,.18,,,.02,,.08,.12,.45,,,,,.08,.6,.15,.4,.03,,.3",
}

// SoundEffectLibrary contains all structured sound effects
var SoundEffectLibrary []*SoundEffect = []*SoundEffect{
	ParseJsfxrString(0, "Player Shoot", "Player", "Basic weapon fire", SfxData[0]),
	ParseJsfxrString(1, "Player Hurt", "Player", "Damage taken", SfxData[1]),
	ParseJsfxrString(2, "Shield Hit", "Player", "Shield absorbs damage", SfxData[2]),
	ParseJsfxrString(3, "Shield On", "Player", "Shield activated", SfxData[3]),
	ParseJsfxrString(4, "Shield Off", "Player", "Shield deactivated", SfxData[4]),
	ParseJsfxrString(5, "Weapon Up", "Pickup", "Weapon upgrade collected", SfxData[5]),
	ParseJsfxrString(6, "Bonus N/A", "Pickup", "Non-applicable bonus", SfxData[6]),
	ParseJsfxrString(7, "Money", "Pickup", "Money/points collected", SfxData[7]),
	ParseJsfxrString(8, "Wave Alarm", "UI", "New wave starting", SfxData[8]),
	ParseJsfxrString(9, "Enemy Hit", "Enemy", "Enemy takes damage", SfxData[9]),
	ParseJsfxrString(10, "Enemy Death", "Enemy", "Enemy destroyed", SfxData[10]),
	ParseJsfxrString(11, "Torpedo Explode", "Enemy", "Torpedo destroyed", SfxData[11]),
	ParseJsfxrString(12, "Boss Shoot", "Enemy", "Boss fires weapon", SfxData[12]),
	ParseJsfxrString(13, "Bomb", "Player", "Bomb explosion", SfxData[13]),
	ParseJsfxrString(14, "Game Over", "UI", "Player death", SfxData[14]),
	ParseJsfxrString(15, "Weak Shoot", "Player", "Weak weapon fire", SfxData[15]),
	ParseJsfxrString(16, "Strong Shoot", "Player", "Strong weapon fire", SfxData[16]),
	ParseJsfxrString(17, "Low Energy", "UI", "Energy warning", SfxData[17]),
	ParseJsfxrString(18, "Turret Shoot", "Enemy", "Turret fires", SfxData[18]),
	ParseJsfxrString(19, "Small Shoot", "Enemy", "Small enemy fires", SfxData[19]),
	ParseJsfxrString(20, "Medium Shoot", "Enemy", "Medium enemy fires", SfxData[20]),
	ParseJsfxrString(21, "Intro", "UI", "Title screen intro", SfxData[21]),
	ParseJsfxrString(22, "Target Lock", "Player", "Target locking sound", SfxData[22]),
	ParseJsfxrString(23, "Thrust", "Player", "Ship engine thrust", SfxData[23]),
}
