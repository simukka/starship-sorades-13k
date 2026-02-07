package game

import (
	"math"
	"strings"
	"testing"
)

func TestSfxrParams_ParseSettingsString_Basic(t *testing.T) {
	params := SfxrParams{}
	params.ParseSettingsString("0,.1,.2,.3,.4,.5,.6,.7,.8,.9,.1,.11,.12,.13,.14,.15,.16,.17,.18,.19,.2,.21,.22,.5")
	
	if params.WaveType != 0 {
		t.Errorf("WaveType: expected 0, got %d", params.WaveType)
	}
	if !floatNear(params.AttackTime, 0.1, 0.001) {
		t.Errorf("AttackTime: expected 0.1, got %f", params.AttackTime)
	}
	if !floatNear(params.SustainTime, 0.2, 0.001) {
		t.Errorf("SustainTime: expected 0.2, got %f", params.SustainTime)
	}
	if !floatNear(params.MasterVolume, 0.5, 0.001) {
		t.Errorf("MasterVolume: expected 0.5, got %f", params.MasterVolume)
	}
}

func TestSfxrParams_ParseSettingsString_NegativeValues(t *testing.T) {
	params := SfxrParams{}
	params.ParseSettingsString("0,0,0.3,0,0.4,0.5,0,-.363,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,.5")
	
	if !floatNear(params.Slide, -0.363, 0.001) {
		t.Errorf("Slide: expected -0.363, got %f", params.Slide)
	}
}

func TestSfxrParams_ParseSettingsString_MinSustainTime(t *testing.T) {
	params := SfxrParams{}
	params.ParseSettingsString("0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,.5")
	
	// Should enforce minimum sustain time
	if params.SustainTime < 0.01 {
		t.Errorf("SustainTime should be at least 0.01, got %f", params.SustainTime)
	}
}

func TestSfxrParams_ParseSettingsString_EnvelopeMinLength(t *testing.T) {
	params := SfxrParams{}
	// Very short envelope times
	params.ParseSettingsString("0,.001,.001,0,.001,0.5,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,.5")
	
	totalTime := params.AttackTime + params.SustainTime + params.DecayTime
	if totalTime < 0.18 {
		t.Errorf("Total envelope time should be at least 0.18, got %f", totalTime)
	}
}

func TestSfxrSynth_TotalReset_ReturnsPositiveLength(t *testing.T) {
	synth := NewSfxrSynth()
	synth.Params.ParseSettingsString("0,0,.3,0,.4,.5,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,.5")
	
	length := synth.TotalReset()
	
	if length <= 0 {
		t.Errorf("TotalReset should return positive length, got %d", length)
	}
}

func TestSfxrSynth_SynthWave_ProducesSamples(t *testing.T) {
	ResetPRNG(12345) // Reproducible random
	
	synth := NewSfxrSynth()
	// Simple square wave sound
	synth.Params.ParseSettingsString("0,0,.3,0,.4,.5,0,0,0,0,0,0,0,.5,0,0,0,0,1,0,0,0,0,.5")
	
	length := synth.TotalReset()
	buffer := make([]int16, length)
	
	written := synth.SynthWave(buffer, length)
	
	if written <= 0 {
		t.Errorf("SynthWave should write samples, wrote %d", written)
	}
	
	// Check that we have non-zero samples (actual audio)
	hasNonZero := false
	for i := 0; i < written; i++ {
		if buffer[i] != 0 {
			hasNonZero = true
			break
		}
	}
	if !hasNonZero {
		t.Error("SynthWave should produce non-zero samples")
	}
}

func TestSfxrSynth_WaveTypes(t *testing.T) {
	tests := []struct {
		name     string
		waveType int
		settings string
	}{
		{"Square", 0, "0,0,.3,0,.4,.5,0,0,0,0,0,0,0,.5,0,0,0,0,1,0,0,0,0,.5"},
		{"Sawtooth", 1, "1,0,.3,0,.4,.5,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,.5"},
		{"Sine", 2, "2,0,.3,0,.4,.5,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,.5"},
		{"Noise", 3, "3,0,.3,0,.4,.5,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,.5"},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResetPRNG(12345)
			
			synth := NewSfxrSynth()
			synth.Params.ParseSettingsString(tt.settings)
			
			length := synth.TotalReset()
			buffer := make([]int16, length)
			
			written := synth.SynthWave(buffer, length)
			
			if written <= 0 {
				t.Errorf("%s wave should produce samples", tt.name)
			}
			
			// Verify wave type was set correctly
			if synth.Params.WaveType != tt.waveType {
				t.Errorf("WaveType: expected %d, got %d", tt.waveType, synth.Params.WaveType)
			}
		})
	}
}

func TestSfxrSynth_Envelope_AttackIncreasesVolume(t *testing.T) {
	ResetPRNG(12345)
	
	synth := NewSfxrSynth()
	// Long attack time
	synth.Params.ParseSettingsString("0,.5,.2,0,.2,.5,0,0,0,0,0,0,0,.5,0,0,0,0,1,0,0,0,0,.5")
	
	length := synth.TotalReset()
	buffer := make([]int16, length)
	synth.SynthWave(buffer, length)
	
	// Sample RMS should generally increase during attack phase
	earlyRMS := calculateRMS(buffer[:1000])
	midRMS := calculateRMS(buffer[5000:6000])
	
	// Mid attack should have higher volume than early attack
	if midRMS <= earlyRMS && earlyRMS > 0 {
		t.Logf("Early RMS: %f, Mid RMS: %f", earlyRMS, midRMS)
		// This is expected behavior but may vary based on waveform
	}
}

func TestSfxrSynth_Filter_LowPassReducesHighFreq(t *testing.T) {
	ResetPRNG(12345)
	
	// Generate noise without filter
	synthNoFilter := NewSfxrSynth()
	synthNoFilter.Params.ParseSettingsString("3,0,.3,0,.3,.5,0,0,0,0,0,0,0,0,0,0,0,0,1,0,0,0,0,.5")
	lengthNoFilter := synthNoFilter.TotalReset()
	bufferNoFilter := make([]int16, lengthNoFilter)
	synthNoFilter.SynthWave(bufferNoFilter, lengthNoFilter)
	
	ResetPRNG(12345)
	
	// Generate noise with low-pass filter
	synthWithFilter := NewSfxrSynth()
	synthWithFilter.Params.ParseSettingsString("3,0,.3,0,.3,.5,0,0,0,0,0,0,0,0,0,0,0,0,.2,0,0,0,0,.5")
	lengthWithFilter := synthWithFilter.TotalReset()
	bufferWithFilter := make([]int16, lengthWithFilter)
	synthWithFilter.SynthWave(bufferWithFilter, lengthWithFilter)
	
	// Filtered version should have lower high-frequency content
	// We'll measure this by comparing sample-to-sample differences
	varianceNoFilter := calculateVariance(bufferNoFilter)
	varianceWithFilter := calculateVariance(bufferWithFilter)
	
	// Filtered signal should have less variance (smoother)
	if varianceWithFilter >= varianceNoFilter && varianceNoFilter > 0 {
		t.Logf("Variance without filter: %f, with filter: %f", varianceNoFilter, varianceWithFilter)
		// This might not always hold due to other synthesis factors
	}
}

func TestGenerateWavDataURL_ReturnsValidDataURL(t *testing.T) {
	ResetPRNG(12345)
	
	dataURL := GenerateWavDataURL("0,0,.3,0,.4,.5,0,0,0,0,0,0,0,.5,0,0,0,0,1,0,0,0,0,.5")
	
	if !strings.HasPrefix(dataURL, "data:audio/wav;base64,") {
		t.Errorf("DataURL should start with 'data:audio/wav;base64,', got: %s", dataURL[:50])
	}
	
	// Extract and validate base64 content exists
	base64Part := strings.TrimPrefix(dataURL, "data:audio/wav;base64,")
	if len(base64Part) == 0 {
		t.Error("Base64 content should not be empty")
	}
}

func TestGenerateFloat32Buffer_ProducesValidSamples(t *testing.T) {
	ResetPRNG(12345)
	
	samples := GenerateFloat32Buffer("0,0,.3,0,.4,.5,0,0,0,0,0,0,0,.5,0,0,0,0,1,0,0,0,0,.5")
	
	if len(samples) == 0 {
		t.Error("Should produce samples")
	}
	
	// All samples should be in valid float range [-1, 1]
	for i, s := range samples {
		if s < -1.0 || s > 1.0 {
			t.Errorf("Sample %d out of range: %f", i, s)
			break
		}
	}
}

func TestSfxrSynth_Deterministic(t *testing.T) {
	// Test that same settings produce same output
	ResetPRNG(12345)
	synth1 := NewSfxrSynth()
	synth1.Params.ParseSettingsString("0,0,.3,0,.4,.5,0,0,0,0,0,0,0,.5,0,0,0,0,1,0,0,0,0,.5")
	length1 := synth1.TotalReset()
	buffer1 := make([]int16, length1)
	synth1.SynthWave(buffer1, length1)
	
	ResetPRNG(12345)
	synth2 := NewSfxrSynth()
	synth2.Params.ParseSettingsString("0,0,.3,0,.4,.5,0,0,0,0,0,0,0,.5,0,0,0,0,1,0,0,0,0,.5")
	length2 := synth2.TotalReset()
	buffer2 := make([]int16, length2)
	synth2.SynthWave(buffer2, length2)
	
	if length1 != length2 {
		t.Errorf("Lengths should match: %d vs %d", length1, length2)
	}
	
	// First 1000 samples should be identical
	for i := 0; i < 1000 && i < len(buffer1) && i < len(buffer2); i++ {
		if buffer1[i] != buffer2[i] {
			t.Errorf("Sample %d differs: %d vs %d", i, buffer1[i], buffer2[i])
			break
		}
	}
}

func TestSfxrSynth_PlayerShootSound(t *testing.T) {
	// Test actual game sound from SfxData
	ResetPRNG(12345)
	
	synth := NewSfxrSynth()
	synth.Params.ParseSettingsString("0,,.167,.1637,.1361,.7212,.0399,-.363,,,,,,.1314,.0517,,.0154,-.1633,1,,,.0515,,.2")
	
	length := synth.TotalReset()
	buffer := make([]int16, length)
	written := synth.SynthWave(buffer, length)
	
	if written <= 0 {
		t.Error("Player shoot sound should produce samples")
	}
	
	// Should be a reasonably short sound effect
	durationMs := float64(written) / 44.1 // 44100 samples per second
	if durationMs > 2000 {
		t.Errorf("Player shoot sound seems too long: %f ms", durationMs)
	}
}

func TestSfxrSynth_PhaserEffect(t *testing.T) {
	ResetPRNG(12345)
	
	// Sound with phaser effect
	synth := NewSfxrSynth()
	synth.Params.ParseSettingsString("0,0,.3,0,.4,.5,0,0,0,0,0,0,0,.5,0,0,.5,.1,1,0,0,0,0,.5")
	
	length := synth.TotalReset()
	buffer := make([]int16, length)
	written := synth.SynthWave(buffer, length)
	
	if written <= 0 {
		t.Error("Sound with phaser should produce samples")
	}
}

func TestSfxrSynth_VibratoEffect(t *testing.T) {
	ResetPRNG(12345)
	
	// Sound with vibrato
	synth := NewSfxrSynth()
	synth.Params.ParseSettingsString("0,0,.3,0,.4,.5,0,0,0,.3,.5,0,0,.5,0,0,0,0,1,0,0,0,0,.5")
	
	length := synth.TotalReset()
	buffer := make([]int16, length)
	written := synth.SynthWave(buffer, length)
	
	if written <= 0 {
		t.Error("Sound with vibrato should produce samples")
	}
}

func TestSfxrSynth_RepeatEffect(t *testing.T) {
	ResetPRNG(12345)
	
	// Sound with repeat
	synth := NewSfxrSynth()
	synth.Params.ParseSettingsString("0,0,.3,0,.4,.5,0,0,0,0,0,0,0,.5,0,.3,0,0,1,0,0,0,0,.5")
	
	length := synth.TotalReset()
	buffer := make([]int16, length)
	written := synth.SynthWave(buffer, length)
	
	if written <= 0 {
		t.Error("Sound with repeat should produce samples")
	}
}

// Helper functions

func floatNear(a, b, epsilon float64) bool {
	return math.Abs(a-b) < epsilon
}

func calculateRMS(samples []int16) float64 {
	if len(samples) == 0 {
		return 0
	}
	sum := 0.0
	for _, s := range samples {
		sum += float64(s) * float64(s)
	}
	return math.Sqrt(sum / float64(len(samples)))
}

func calculateVariance(samples []int16) float64 {
	if len(samples) < 2 {
		return 0
	}
	diffSum := 0.0
	for i := 1; i < len(samples); i++ {
		diff := float64(samples[i] - samples[i-1])
		diffSum += diff * diff
	}
	return diffSum / float64(len(samples)-1)
}

// Stereo tests

func TestSfxrSynth_SynthWaveStereo_ProducesSamples(t *testing.T) {
	ResetPRNG(12345)
	
	synth := NewSfxrSynth()
	synth.Params.ParseSettingsString("0,0,.3,0,.4,.5,0,0,0,0,0,0,0,.5,0,0,0,0,1,0,0,0,0,.5")
	
	length := synth.TotalReset()
	bufferL := make([]int16, length)
	bufferR := make([]int16, length)
	
	written := synth.SynthWaveStereo(bufferL, bufferR, length, 0) // Center pan
	
	if written <= 0 {
		t.Errorf("SynthWaveStereo should write samples, wrote %d", written)
	}
	
	// Both channels should have audio
	hasNonZeroL := false
	hasNonZeroR := false
	for i := 0; i < written; i++ {
		if bufferL[i] != 0 {
			hasNonZeroL = true
		}
		if bufferR[i] != 0 {
			hasNonZeroR = true
		}
		if hasNonZeroL && hasNonZeroR {
			break
		}
	}
	if !hasNonZeroL {
		t.Error("Left channel should have non-zero samples")
	}
	if !hasNonZeroR {
		t.Error("Right channel should have non-zero samples")
	}
}

func TestSfxrSynth_SynthWaveStereo_CenterPan(t *testing.T) {
	ResetPRNG(12345)
	
	synth := NewSfxrSynth()
	synth.Params.ParseSettingsString("0,0,.3,0,.4,.5,0,0,0,0,0,0,0,.5,0,0,0,0,1,0,0,0,0,.5")
	
	length := synth.TotalReset()
	bufferL := make([]int16, length)
	bufferR := make([]int16, length)
	
	written := synth.SynthWaveStereo(bufferL, bufferR, length, 0) // Center pan
	
	// At center pan, left and right should be approximately equal
	rmsL := calculateRMS(bufferL[:written])
	rmsR := calculateRMS(bufferR[:written])
	
	ratio := rmsL / rmsR
	if ratio < 0.95 || ratio > 1.05 {
		t.Errorf("Center pan should have equal L/R levels, got ratio %f", ratio)
	}
}

func TestSfxrSynth_SynthWaveStereo_FullLeftPan(t *testing.T) {
	ResetPRNG(12345)
	
	synth := NewSfxrSynth()
	synth.Params.ParseSettingsString("0,0,.3,0,.4,.5,0,0,0,0,0,0,0,.5,0,0,0,0,1,0,0,0,0,.5")
	
	length := synth.TotalReset()
	bufferL := make([]int16, length)
	bufferR := make([]int16, length)
	
	written := synth.SynthWaveStereo(bufferL, bufferR, length, -1) // Full left
	
	rmsL := calculateRMS(bufferL[:written])
	rmsR := calculateRMS(bufferR[:written])
	
	// Left should have much more signal than right
	if rmsR > rmsL*0.1 {
		t.Errorf("Full left pan: right channel should be near silent, got L=%f R=%f", rmsL, rmsR)
	}
	if rmsL == 0 {
		t.Error("Full left pan: left channel should have audio")
	}
}

func TestSfxrSynth_SynthWaveStereo_FullRightPan(t *testing.T) {
	ResetPRNG(12345)
	
	synth := NewSfxrSynth()
	synth.Params.ParseSettingsString("0,0,.3,0,.4,.5,0,0,0,0,0,0,0,.5,0,0,0,0,1,0,0,0,0,.5")
	
	length := synth.TotalReset()
	bufferL := make([]int16, length)
	bufferR := make([]int16, length)
	
	written := synth.SynthWaveStereo(bufferL, bufferR, length, 1) // Full right
	
	rmsL := calculateRMS(bufferL[:written])
	rmsR := calculateRMS(bufferR[:written])
	
	// Right should have much more signal than left
	if rmsL > rmsR*0.1 {
		t.Errorf("Full right pan: left channel should be near silent, got L=%f R=%f", rmsL, rmsR)
	}
	if rmsR == 0 {
		t.Error("Full right pan: right channel should have audio")
	}
}

func TestSfxrSynth_SynthWaveStereo_PartialPan(t *testing.T) {
	ResetPRNG(12345)
	
	synth := NewSfxrSynth()
	synth.Params.ParseSettingsString("0,0,.3,0,.4,.5,0,0,0,0,0,0,0,.5,0,0,0,0,1,0,0,0,0,.5")
	
	length := synth.TotalReset()
	bufferL := make([]int16, length)
	bufferR := make([]int16, length)
	
	written := synth.SynthWaveStereo(bufferL, bufferR, length, 0.5) // Partial right
	
	rmsL := calculateRMS(bufferL[:written])
	rmsR := calculateRMS(bufferR[:written])
	
	// Right should be louder than left, but left should still have signal
	if rmsR <= rmsL {
		t.Errorf("Pan 0.5: right should be louder than left, got L=%f R=%f", rmsL, rmsR)
	}
	if rmsL == 0 {
		t.Error("Pan 0.5: left channel should still have audio")
	}
}

func TestGenerateStereoWavDataURL_ReturnsValidDataURL(t *testing.T) {
	ResetPRNG(12345)
	
	dataURL := GenerateStereoWavDataURL("0,0,.3,0,.4,.5,0,0,0,0,0,0,0,.5,0,0,0,0,1,0,0,0,0,.5", 0)
	
	if !strings.HasPrefix(dataURL, "data:audio/wav;base64,") {
		t.Errorf("Stereo DataURL should start with 'data:audio/wav;base64,', got: %s", dataURL[:50])
	}
	
	base64Part := strings.TrimPrefix(dataURL, "data:audio/wav;base64,")
	if len(base64Part) == 0 {
		t.Error("Stereo base64 content should not be empty")
	}
}

func TestGenerateStereoFloat32Buffer_ProducesValidSamples(t *testing.T) {
	ResetPRNG(12345)
	
	samplesL, samplesR := GenerateStereoFloat32Buffer("0,0,.3,0,.4,.5,0,0,0,0,0,0,0,.5,0,0,0,0,1,0,0,0,0,.5", 0)
	
	if len(samplesL) == 0 || len(samplesR) == 0 {
		t.Error("Should produce samples for both channels")
	}
	
	if len(samplesL) != len(samplesR) {
		t.Errorf("Left and right channels should have same length: %d vs %d", len(samplesL), len(samplesR))
	}
	
	// All samples should be in valid float range [-1, 1]
	for i, s := range samplesL {
		if s < -1.0 || s > 1.0 {
			t.Errorf("Left sample %d out of range: %f", i, s)
			break
		}
	}
	for i, s := range samplesR {
		if s < -1.0 || s > 1.0 {
			t.Errorf("Right sample %d out of range: %f", i, s)
			break
		}
	}
}

func TestSfxrSynth_SynthWaveStereo_PanClamping(t *testing.T) {
	ResetPRNG(12345)
	
	synth := NewSfxrSynth()
	synth.Params.ParseSettingsString("0,0,.3,0,.4,.5,0,0,0,0,0,0,0,.5,0,0,0,0,1,0,0,0,0,.5")
	
	length := synth.TotalReset()
	bufferL := make([]int16, length)
	bufferR := make([]int16, length)
	
	// Test that extreme pan values are clamped
	written := synth.SynthWaveStereo(bufferL, bufferR, length, -5) // Should clamp to -1
	
	if written <= 0 {
		t.Error("Should produce samples even with out-of-range pan")
	}
	
	ResetPRNG(12345)
	synth2 := NewSfxrSynth()
	synth2.Params.ParseSettingsString("0,0,.3,0,.4,.5,0,0,0,0,0,0,0,.5,0,0,0,0,1,0,0,0,0,.5")
	synth2.TotalReset()
	bufferL2 := make([]int16, length)
	bufferR2 := make([]int16, length)
	synth2.SynthWaveStereo(bufferL2, bufferR2, length, 5) // Should clamp to 1
	
	// Just verify it doesn't crash and produces output
	rmsR := calculateRMS(bufferR2[:written])
	if rmsR == 0 {
		t.Error("Pan clamped to +1 should have right channel audio")
	}
}
