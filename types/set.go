package types

type Set map[string]Void

func (s *Set) Has(key string) bool {
	_, exists := s[key]
	return exists
}

func (s *Set) Add(key string, value Void) {
	s[key] = value
}

func (s *Set) Del(key string) {
	delete(s, key)
}
