package utils

import (
	"testing"
)

func TestYeast(t *testing.T) {
	yest := &Yeast{}

	t.Run("Encode", func(t *testing.T) {
		data := yest.Encode(1234567890)
		if data != "19bWBI" {
			t.Fatalf(`Encode value not as expected: %s, want match for %s`, data, "19bWBI")
		}
	})
}
