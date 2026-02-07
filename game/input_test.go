package game

import (
	"testing"

	"github.com/gopherjs/gopherjs/js"
)

// createMockImageObject creates a mock js.Object with a width property for testing
func createMockImageObject(width int) *js.Object {
	obj := js.Global.Get("Object").New()
	obj.Set("width", width)
	return obj
}

func TestKeyMap_ZMapsToX(t *testing.T) {
	// Z (90) should map to X (88)
	if mapped, ok := KeyMap[90]; !ok || mapped != 88 {
		t.Errorf("Expected KeyMap[90] (Z) to be 88 (X), got %d", mapped)
	}
}

func TestKeyMap_SpaceMapsToX(t *testing.T) {
	// Space (32) should map to X (88)
	if mapped, ok := KeyMap[32]; !ok || mapped != 88 {
		t.Errorf("Expected KeyMap[32] (Space) to be 88 (X), got %d", mapped)
	}
}

func TestKeyMap_WASDMapsToArrows(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"W maps to Up", 87, 38},
		{"A maps to Left", 65, 37},
		{"S maps to Down", 83, 40},
		{"D maps to Right", 68, 39},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if mapped, ok := KeyMap[tt.input]; !ok || mapped != tt.expected {
				t.Errorf("Expected KeyMap[%d] to be %d, got %d", tt.input, tt.expected, mapped)
			}
		})
	}
}

func TestTranslateKeyCode(t *testing.T) {
	tests := []struct {
		name     string
		input    int
		expected int
	}{
		{"Z translates to X", 90, 88},
		{"Space translates to X", 32, 88},
		{"W translates to Up", 87, 38},
		{"Arrow Left stays Arrow Left", 37, 37},
		{"X stays X", 88, 88},
		{"Unknown key stays same", 999, 999},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TranslateKeyCode(tt.input)
			if result != tt.expected {
				t.Errorf("TranslateKeyCode(%d) = %d, expected %d", tt.input, result, tt.expected)
			}
		})
	}
}

func TestSpawnText_DurationZero_UsesMaxT(t *testing.T) {
	g := &Game{
		Level: Level{
			Text: TextDisplay{MaxT: 90},
		},
	}

	// Create a mock image object with width
	mockImage := createMockImageObject(100)
	g.Level.Text.Image = mockImage

	// Call the logic that SpawnText uses (without RenderTextImage which needs browser)
	g.spawnTextWithDuration(0)

	if g.Level.Text.T != 90 {
		t.Errorf("Expected Text.T to be MaxT (90) when duration is 0, got %d", g.Level.Text.T)
	}
	if g.Level.Text.Y != 16 {
		t.Errorf("Expected Text.Y to be 16, got %d", g.Level.Text.Y)
	}
	if g.Level.Text.YAcc != Speed/2 {
		t.Errorf("Expected Text.YAcc to be Speed/2 (%f), got %f", Speed/2.0, g.Level.Text.YAcc)
	}
}

func TestSpawnText_PositiveDuration_UsesDuration(t *testing.T) {
	g := &Game{
		Level: Level{
			Text: TextDisplay{MaxT: 90},
		},
	}

	mockImage := createMockImageObject(100)
	g.Level.Text.Image = mockImage

	g.spawnTextWithDuration(45)

	if g.Level.Text.T != 45 {
		t.Errorf("Expected Text.T to be 45, got %d", g.Level.Text.T)
	}
}

func TestSpawnText_NegativeDuration_SetsZeroT(t *testing.T) {
	g := &Game{
		Level: Level{
			Text: TextDisplay{MaxT: 90},
		},
	}

	mockImage := createMockImageObject(100)
	g.Level.Text.Image = mockImage

	g.spawnTextWithDuration(-1)

	if g.Level.Text.T != 0 {
		t.Errorf("Expected Text.T to be 0 for negative duration, got %d", g.Level.Text.T)
	}
}

func TestSpawnText_CentersText(t *testing.T) {
	g := &Game{
		Level: Level{
			Text: TextDisplay{MaxT: 90},
		},
	}

	// Image width of 200, should center at (1024-200)/2 = 412
	mockImage := createMockImageObject(200)
	g.Level.Text.Image = mockImage

	g.spawnTextWithDuration(0)

	expectedX := (WIDTH - 200) / 2
	if g.Level.Text.X != expectedX {
		t.Errorf("Expected Text.X to be %d (centered), got %d", expectedX, g.Level.Text.X)
	}
}
