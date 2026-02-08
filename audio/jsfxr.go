package audio

import (
	"encoding/base64"
	"math"
	"strconv"
	"strings"
)

// SfxrParams holds all configurable parameters for sound synthesis.
type SfxrParams struct {
	WaveType            int     // 0=square, 1=sawtooth, 2=sine, 3=noise
	AttackTime          float64 // Time for volume to ramp up (0-1)
	SustainTime         float64 // Time at full volume (0-1)
	SustainPunch        float64 // Extra volume boost at sustain start (0-1)
	DecayTime           float64 // Time for volume to fade out (0-1)
	StartFrequency      float64 // Base frequency of the sound (0-1)
	MinFrequency        float64 // Frequency cutoff
	Slide               float64 // Frequency slide
	DeltaSlide          float64 // Acceleration of frequency slide
	VibratoDepth        float64 // Depth of vibrato effect
	VibratoSpeed        float64 // Speed of vibrato oscillation
	ChangeAmount        float64 // Amount to change pitch mid-sound
	ChangeSpeed         float64 // When to apply the pitch change
	SquareDuty          float64 // Duty cycle for square wave (0-1)
	DutySweep           float64 // Sweep of square wave duty cycle
	RepeatSpeed         float64 // Speed of sound repeat
	PhaserOffset        float64 // Initial phaser offset
	PhaserSweep         float64 // Phaser offset sweep
	LpFilterCutoff      float64 // Low-pass filter cutoff frequency (0-1)
	LpFilterCutoffSweep float64 // Low-pass filter cutoff sweep
	LpFilterResonance   float64 // Low-pass filter resonance (0-1)
	HpFilterCutoff      float64 // High-pass filter cutoff frequency (0-1)
	HpFilterCutoffSweep float64 // High-pass filter cutoff sweep
	MasterVolume        float64 // Master volume (0-1)
}

// ParseSettingsString parses a comma-separated settings string into parameters.
func (p *SfxrParams) ParseSettingsString(s string) {
	values := strings.Split(s, ",")

	parseFloat := func(idx int) float64 {
		if idx >= len(values) || values[idx] == "" {
			return 0
		}
		f, _ := strconv.ParseFloat(values[idx], 64)
		return f
	}

	parseInt := func(idx int) int {
		if idx >= len(values) || values[idx] == "" {
			return 0
		}
		i, _ := strconv.Atoi(values[idx])
		return i
	}

	p.WaveType = parseInt(0)
	p.AttackTime = parseFloat(1)
	p.SustainTime = parseFloat(2)
	p.SustainPunch = parseFloat(3)
	p.DecayTime = parseFloat(4)
	p.StartFrequency = parseFloat(5)
	p.MinFrequency = parseFloat(6)
	p.Slide = parseFloat(7)
	p.DeltaSlide = parseFloat(8)
	p.VibratoDepth = parseFloat(9)
	p.VibratoSpeed = parseFloat(10)
	p.ChangeAmount = parseFloat(11)
	p.ChangeSpeed = parseFloat(12)
	p.SquareDuty = parseFloat(13)
	p.DutySweep = parseFloat(14)
	p.RepeatSpeed = parseFloat(15)
	p.PhaserOffset = parseFloat(16)
	p.PhaserSweep = parseFloat(17)
	p.LpFilterCutoff = parseFloat(18)
	p.LpFilterCutoffSweep = parseFloat(19)
	p.LpFilterResonance = parseFloat(20)
	p.HpFilterCutoff = parseFloat(21)
	p.HpFilterCutoffSweep = parseFloat(22)
	p.MasterVolume = parseFloat(23)

	// Ensure minimum sustain time for audible sound
	if p.SustainTime < 0.01 {
		p.SustainTime = 0.01
	}

	// Ensure minimum total envelope length to prevent clicks/pops
	totalTime := p.AttackTime + p.SustainTime + p.DecayTime
	if totalTime < 0.18 {
		multiplier := 0.18 / totalTime
		p.AttackTime *= multiplier
		p.SustainTime *= multiplier
		p.DecayTime *= multiplier
	}
}

// SfxrSynth is the sound synthesizer engine.
type SfxrSynth struct {
	Params SfxrParams

	// Envelope lengths
	envelopeLength0 float64
	envelopeLength1 float64
	envelopeLength2 float64

	// Oscillator state
	period       float64
	maxPeriod    float64
	slide        float64
	deltaSlide   float64
	changeAmount float64
	changeTime   float64
	changeLimit  float64
	squareDuty   float64
	dutySweep    float64

	// Buffers
	phaserBuffer []float64
	noiseBuffer  []float64
}

// NewSfxrSynth creates a new synthesizer instance.
func NewSfxrSynth() *SfxrSynth {
	return &SfxrSynth{
		phaserBuffer: make([]float64, 1024),
		noiseBuffer:  make([]float64, 32),
	}
}

// Reset resets oscillator state for partial reset (used for repeat effect).
func (s *SfxrSynth) Reset() {
	p := &s.Params

	// Calculate period from frequency
	s.period = 100 / (p.StartFrequency*p.StartFrequency + 0.001)
	s.maxPeriod = 100 / (p.MinFrequency*p.MinFrequency + 0.001)

	// Calculate slide as a multiplier
	s.slide = 1 - p.Slide*p.Slide*p.Slide*0.01
	s.deltaSlide = -p.DeltaSlide * p.DeltaSlide * p.DeltaSlide * 0.000001

	// Square wave duty cycle
	if p.WaveType == 0 {
		s.squareDuty = 0.5 - p.SquareDuty/2
		s.dutySweep = -p.DutySweep * 0.00005
	}

	// Pitch change calculation
	if p.ChangeAmount > 0 {
		s.changeAmount = 1 - p.ChangeAmount*p.ChangeAmount*0.9
	} else {
		s.changeAmount = 1 + p.ChangeAmount*p.ChangeAmount*10
	}
	s.changeTime = 0
	if p.ChangeSpeed == 1 {
		s.changeLimit = 0
	} else {
		s.changeLimit = (1-p.ChangeSpeed)*(1-p.ChangeSpeed)*20000 + 32
	}
}

// TotalReset performs full reset including envelope calculation.
func (s *SfxrSynth) TotalReset() int {
	s.Reset()
	p := &s.Params

	// Calculate envelope lengths
	s.envelopeLength0 = p.AttackTime * p.AttackTime * 100000
	s.envelopeLength1 = p.SustainTime * p.SustainTime * 100000
	s.envelopeLength2 = p.DecayTime*p.DecayTime*100000 + 10

	return int(s.envelopeLength0 + s.envelopeLength1 + s.envelopeLength2)
}

// SynthWave synthesizes audio samples into the provided buffer.
func (s *SfxrSynth) SynthWave(buffer []int16, length int) int {
	p := &s.Params

	// Filter configuration
	filtersEnabled := p.LpFilterCutoff != 1 || p.HpFilterCutoff != 0
	hpFilterCutoff := p.HpFilterCutoff * p.HpFilterCutoff * 0.1
	hpFilterDeltaCutoff := 1 + p.HpFilterCutoffSweep*0.0003
	lpFilterCutoff := p.LpFilterCutoff * p.LpFilterCutoff * p.LpFilterCutoff * 0.1
	lpFilterDeltaCutoff := 1 + p.LpFilterCutoffSweep*0.0001
	lpFilterOn := p.LpFilterCutoff != 1
	masterVolume := p.MasterVolume * p.MasterVolume
	minFrequency := p.MinFrequency
	phaserEnabled := p.PhaserOffset != 0 || p.PhaserSweep != 0
	phaserDeltaOffset := p.PhaserSweep * p.PhaserSweep * p.PhaserSweep * 0.2

	phaserOffset := p.PhaserOffset * p.PhaserOffset
	if p.PhaserOffset < 0 {
		phaserOffset *= -1020
	} else {
		phaserOffset *= 1020
	}

	var repeatLimit int
	if p.RepeatSpeed != 0 {
		repeatLimit = int((1-p.RepeatSpeed)*(1-p.RepeatSpeed)*20000) + 32
	}

	sustainPunch := p.SustainPunch
	vibratoAmplitude := p.VibratoDepth / 2
	vibratoSpeed := p.VibratoSpeed * p.VibratoSpeed * 0.01
	waveType := p.WaveType

	// Envelope state
	envelopeLength := s.envelopeLength0
	envelopeOverLength0 := 1 / s.envelopeLength0
	envelopeOverLength1 := 1 / s.envelopeLength1
	envelopeOverLength2 := 1 / s.envelopeLength2

	// Low-pass filter damping
	lpFilterDamping := 5 / (1 + p.LpFilterResonance*p.LpFilterResonance*20) * (0.01 + lpFilterCutoff)
	if lpFilterDamping > 0.8 {
		lpFilterDamping = 0.8
	}
	lpFilterDamping = 1 - lpFilterDamping

	// Synthesis state
	finished := false
	envelopeStage := 0
	envelopeTime := 0.0
	envelopeVolume := 0.0
	hpFilterPos := 0.0
	lpFilterDeltaPos := 0.0
	lpFilterOldPos := 0.0
	lpFilterPos := 0.0
	periodTemp := 0.0
	phase := 0.0
	phaserInt := 0
	phaserPos := 0
	repeatTime := 0
	sample := 0.0
	superSample := 0.0
	vibratoPhase := 0.0

	// Clear buffers
	for i := range s.phaserBuffer {
		s.phaserBuffer[i] = 0
	}
	for i := range s.noiseBuffer {
		s.noiseBuffer[i] = pseudoRandom()*2 - 1
	}

	// Cache state
	period := s.period
	maxPeriod := s.maxPeriod
	slide := s.slide
	deltaSlide := s.deltaSlide
	changeAmount := s.changeAmount
	changeTime := s.changeTime
	changeLimit := s.changeLimit
	squareDuty := s.squareDuty
	dutySweep := s.dutySweep

	// Use deterministic random for reproducibility
	randSeed := uint32(12345)

	for i := 0; i < length; i++ {
		if finished {
			return i
		}

		// Handle repeat effect
		if repeatLimit != 0 {
			repeatTime++
			if repeatTime >= repeatLimit {
				repeatTime = 0
				s.Reset()
				period = s.period
				maxPeriod = s.maxPeriod
				slide = s.slide
				deltaSlide = s.deltaSlide
				changeAmount = s.changeAmount
				changeTime = s.changeTime
				changeLimit = s.changeLimit
				squareDuty = s.squareDuty
				dutySweep = s.dutySweep
			}
		}

		// Handle pitch change
		if changeLimit != 0 {
			changeTime++
			if changeTime >= changeLimit {
				changeLimit = 0
				period *= changeAmount
			}
		}

		// Apply frequency slide
		slide += deltaSlide
		period *= slide

		// Check for minimum frequency cutoff
		if period > maxPeriod {
			period = maxPeriod
			if minFrequency > 0 {
				finished = true
			}
		}

		periodTemp = period

		// Apply vibrato
		if vibratoAmplitude > 0 {
			vibratoPhase += vibratoSpeed
			periodTemp *= 1 + math.Sin(vibratoPhase)*vibratoAmplitude
		}

		// Clamp period to minimum
		periodTempInt := int(periodTemp)
		if periodTempInt < 8 {
			periodTempInt = 8
		}
		periodTemp = float64(periodTempInt)

		// Square wave duty sweep
		if waveType == 0 {
			squareDuty += dutySweep
			if squareDuty < 0 {
				squareDuty = 0
			}
			if squareDuty > 0.5 {
				squareDuty = 0.5
			}
		}

		// Envelope stage progression
		envelopeTime++
		if envelopeTime > envelopeLength {
			envelopeTime = 0
			envelopeStage++

			if envelopeStage == 1 {
				envelopeLength = s.envelopeLength1
			} else if envelopeStage == 2 {
				envelopeLength = s.envelopeLength2
			}
		}

		// Calculate envelope volume
		switch envelopeStage {
		case 0: // Attack
			envelopeVolume = envelopeTime * envelopeOverLength0
		case 1: // Sustain
			envelopeVolume = 1 + (1-envelopeTime*envelopeOverLength1)*2*sustainPunch
		case 2: // Decay
			envelopeVolume = 1 - envelopeTime*envelopeOverLength2
		case 3: // Finished
			envelopeVolume = 0
			finished = true
		}

		// Update phaser offset
		if phaserEnabled {
			phaserOffset += phaserDeltaOffset
			phaserInt = int(math.Abs(phaserOffset))
			if phaserInt > 1023 {
				phaserInt = 1023
			}
		}

		// Update high-pass filter cutoff
		if filtersEnabled && hpFilterDeltaCutoff != 1 {
			hpFilterCutoff *= hpFilterDeltaCutoff
			if hpFilterCutoff < 0.00001 {
				hpFilterCutoff = 0.00001
			}
			if hpFilterCutoff > 0.1 {
				hpFilterCutoff = 0.1
			}
		}

		// 8x Oversampling loop
		superSample = 0

		for j := 0; j < 8; j++ {
			// Advance phase
			phase++
			if phase >= periodTemp {
				phase = math.Mod(phase, periodTemp)

				// Generate new noise for noise wave type
				if waveType == 3 {
					for n := 0; n < 32; n++ {
						randSeed = randSeed*1103515245 + 12345
						s.noiseBuffer[n] = float64(randSeed)/float64(1<<31) - 1
					}
				}
			}

			// Generate sample based on wave type
			switch waveType {
			case 0: // Square wave
				if phase/periodTemp < squareDuty {
					sample = 0.5
				} else {
					sample = -0.5
				}
			case 1: // Sawtooth wave
				sample = 1 - (phase/periodTemp)*2
			case 2: // Sine wave (polynomial approximation)
				pos := phase / periodTemp
				if pos > 0.5 {
					pos = (pos - 1) * 6.28318531
				} else {
					pos = pos * 6.28318531
				}
				if pos < 0 {
					sample = 1.27323954*pos + 0.405284735*pos*pos
				} else {
					sample = 1.27323954*pos - 0.405284735*pos*pos
				}
				if sample < 0 {
					sample = 0.225*(sample*-sample-sample) + sample
				} else {
					sample = 0.225*(sample*sample-sample) + sample
				}
			case 3: // White noise
				idx := int(math.Abs(phase*32/periodTemp)) % 32
				sample = s.noiseBuffer[idx]
			}

			// Apply filters
			if filtersEnabled {
				lpFilterOldPos = lpFilterPos
				lpFilterCutoff *= lpFilterDeltaCutoff
				if lpFilterCutoff < 0 {
					lpFilterCutoff = 0
				}
				if lpFilterCutoff > 0.1 {
					lpFilterCutoff = 0.1
				}

				if lpFilterOn {
					lpFilterDeltaPos += (sample - lpFilterPos) * lpFilterCutoff
					lpFilterDeltaPos *= lpFilterDamping
				} else {
					lpFilterPos = sample
					lpFilterDeltaPos = 0
				}

				lpFilterPos += lpFilterDeltaPos

				hpFilterPos += lpFilterPos - lpFilterOldPos
				hpFilterPos *= 1 - hpFilterCutoff
				sample = hpFilterPos
			}

			// Apply phaser effect
			if phaserEnabled {
				s.phaserBuffer[phaserPos&1023] = sample
				sample += s.phaserBuffer[(phaserPos-phaserInt+1024)&1023]
				phaserPos++
			}

			superSample += sample
		}

		// Finalize sample
		superSample *= 0.125 * envelopeVolume * masterVolume

		// Convert to 16-bit PCM with clipping
		if superSample >= 1 {
			buffer[i] = 32767
		} else if superSample <= -1 {
			buffer[i] = -32768
		} else {
			buffer[i] = int16(superSample * 32767)
		}
	}

	return length
}

// StereoSample holds left and right channel values.
type StereoSample struct {
	Left  int16
	Right int16
}

// SynthWaveStereo synthesizes stereo audio samples with panning.
// pan: -1.0 = full left, 0.0 = center, 1.0 = full right
func (s *SfxrSynth) SynthWaveStereo(bufferL, bufferR []int16, length int, pan float64) int {
	// Clamp pan to valid range
	if pan < -1 {
		pan = -1
	}
	if pan > 1 {
		pan = 1
	}

	// Calculate left/right gains using constant power panning
	// This maintains perceived loudness across the stereo field
	angle := (pan + 1) * math.Pi / 4 // 0 to π/2
	leftGain := math.Cos(angle)
	rightGain := math.Sin(angle)

	p := &s.Params

	// Filter configuration (same as SynthWave)
	filtersEnabled := p.LpFilterCutoff != 1 || p.HpFilterCutoff != 0
	hpFilterCutoff := p.HpFilterCutoff * p.HpFilterCutoff * 0.1
	hpFilterDeltaCutoff := 1 + p.HpFilterCutoffSweep*0.0003
	lpFilterCutoff := p.LpFilterCutoff * p.LpFilterCutoff * p.LpFilterCutoff * 0.1
	lpFilterDeltaCutoff := 1 + p.LpFilterCutoffSweep*0.0001
	lpFilterOn := p.LpFilterCutoff != 1
	masterVolume := p.MasterVolume * p.MasterVolume
	minFrequency := p.MinFrequency
	phaserEnabled := p.PhaserOffset != 0 || p.PhaserSweep != 0
	phaserDeltaOffset := p.PhaserSweep * p.PhaserSweep * p.PhaserSweep * 0.2

	phaserOffset := p.PhaserOffset * p.PhaserOffset
	if p.PhaserOffset < 0 {
		phaserOffset *= -1020
	} else {
		phaserOffset *= 1020
	}

	var repeatLimit int
	if p.RepeatSpeed != 0 {
		repeatLimit = int((1-p.RepeatSpeed)*(1-p.RepeatSpeed)*20000) + 32
	}

	sustainPunch := p.SustainPunch
	vibratoAmplitude := p.VibratoDepth / 2
	vibratoSpeed := p.VibratoSpeed * p.VibratoSpeed * 0.01
	waveType := p.WaveType

	// Envelope state
	envelopeLength := s.envelopeLength0
	envelopeOverLength0 := 1 / s.envelopeLength0
	envelopeOverLength1 := 1 / s.envelopeLength1
	envelopeOverLength2 := 1 / s.envelopeLength2

	// Low-pass filter damping
	lpFilterDamping := 5 / (1 + p.LpFilterResonance*p.LpFilterResonance*20) * (0.01 + lpFilterCutoff)
	if lpFilterDamping > 0.8 {
		lpFilterDamping = 0.8
	}
	lpFilterDamping = 1 - lpFilterDamping

	// Synthesis state
	finished := false
	envelopeStage := 0
	envelopeTime := 0.0
	envelopeVolume := 0.0
	hpFilterPos := 0.0
	lpFilterDeltaPos := 0.0
	lpFilterOldPos := 0.0
	lpFilterPos := 0.0
	periodTemp := 0.0
	phase := 0.0
	phaserInt := 0
	phaserPos := 0
	repeatTime := 0
	sample := 0.0
	superSample := 0.0
	vibratoPhase := 0.0

	// Clear buffers
	for i := range s.phaserBuffer {
		s.phaserBuffer[i] = 0
	}
	for i := range s.noiseBuffer {
		s.noiseBuffer[i] = pseudoRandom()*2 - 1
	}

	// Cache state
	period := s.period
	maxPeriod := s.maxPeriod
	slide := s.slide
	deltaSlide := s.deltaSlide
	changeAmount := s.changeAmount
	changeTime := s.changeTime
	changeLimit := s.changeLimit
	squareDuty := s.squareDuty
	dutySweep := s.dutySweep

	randSeed := uint32(12345)

	for i := 0; i < length; i++ {
		if finished {
			return i
		}

		// Handle repeat effect
		if repeatLimit != 0 {
			repeatTime++
			if repeatTime >= repeatLimit {
				repeatTime = 0
				s.Reset()
				period = s.period
				maxPeriod = s.maxPeriod
				slide = s.slide
				deltaSlide = s.deltaSlide
				changeAmount = s.changeAmount
				changeTime = s.changeTime
				changeLimit = s.changeLimit
				squareDuty = s.squareDuty
				dutySweep = s.dutySweep
			}
		}

		// Handle pitch change
		if changeLimit != 0 {
			changeTime++
			if changeTime >= changeLimit {
				changeLimit = 0
				period *= changeAmount
			}
		}

		// Apply frequency slide
		slide += deltaSlide
		period *= slide

		if period > maxPeriod {
			period = maxPeriod
			if minFrequency > 0 {
				finished = true
			}
		}

		periodTemp = period

		// Apply vibrato
		if vibratoAmplitude > 0 {
			vibratoPhase += vibratoSpeed
			periodTemp *= 1 + math.Sin(vibratoPhase)*vibratoAmplitude
		}

		periodTempInt := int(periodTemp)
		if periodTempInt < 8 {
			periodTempInt = 8
		}
		periodTemp = float64(periodTempInt)

		// Square wave duty sweep
		if waveType == 0 {
			squareDuty += dutySweep
			if squareDuty < 0 {
				squareDuty = 0
			}
			if squareDuty > 0.5 {
				squareDuty = 0.5
			}
		}

		// Envelope stage progression
		envelopeTime++
		if envelopeTime > envelopeLength {
			envelopeTime = 0
			envelopeStage++

			if envelopeStage == 1 {
				envelopeLength = s.envelopeLength1
			} else if envelopeStage == 2 {
				envelopeLength = s.envelopeLength2
			}
		}

		// Calculate envelope volume
		switch envelopeStage {
		case 0:
			envelopeVolume = envelopeTime * envelopeOverLength0
		case 1:
			envelopeVolume = 1 + (1-envelopeTime*envelopeOverLength1)*2*sustainPunch
		case 2:
			envelopeVolume = 1 - envelopeTime*envelopeOverLength2
		case 3:
			envelopeVolume = 0
			finished = true
		}

		// Update phaser offset
		if phaserEnabled {
			phaserOffset += phaserDeltaOffset
			phaserInt = int(math.Abs(phaserOffset))
			if phaserInt > 1023 {
				phaserInt = 1023
			}
		}

		// Update high-pass filter cutoff
		if filtersEnabled && hpFilterDeltaCutoff != 1 {
			hpFilterCutoff *= hpFilterDeltaCutoff
			if hpFilterCutoff < 0.00001 {
				hpFilterCutoff = 0.00001
			}
			if hpFilterCutoff > 0.1 {
				hpFilterCutoff = 0.1
			}
		}

		// 8x Oversampling loop
		superSample = 0

		for j := 0; j < 8; j++ {
			phase++
			if phase >= periodTemp {
				phase = math.Mod(phase, periodTemp)

				if waveType == 3 {
					for n := 0; n < 32; n++ {
						randSeed = randSeed*1103515245 + 12345
						s.noiseBuffer[n] = float64(randSeed)/float64(1<<31) - 1
					}
				}
			}

			switch waveType {
			case 0:
				if phase/periodTemp < squareDuty {
					sample = 0.5
				} else {
					sample = -0.5
				}
			case 1:
				sample = 1 - (phase/periodTemp)*2
			case 2:
				pos := phase / periodTemp
				if pos > 0.5 {
					pos = (pos - 1) * 6.28318531
				} else {
					pos = pos * 6.28318531
				}
				if pos < 0 {
					sample = 1.27323954*pos + 0.405284735*pos*pos
				} else {
					sample = 1.27323954*pos - 0.405284735*pos*pos
				}
				if sample < 0 {
					sample = 0.225*(sample*-sample-sample) + sample
				} else {
					sample = 0.225*(sample*sample-sample) + sample
				}
			case 3:
				idx := int(math.Abs(phase*32/periodTemp)) % 32
				sample = s.noiseBuffer[idx]
			}

			if filtersEnabled {
				lpFilterOldPos = lpFilterPos
				lpFilterCutoff *= lpFilterDeltaCutoff
				if lpFilterCutoff < 0 {
					lpFilterCutoff = 0
				}
				if lpFilterCutoff > 0.1 {
					lpFilterCutoff = 0.1
				}

				if lpFilterOn {
					lpFilterDeltaPos += (sample - lpFilterPos) * lpFilterCutoff
					lpFilterDeltaPos *= lpFilterDamping
				} else {
					lpFilterPos = sample
					lpFilterDeltaPos = 0
				}

				lpFilterPos += lpFilterDeltaPos

				hpFilterPos += lpFilterPos - lpFilterOldPos
				hpFilterPos *= 1 - hpFilterCutoff
				sample = hpFilterPos
			}

			if phaserEnabled {
				s.phaserBuffer[phaserPos&1023] = sample
				sample += s.phaserBuffer[(phaserPos-phaserInt+1024)&1023]
				phaserPos++
			}

			superSample += sample
		}

		// Finalize sample
		superSample *= 0.125 * envelopeVolume * masterVolume

		// Apply panning and convert to 16-bit PCM
		leftSample := superSample * leftGain
		rightSample := superSample * rightGain

		if leftSample >= 1 {
			bufferL[i] = 32767
		} else if leftSample <= -1 {
			bufferL[i] = -32768
		} else {
			bufferL[i] = int16(leftSample * 32767)
		}

		if rightSample >= 1 {
			bufferR[i] = 32767
		} else if rightSample <= -1 {
			bufferR[i] = -32768
		} else {
			bufferR[i] = int16(rightSample * 32767)
		}
	}

	return length
}

// Deterministic pseudo-random for reproducible tests
var prngState uint32 = 12345

func pseudoRandom() float64 {
	prngState = prngState*1103515245 + 12345
	return float64(prngState) / float64(1<<32)
}

// ResetPRNG resets the pseudo-random number generator for reproducible results.
func ResetPRNG(seed uint32) {
	prngState = seed
}

// GenerateWavDataURL generates a sound effect and returns it as a data URL.
func GenerateWavDataURL(settings string) string {
	synth := NewSfxrSynth()
	synth.Params.ParseSettingsString(settings)

	envelopeFullLength := synth.TotalReset()

	// Calculate buffer size
	sampleCount := (envelopeFullLength + 1) / 2
	dataSize := sampleCount*4 + 44

	// Allocate buffer
	data := make([]byte, dataSize)

	// Generate audio samples
	samples := make([]int16, envelopeFullLength)
	samplesWritten := synth.SynthWave(samples, envelopeFullLength)
	bytesUsed := samplesWritten * 2

	// Write WAV header
	writeWavHeader(data, bytesUsed)

	// Write sample data (little-endian)
	for i := 0; i < samplesWritten; i++ {
		data[44+i*2] = byte(samples[i])
		data[44+i*2+1] = byte(samples[i] >> 8)
	}

	totalBytes := bytesUsed + 44

	// Base64 encode
	encoded := base64.StdEncoding.EncodeToString(data[:totalBytes])

	return "data:audio/wav;base64," + encoded
}

// writeWavHeader writes a WAV file header to the buffer.
func writeWavHeader(data []byte, dataSize int) {
	// RIFF header
	data[0] = 'R'
	data[1] = 'I'
	data[2] = 'F'
	data[3] = 'F'
	writeUint32LE(data, 4, uint32(dataSize+36))
	data[8] = 'W'
	data[9] = 'A'
	data[10] = 'V'
	data[11] = 'E'

	// fmt sub-chunk
	data[12] = 'f'
	data[13] = 'm'
	data[14] = 't'
	data[15] = ' '
	writeUint32LE(data, 16, 16)    // Sub-chunk size
	writeUint16LE(data, 20, 1)     // Audio format (PCM)
	writeUint16LE(data, 22, 1)     // Channels (mono)
	writeUint32LE(data, 24, 44100) // Sample rate
	writeUint32LE(data, 28, 88200) // Byte rate
	writeUint16LE(data, 32, 2)     // Block align
	writeUint16LE(data, 34, 16)    // Bits per sample

	// data sub-chunk
	data[36] = 'd'
	data[37] = 'a'
	data[38] = 't'
	data[39] = 'a'
	writeUint32LE(data, 40, uint32(dataSize))
}

func writeUint16LE(data []byte, offset int, value uint16) {
	data[offset] = byte(value)
	data[offset+1] = byte(value >> 8)
}

func writeUint32LE(data []byte, offset int, value uint32) {
	data[offset] = byte(value)
	data[offset+1] = byte(value >> 8)
	data[offset+2] = byte(value >> 16)
	data[offset+3] = byte(value >> 24)
}

// GenerateFloat32Buffer generates a sound effect as float32 samples for Web Audio API.
func GenerateFloat32Buffer(settings string) []float32 {
	synth := NewSfxrSynth()
	synth.Params.ParseSettingsString(settings)

	length := synth.TotalReset()

	// Generate samples
	tempBuffer := make([]int16, length)
	samplesWritten := synth.SynthWave(tempBuffer, length)

	// Convert to float32
	result := make([]float32, samplesWritten)
	for i := 0; i < samplesWritten; i++ {
		result[i] = float32(tempBuffer[i]) / 32768
	}

	return result
}

// GenerateStereoWavDataURL generates a stereo sound effect with panning.
// pan: -1.0 = full left, 0.0 = center, 1.0 = full right
func GenerateStereoWavDataURL(settings string, pan float64) string {
	synth := NewSfxrSynth()
	synth.Params.ParseSettingsString(settings)

	envelopeFullLength := synth.TotalReset()

	// Calculate buffer size for stereo (2 channels)
	sampleCount := (envelopeFullLength + 1) / 2
	dataSize := sampleCount*8 + 44 // 4 bytes per stereo sample pair

	data := make([]byte, dataSize)

	// Generate stereo audio samples
	samplesL := make([]int16, envelopeFullLength)
	samplesR := make([]int16, envelopeFullLength)
	samplesWritten := synth.SynthWaveStereo(samplesL, samplesR, envelopeFullLength, pan)
	bytesUsed := samplesWritten * 4 // 2 bytes per channel * 2 channels

	// Write stereo WAV header
	writeStereoWavHeader(data, bytesUsed)

	// Write interleaved stereo sample data (L, R, L, R, ...)
	for i := 0; i < samplesWritten; i++ {
		offset := 44 + i*4
		data[offset] = byte(samplesL[i])
		data[offset+1] = byte(samplesL[i] >> 8)
		data[offset+2] = byte(samplesR[i])
		data[offset+3] = byte(samplesR[i] >> 8)
	}

	totalBytes := bytesUsed + 44
	encoded := base64.StdEncoding.EncodeToString(data[:totalBytes])

	return "data:audio/wav;base64," + encoded
}

// writeStereoWavHeader writes a stereo WAV file header.
func writeStereoWavHeader(data []byte, dataSize int) {
	// RIFF header
	data[0] = 'R'
	data[1] = 'I'
	data[2] = 'F'
	data[3] = 'F'
	writeUint32LE(data, 4, uint32(dataSize+36))
	data[8] = 'W'
	data[9] = 'A'
	data[10] = 'V'
	data[11] = 'E'

	// fmt sub-chunk
	data[12] = 'f'
	data[13] = 'm'
	data[14] = 't'
	data[15] = ' '
	writeUint32LE(data, 16, 16)     // Sub-chunk size
	writeUint16LE(data, 20, 1)      // Audio format (PCM)
	writeUint16LE(data, 22, 2)      // Channels (stereo)
	writeUint32LE(data, 24, 44100)  // Sample rate
	writeUint32LE(data, 28, 176400) // Byte rate (44100 * 2 channels * 2 bytes)
	writeUint16LE(data, 32, 4)      // Block align (2 channels * 2 bytes)
	writeUint16LE(data, 34, 16)     // Bits per sample

	// data sub-chunk
	data[36] = 'd'
	data[37] = 'a'
	data[38] = 't'
	data[39] = 'a'
	writeUint32LE(data, 40, uint32(dataSize))
}

// GenerateStereoFloat32Buffer generates stereo float32 samples.
// Returns two slices: left channel and right channel.
func GenerateStereoFloat32Buffer(settings string, pan float64) ([]float32, []float32) {
	synth := NewSfxrSynth()
	synth.Params.ParseSettingsString(settings)

	length := synth.TotalReset()

	// Generate stereo samples
	tempL := make([]int16, length)
	tempR := make([]int16, length)
	samplesWritten := synth.SynthWaveStereo(tempL, tempR, length, pan)

	// Convert to float32
	resultL := make([]float32, samplesWritten)
	resultR := make([]float32, samplesWritten)
	for i := 0; i < samplesWritten; i++ {
		resultL[i] = float32(tempL[i]) / 32768
		resultR[i] = float32(tempR[i]) / 32768
	}

	return resultL, resultR
}
