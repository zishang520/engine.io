package log

import (
	_log "log"
	"os"
	"regexp"
	"strings"
	"sync"

	"github.com/gookit/color"
)

type Log struct {
	*_log.Logger

	DEBUG bool

	mu              sync.RWMutex // ensures atomic writes; protects the following fields
	prefix          string
	namespaceRegexp *regexp.Regexp
}

func NewLog(prefix string) *Log {
	l := &Log{
		Logger: _log.New(os.Stderr, "", 0),
		DEBUG:  false,
	}

	if prefix != "" {
		l.SetPrefix(prefix)
	}

	if debug := os.Getenv("DEBUG"); debug != "" {
		l.namespaceRegexp = regexp.MustCompile(strings.ReplaceAll(regexp.QuoteMeta(strings.TrimSpace(debug)), `\*`, `.*`))
	}
	return l
}

func (d *Log) checkNamespace(namespace string) bool {
	if d.namespaceRegexp != nil {
		return d.namespaceRegexp.MatchString(namespace)
	}
	return false
}

// Console log Println.
func (d *Log) Println(message string, args ...any) {
	d.Logger.Println(color.Sprintf(message, args...))
}

// Console log Default.
func (d *Log) Default(message string, args ...any) {
	d.Logger.Println(color.Tag("default").Sprintf(message, args...))
}

// Console log Info.
func (d *Log) Info(message string, args ...any) {
	d.Logger.Println(color.Info.Sprintf(message, args...))
}

// Console Debug Debug.
func (d *Log) Debug(message string, args ...any) {
	if d.DEBUG || d.checkNamespace(d.Prefix()) {
		d.Logger.Println(color.Debug.Sprintf(message, args...))
	}
}

// Console log Success.
func (d *Log) Success(message string, args ...any) {
	d.Logger.Println(color.Success.Sprintf(message, args...))
}

// Console log Error.
func (d *Log) Error(message string, args ...any) {
	d.Logger.Println(color.Danger.Sprintf(message, args...))
}

// Console log Warning.
func (d *Log) Warning(message string, args ...any) {
	d.Logger.Println(color.Warn.Sprintf(message, args...))
}

// Console log Secondary.
func (d *Log) Secondary(message string, args ...any) {
	d.Logger.Println(color.Secondary.Sprintf(message, args...))
}

// Console log Secondary.
func (d *Log) Question(message string, args ...any) {
	d.Logger.Println(color.Question.Sprintf(message, args...))
}

// Console log Fatal.
func (d *Log) Fatal(message string, args ...any) {
	d.Logger.Fatal(color.Error.Sprintf(message, args...))
}

// Prefix returns the output prefix for the logger.
func (d *Log) Prefix() string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return d.prefix
}

// SetPrefix sets the output prefix for the logger.
func (d *Log) SetPrefix(prefix string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.prefix = prefix

	d.Logger.SetPrefix(prefix + " ")
}
