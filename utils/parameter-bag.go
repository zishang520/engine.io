package utils

import (
	"sync"
)

type ParameterBag struct {
	parameters map[string][]string

	mu sync.RWMutex
}

func NewParameterBag(parameters map[string][]string) *ParameterBag {
	if parameters == nil {
		parameters = make(map[string][]string)
	}
	return &ParameterBag{parameters: parameters}
}

// Returns the parameters.
func (p *ParameterBag) All() map[string][]string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	_tmp := map[string][]string{}
	for k, v := range p.parameters {
		_tmp[k] = append([]string{}, v...)
	}

	return _tmp
}

// Returns the parameter keys.
func (p *ParameterBag) Keys() []string {
	p.mu.RLock()
	defer p.mu.RUnlock()

	keys := []string{}
	for k := range p.parameters {
		keys = append(keys, k)
	}
	return keys
}

// Replaces the current parameters by a new set.
func (p *ParameterBag) Replace(parameters map[string][]string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.parameters = parameters
}

// Replaces the current parameters by a new set.
func (p *ParameterBag) With(parameters map[string][]string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for k, v := range parameters {
		p.parameters[k] = append([]string{}, v...)
	}
}

// Add adds the value to key. It appends to any existing
// values associated with key.
func (p *ParameterBag) Add(key string, value string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.parameters[key] = append(p.parameters[key], value)
}

// Returns a parameter by name.
func (p *ParameterBag) Get(key string, _default ...string) (string, bool) {
	return p.GetLast(key, _default...)
}

func (p *ParameterBag) Peek(key string, _default ...string) string {
	v, _ := p.GetLast(key, _default...)
	return v
}

// Returns a parameter by name.
func (p *ParameterBag) GetFirst(key string, _default ...string) (string, bool) {
	_default = append(_default, "")

	if value, ok := p.Gets(key); ok && len(value) > 0 {
		return value[0], ok
	}
	return _default[0], false
}

// Returns a parameter by name.
func (p *ParameterBag) GetLast(key string, _default ...string) (string, bool) {
	_default = append(_default, "")

	if value, ok := p.Gets(key); ok {
		if l := len(value); l > 0 {
			return value[l-1], ok
		}
	}
	return _default[0], false
}

// Returns a parameter by name.
func (p *ParameterBag) Gets(key string, _default ...[]string) ([]string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	_default = append(_default, []string{})
	if v, ok := p.parameters[key]; ok {
		return v, ok
	}
	return _default[0], false
}

// Sets a parameter by name.
func (p *ParameterBag) Set(key string, value string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.parameters[key] = []string{value}
}

// Returns true if the parameter is defined.
func (p *ParameterBag) Has(key string) bool {
	p.mu.RLock()
	defer p.mu.RUnlock()

	_, ok := p.parameters[key]
	return ok
}

// Removes a parameter.
func (p *ParameterBag) Remove(key string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.parameters, key)
}

// Returns the number of parameters.
func (p *ParameterBag) Count() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.parameters)
}
