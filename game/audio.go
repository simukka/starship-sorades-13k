package game

import (
	"github.com/gopherjs/gopherjs/js"
)

// AudioManager manages sound effects using the Web Audio API.
type AudioManager struct {
	ctx         *js.Object
	masterGain  *js.Object
	buffers     map[int]*js.Object
	ready       bool
	AudioCtx    *js.Object // Exposed for state checking
	musicSource *js.Object
	musicGain   *js.Object
	titleSource *js.Object
	titleGain   *js.Object
}

// NewAudioManager creates a new audio manager.
func NewAudioManager() *AudioManager {
	return &AudioManager{
		buffers: make(map[int]*js.Object),
	}
}

// Init initializes the Web Audio context.
func (am *AudioManager) Init() {
	if am.ctx != nil {
		return
	}

	// Try to create AudioContext
	audioCtx := js.Global.Get("AudioContext")
	if audioCtx == nil || audioCtx == js.Undefined {
		audioCtx = js.Global.Get("webkitAudioContext")
	}
	if audioCtx == nil || audioCtx == js.Undefined {
		return
	}

	am.ctx = audioCtx.New()
	am.AudioCtx = am.ctx // Expose for state checking
	am.masterGain = am.ctx.Call("createGain")
	am.masterGain.Call("connect", am.ctx.Get("destination"))
	am.masterGain.Get("gain").Set("value", 0.7)
	am.ready = true
}

// LoadSound loads and decodes a sound effect from a jsfxr data URL.
func (am *AudioManager) LoadSound(id int, dataURL string) {
	if am.ctx == nil {
		return
	}

	// Fetch and decode the audio data
	fetchPromise := js.Global.Call("fetch", dataURL)
	fetchPromise.Call("then", func(response *js.Object) {
		arrayBufferPromise := response.Call("arrayBuffer")
		arrayBufferPromise.Call("then", func(arrayBuffer *js.Object) {
			decodePromise := am.ctx.Call("decodeAudioData", arrayBuffer)
			decodePromise.Call("then", func(audioBuffer *js.Object) {
				am.buffers[id] = audioBuffer
			})
		})
	})
}

// Play plays a sound effect by ID.
func (am *AudioManager) Play(id int) {
	if !am.ready {
		return
	}
	buffer, ok := am.buffers[id]
	if !ok || buffer == nil {
		return
	}

	// Resume context if suspended
	if am.ctx.Get("state").String() == "suspended" {
		am.ctx.Call("resume")
	}

	// Create and play buffer source
	source := am.ctx.Call("createBufferSource")
	source.Set("buffer", buffer)
	source.Call("connect", am.masterGain)
	source.Call("start", 0)
}

// SetVolume sets the master volume (0.0 to 1.0).
func (am *AudioManager) SetVolume(volume float64) {
	if am.masterGain == nil {
		return
	}
	if volume < 0 {
		volume = 0
	}
	if volume > 1 {
		volume = 1
	}
	am.masterGain.Get("gain").Set("value", volume)
}

// PlayTitle plays the title screen music loop.
func (am *AudioManager) PlayTitle() {
	if !am.ready {
		return
	}
	// Play intro sound
	am.Play(21)
}

// FadeOutTitle fades out the title music.
func (am *AudioManager) FadeOutTitle() {
	if !am.ready || am.titleGain == nil {
		return
	}
	// Fade out title gain
	currentTime := am.ctx.Get("currentTime").Float()
	am.titleGain.Get("gain").Call("linearRampToValueAtTime", 0, currentTime+0.5)
}

// PlayMusic starts the game music.
func (am *AudioManager) PlayMusic() {
	// Game music would be loaded separately
	// For now, this is a stub
}

// StopMusic stops the game music.
func (am *AudioManager) StopMusic() {
	if !am.ready || am.musicSource == nil {
		return
	}
	am.musicSource.Call("stop")
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
}
