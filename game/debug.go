package game

import (
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
