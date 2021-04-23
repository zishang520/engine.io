package types

type Void struct{}

type Set map[string]Void

type Kv struct {
	Key   string
	Value string
}

var NULL Void = struct{}{}
