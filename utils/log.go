package utils

import (
	"fmt"
	"github.com/gookit/color"
	"log"
)

type _log struct {
	DEBUG bool
}

var (
	Log = &_log{false}
)

/**
 * Console log line.
 *
 * @param  {...interface{}} message
 * @return {void}
 */
func (*_log) Comment(message string, args ...interface{}) {
	fmt.Print(color.Tag("default").Sprint(fmt.Sprintf(message, args...)))
}

/**
 * Console log line.
 *
 * @param  {...interface{}} message
 * @return {void}
 */
func (*_log) Line(message string, args ...interface{}) {
	fmt.Println(color.Tag("default").Sprint(fmt.Sprintf(message, args...)))
}

/**
 * Console log default.
 *
 * @param  {...interface{}} message
 * @return {void}
 */
func (*_log) Default(message string, args ...interface{}) {
	log.Println(color.Tag("default").Sprint(fmt.Sprintf(message, args...)))
}

/**
 * Console log info.
 *
 * @param  {...interface{}} message
 * @return {void}
 */
func (*_log) Info(message string, args ...interface{}) {
	log.Println(color.Info.Sprint(fmt.Sprintf(message, args...)))
}

/**
 * Console debug info.
 *
 * @param  {...interface{}} message
 * @return {void}
 */
func (l *_log) Debug(message string, args ...interface{}) {
	if l.DEBUG {
		fmt.Println(color.Debug.Sprint(fmt.Sprintf(message, args...)))
	}
}

/**
 * Console log success.
 *
 * @param  {...interface{}} message
 * @return {void}
 */
func (*_log) Success(message string, args ...interface{}) {
	log.Println(color.Success.Sprint(fmt.Sprintf(message, args...)))
}

/**
 * Console log info.
 *
 * @param  {...interface{}} message
 * @return {void}
 */
func (*_log) Error(message string, args ...interface{}) {
	log.Println(color.Danger.Sprint(fmt.Sprintf(message, args...)))
}

/**
 * Console log warning.
 *
 * @param  {...interface{}} message
 * @return {void}
 */
func (*_log) Warning(message string, args ...interface{}) {
	log.Println(color.Warn.Sprint(fmt.Sprintf(message, args...)))
}

/**
 * Console log fatal.
 *
 * @param  {...interface{}} message
 * @return {void}
 */
func (*_log) Fatal(message string, args ...interface{}) {
	log.Fatal(color.Error.Sprint(fmt.Sprintf(message, args...)))
}
