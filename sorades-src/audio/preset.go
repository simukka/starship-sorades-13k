package audio

import (
	"sort"
	"strings"
)

// PresetInfo holds music preset data for template rendering
type PresetInfo struct {
	Level int
	Key   string
	Name  string
}

// LevelMusicPreset defines musical notes/frequencies for a level's atmosphere
type LevelMusicPreset struct {
	Name string // Preset name for display

	// Drone pad frequencies (typically root, fifth, octave, upper fifth)
	DroneFreqs  []float64
	SubBassFreq float64 // Deep sub bass (one octave below root)

	// Bass line sequence (8 notes per cycle)
	BassNotes []float64

	// Arpeggio note sequence (pentatonic or similar scale)
	ArpNotes []float64

	// Reactive layer frequencies
	ShimmerFreqs        []float64 // Detuned shimmer harmonics
	PadFreqs            []float64 // Chord tones (e.g., Dm7)
	TensionBaseFreq     float64   // High tension oscillator base
	SirenBaseFreq       float64   // Warning siren frequency
	SubBassReactiveFreq float64   // Danger sub bass

	// Tempo/timing (ms)
	ArpTempo  int // Arpeggio note duration
	BassTempo int // Bass note duration
}

// GetLevelPreset returns the music preset for a level (defaults to level 1)
func GetLevelPreset(level int) *LevelMusicPreset {
	if preset, ok := LevelMusicPresets[level]; ok {
		return preset
	}
	return LevelMusicPresets[1] // Default to level 1
}

// GetAllPresetInfo returns preset info for all available music presets, sorted by level
func GetAllPresetInfo() []PresetInfo {
	presets := make([]PresetInfo, 0, len(LevelMusicPresets))
	for level, preset := range LevelMusicPresets {
		// Extract key from name (e.g., "A Minor - Deep Space" -> "A Minor")
		name := preset.Name
		key := name
		description := ""
		if idx := strings.Index(name, " - "); idx > 0 {
			key = name[:idx]
			description = name[idx+3:]
		}
		presets = append(presets, PresetInfo{
			Level: level,
			Key:   key,
			Name:  description,
		})
	}
	// Sort by level
	sort.Slice(presets, func(i, j int) bool {
		return presets[i].Level < presets[j].Level
	})
	return presets
}
