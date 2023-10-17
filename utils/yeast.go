package utils

import (
	"sync"
	"sync/atomic"
	"time"
)

var (
	utils_alphabet [64]string = [64]string{"0", "1", "2", "3", "4", "5", "6", "7", "8", "9", "A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k", "l", "m", "n", "o", "p", "q", "r", "s", "t", "u", "v", "w", "x", "y", "z", "-", "_"}
)

type Yeast struct {
	seed int64
	prev string

	mu sync.RWMutex
}

func (y *Yeast) SetPrev(prev string) {
	y.mu.Lock()
	defer y.mu.Unlock()

	y.prev = prev
}

func (y *Yeast) Prev() string {
	y.mu.RLock()
	defer y.mu.RUnlock()

	return y.prev
}

func (y *Yeast) Seed() int64 {
	return atomic.AddInt64(&y.seed, 1)
}

func (y *Yeast) ResetSeed() {
	atomic.StoreInt64(&y.seed, 0)
}

func (y *Yeast) Encode(num int64) (encoded string) {
	for {
		encoded = utils_alphabet[num%64] + encoded
		num = num / 64
		if num <= 0 {
			break
		}
	}

	return encoded
}

var DefaultYeast = &Yeast{}

func YeastDate() (now string) {
	now = DefaultYeast.Encode(time.Now().UnixMilli())
	if now != DefaultYeast.Prev() {
		DefaultYeast.ResetSeed()
		DefaultYeast.SetPrev(now)
		return now
	}
	return now + "." + DefaultYeast.Encode(DefaultYeast.Seed())
}
