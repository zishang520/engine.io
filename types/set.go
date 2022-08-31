package types

import (
	"sync"
)

type Set[T comparable] struct {
	cache map[T]Void
	// mu
	mu sync.RWMutex
}

func NewSet[T comparable](keys ...T) *Set[T] {
	s := &Set[T]{cache: map[T]Void{}}
	s.Add(keys...)
	return s
}

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

func (s *Set[T]) Clear() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.cache = map[T]Void{}
	return true
}

func (s *Set[T]) Has(key T) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.cache[key]
	return exists
}

func (s *Set[T]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.cache)
}

func (s *Set[T]) All() map[T]Void {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_tmp := map[T]Void{}

	for k := range s.cache {
		_tmp[k] = NULL
	}

	return _tmp
}

func (s *Set[T]) Keys() (list []T) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for k := range s.cache {
		list = append(list, k)
	}

	return list
}
