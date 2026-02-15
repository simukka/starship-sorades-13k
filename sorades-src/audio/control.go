package audio

import (
	"bytes"
	_ "embed"
	"text/template"

	"github.com/gopherjs/gopherjs/js"
)

// ControlPanelData holds all data needed to render the control panel template
type ControlPanelData struct {
	Categories    []SfxCategory
	CurrentPreset string
	Presets       []PresetInfo
}

// SfxCategory groups sound effects by category for display
type SfxCategory struct {
	Name    string
	Effects []SfxTemplateData
}

// SfxTemplateData holds sound effect data for template rendering
type SfxTemplateData struct {
	ID           int
	Name         string
	WaveTypeName string
}

//go:embed control.gohtml
var controlHtml string

// InitControlPanel creates the audio control panel and attaches right-click handler.
func (am *AudioManager) InitControlPanel(canvas *js.Object) {
	doc := js.Global.Get("document")

	// Create control panel container
	panel := doc.Call("createElement", "div")
	panel.Set("id", "audio-control-panel")
	panel.Get("style").Set("cssText", `
		position: fixed;
		top: 50%;
		left: 50%;
		transform: translate(-50%, -50%);
		background: rgba(20, 20, 30, 0.95);
		border: 2px solid #4a9eff;
		border-radius: 8px;
		padding: 20px;
		color: #fff;
		font-family: 'Courier New', monospace;
		font-size: 12px;
		z-index: 10000;
		display: none;
		max-height: 80vh;
		overflow-y: auto;
		min-width: 400px;
		box-shadow: 0 0 30px rgba(74, 158, 255, 0.3);
	`)

	// Build panel content
	panel.Set("innerHTML", am.buildControlPanelHTML())

	doc.Get("body").Call("appendChild", panel)
	am.controlPanel = panel

	// Right-click handler to show panel
	canvas.Call("addEventListener", "contextmenu", func(e *js.Object) {
		e.Call("preventDefault")
		am.toggleControlPanel()
	})

	// Close button handler
	closeBtn := doc.Call("getElementById", "audio-panel-close")
	if closeBtn != nil {
		closeBtn.Call("addEventListener", "click", func() {
			am.hideControlPanel()
		})
	}

	// Attach slider handlers after panel is in DOM
	am.attachSliderHandlers()
}

// buildControlPanelHTML generates the HTML for the control panel.
func (am *AudioManager) buildControlPanelHTML() string {
	// Build template data
	data := ControlPanelData{
		CurrentPreset: am.GetCurrentPresetName(),
		Categories:    make([]SfxCategory, 0),
		Presets:       GetAllPresetInfo(),
	}

	// Group SFX by category
	categoryNames := []string{"Player", "Enemy", "Pickup", "UI"}
	for _, catName := range categoryNames {
		cat := SfxCategory{Name: catName, Effects: make([]SfxTemplateData, 0)}
		for _, sfx := range SoundEffectLibrary {
			if sfx.Category == catName {
				cat.Effects = append(cat.Effects, SfxTemplateData{
					ID:           sfx.ID,
					Name:         sfx.Name,
					WaveTypeName: sfx.WaveType.String(),
				})
			}
		}
		if len(cat.Effects) > 0 {
			data.Categories = append(data.Categories, cat)
		}
	}

	// Parse and execute template
	tmpl, err := template.New("controlPanel").Parse(controlHtml)
	if err != nil {
		return "<div style='color:red'>Template error: " + err.Error() + "</div>"
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "<div style='color:red'>Execute error: " + err.Error() + "</div>"
	}

	return buf.String()
}

// attachSliderHandlers connects all sliders to their respective audio parameters.
func (am *AudioManager) attachSliderHandlers() {
	doc := js.Global.Get("document")

	// Helper to attach a slider
	attachSlider := func(id string, suffix string, handler func(float64)) {
		slider := doc.Call("getElementById", id)
		valSpan := doc.Call("getElementById", id+"-val")
		if slider == nil || slider == js.Undefined {
			return
		}
		slider.Call("addEventListener", "input", func(e *js.Object) {
			val := e.Get("target").Get("value").Float()
			if valSpan != nil && valSpan != js.Undefined {
				if suffix == "Hz" {
					valSpan.Set("textContent", js.Global.Get("Math").Call("round", val).String()+suffix)
				} else {
					valSpan.Set("textContent", js.Global.Get("Math").Call("round", val).String()+suffix)
				}
			}
			handler(val)
		})
	}

	// Master controls
	attachSlider("ctrl-master-vol", "%", func(v float64) {
		if am.masterGain != nil {
			am.masterGain.Get("gain").Set("value", v/100)
		}
	})

	attachSlider("ctrl-reverb-mix", "%", func(v float64) {
		if am.reverbGain != nil {
			am.reverbGain.Get("gain").Set("value", v/100)
		}
	})

	// Synth controls
	attachSlider("ctrl-synth-vol", "%", func(v float64) {
		if am.synthGain != nil {
			am.synthGain.Get("gain").Set("value", v/100)
		}
	})

	attachSlider("ctrl-drone-vol", "%", func(v float64) {
		// Update drone pad gains (first 5 gains in synthGains are drone)
		for i := 0; i < 5 && i < len(am.synthGains); i++ {
			if am.synthGains[i] != nil {
				am.synthGains[i].Get("gain").Set("value", v/100)
			}
		}
	})

	attachSlider("ctrl-arp-vol", "%", func(v float64) {
		// Arpeggio volume is handled dynamically, store for reference
		// This affects the envelope peak in playArpNote
	})

	attachSlider("ctrl-bass-vol", "%", func(v float64) {
		// Bass gains are indices 5-6 in synthGains (if bass line started)
		for i := 5; i < 7 && i < len(am.synthGains); i++ {
			if am.synthGains[i] != nil {
				am.synthGains[i].Get("gain").Set("value", v/100)
			}
		}
	})

	// Reactive layer controls
	attachSlider("ctrl-reactive-vol", "%", func(v float64) {
		if am.reactiveGain != nil {
			am.reactiveGain.Get("gain").Set("value", v/100)
		}
	})

	attachSlider("ctrl-reactive-reverb", "%", func(v float64) {
		if am.reactiveReverb != nil {
			am.reactiveReverb.Get("gain").Set("value", v/100)
		}
	})

	attachSlider("ctrl-tension-vol", "%", func(v float64) {
		if am.tensionGain != nil {
			am.tensionGain.Get("gain").Set("value", v/100)
		}
	})

	attachSlider("ctrl-tension-filter", "Hz", func(v float64) {
		if am.tensionFilter != nil {
			am.tensionFilter.Get("frequency").Set("value", v)
		}
	})

	// Blade Runner FX controls
	attachSlider("ctrl-shimmer-vol", "%", func(v float64) {
		if am.shimmerGain != nil {
			am.shimmerGain.Get("gain").Set("value", v/100)
		}
	})

	attachSlider("ctrl-siren-vol", "%", func(v float64) {
		if am.sirenGain != nil {
			am.sirenGain.Get("gain").Set("value", v/100)
		}
	})

	attachSlider("ctrl-subbass-vol", "%", func(v float64) {
		if am.subBassGain != nil {
			am.subBassGain.Get("gain").Set("value", v/100)
		}
	})

	attachSlider("ctrl-pad-vol", "%", func(v float64) {
		if am.padGain != nil {
			am.padGain.Get("gain").Set("value", v/100)
		}
	})

	attachSlider("ctrl-pad-filter", "Hz", func(v float64) {
		if am.padFilter != nil {
			am.padFilter.Get("frequency").Set("value", v)
		}
	})

	attachSlider("ctrl-pulse-vol", "%", func(v float64) {
		if am.pulseGain != nil {
			am.pulseGain.Get("gain").Set("value", v/100)
		}
	})

	// Tab navigation
	tabBtns := doc.Call("querySelectorAll", ".tab-btn")
	for i := 0; i < tabBtns.Length(); i++ {
		btn := tabBtns.Index(i)
		btn.Call("addEventListener", "click", func(e *js.Object) {
			tabName := e.Get("currentTarget").Call("getAttribute", "data-tab").String()

			// Update button states
			allBtns := doc.Call("querySelectorAll", ".tab-btn")
			for j := 0; j < allBtns.Length(); j++ {
				allBtns.Index(j).Get("classList").Call("remove", "active")
			}
			e.Get("currentTarget").Get("classList").Call("add", "active")

			// Update tab content
			allTabs := doc.Call("querySelectorAll", ".tab-content")
			for j := 0; j < allTabs.Length(); j++ {
				allTabs.Index(j).Get("style").Set("display", "none")
				allTabs.Index(j).Get("classList").Call("remove", "active")
			}
			activeTab := doc.Call("getElementById", "tab-"+tabName)
			if activeTab != nil && activeTab != js.Undefined {
				activeTab.Get("style").Set("display", "block")
				activeTab.Get("classList").Call("add", "active")
			}
		})
	}

	// SFX play buttons
	playBtns := doc.Call("querySelectorAll", ".sfx-play-btn")
	for i := 0; i < playBtns.Length(); i++ {
		btn := playBtns.Index(i)
		btn.Call("addEventListener", "click", func(e *js.Object) {
			idStr := e.Get("currentTarget").Call("getAttribute", "data-id").String()
			id := int(js.Global.Call("parseInt", idStr).Int())
			am.Play(id)
		})
	}

	// SFX edit buttons
	editBtns := doc.Call("querySelectorAll", ".sfx-edit-btn")
	for i := 0; i < editBtns.Length(); i++ {
		btn := editBtns.Index(i)
		btn.Call("addEventListener", "click", func(e *js.Object) {
			idStr := e.Get("currentTarget").Call("getAttribute", "data-id").String()
			id := int(js.Global.Call("parseInt", idStr).Int())
			am.openSfxEditor(id)
		})
	}

	// SFX editor close button
	sfxEditorClose := doc.Call("getElementById", "sfx-editor-close")
	if sfxEditorClose != nil && sfxEditorClose != js.Undefined {
		sfxEditorClose.Call("addEventListener", "click", func() {
			doc.Call("getElementById", "sfx-editor").Get("style").Set("display", "none")
		})
	}

	// SFX preview button
	sfxPreview := doc.Call("getElementById", "sfx-preview")
	if sfxPreview != nil && sfxPreview != js.Undefined {
		sfxPreview.Call("addEventListener", "click", func() {
			am.previewEditedSfx()
		})
	}

	// SFX apply button
	sfxApply := doc.Call("getElementById", "sfx-apply")
	if sfxApply != nil && sfxApply != js.Undefined {
		sfxApply.Call("addEventListener", "click", func() {
			am.applyEditedSfx()
		})
	}

	// SFX reset button
	sfxReset := doc.Call("getElementById", "sfx-reset")
	if sfxReset != nil && sfxReset != js.Undefined {
		sfxReset.Call("addEventListener", "click", func() {
			if am.editingSfxID >= 0 {
				am.openSfxEditor(am.editingSfxID)
			}
		})
	}

	// Music preset buttons
	presetBtns := doc.Call("querySelectorAll", ".preset-btn")
	for i := 0; i < presetBtns.Length(); i++ {
		btn := presetBtns.Index(i)
		btn.Call("addEventListener", "click", func(e *js.Object) {
			levelStr := e.Get("currentTarget").Call("getAttribute", "data-level").String()
			level := int(js.Global.Call("parseInt", levelStr).Int())
			am.SetMusicPreset(level)

			// Update display
			presetName := doc.Call("getElementById", "current-preset-name")
			if presetName != nil && presetName != js.Undefined {
				presetName.Set("textContent", "Current: "+am.GetCurrentPresetName())
			}
		})
	}

	// SFX editor slider handlers
	am.attachSfxEditorSliders()
}

// openSfxEditor opens the SFX editor for a specific sound effect
func (am *AudioManager) openSfxEditor(id int) {
	if id < 0 || id >= len(SoundEffectLibrary) {
		return
	}

	am.editingSfxID = id
	sfx := SoundEffectLibrary[id]
	doc := js.Global.Get("document")

	// Show editor
	editor := doc.Call("getElementById", "sfx-editor")
	if editor != nil && editor != js.Undefined {
		editor.Get("style").Set("display", "block")
	}

	// Set title
	title := doc.Call("getElementById", "sfx-editor-title")
	if title != nil && title != js.Undefined {
		title.Set("textContent", "Edit: "+sfx.Name)
	}

	// Set wave type
	waveSelect := doc.Call("getElementById", "sfx-wave")
	if waveSelect != nil && waveSelect != js.Undefined {
		waveSelect.Set("value", int(sfx.WaveType))
	}

	// Helper to set slider
	setSlider := func(baseID string, value float64, scale float64) {
		slider := doc.Call("getElementById", baseID)
		valSpan := doc.Call("getElementById", baseID+"-val")
		if slider != nil && slider != js.Undefined {
			slider.Set("value", value*scale)
		}
		if valSpan != nil && valSpan != js.Undefined {
			valSpan.Set("textContent", js.Global.Get("Math").Call("round", value*scale).String()+"%")
		}
	}

	setSlider("sfx-volume", sfx.MasterVolume, 100)
	setSlider("sfx-attack", sfx.AttackTime, 100)
	setSlider("sfx-sustain", sfx.SustainTime, 100)
	setSlider("sfx-decay", sfx.DecayTime, 100)
	setSlider("sfx-freq", sfx.StartFrequency, 100)

	// Slide is -1 to 1, map to -100 to 100
	slideSlider := doc.Call("getElementById", "sfx-slide")
	slideVal := doc.Call("getElementById", "sfx-slide-val")
	if slideSlider != nil && slideSlider != js.Undefined {
		slideSlider.Set("value", sfx.Slide*100)
	}
	if slideVal != nil && slideVal != js.Undefined {
		slideVal.Set("textContent", js.Global.Get("Math").Call("round", sfx.Slide*100).String())
	}

	setSlider("sfx-vibrato", sfx.VibratoDepth, 100)
}

// attachSfxEditorSliders attaches handlers to SFX editor sliders
func (am *AudioManager) attachSfxEditorSliders() {
	doc := js.Global.Get("document")

	// Generic slider handler for percentage sliders
	attachSfxSlider := func(id string) {
		slider := doc.Call("getElementById", id)
		valSpan := doc.Call("getElementById", id+"-val")
		if slider == nil || slider == js.Undefined {
			return
		}
		slider.Call("addEventListener", "input", func(e *js.Object) {
			val := e.Get("target").Get("value").Float()
			if valSpan != nil && valSpan != js.Undefined {
				if id == "sfx-slide" {
					valSpan.Set("textContent", js.Global.Get("Math").Call("round", val).String())
				} else {
					valSpan.Set("textContent", js.Global.Get("Math").Call("round", val).String()+"%")
				}
			}
		})
	}

	attachSfxSlider("sfx-volume")
	attachSfxSlider("sfx-attack")
	attachSfxSlider("sfx-sustain")
	attachSfxSlider("sfx-decay")
	attachSfxSlider("sfx-freq")
	attachSfxSlider("sfx-slide")
	attachSfxSlider("sfx-vibrato")
}

// previewEditedSfx plays the sound with current editor values
func (am *AudioManager) previewEditedSfx() {
	if am.editingSfxID < 0 {
		return
	}

	// Build temporary jsfxr string from editor values and play it
	doc := js.Global.Get("document")

	getVal := func(id string) float64 {
		el := doc.Call("getElementById", id)
		if el == nil || el == js.Undefined {
			return 0
		}
		return el.Get("value").Float() / 100
	}

	waveType := 0
	waveSelect := doc.Call("getElementById", "sfx-wave")
	if waveSelect != nil && waveSelect != js.Undefined {
		waveType = int(js.Global.Call("parseInt", waveSelect.Get("value").String()).Int())
	}

	// Get original SFX for values we're not editing
	origSfx := SoundEffectLibrary[am.editingSfxID]

	// Build jsfxr string
	params := js.Global.Get("Array").Call("of",
		waveType,
		getVal("sfx-attack"),
		getVal("sfx-sustain"),
		origSfx.SustainPunch,
		getVal("sfx-decay"),
		getVal("sfx-freq"),
		origSfx.MinFrequency,
		getVal("sfx-slide"),
		origSfx.DeltaSlide,
		getVal("sfx-vibrato"),
		origSfx.VibratoSpeed,
		origSfx.ArpChange,
		origSfx.ArpSpeed,
		origSfx.SquareDuty,
		origSfx.DutySweep,
		origSfx.RepeatSpeed,
		origSfx.PhaserOffset,
		origSfx.PhaserSweep,
		origSfx.LPFilterCutoff,
		origSfx.LPFilterCutoffSweep,
		origSfx.LPFilterResonance,
		origSfx.HPFilterCutoff,
		origSfx.HPFilterCutoffSweep,
		getVal("sfx-volume"),
	).Call("join", ",").String()

	// Generate and play the sound
	am.playJsfxrPreview(params)
}

// playJsfxrPreview plays a jsfxr parameter string directly
func (am *AudioManager) playJsfxrPreview(params string) {
	if !am.ready {
		return
	}

	// Use jsfxr to generate audio
	jsfxr := js.Global.Get("jsfxr")
	if jsfxr == nil || jsfxr == js.Undefined {
		return
	}

	dataURL := jsfxr.Invoke(params)
	if dataURL == nil || dataURL == js.Undefined {
		return
	}

	// Create audio element and play
	audio := js.Global.Get("Audio").New(dataURL.String())
	audio.Call("play")
}

// applyEditedSfx applies the edited values to the sound effect
func (am *AudioManager) applyEditedSfx() {
	if am.editingSfxID < 0 || am.editingSfxID >= len(SoundEffectLibrary) {
		return
	}

	doc := js.Global.Get("document")
	sfx := SoundEffectLibrary[am.editingSfxID]

	getVal := func(id string) float64 {
		el := doc.Call("getElementById", id)
		if el == nil || el == js.Undefined {
			return 0
		}
		return el.Get("value").Float() / 100
	}

	waveSelect := doc.Call("getElementById", "sfx-wave")
	if waveSelect != nil && waveSelect != js.Undefined {
		sfx.WaveType = WaveType(js.Global.Call("parseInt", waveSelect.Get("value").String()).Int())
	}

	sfx.MasterVolume = getVal("sfx-volume")
	sfx.AttackTime = getVal("sfx-attack")
	sfx.SustainTime = getVal("sfx-sustain")
	sfx.DecayTime = getVal("sfx-decay")
	sfx.StartFrequency = getVal("sfx-freq")
	sfx.Slide = getVal("sfx-slide")
	sfx.VibratoDepth = getVal("sfx-vibrato")

	// Regenerate and reload the sound
	newParams := sfx.ToJsfxrString()
	SfxData[am.editingSfxID] = newParams
	am.ReloadSound(am.editingSfxID)
}

// ReloadSound reloads a sound effect from its SfxData
func (am *AudioManager) ReloadSound(id int) {
	if id < 0 || id >= len(SfxData) || !am.ready {
		return
	}

	jsfxr := js.Global.Get("jsfxr")
	if jsfxr == nil || jsfxr == js.Undefined {
		return
	}

	dataURL := jsfxr.Invoke(SfxData[id])
	if dataURL == nil || dataURL == js.Undefined {
		return
	}

	am.LoadSound(id, dataURL.String())
}

// toggleControlPanel shows or hides the control panel.
func (am *AudioManager) toggleControlPanel() {
	if am.controlPanel == nil {
		return
	}
	current := am.controlPanel.Get("style").Get("display").String()
	if current == "none" {
		am.controlPanel.Get("style").Set("display", "block")
	} else {
		am.controlPanel.Get("style").Set("display", "none")
	}
}

// hideControlPanel hides the control panel.
func (am *AudioManager) hideControlPanel() {
	if am.controlPanel != nil {
		am.controlPanel.Get("style").Set("display", "none")
	}
}

// showControlPanel shows the control panel.
func (am *AudioManager) showControlPanel() {
	if am.controlPanel != nil {
		am.controlPanel.Get("style").Set("display", "block")
	}
}
