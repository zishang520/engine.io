package utils

import (
	"github.com/gookit/color"
	"log"
)

type Debug struct {
	DEBUG bool
	Log   *log.Logger
}

var _log = &Debug{DEBUG: false, Log: log.Default()}

func Log() *Debug {
	return _log
}

// Console log Println.
func (d *Debug) Println(message string, args ...interface{}) {
	d.Log.Println(color.Sprintf(message, args...))
}

// Console log Default.
func (d *Debug) Default(message string, args ...interface{}) {
	d.Log.Println(color.Tag("default").Sprintf(message, args...))
}

// Console log Info.
func (d *Debug) Info(message string, args ...interface{}) {
	d.Log.Println(color.Info.Sprintf(message, args...))
}

// Console Debug Debug.
func (d *Debug) Debug(message string, args ...interface{}) {
	if d.DEBUG {
		d.Log.Println(color.Debug.Sprintf(message, args...))
	}
}

// Console log Success.
func (d *Debug) Success(message string, args ...interface{}) {
	d.Log.Println(color.Success.Sprintf(message, args...))
}

// Console log Error.
func (d *Debug) Error(message string, args ...interface{}) {
	d.Log.Println(color.Danger.Sprintf(message, args...))
}

// Console log Warning.
func (d *Debug) Warning(message string, args ...interface{}) {
	d.Log.Println(color.Warn.Sprintf(message, args...))
}

// Console log Secondary.
func (d *Debug) Secondary(message string, args ...interface{}) {
	d.Log.Println(color.Secondary.Sprintf(message, args...))
}

// Console log Secondary.
func (d *Debug) Question(message string, args ...interface{}) {
	d.Log.Println(color.Question.Sprintf(message, args...))
}

// Console log Fatal.
func (d *Debug) Fatal(message string, args ...interface{}) {
	d.Log.Fatal(color.Error.Sprintf(message, args...))
}
