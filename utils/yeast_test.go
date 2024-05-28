package utils

import (
	"testing"
)

func TestNewYeast(t *testing.T) {
	y := NewYeast()
	if y.seed.Load() != 0 {
		t.Errorf("Expected seed to be 0, got %d", y.seed)
	}
	if prev := y.Prev(); prev != "" {
		t.Errorf("Expected prev to be empty, got %s", prev)
	}
}

func TestSetPrevAndGetPrev(t *testing.T) {
	y := NewYeast()
	prev := "some previous value"
	y.SetPrev(prev)

	if got := y.Prev(); got != prev {
		t.Errorf("Expected Prev() to return %s, got %s", prev, got)
	}
}

func TestSeed(t *testing.T) {
	y := NewYeast()
	seed1 := y.Seed()
	seed2 := y.Seed()

	if seed1 != 1 {
		t.Errorf("Expected first Seed() to return 1, got %d", seed1)
	}
	if seed2 != 2 {
		t.Errorf("Expected second Seed() to return 2, got %d", seed2)
	}
}

func TestResetSeed(t *testing.T) {
	y := NewYeast()
	y.Seed()
	y.ResetSeed()

	if seed := y.Seed(); seed != 1 {
		t.Errorf("Expected Seed() after ResetSeed() to be 1, got %d", seed)
	}
}

func TestEncode(t *testing.T) {
	tests := []struct {
		num  int64
		want string
	}{
		{0, "0"},
		{1, "1"},
		{63, "_"},
		{64, "10"},
		{1023, "F_"},
	}

	for _, tc := range tests {
		got := DefaultYeast.Encode(tc.num)
		if got != tc.want {
			t.Errorf("Encode(%d) expected %s, got %s", tc.num, tc.want, got)
		}
	}
}
