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
func (*_log) Comment(message ...interface{}) {
	fmt.Print(color.Tag("default").Sprint(fmt.Sprint(message...)))
}

/**
 * Console log line.
 *
 * @param  {...interface{}} message
 * @return {void}
 */
func (*_log) Line(message ...interface{}) {
	fmt.Println(color.Tag("default").Sprint(fmt.Sprint(message...)))
}

/**
 * Console log default.
 *
 * @param  {...interface{}} message
 * @return {void}
 */
func (*_log) Default(message ...interface{}) {
	log.Println(color.Tag("default").Sprint(fmt.Sprint(message...)))
}

/**
 * Console log info.
 *
 * @param  {...interface{}} message
 * @return {void}
 */
func (*_log) Info(message ...interface{}) {
	log.Println(color.Info.Sprint(fmt.Sprint(message...)))
}

/**
 * Console debug info.
 *
 * @param  {...interface{}} message
 * @return {void}
 */
func (l *_log) Debug(message ...interface{}) {
	if l.DEBUG {
		fmt.Println(color.Debug.Sprint(fmt.Sprint(message...)))
	}
}

/**
 * Console log success.
 *
 * @param  {...interface{}} message
 * @return {void}
 */
func (*_log) Success(message ...interface{}) {
	log.Println(color.Success.Sprint(fmt.Sprint(message...)))
}

/**
 * Console log info.
 *
 * @param  {...interface{}} message
 * @return {void}
 */
func (*_log) Error(message ...interface{}) {
	log.Println(color.Danger.Sprint(fmt.Sprint(message...)))
}

/**
 * Console log warning.
 *
 * @param  {...interface{}} message
 * @return {void}
 */
func (*_log) Warning(message ...interface{}) {
	log.Println(color.Warn.Sprint(fmt.Sprint(message...)))
}

/**
 * Console log fatal.
 *
 * @param  {...interface{}} message
 * @return {void}
 */
func (*_log) Fatal(message ...interface{}) {
	log.Fatal(color.Error.Sprint(fmt.Sprint(message...)))
}
