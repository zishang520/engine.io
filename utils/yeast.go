package utils

import (
	"sync"
	"sync/atomic"
	"time"
)

var utils_alphabet = [64]byte{'0', '1', '2', '3', '4', '5', '6', '7', '8', '9', 'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J', 'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z', 'a', 'b', 'c', 'd', 'e', 'f', 'g', 'h', 'i', 'j', 'k', 'l', 'm', 'n', 'o', 'p', 'q', 'r', 's', 't', 'u', 'v', 'w', 'x', 'y', 'z', '-', '_'}

type Yeast struct {
	seed atomic.Int64
	prev atomic.Value
	mu   sync.Mutex
}

func NewYeast() *Yeast {
	return &Yeast{}
}

func (y *Yeast) SetPrev(prev string) {
	y.prev.Store(prev)
}

func (y *Yeast) Prev() string {
	if v, ok := y.prev.Load().(string); ok {
		return v
	}
	return ""
}

func (y *Yeast) Seed() int64 {
	return y.seed.Add(1)
}

func (y *Yeast) ResetSeed() {
	y.seed.Store(0)
}

func (y *Yeast) Encode(num int64) string {
	if num == 0 {
		return string(utils_alphabet[0])
	}

	encoded := make([]byte, 0, 11) // Pre-allocate slice with a reasonable size
	for num > 0 {
		encoded = append([]byte{utils_alphabet[num%64]}, encoded...)
		num /= 64
	}

	return string(encoded)
}

var DefaultYeast = NewYeast()

func YeastDate() string {
	now := DefaultYeast.Encode(time.Now().UnixMilli())

	if prev := DefaultYeast.Prev(); now != prev {
		DefaultYeast.mu.Lock()
		if now != DefaultYeast.Prev() { // Double check to avoid race condition
			DefaultYeast.ResetSeed()
			DefaultYeast.SetPrev(now)
		}
		DefaultYeast.mu.Unlock()
		return now
	}

	return now + "." + DefaultYeast.Encode(DefaultYeast.Seed())
}
