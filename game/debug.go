package game

import (
	"strconv"

	"github.com/gopherjs/gopherjs/js"
)

var EnableDebug = true

// Debug logs a message to the browser console if debug mode is enabled.
func Debug(args ...interface{}) {
	if EnableDebug {
		js.Global.Get("console").Call("log", args...)
	}
}

// Debugf logs a formatted message to the browser console if debug mode is enabled.
func Debugf(format string, args ...interface{}) {
	if EnableDebug {
		js.Global.Get("console").Call("log", format, args)
	}
}

// DebugWarn logs a warning to the browser console if debug mode is enabled.
func DebugWarn(args ...interface{}) {
	if EnableDebug {
		js.Global.Get("console").Call("warn", args...)
	}
}

// DebugError logs an error to the browser console if debug mode is enabled.
func DebugError(args ...interface{}) {
	if EnableDebug {
		js.Global.Get("console").Call("error", args...)
	}
}

// DebugUI holds the state for the enemy config debug panel
type DebugUI struct {
	Visible          bool
	SelectedEnemy    EnemyKind
	SelectedField    int
	PanelX           int
	PanelY           int
	PanelWidth       int
	PanelHeight      int
	FieldNames       []string
	ScrollOffset     int
	MaxVisibleFields int
}

// NewDebugUI creates a new debug UI instance
func NewDebugUI() *DebugUI {
	return &DebugUI{
		Visible:       false,
		SelectedEnemy: SmallFighter,
		SelectedField: 0,
		PanelX:        16,
		PanelY:        16,
		PanelWidth:    350,
		PanelHeight:   400,
		FieldNames: []string{
			"CountBase",
			"CountPerLevel",
			"MaxAngle",
			"HealthBase",
			"HealthPerLevel",
			"TBase",
			"TRange",
			"TStart",
			"YStopBase",
			"YStopRange",
			"YOffsetMult",
			"UseYStep",
			"HasFireDir",
		},
		MaxVisibleFields: 10,
	}
}

// Toggle toggles the debug UI visibility
func (d *DebugUI) Toggle() {
	d.Visible = !d.Visible
}

// NextEnemy cycles to the next enemy type
func (d *DebugUI) NextEnemy() {
	d.SelectedEnemy = (d.SelectedEnemy + 1) % 4
	d.SelectedField = 0
	d.ScrollOffset = 0
}

// PrevEnemy cycles to the previous enemy type
func (d *DebugUI) PrevEnemy() {
	d.SelectedEnemy = (d.SelectedEnemy + 3) % 4 // +3 is same as -1 mod 4
	d.SelectedField = 0
	d.ScrollOffset = 0
}

// NextField moves to the next field
func (d *DebugUI) NextField() {
	d.SelectedField = (d.SelectedField + 1) % len(d.FieldNames)
	// Adjust scroll if needed
	if d.SelectedField >= d.ScrollOffset+d.MaxVisibleFields {
		d.ScrollOffset = d.SelectedField - d.MaxVisibleFields + 1
	} else if d.SelectedField < d.ScrollOffset {
		d.ScrollOffset = d.SelectedField
	}
}

// PrevField moves to the previous field
func (d *DebugUI) PrevField() {
	d.SelectedField--
	if d.SelectedField < 0 {
		d.SelectedField = len(d.FieldNames) - 1
	}
	// Adjust scroll if needed
	if d.SelectedField >= d.ScrollOffset+d.MaxVisibleFields {
		d.ScrollOffset = d.SelectedField - d.MaxVisibleFields + 1
	} else if d.SelectedField < d.ScrollOffset {
		d.ScrollOffset = d.SelectedField
	}
}

// AdjustValue adjusts the currently selected field value
func (d *DebugUI) AdjustValue(delta float64) {
	config := enemyConfigs[d.SelectedEnemy]
	fieldName := d.FieldNames[d.SelectedField]

	switch fieldName {
	case "CountBase":
		config.CountBase = max(0, config.CountBase+int(delta))
	case "CountPerLevel":
		config.CountPerLevel = max(0, config.CountPerLevel+int(delta))
	case "MaxAngle":
		newVal := config.MaxAngle + delta*0.1
		if newVal < 0 {
			newVal = 0
		}
		config.MaxAngle = newVal
	case "HealthBase":
		config.HealthBase = max(1, config.HealthBase+int(delta))
	case "HealthPerLevel":
		config.HealthPerLevel = max(0, config.HealthPerLevel+int(delta))
	case "TBase":
		config.TBase = max(0, config.TBase+int(delta)*10)
	case "TRange":
		config.TRange = max(0, config.TRange+int(delta)*10)
	case "TStart":
		config.TStart = max(0, config.TStart+int(delta)*10)
	case "YStopBase":
		newVal := config.YStopBase + delta*10
		if newVal < 0 {
			newVal = 0
		}
		config.YStopBase = newVal
	case "YStopRange":
		newVal := config.YStopRange + delta*10
		if newVal < 0 {
			newVal = 0
		}
		config.YStopRange = newVal
	case "YOffsetMult":
		newVal := config.YOffsetMult + delta*0.1
		if newVal < 0 {
			newVal = 0
		} else if newVal > 1 {
			newVal = 1
		}
		config.YOffsetMult = newVal
	case "UseYStep":
		config.UseYStep = !config.UseYStep
	case "HasFireDir":
		config.HasFireDir = !config.HasFireDir
	}

	enemyConfigs[d.SelectedEnemy] = config
}

// GetFieldValue returns the current value of a field as a string
func (d *DebugUI) GetFieldValue(kind EnemyKind, fieldName string) string {
	config := enemyConfigs[kind]

	switch fieldName {
	case "CountBase":
		return strconv.Itoa(config.CountBase)
	case "CountPerLevel":
		return strconv.Itoa(config.CountPerLevel)
	case "MaxAngle":
		return strconv.FormatFloat(config.MaxAngle, 'f', 3, 64)
	case "HealthBase":
		return strconv.Itoa(config.HealthBase)
	case "HealthPerLevel":
		return strconv.Itoa(config.HealthPerLevel)
	case "TBase":
		return strconv.Itoa(config.TBase)
	case "TRange":
		return strconv.Itoa(config.TRange)
	case "TStart":
		return strconv.Itoa(config.TStart)
	case "YStopBase":
		return strconv.FormatFloat(config.YStopBase, 'f', 1, 64)
	case "YStopRange":
		return strconv.FormatFloat(config.YStopRange, 'f', 1, 64)
	case "YOffsetMult":
		return strconv.FormatFloat(config.YOffsetMult, 'f', 2, 64)
	case "UseYStep":
		if config.UseYStep {
			return "true"
		}
		return "false"
	case "HasFireDir":
		if config.HasFireDir {
			return "true"
		}
		return "false"
	}
	return ""
}

// Render draws the debug UI panel
func (d *DebugUI) Render(ctx *js.Object) {
	if !d.Visible {
		return
	}

	// Draw panel background
	ctx.Set("fillStyle", "rgba(0, 0, 0, 0.85)")
	ctx.Call("fillRect", d.PanelX, d.PanelY, d.PanelWidth, d.PanelHeight)

	// Draw panel border
	ctx.Set("strokeStyle", "#00ff00")
	ctx.Set("lineWidth", 2)
	ctx.Call("strokeRect", d.PanelX, d.PanelY, d.PanelWidth, d.PanelHeight)

	// Draw title
	ctx.Set("fillStyle", "#00ff00")
	ctx.Set("font", "bold 16px monospace")
	ctx.Set("textAlign", "left")
	ctx.Call("fillText", "ENEMY CONFIG DEBUG", d.PanelX+10, d.PanelY+25)

	// Draw enemy type selector
	ctx.Set("font", "14px monospace")
	ctx.Set("fillStyle", "#ffff00")
	enemyName := EnemyKindNames[d.SelectedEnemy]
	ctx.Call("fillText", "< "+enemyName+" >", d.PanelX+10, d.PanelY+50)

	// Draw instructions
	ctx.Set("fillStyle", "#888888")
	ctx.Set("font", "11px monospace")
	ctx.Call("fillText", "Q/E: Enemy | W/S: Field | A/D: Value | F9: Close", d.PanelX+10, d.PanelY+70)

	// Draw separator
	ctx.Set("strokeStyle", "#444444")
	ctx.Call("beginPath")
	ctx.Call("moveTo", d.PanelX+10, d.PanelY+80)
	ctx.Call("lineTo", d.PanelX+d.PanelWidth-10, d.PanelY+80)
	ctx.Call("stroke")

	// Draw fields
	ctx.Set("font", "13px monospace")
	startY := d.PanelY + 100
	lineHeight := 28

	endIdx := min(d.ScrollOffset+d.MaxVisibleFields, len(d.FieldNames))
	for i := d.ScrollOffset; i < endIdx; i++ {
		fieldName := d.FieldNames[i]
		value := d.GetFieldValue(d.SelectedEnemy, fieldName)
		yPos := startY + (i-d.ScrollOffset)*lineHeight

		// Highlight selected field
		if i == d.SelectedField {
			ctx.Set("fillStyle", "rgba(0, 255, 0, 0.2)")
			ctx.Call("fillRect", d.PanelX+5, yPos-15, d.PanelWidth-10, lineHeight-2)
			ctx.Set("fillStyle", "#00ff00")
		} else {
			ctx.Set("fillStyle", "#cccccc")
		}

		// Draw field name
		ctx.Call("fillText", fieldName+":", d.PanelX+15, yPos)

		// Draw value (right-aligned)
		ctx.Set("textAlign", "right")
		if i == d.SelectedField {
			ctx.Set("fillStyle", "#ffffff")
		}
		ctx.Call("fillText", value, d.PanelX+d.PanelWidth-15, yPos)
		ctx.Set("textAlign", "left")
	}

	// Draw scroll indicator if needed
	if len(d.FieldNames) > d.MaxVisibleFields {
		ctx.Set("fillStyle", "#666666")
		ctx.Set("font", "11px monospace")
		scrollInfo := strconv.Itoa(d.ScrollOffset+1) + "-" + strconv.Itoa(endIdx) + " of " + strconv.Itoa(len(d.FieldNames))
		ctx.Call("fillText", scrollInfo, d.PanelX+10, d.PanelY+d.PanelHeight-10)
	}
}
