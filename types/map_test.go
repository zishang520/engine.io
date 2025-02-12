package types

import (
	"testing"
)

func TestMap(t *testing.T) {
	_map := &Map[string, string]{}

	t.Run("Swap", func(t *testing.T) {
		if _, b := _map.Swap("Swap", "Value"); b != false {
			t.Fatalf(`*Map.Swap("Swap") = %t, want match for %t`, b, false)
		}
		if _, b := _map.Swap("Swap", "123"); b != true {
			t.Fatalf(`*Map.Swap("Swap") = %t, want match for %t`, b, true)
		}
		if expunged := _map.dirty["Swap"].expunged; expunged == nil {
			t.Fatalf(`*Map.Swap("Swap").expunged = %v, want to match not nil`, expunged)
		}
		_map.Delete("Swap")
		if n := _map.Len(); n != 0 {
			t.Fatalf(`*Map.Len() = %d, want match for %d`, n, 0)
		}
		if keys := _map.Keys(); len(keys) != 0 {
			t.Fatalf(`*Map.Keys() = %v, want match for []`, keys)
		}
		_map.Store("key", "value")
		if values := _map.Values(); len(values) != 1 {
			t.Fatalf(`*Map.Values() = %v, want match for []`, values)
		}
	})
}
