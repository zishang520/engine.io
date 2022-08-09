package log

import (
	"bytes"
	"os"
	"regexp"
	"testing"
)

func TestLog(t *testing.T) {
	os.Setenv("DEBUG", "")
	_log := NewLog("namespace")
	buf := new(bytes.Buffer)

	t.Run("prefix", func(t *testing.T) {
		if _log.Prefix() != "namespace" && _log.Logger.Prefix() == "namespace " {
			t.Fatalf(`*Log.Prefix() = %q, want match for %#q`, _log.Prefix(), "namespace")
		}
	})

	_log.SetFlags(0)
	_log.SetOutput(buf)

	_log.Debug("Test")

	if buf.Len() > 0 {
		t.Fatal(`_log.Debug("Test") There should be no output here, but got the output.`)
	}

	buf.Reset()

	_log.Printf("hello %d world", 23)
	line := buf.String()
	line = line[0 : len(line)-1]
	pattern := "^" + _log.Logger.Prefix() + "hello 23 world$"
	matched, err := regexp.MatchString(pattern, line)
	if err != nil {
		t.Fatal("pattern did not compile:", err)
	}
	if !matched {
		t.Errorf("log output should match %q is %q", pattern, line)
	}
	_log.SetOutput(os.Stderr)
}
