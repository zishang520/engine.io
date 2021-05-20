package types

type Set map[string]Void

func (s Set) Has(key string) bool {
	_, exists := s[key]
	return exists
}
