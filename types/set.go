package types

import (
	"sync"
)

type Set[T comparable] struct {
	mu    sync.RWMutex
	cache map[T]Void
}

// NewSet creates a new Set and initializes it with the provided keys.
func NewSet[T comparable](keys ...T) *Set[T] {
	s := &Set[T]{cache: make(map[T]Void, len(keys))}
	for _, key := range keys {
		s.cache[key] = NULL
	}
	return s
}

// Add adds the provided keys to the set.
func (s *Set[T]) Add(keys ...T) bool {
	if len(keys) == 0 {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, key := range keys {
		s.cache[key] = NULL
	}
	return true
}

// Delete removes the provided keys from the set.
func (s *Set[T]) Delete(keys ...T) bool {
	if len(keys) == 0 {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	for _, key := range keys {
		delete(s.cache, key)
	}
	return true
}

// Clear removes all items from the set.
func (s *Set[T]) Clear() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache = map[T]Void{}
	return true
}

// Has checks if the set contains the provided key.
func (s *Set[T]) Has(key T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.cache[key]
	return exists
}

// Len returns the number of items in the set.
func (s *Set[T]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.cache)
}

// All returns a copy of the set's internal map.
func (s *Set[T]) All() map[T]Void {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_tmp := make(map[T]Void, len(s.cache))
	for k := range s.cache {
		_tmp[k] = NULL
	}

	return _tmp
}

// Keys returns a slice containing all keys in the set.
func (s *Set[T]) Keys() []T {
	s.mu.RLock()
	defer s.mu.RUnlock()

	list := make([]T, 0, len(s.cache))
	for k := range s.cache {
		list = append(list, k)
	}

	return list
}
